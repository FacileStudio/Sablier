package mcp

import (
	"time"

	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// PendingItem is one action waiting on the confirm page.
type PendingItem struct {
	mcpkit.PendingItem
	Entry *PendingEntry `json:"entry,omitempty"`
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
