package antenne

import (
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

var remoteStatusAliases = map[string]string{
	"to-do":       schemas.StatusTodo,
	"todo":        schemas.StatusTodo,
	"open":        schemas.StatusTodo,
	"new":         schemas.StatusTodo,
	"backlog":     schemas.StatusTodo,
	"planned":     schemas.StatusTodo,
	"in-progress": schemas.StatusInProgress,
	"doing":       schemas.StatusInProgress,
	"started":     schemas.StatusInProgress,
	"wip":         schemas.StatusInProgress,
	"active":      schemas.StatusInProgress,
	"in-review":   schemas.StatusInReview,
	"review":      schemas.StatusInReview,
	"reviewing":   schemas.StatusInReview,
	"testing":     schemas.StatusInReview,
	"qa":          schemas.StatusInReview,
	"done":        schemas.StatusDone,
	"closed":      schemas.StatusDone,
	"complete":    schemas.StatusDone,
	"completed":   schemas.StatusDone,
	"resolved":    schemas.StatusDone,
	"finished":    schemas.StatusDone,
}

func mapRemoteStatus(raw string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(raw))
	key = strings.NewReplacer(" ", "-", "_", "-").Replace(key)
	status, ok := remoteStatusAliases[key]
	return status, ok
}
