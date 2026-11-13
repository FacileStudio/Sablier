package notifications

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/schemas"

	webpush "github.com/SherClockHolmes/webpush-go"
	"gorm.io/gorm"
)

type reminderText struct{ title, body string }

var reminders = map[string]reminderText{
	"en": {"Are you still working?", "You have an active timer. Don't forget to stop it when you're done!"},
	"fr": {"Vous travaillez toujours ?", "Un minuteur est en cours. N’oubliez pas de l’arrêter quand vous avez fini !"},
	"es": {"¿Sigues trabajando?", "Tienes un temporizador activo. ¡No olvides detenerlo cuando termines!"},
	"de": {"Arbeitest du noch?", "Es läuft ein Timer. Vergiss nicht, ihn zu stoppen, wenn du fertig bist!"},
}

// reminderPayload is the web-push body asking someone whether their running timer is meant
// to run, in their Registre locale and in English for any other value.
func reminderPayload(locale string) ([]byte, error) {
	text, ok := reminders[strings.ToLower(locale)]
	if !ok {
		text = reminders["en"]
	}
	return json.Marshal(map[string]string{"title": text.title, "body": text.body, "icon": "/favicon.svg"})
}

func (s *Service) remind(ctx context.Context, opts *webpush.Options, target reminderTarget) bool {
	log := s.logger.With("user_id", target.UserID, "subscription_id", target.SubscriptionID, "endpoint_host", endpointHost(target.Endpoint))
	payload, err := reminderPayload(target.Locale)
	if err != nil {
		log.Error("failed to marshal notification payload", "error", err)
		return false
	}
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: target.Endpoint,
		Keys:     webpush.Keys{Auth: target.Auth, P256dh: target.P256DH},
	}, opts)
	if err != nil {
		log.Warn("push notification request failed, retrying in an hour", "error", err)
		s.markReminded(ctx, log, target.UserID)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 400 {
		s.markReminded(ctx, log, target.UserID)
		return true
	}
	s.rejected(ctx, log, target, resp)
	return false
}

func (s *Service) rejected(ctx context.Context, log *slog.Logger, target reminderTarget, resp *http.Response) {
	detail := readDetail(resp.Body)
	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusGone {
		log.Warn("push service rejected notification, retrying in an hour", "status", resp.StatusCode, "detail", detail)
		s.markReminded(ctx, log, target.UserID)
		return
	}
	if err := s.orm.WithContext(ctx).Where("id = ? AND endpoint = ?", target.SubscriptionID, target.Endpoint).Delete(&schemas.PushSubscription{}).Error; err != nil {
		log.Error("failed to delete expired push subscription", "status", resp.StatusCode, "error", err)
		return
	}
	log.Info("push subscription expired, deleted it", "status", resp.StatusCode, "detail", detail)
}

func (s *Service) markReminded(ctx context.Context, log *slog.Logger, userID int64) {
	if err := s.orm.WithContext(ctx).Model(&schemas.TimeEntry{}).
		Where("user_id = ? AND stopped_at IS NULL", userID).
		Update("last_notification_at", gorm.Expr("NOW()")).Error; err != nil {
		log.Error("failed to update last_notification_at", "error", err)
	}
}

func endpointHost(endpoint string) string {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}
	return parsed.Host
}

func readDetail(body io.Reader) string {
	detail, _ := io.ReadAll(io.LimitReader(body, 256))
	return strings.TrimSpace(string(detail))
}
