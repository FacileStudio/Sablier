package heartbeats

import (
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	mergeGap      = 15 * time.Minute
	loneHeartbeat = 2 * time.Minute
	minDraft      = time.Minute
)

type beat struct {
	At        time.Time
	Source    string
	ProjectID *int64
	Hint      string
}

type interval struct {
	Start time.Time
	Stop  time.Time
}

type span struct {
	Start     time.Time
	Stop      time.Time
	ProjectID *int64
	Hint      string
	Sources   []string
	Beats     []beat
}

// projectKey names the project a heartbeat resolved to; the empty key means it
// carried no project and belongs to whatever span it falls into.
func projectKey(projectID *int64, hint string) string {
	if projectID != nil {
		return "id:" + strconv.FormatInt(*projectID, 10)
	}
	if hint != "" {
		return "hint:" + strings.ToLower(hint)
	}
	return ""
}

func openSpan(b beat) span {
	s := span{Start: b.At, Stop: b.At}
	s.add(b)
	return s
}

func (s *span) add(b beat) {
	s.Beats = append(s.Beats, b)
	if !slices.Contains(s.Sources, b.Source) {
		s.Sources = append(s.Sources, b.Source)
	}
	if end := b.At.Add(loneHeartbeat); end.After(s.Stop) {
		s.Stop = end
	}
	if projectKey(s.ProjectID, s.Hint) == "" {
		s.ProjectID, s.Hint = b.ProjectID, b.Hint
	}
}

func (s span) switchesProject(b beat) bool {
	current, next := projectKey(s.ProjectID, s.Hint), projectKey(b.ProjectID, b.Hint)
	return current != "" && next != "" && current != next
}

// buildSpans merges heartbeats closer than mergeGap into spans, each running to its
// last heartbeat plus loneHeartbeat, and cuts a span where the resolved project changes.
func buildSpans(beats []beat) []span {
	sorted := slices.Clone(beats)
	slices.SortStableFunc(sorted, func(a, b beat) int { return a.At.Compare(b.At) })
	var spans []span
	for _, b := range sorted {
		if len(spans) == 0 {
			spans = append(spans, openSpan(b))
			continue
		}
		current := &spans[len(spans)-1]
		last := current.Beats[len(current.Beats)-1].At
		switch {
		case b.At.Sub(last) > mergeGap:
			spans = append(spans, openSpan(b))
		case current.switchesProject(b):
			current.Stop = b.At
			spans = append(spans, openSpan(b))
		default:
			current.add(b)
		}
	}
	return spans
}

// carve removes busy from every piece, splitting a piece it lands in the middle of.
func carve(pieces []interval, busy interval) []interval {
	var out []interval
	for _, p := range pieces {
		if !busy.Start.Before(p.Stop) || !busy.Stop.After(p.Start) {
			out = append(out, p)
			continue
		}
		if busy.Start.After(p.Start) {
			out = append(out, interval{Start: p.Start, Stop: busy.Start})
		}
		if busy.Stop.Before(p.Stop) {
			out = append(out, interval{Start: busy.Stop, Stop: p.Stop})
		}
	}
	return out
}

func (s span) within(p interval) span {
	piece := span{Start: p.Start, Stop: p.Stop, ProjectID: s.ProjectID, Hint: s.Hint}
	for _, b := range s.Beats {
		if !b.At.Before(p.Start) && b.At.Before(p.Stop) {
			piece.Beats = append(piece.Beats, b)
			if !slices.Contains(piece.Sources, b.Source) {
				piece.Sources = append(piece.Sources, b.Source)
			}
		}
	}
	if len(piece.Sources) == 0 {
		piece.Sources = slices.Clone(s.Sources)
	}
	return piece
}

// computeDrafts turns heartbeats into the spans left once busy time is subtracted,
// dropping any piece shorter than minDraft.
func computeDrafts(beats []beat, busy []interval) []span {
	var drafts []span
	for _, s := range buildSpans(beats) {
		pieces := []interval{{Start: s.Start, Stop: s.Stop}}
		for _, b := range busy {
			pieces = carve(pieces, b)
		}
		for _, p := range pieces {
			if p.Stop.Sub(p.Start) >= minDraft {
				drafts = append(drafts, s.within(p))
			}
		}
	}
	return drafts
}
