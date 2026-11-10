package mcp

import "time"

// PendingItem is one action waiting on the confirm page.
type PendingItem struct {
	ID        string        `json:"id"`
	Kind      string        `json:"kind"`
	ClientID  string        `json:"client_id"`
	Summary   string        `json:"summary"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`
	State     string        `json:"state"`
	Entry     *PendingEntry `json:"entry,omitempty"`
}

// PendingEntry is the time entry an action targets, as it is now.
type PendingEntry struct {
	ID        int64      `json:"id"`
	Project   string     `json:"project"`
	Task      string     `json:"task"`
	StartedAt time.Time  `json:"started_at"`
	StoppedAt *time.Time `json:"stopped_at"`
	Minutes   int        `json:"minutes"`
}

// PendingList is the body of GET /api/mcp/pending.
type PendingList struct {
	Actions []PendingItem `json:"actions"`
}

// PendingResult is the body answered when an action is confirmed or dismissed.
type PendingResult struct {
	Status string `json:"status"`
}

// EmptyRequest documents a route that takes no body.
type EmptyRequest struct{}
