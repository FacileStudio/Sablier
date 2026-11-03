package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/modules/heartbeats"
	"github.com/FacileStudio/Sablier/apps/api/modules/notifications"
)

// RunNotificationWorker runs a ticking loop that sends active-timer
// reminders once per minute until ctx is canceled.
func RunNotificationWorker(ctx context.Context, notifService *notifications.Service, appLogger *slog.Logger) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			appLogger.Info("Sending active timer reminders")
			sent, failed := notifService.SendActiveTimerReminders(ctx)
			appLogger.Info("Reminder sent", slog.Int("sent", sent), slog.Int("failed", failed))
		case <-ctx.Done():
			return
		}
	}
}

// RunHeartbeatWorker auto-confirms stale drafts and prunes old heartbeats once
// per hour until ctx is canceled.
func RunHeartbeatWorker(ctx context.Context, heartbeatService *heartbeats.Service, appLogger *slog.Logger) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			confirmed, err := heartbeatService.RunMaintenance(ctx, appLogger)
			if err != nil {
				appLogger.Error("heartbeats: maintenance failed", slog.Any("error", err))
			}
			appLogger.Info("heartbeats: drafts auto-confirmed", slog.Int("confirmed", confirmed))
		case <-ctx.Done():
			return
		}
	}
}
