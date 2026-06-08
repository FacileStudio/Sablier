package worker

import (
	"context"
	"log/slog"
	"time"

	"api/modules/notifications"
)

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
