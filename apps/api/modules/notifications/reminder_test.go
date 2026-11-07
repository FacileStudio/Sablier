package notifications

import (
	"strings"
	"testing"
)

func TestReminderFollowsTheLocaleAndFallsBackToEnglish(t *testing.T) {
	for locale, want := range map[string]string{"fr": "Vous travaillez", "de": "Arbeitest du", "es": "Sigues trabajando", "": "still working", "it": "still working"} {
		payload, err := reminderPayload(locale)
		if err != nil || !strings.Contains(string(payload), want) {
			t.Errorf("%q: %s %v, want %q", locale, payload, err, want)
		}
	}
}
