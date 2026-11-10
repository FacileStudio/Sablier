package mcp

import (
	"math"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

const (
	stateRunning = "running"
	statePaused  = "paused"
)

// view renders entries in a call's timezone, for the person calling.
type view struct {
	clock    clock
	callerID int64
	detailed bool
}

// entry renders one entry; the person is named only when it is someone else.
func (v view) entry(e timeentries.TimeEntryResponse) entryOut {
	net, _ := worked(e, v.clock.now)
	out := entryOut{
		ID: e.ID, ProjectID: e.ProjectID, Project: e.ProjectName, TaskID: e.TaskID, Task: e.TaskName,
		StartedAt: v.stamp(e.StartedAt), Minutes: minutes(net), NeedsReview: e.NeedsReview,
	}
	switch {
	case e.StoppedAt != nil:
		out.StoppedAt = v.stamp(*e.StoppedAt)
	case e.PausedAt != nil:
		out.State = statePaused
	default:
		out.State = stateRunning
	}
	if e.UserID != v.callerID {
		out.Person = personOf(e)
	}
	if v.detailed {
		out.Detail = &entryDetail{
			UserID: e.UserID, SpaceID: e.SpaceID, PausedMinutes: int(e.PausedDuration / 60_000),
			Attribution: e.Source, Confidence: e.Confidence,
		}
	}
	return out
}

// stamp renders an instant in the call's timezone.
func (v view) stamp(at time.Time) string {
	return at.In(v.clock.loc).Format(time.RFC3339)
}

// personOf names who booked an entry, marking agents.
func personOf(e timeentries.TimeEntryResponse) string {
	name := e.UserName
	if name == "" {
		name = e.UserEmail
	}
	if e.UserKind == schemas.UserKindAgent {
		name += " (agent)"
	}
	return name
}

// hoursOf converts whole minutes to hours, two decimals.
func hoursOf(m int) float64 {
	return math.Round(float64(m)/60*100) / 100
}

// totalOf is an amount of time over n entries.
func totalOf(d time.Duration, n int) total {
	m := minutes(d)
	return total{Minutes: m, Hours: hoursOf(m), Entries: n}
}
