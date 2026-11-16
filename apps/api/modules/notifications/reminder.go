package notifications

import (
	"context"
	"encoding/json"
	"log/slog"
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

var idleStops = map[string]reminderText{
	"en": {"Timer stopped", "Your timer ran 10 hours with no activity, so Sablier stopped it at your last activity. Fix the entry if that is wrong."},
	"fr": {"Minuteur arrêté", "Votre minuteur a tourné 10 heures sans activité : Sablier l’a arrêté à votre dernière activité. Corrigez l’entrée si besoin."},
	"es": {"Temporizador detenido", "Tu temporizador estuvo 10 horas sin actividad, así que Sablier lo detuvo en tu última actividad. Corrige la entrada si hace falta."},
	"de": {"Timer gestoppt", "Dein Timer lief 10 Stunden ohne Aktivität, deshalb hat Sablier ihn bei deiner letzten Aktivität gestoppt. Korrigiere den Eintrag bei Bedarf."},
}

// reminderPayload is the web-push body asking someone whether their running timer is meant
// to run, in their Registre locale and in English for any other value.
func reminderPayload(locale string) ([]byte, error) {
	return pushPayload(reminders, locale)
}

func pushPayload(texts map[string]reminderText, locale string) ([]byte, error) {
	text, ok := texts[strings.ToLower(locale)]
	if !ok {
		text = texts["en"]
	}
	return json.Marshal(map[string]string{"title": text.title, "body": text.body, "icon": "/favicon.svg"})
}

func (s *Service) remind(ctx context.Context, opts *webpush.Options, target reminderTarget) bool {
	log := targetLog(s.logger, target)
	payload, err := reminderPayload(target.Locale)
	if err != nil {
		log.Error("failed to marshal notification payload", "error", err)
		return false
	}
	return s.deliver(ctx, opts, target, payload, func() { s.markReminded(ctx, log, target.UserID) })
}

func (s *Service) markReminded(ctx context.Context, log *slog.Logger, userID int64) {
	if err := s.orm.WithContext(ctx).Model(&schemas.TimeEntry{}).
		Where("user_id = ? AND stopped_at IS NULL", userID).
		Update("last_notification_at", gorm.Expr("NOW()")).Error; err != nil {
		log.Error("failed to update last_notification_at", "error", err)
	}
}
