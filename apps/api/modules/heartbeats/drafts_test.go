package heartbeats

import (
	"testing"
	"time"
)

type wantDraft struct {
	start, stop time.Duration
	project     string
	beats       int
}

func TestComputeDrafts(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	id := func(v int64) *int64 { return &v }
	at := func(minutes float64) time.Time { return t0.Add(time.Duration(minutes * float64(time.Minute))) }
	on := func(minutes float64, projectID *int64, hint string) beat {
		return beat{At: at(minutes), Source: "editor", ProjectID: projectID, Hint: hint}
	}
	m := func(minutes float64) time.Duration { return time.Duration(minutes * float64(time.Minute)) }
	cases := []struct {
		name  string
		beats []beat
		busy  []interval
		want  []wantDraft
	}{
		{"lone heartbeat counts two minutes", []beat{on(0, id(1), "")}, nil,
			[]wantDraft{{0, m(2), "id:1", 1}}},
		{"heartbeats within the gap merge", []beat{on(10, id(1), ""), on(0, id(1), "")}, nil,
			[]wantDraft{{0, m(12), "id:1", 2}}},
		{"a gap over fifteen minutes splits", []beat{on(0, id(1), ""), on(20, id(1), "")}, nil,
			[]wantDraft{{0, m(2), "id:1", 1}, {m(20), m(22), "id:1", 1}}},
		{"a project change splits the span", []beat{on(0, id(1), ""), on(5, id(1), ""), on(10, nil, "Other")}, nil,
			[]wantDraft{{0, m(10), "id:1", 2}, {m(10), m(12), "hint:other", 1}}},
		{"a heartbeat without project joins the span", []beat{on(0, nil, ""), on(5, id(2), ""), on(8, nil, "")}, nil,
			[]wantDraft{{0, m(10), "id:2", 3}}},
		{"an entry in the middle cuts the span", []beat{on(0, id(1), ""), on(5, id(1), ""), on(10, id(1), "")},
			[]interval{{at(3), at(8)}},
			[]wantDraft{{0, m(3), "id:1", 1}, {m(8), m(12), "id:1", 1}}},
		{"a sliver under a minute is dropped", []beat{on(0, id(1), "")},
			[]interval{{at(0.5), at(5)}}, nil},
		{"a fully covered span disappears", []beat{on(0, id(1), ""), on(4, id(1), "")},
			[]interval{{at(-1), at(3)}, {at(2), at(30)}}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeDrafts(tc.beats, tc.busy)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d drafts, want %d: %+v", len(got), len(tc.want), got)
			}
			for i, w := range tc.want {
				checkDraft(t, i, got[i], w, t0)
			}
		})
	}
}

func checkDraft(t *testing.T, i int, g span, w wantDraft, t0 time.Time) {
	t.Helper()
	if !g.Start.Equal(t0.Add(w.start)) || !g.Stop.Equal(t0.Add(w.stop)) {
		t.Errorf("draft %d spans %s-%s, want %s-%s", i, g.Start, g.Stop, t0.Add(w.start), t0.Add(w.stop))
	}
	if key := projectKey(g.ProjectID, g.Hint); key != w.project {
		t.Errorf("draft %d project %q, want %q", i, key, w.project)
	}
	if len(g.Beats) != w.beats {
		t.Errorf("draft %d has %d heartbeats, want %d", i, len(g.Beats), w.beats)
	}
}
