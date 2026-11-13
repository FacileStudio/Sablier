package notifications

import (
	"context"
	stderrors "errors"
	"log/slog"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	webpush "github.com/SherClockHolmes/webpush-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Service manages web-push subscriptions and sends notifications.
type Service struct {
	orm             *gorm.DB
	vapidPublicKey  string
	vapidPrivateKey string
	vapidSubject    string
	controller      *Controller
	logger          *slog.Logger
}

// NewService wires the notification service with its VAPID credentials.
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

// SendActiveTimerReminders pushes a reminder to everyone whose timer has run for an hour
// since it started or since the last attempt. A failed attempt also waits an hour, and a
// subscription the push service reports gone is deleted so it is never tried again.
func (s *Service) SendActiveTimerReminders(ctx context.Context) (sent, failed int) {
	var targets []reminderTarget

	deltaNotifyTime := "NOW() - INTERVAL '1 HOUR'"
	if err := s.orm.WithContext(ctx).Table("time_entries").
		Distinct("time_entries.user_id, users.locale, push_subscriptions.id AS subscription_id, push_subscriptions.endpoint, push_subscriptions.p256dh, push_subscriptions.auth").
		Joins("JOIN push_subscriptions ON time_entries.user_id = push_subscriptions.user_id").
		Joins("JOIN users ON users.id = time_entries.user_id").
		Where("time_entries.stopped_at IS NULL").
		Where("(time_entries.last_notification_at IS NULL AND time_entries.started_at <= ? OR time_entries.last_notification_at <= ?)", gorm.Expr(deltaNotifyTime), gorm.Expr(deltaNotifyTime)).
		Scan(&targets).Error; err != nil {
		s.logger.Error("failed to list active timers to remind", "error", err)
		return 0, 0
	}

	s.logger.Info("found active timers for users", "count", len(targets))

	opts := newWebPushOptions(s)
	for _, target := range targets {
		if s.remind(ctx, opts, target) {
			sent++
		} else {
			failed++
		}
	}
	return sent, failed
}
