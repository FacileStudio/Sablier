package mcp

import (
	"context"
	"strconv"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// kindDelete is the one sensitive action: deleting a time entry, which cannot be undone.
const kindDelete = "delete_entry"

var errTarget = errors.NotFound("this action names no time entry")

// kinds are the sensitive actions. The fingerprint covers the entry as it was when the
// person was asked, so an edit in between closes the action instead of deleting time the
// person did not see.
func (t *toolset) kinds() map[string]mcpkit.Kind {
	return map[string]mcpkit.Kind{
		kindDelete: {Tool: toolDelete, Fingerprint: t.fingerprint, Run: t.runDelete},
	}
}

// refOf parses an action's user and entry.
func refOf(a mcpkit.Action) (int64, int64, error) {
	return mcpkit.TargetIDs(a, errTarget)
}

// fingerprint digests the targeted entry as it is now, through the reach of DELETE
// /time-entries/{id}: the person's own entries only.
func (t *toolset) fingerprint(ctx context.Context, a mcpkit.Action) (string, error) {
	userID, entryID, err := refOf(a)
	if err != nil {
		return "", err
	}
	entry, err := t.Entries.Own(ctx, userID, entryID)
	if err != nil {
		return "", err
	}
	return digest(entry), nil
}

// digest fingerprints an entry.
func digest(e *timeentries.TimeEntryResponse) string {
	fields := []string{
		strconv.FormatInt(e.ID, 10), strconv.FormatInt(e.ProjectID, 10), strconv.FormatInt(e.TaskID, 10),
		e.StartedAt.UTC().Format(time.RFC3339Nano), stampOf(e.StoppedAt), strconv.FormatInt(e.PausedDuration, 10),
		e.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	return mcpkit.Digest(fields...)
}

// stampOf renders an optional instant.
func stampOf(at *time.Time) string {
	if at == nil {
		return "-"
	}
	return at.UTC().Format(time.RFC3339Nano)
}

// runDelete deletes the targeted entry.
func (t *toolset) runDelete(ctx context.Context, a mcpkit.Action) error {
	userID, entryID, err := refOf(a)
	if err != nil {
		return err
	}
	return t.Entries.Delete(ctx, userID, entryID)
}
