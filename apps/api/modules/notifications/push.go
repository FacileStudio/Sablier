package notifications

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/schemas"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// deliver pushes payload to target. after runs once the push service accepted it or failed
// in a way worth retrying later; a subscription it reports gone is deleted instead.
func (s *Service) deliver(ctx context.Context, opts *webpush.Options, target reminderTarget, payload []byte, after func()) bool {
	log := targetLog(s.logger, target)
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: target.Endpoint,
		Keys:     webpush.Keys{Auth: target.Auth, P256dh: target.P256DH},
	}, opts)
	if err != nil {
		log.Warn("push notification request failed", "error", err)
		after()
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 400 {
		after()
		return true
	}
	s.rejected(ctx, log, target, resp, after)
	return false
}

func targetLog(logger *slog.Logger, target reminderTarget) *slog.Logger {
	return logger.With("user_id", target.UserID, "subscription_id", target.SubscriptionID, "endpoint_host", endpointHost(target.Endpoint))
}

func (s *Service) rejected(ctx context.Context, log *slog.Logger, target reminderTarget, resp *http.Response, after func()) {
	detail := readDetail(resp.Body)
	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusGone {
		log.Warn("push service rejected notification", "status", resp.StatusCode, "detail", detail)
		after()
		return
	}
	if err := s.orm.WithContext(ctx).Where("id = ? AND endpoint = ?", target.SubscriptionID, target.Endpoint).Delete(&schemas.PushSubscription{}).Error; err != nil {
		log.Error("failed to delete expired push subscription", "status", resp.StatusCode, "error", err)
		return
	}
	log.Info("push subscription expired, deleted it", "status", resp.StatusCode, "detail", detail)
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
