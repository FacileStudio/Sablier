package mcp

import (
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
)

// TestWithinSplitsAcrossDaysAndPauses checks an entry over midnight splits by the hours
// on each side, with its pause spread evenly, and a running one counts up to now.
func TestWithinSplitsAcrossDaysAndPauses(t *testing.T) {
	paris, _ := time.LoadLocation("Europe/Paris")
	start := time.Date(2026, 11, 9, 22, 0, 0, 0, paris)
	stop := start.Add(4 * time.Hour)
	e := timeentries.TimeEntryResponse{StartedAt: start, StoppedAt: &stop, PausedDuration: int64(time.Hour / time.Millisecond)}
	midnight := time.Date(2026, 11, 10, 0, 0, 0, 0, paris)
	if got := within(e, start.Add(-time.Hour), midnight, stop); got != 90*time.Minute {
		t.Errorf("before midnight: got %v, want 1h30m", got)
	}
	if got := within(e, midnight, midnight.AddDate(0, 0, 1), stop); got != 90*time.Minute {
		t.Errorf("after midnight: got %v, want 1h30m", got)
	}
	e.StoppedAt = nil
	if got, _ := worked(e, start.Add(2*time.Hour)); got != time.Hour {
		t.Errorf("a running entry counts until now minus pauses: got %v", got)
	}
}

// TestWeekAndRange checks weeks start on Monday across a DST change, and a date as to
// includes that day.
func TestWeekAndRange(t *testing.T) {
	c, err := clockOf("Europe/Paris", time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	w, _ := c.weekOf("2026-10-25")
	if w.start.Format(time.RFC3339) != "2026-10-19T00:00:00+02:00" || w.days[6].Format(layoutDate) != "2026-10-25" {
		t.Errorf("a Sunday belongs to the week from Monday, got %s", w.start)
	}
	if w.end().Sub(w.start) != 7*24*time.Hour+time.Hour {
		t.Errorf("the week of the change to winter time lasts 169 hours, got %v", w.end().Sub(w.start))
	}
	if now, _ := c.weekOf(""); now.start.Format(layoutDate) != "2026-10-26" {
		t.Errorf("this week starts on Monday 26 October, got %s", now.start)
	}
	from, to, err := c.rangeOf("2026-10-01", "2026-10-02")
	if err != nil || to.Sub(from) != 48*time.Hour {
		t.Errorf("from and to dates cover both days, got %v %v %v", from, to, err)
	}
	if _, _, err := c.rangeOf("2026-10-02", "2026-10-01"); err == nil {
		t.Error("a reversed range must be refused")
	}
	if _, err := clockOf("Mars/Olympus", time.Now()); err == nil {
		t.Error("an unknown timezone must be refused")
	}
}
