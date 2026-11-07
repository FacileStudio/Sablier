package notifications

import (
	"encoding/json"
	"strings"
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
