package agents

import (
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

var base = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func entry(userID int64, fromMin, toMin int, cowork bool) schemas.TimeEntry {
	stopped := base.Add(time.Duration(toMin) * time.Minute)
	return schemas.TimeEntry{UserID: userID, StartedAt: base.Add(time.Duration(fromMin) * time.Minute), StoppedAt: &stopped, Cowork: cowork}
}

func TestSummarizePairedOverlap(t *testing.T) {
	const owner, agent = 1, 2
	entries := []schemas.TimeEntry{
		entry(owner, 0, 60, false),
		entry(owner, 30, 90, false),
		entry(owner, 200, 230, false),
		entry(agent, 45, 120, true),
		entry(agent, 100, 130, false),
		entry(agent, 220, 300, false),
	}
	window := []time.Time{base, base.Add(240 * time.Minute)}
	rows := summarize(owner, []schemas.User{{ID: agent}}, entries, window[0], window[1], base.Add(10*time.Hour))
	if len(rows) != 1 {
		t.Fatalf("rows: %+v", rows)
	}
	got := rows[0]
	want := PairedResponse{AgentID: agent, AgentSeconds: (75 + 30 + 20) * 60, CoworkSeconds: 75 * 60, PairedSeconds: (45 + 10) * 60}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestPausedTimeIsSpreadOverTheEntry(t *testing.T) {
	e := entry(2, 0, 100, false)
	e.PausedDurationMs = 50 * 60 * 1000
	clipped, ok := clip(e, base.Add(50*time.Minute), base.Add(10*time.Hour), base)
	if !ok {
		t.Fatal("entry should overlap the window")
	}
	if got := activeSeconds(e, clipped, base); got != 25*60 {
		t.Fatalf("got %v seconds, want 1500", got)
	}
}
