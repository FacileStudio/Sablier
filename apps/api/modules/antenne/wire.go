package antenne

import (
	"encoding/json"
	"strings"
	"time"
)

type spacePayload struct {
	RegistreSpaceID string `json:"registre_space_id"`
	Personal        bool   `json:"personal"`
	OwnerSubject    string `json:"owner_subject"`
	OwnerEmail      string `json:"owner_email"`
}

type projectPayload struct {
	FacileID    string       `json:"facile_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Icon        *string      `json:"icon"`
	Space       spacePayload `json:"space"`
	UpdatedAt   string       `json:"updated_at"`
}

type taskPayload struct {
	FacileID        string  `json:"facile_id"`
	ProjectFacileID string  `json:"project_facile_id"`
	Name            string  `json:"name"`
	Status          string  `json:"status"`
	ActorEmail      string  `json:"actor_email"`
	ActorSubject    string  `json:"actor_subject"`
	Description     string  `json:"description"`
	DueDate         *string `json:"due_date"`
	Priority        string  `json:"priority"`
	UpdatedAt       string  `json:"updated_at"`
}

type timeEntryPayload struct {
	FacileID         string  `json:"facile_id"`
	TaskFacileID     string  `json:"task_facile_id"`
	ProjectFacileID  string  `json:"project_facile_id"`
	UserEmail        string  `json:"user_email"`
	UserSubject      string  `json:"user_subject"`
	StartedAt        string  `json:"started_at"`
	StoppedAt        *string `json:"stopped_at"`
	PausedDurationMs int64   `json:"paused_duration_ms"`
	Description      string  `json:"description"`
	UpdatedAt        string  `json:"updated_at"`
}

type projectWire struct {
	FacileID    string        `json:"facile_id"`
	Name        string        `json:"name"`
	Title       string        `json:"title"`
	Description *string       `json:"description"`
	Icon        *string       `json:"icon"`
	Space       *spacePayload `json:"space"`
	UpdatedAt   string        `json:"updated_at"`
}

type taskWire struct {
	FacileID        string    `json:"facile_id"`
	ProjectFacileID string    `json:"project_facile_id"`
	Name            string    `json:"name"`
	Title           string    `json:"title"`
	Status          string    `json:"status"`
	ActorEmail      string    `json:"actor_email"`
	ActorSubject    string    `json:"actor_subject"`
	Description     *string   `json:"description"`
	DueDate         optString `json:"due_date"`
	Priority        *string   `json:"priority"`
	UpdatedAt       string    `json:"updated_at"`
}

// optString tells a key sent as null, which clears, from a key left out, which keeps.
type optString struct {
	set   bool
	value *string
}

// UnmarshalJSON records that the key was present, null included.
func (o *optString) UnmarshalJSON(data []byte) error {
	o.set = true
	return json.Unmarshal(data, &o.value)
}

func (p projectWire) name() string {
	if p.Name != "" {
		return p.Name
	}
	return p.Title
}

func (t taskWire) name() string {
	if t.Name != "" {
		return t.Name
	}
	return t.Title
}

var priorityAliases = map[string]string{
	"none": "none", "": "none", "no": "none",
	"low": "low", "medium": "medium", "normal": "medium",
	"high": "high", "urgent": "urgent", "critical": "urgent",
}

func normalizePriority(raw string) (string, bool) {
	priority, ok := priorityAliases[strings.ToLower(strings.TrimSpace(raw))]
	return priority, ok
}

func normalizeDueDate(raw *string) (*string, bool) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, true
	}
	value := strings.TrimSpace(*raw)
	if len(value) < 10 {
		return nil, false
	}
	if _, err := time.Parse(time.DateOnly, value[:10]); err != nil {
		return nil, false
	}
	day := value[:10]
	return &day, true
}

func wireTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// eventTime is the instant an inbound change happened: the payload's updated_at, or the
// envelope's second-precision timestamp from an older sender, flagged as coarse.
func eventTime(updatedAt, timestamp string) (time.Time, bool) {
	if at, err := time.Parse(time.RFC3339Nano, updatedAt); err == nil {
		return at, false
	}
	at, _ := time.Parse(time.RFC3339, timestamp)
	return at, true
}

// isOlder reports whether an inbound change at `at` predates the local row.
func isOlder(at time.Time, coarse bool, local time.Time) bool {
	if coarse {
		return at.Before(local.Truncate(time.Second))
	}
	return at.Before(local)
}
