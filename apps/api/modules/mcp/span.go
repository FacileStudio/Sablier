package mcp

import (
	"errors"
	"fmt"
	"math"
	"time"
	_ "time/tzdata"

	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
)

const (
	layoutDate  = "2006-01-02"
	layoutLocal = "2006-01-02T15:04"
	maxRange    = 366 * 24 * time.Hour
	defaultSpan = 7 * 24 * time.Hour
)

var (
	errZone  = errors.New(`timezone must be an IANA name such as "Europe/Paris" or "UTC"`)
	errOrder = errors.New("to must be after from")
	errWide  = errors.New("the range is longer than 366 days. Narrow from and to")
)

// clock is the timezone a call reads and writes times in, and the instant it runs at.
type clock struct {
	loc *time.Location
	now time.Time
}

// clockOf loads timezone, UTC when empty.
func clockOf(timezone string, now time.Time) (clock, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return clock{}, errZone
	}
	return clock{loc: loc, now: now.In(loc)}, nil
}

// parse reads an RFC 3339 time, a local time like 2026-11-10T09:30, or a date, which is
// the start of that day. dateOnly reports the last form.
func (c clock) parse(field, value string) (at time.Time, dateOnly bool, err error) {
	if at, err = time.Parse(time.RFC3339, value); err == nil {
		return at.In(c.loc), false, nil
	}
	if at, err = time.ParseInLocation(layoutLocal, value, c.loc); err == nil {
		return at, false, nil
	}
	if at, err = time.ParseInLocation(layoutDate, value, c.loc); err == nil {
		return at, true, nil
	}
	return time.Time{}, false, fmt.Errorf("%s must be a date (2026-11-10), a local time (2026-11-10T09:30) "+
		"or an RFC 3339 time (2026-11-10T09:30:00+01:00)", field)
}

// rangeOf reads from and to, the last 7 days by default. A date as to includes that day.
func (c clock) rangeOf(from, to string) (time.Time, time.Time, error) {
	end, start := c.now, c.now.Add(-defaultSpan)
	if to != "" {
		at, dateOnly, err := c.parse("to", to)
		if err != nil {
			return start, end, err
		}
		end = at
		if dateOnly {
			end = at.AddDate(0, 0, 1)
		}
	}
	if from != "" {
		at, _, err := c.parse("from", from)
		if err != nil {
			return start, end, err
		}
		start = at
	} else if to != "" {
		start = end.Add(-defaultSpan)
	}
	switch {
	case !end.After(start):
		return start, end, errOrder
	case end.Sub(start) > maxRange:
		return start, end, errWide
	}
	return start, end, nil
}

// worked is the time an entry counts: until it stopped, or until now while it runs,
// minus its pauses, including the one in progress.
func worked(e timeentries.TimeEntryResponse, now time.Time) (time.Duration, time.Time) {
	end := now
	if e.StoppedAt != nil {
		end = *e.StoppedAt
	}
	paused := time.Duration(e.PausedDuration) * time.Millisecond
	if e.PausedAt != nil && e.StoppedAt == nil {
		paused += now.Sub(*e.PausedAt)
	}
	return max(end.Sub(e.StartedAt)-paused, 0), end
}

// within is the share of an entry's worked time that falls in [from, to), pauses spread
// evenly over the entry.
func within(e timeentries.TimeEntryResponse, from, to, now time.Time) time.Duration {
	net, end := worked(e, now)
	span := end.Sub(e.StartedAt)
	lo, hi := later(e.StartedAt, from), earlier(end, to)
	if span <= 0 || !hi.After(lo) {
		return 0
	}
	return time.Duration(float64(net) * float64(hi.Sub(lo)) / float64(span))
}

// later is the later of two instants.
func later(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// earlier is the earlier of two instants.
func earlier(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// minutes rounds a duration to whole minutes.
func minutes(d time.Duration) int {
	return int(math.Round(d.Minutes()))
}
