package mcp

import (
	"cmp"
	"slices"
	"time"
)

// bucket accumulates the time of one project, task, person or day.
type bucket struct {
	id    int64
	name  string
	d     time.Duration
	tasks *tally
}

// tally accumulates time per key.
type tally struct {
	items map[int64]*bucket
}

// newTally returns an empty tally.
func newTally() *tally {
	return &tally{items: map[int64]*bucket{}}
}

// add counts d under id, named name, and returns its bucket.
func (t *tally) add(id int64, name string, d time.Duration) *bucket {
	b, ok := t.items[id]
	if !ok {
		b = &bucket{id: id, name: name, tasks: nil}
		t.items[id] = b
	}
	b.d += d
	return b
}

// addTask counts d under a project and, inside it, a task.
func (t *tally) addTask(projectID int64, project string, taskID int64, task string, d time.Duration) {
	b := t.add(projectID, project, d)
	if b.tasks == nil {
		b.tasks = newTally()
	}
	b.tasks.add(taskID, task, d)
}

// sorted lists the buckets, most time first, then by name.
func (t *tally) sorted() []*bucket {
	list := make([]*bucket, 0, len(t.items))
	for _, b := range t.items {
		list = append(list, b)
	}
	slices.SortFunc(list, func(a, b *bucket) int {
		return cmp.Or(cmp.Compare(b.d, a.d), cmp.Compare(a.name, b.name), cmp.Compare(a.id, b.id))
	})
	return list
}

// lines renders the buckets.
func (t *tally) lines() []line {
	list := t.sorted()
	out := make([]line, len(list))
	for i, b := range list {
		m := minutes(b.d)
		out[i] = line{ID: b.id, Name: b.name, Minutes: m, Hours: hoursOf(m)}
	}
	return out
}

// projects renders the buckets as projects, with their tasks when withTasks is set.
func (t *tally) projects(withTasks bool) []projectLine {
	list := t.sorted()
	out := make([]projectLine, len(list))
	for i, b := range list {
		m := minutes(b.d)
		out[i] = projectLine{ID: b.id, Name: b.name, Minutes: m, Hours: hoursOf(m)}
		if withTasks && b.tasks != nil {
			out[i].Tasks = b.tasks.lines()
		}
	}
	return out
}
