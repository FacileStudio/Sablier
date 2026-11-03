package agents

import (
	"sort"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

type span struct {
	start time.Time
	end   time.Time
}

// clip bounds an entry to [from, to). A running entry runs until now.
func clip(entry schemas.TimeEntry, from, to, now time.Time) (span, bool) {
	end := now
	if entry.StoppedAt != nil {
		end = *entry.StoppedAt
	}
	clipped := span{start: entry.StartedAt, end: end}
	if clipped.start.Before(from) {
		clipped.start = from
	}
	if clipped.end.After(to) {
		clipped.end = to
	}
	return clipped, clipped.end.After(clipped.start)
}

// activeSeconds is the clipped part of an entry minus its pauses, which are
// spread evenly over the entry since only their total is stored.
func activeSeconds(entry schemas.TimeEntry, clipped span, now time.Time) float64 {
	end := now
	if entry.StoppedAt != nil {
		end = *entry.StoppedAt
	}
	full := end.Sub(entry.StartedAt).Seconds()
	if full <= 0 {
		return 0
	}
	paused := float64(entry.PausedDurationMs) / 1000
	ratio := (full - paused) / full
	if ratio < 0 {
		ratio = 0
	}
	return clipped.end.Sub(clipped.start).Seconds() * ratio
}

// merge sorts spans and joins the ones that touch or overlap.
func merge(spans []span) []span {
	sorted := append([]span(nil), spans...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].start.Before(sorted[j].start) })
	var merged []span
	for _, current := range sorted {
		last := len(merged) - 1
		if last >= 0 && !current.start.After(merged[last].end) {
			if current.end.After(merged[last].end) {
				merged[last].end = current.end
			}
			continue
		}
		merged = append(merged, current)
	}
	return merged
}

// overlapSeconds is the length of time covered by both a and b, each counted once.
func overlapSeconds(a, b []span) float64 {
	left, right := merge(a), merge(b)
	total := 0.0
	for i, j := 0, 0; i < len(left) && j < len(right); {
		start := later(left[i].start, right[j].start)
		end := earlier(left[i].end, right[j].end)
		if end.After(start) {
			total += end.Sub(start).Seconds()
		}
		if left[i].end.Before(right[j].end) {
			i++
		} else {
			j++
		}
	}
	return total
}

func later(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func earlier(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
