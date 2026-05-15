package notifications

import (
	"context"
	stderrors "errors"
	"encoding/json"
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
}

func NewService(orm *gorm.DB, vapidPublicKey, vapidPrivateKey, vapidSubject string) *Service {
	s := &Service{
		orm:             orm,
		vapidPublicKey:  vapidPublicKey,
		vapidPrivateKey: vapidPrivateKey,
		vapidSubject:    vapidSubject,
	}
	s.controller = newController(s)
	return s
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

	opts := &webpush.Options{
		Subscriber:      s.vapidSubject,
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivateKey,
		TTL:             30,
	}

	for _, sub := range subs {
		resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				Auth:   sub.Auth,
				P256dh: sub.P256DH,
			},
		}, opts)
		if err != nil || resp.StatusCode >= 400 {
			failed++
			continue
		}
		resp.Body.Close()
		sent++
	}
	return sent, failed
}
