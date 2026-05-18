package notifications

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"log/slog"
	"strconv"

	"api/internal/authcontext"
	"api/internal/errors"
	"api/schemas"

	webpush "github.com/SherClockHolmes/webpush-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	orm             *gorm.DB
	vapidPublicKey  string
	vapidPrivateKey string
	vapidSubject    string
	controller      *Controller
	logger          *slog.Logger
}

func NewService(orm *gorm.DB, vapidPublicKey, vapidPrivateKey, vapidSubject string, logger *slog.Logger) *Service {
	s := &Service{
		orm:             orm,
		vapidPublicKey:  vapidPublicKey,
		vapidPrivateKey: vapidPrivateKey,
		vapidSubject:    vapidSubject,
		logger:          logger,
	}
	s.controller = newController(s)
	return s
}

func newWebPushOptions(s *Service) *webpush.Options {
	return &webpush.Options{
		Subscriber:      s.vapidSubject,
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivateKey,
		TTL:             30,
	}
}

func (s *Service) getVAPIDPublicKey() string {
	return s.vapidPublicKey
}

func (s *Service) saveSubscription(ctx context.Context, req *SaveSubscriptionRequest) error {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return errors.Unauthorized("missing identity")
	}
	userID, err := strconv.ParseInt(identity.UserID, 10, 64)
	if err != nil {
		return errors.Internal("invalid user id", err)
	}
	record := schemas.PushSubscription{
		UserID:   userID,
		Endpoint: req.Endpoint,
		P256DH:   req.P256DH,
		Auth:     req.Auth,
	}
	if err := s.orm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"endpoint", "p256dh", "auth"}),
	}).Create(&record).Error; err != nil {
		return errors.Internal("failed to save push subscription", err)
	}
	return nil
}

func (s *Service) deleteSubscription(ctx context.Context) error {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return errors.Unauthorized("missing identity")
	}
	userID, err := strconv.ParseInt(identity.UserID, 10, 64)
	if err != nil {
		return errors.Internal("invalid user id", err)
	}
	result := s.orm.WithContext(ctx).Where("user_id = ?", userID).Delete(&schemas.PushSubscription{})
	if result.Error != nil && !stderrors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.Internal("failed to delete push subscription", result.Error)
	}
	return nil
}

func (s *Service) SendActiveTimerReminders(ctx context.Context) (sent, failed int) {
	var entries []struct {
		UserID   int64  `gorm:"column:user_id"`
		Endpoint string `gorm:"column:endpoint"`
		P256DH   string `gorm:"column:p256dh"`
		Auth     string `gorm:"column:auth"`
	}

	deltaNotifyTime := "NOW() - INTERVAL '1 HOUR'"
	if err := s.orm.WithContext(ctx).Table("time_entries").
		Distinct("time_entries.user_id, push_subscriptions.endpoint, push_subscriptions.p256dh, push_subscriptions.auth").
		Joins("JOIN push_subscriptions ON time_entries.user_id = push_subscriptions.user_id").
		Where("time_entries.stopped_at IS NULL").
		Where("(time_entries.last_notification_at IS NULL AND time_entries.started_at <= ? OR time_entries.last_notification_at <= ?)", gorm.Expr(deltaNotifyTime), gorm.Expr(deltaNotifyTime)).
		Scan(&entries).Error; err != nil {
		return 0, 0
	}

	s.logger.Info("found active timers for users", "count", len(entries))

	opts := newWebPushOptions(s)

	payload, err := json.Marshal(map[string]string{
		"title": "Are you working Son?",
		"body":  "You have an active timer. Don't forget to stop it when you're done!",
		"icon":  "/favicon.svg",
	})

	if err != nil {
		s.logger.Error("failed to marshal notification payload", "error", err)
		return 0, 0
	}

	for _, entry := range entries {

		//log the whole entry for debugging
		s.logger.Info("sending reminder notification", "entry", entry)

		resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
			Endpoint: entry.Endpoint,
			Keys: webpush.Keys{
				Auth:   entry.Auth,
				P256dh: entry.P256DH,
			},
		}, opts)

		if resp != nil {
			resp.Body.Close()
		}
		if err != nil || resp.StatusCode >= 400 {

			s.logger.Error("failed to send notification", "entry", entry, "error", err)
			failed++
			continue
		}
		sent++
	}

	for _, entry := range entries {
		if err := s.orm.WithContext(ctx).Model(&schemas.TimeEntry{}).
			Where("user_id = ? AND stopped_at IS NULL", entry.UserID).
			Update("last_notification_at", gorm.Expr("NOW()")).Error; err != nil {
			s.logger.Error("failed to update last_notification_at", "user_id", entry.UserID, "error", err)
		}
	}

	return sent, failed
}

func (s *Service) broadcastNotification(ctx context.Context, req *BroadcastRequest) (sent, failed int) {
	var subs []schemas.PushSubscription
	if err := s.orm.WithContext(ctx).Find(&subs).Error; err != nil {
		return 0, 0
	}

	payload, _ := json.Marshal(map[string]string{
		"title": req.Title,
		"body":  req.Body,
		"icon":  req.Icon,
	})

	opts := newWebPushOptions(s)

	for _, sub := range subs {
		resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				Auth:   sub.Auth,
				P256dh: sub.P256DH,
			},
		}, opts)
		if err != nil || resp.StatusCode >= 400 {
			s.orm.Logger.Error(ctx, "failed to send notification", "user_id", sub.UserID, "error", err, "status_code", resp.StatusCode)
			failed++
			continue
		}
		resp.Body.Close()
		sent++
	}
	return sent, failed
}
