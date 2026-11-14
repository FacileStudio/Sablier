package mcp

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	pageSize      = 50
	maxRows       = 5000
	statusCreated = "created"
	statusUpdated = "updated"
)

var (
	errStop    = stderrors.New("pass stopped_at, or duration_minutes, to say when the work stopped")
	errMove    = stderrors.New("moving an entry to another project needs task_id or task for that project too")
	errNothing = stderrors.New("pass at least one of project_id, task_id, task, started_at or stopped_at to change")
	errTooMany = fmt.Errorf("more than %d entries match. Narrow from and to, or pass project_id", maxRows)
)

// listEntries is sablier_list_entries.
func (t *toolset) listEntries(ctx context.Context, req *sdk.CallToolRequest, in entriesInput) (*sdk.CallToolResult, entriesOutput, error) {
	userID, c, err := t.callerClock(ctx, req, in.Timezone)
	if err != nil {
		return nil, entriesOutput{}, err
	}
	detailed, err := mcpkit.Detailed(in.ResponseFormat)
	if err != nil {
		return nil, entriesOutput{}, err
	}
	offset, err := mcpkit.Offset(in.Cursor)
	if err != nil {
		return nil, entriesOutput{}, err
	}
	from, to, err := c.rangeOf(in.From, in.To)
	if err != nil {
		return nil, entriesOutput{}, err
	}
	entries, err := t.inScope(ctx, userID, in.scopeInput, from, to)
	if err != nil {
		return nil, entriesOutput{}, err
	}
	v := view{clock: c, callerID: userID, detailed: detailed}
	return nil, entriesPage(v, entries, from, to, offset), nil
}

// entriesPage renders the totals of the whole range and one page of entries from offset.
func entriesPage(v view, entries []timeentries.TimeEntryResponse, from, to time.Time, offset int) entriesOutput {
	projects, sum := newTally(), time.Duration(0)
	for _, e := range entries {
		d := within(e, from, to, v.clock.now)
		projects.add(e.ProjectID, e.ProjectName, d)
		sum += d
	}
	out := entriesOutput{From: v.stamp(from), To: v.stamp(to), Total: totalOf(sum, len(entries)), ByProject: projects.projects(false)}
	rest := entries[min(offset, len(entries)):]
	page := make([]entryOut, 0, min(len(rest), pageSize))
	for _, e := range rest[:min(len(rest), pageSize)] {
		page = append(page, v.entry(e))
	}
	n := mcpkit.FitCount(len(page), func(n int) int { out.Entries = page[:n]; return mcpkit.EncodedSize(out) })
	out.Entries = page[:max(n, min(1, len(page)))]
	out.NextCursor = mcpkit.Cursor(offset+len(out.Entries), len(entries))
	return out
}

// inScope reads the entries a report covers, through the REST list's space scope.
func (t *toolset) inScope(ctx context.Context, userID int64, s scopeInput, from, to time.Time) ([]timeentries.TimeEntryResponse, error) {
	who := userID
	switch {
	case s.UserID > 0:
		who = s.UserID
	case s.Everyone:
		who = 0
	}
	entries, err := t.Entries.InRange(ctx, userID, timeentries.RangeFilter{
		Space: scopeOf(s.Space), ProjectID: s.ProjectID, UserID: who, From: from, To: to, Limit: maxRows + 1,
	})
	if err != nil {
		return nil, mcpkit.Failure(err, hintSpace)
	}
	if len(entries) > maxRows {
		return nil, errTooMany
	}
	return entries, nil
}

// addEntry is sablier_add_entry.
func (t *toolset) addEntry(ctx context.Context, req *sdk.CallToolRequest, in addInput) (*sdk.CallToolResult, entryOutput, error) {
	userID, c, err := t.callerClock(ctx, req, in.Timezone)
	if err != nil {
		return nil, entryOutput{}, err
	}
	started, stopped, err := c.interval(in.StartedAt, in.StoppedAt, in.DurationMinutes)
	if err != nil {
		return nil, entryOutput{}, err
	}
	taskID, err := t.taskFor(ctx, userID, in.ProjectID, in.TaskID, in.Task)
	if err != nil {
		return nil, entryOutput{}, err
	}
	entry, err := t.Entries.CreateEntry(ctx, userID, &timeentries.CreateEntryRequest{
		ProjectID: in.ProjectID, TaskID: taskID, StartedAt: started, StoppedAt: stopped,
	})
	if err != nil {
		return nil, entryOutput{}, mcpkit.Failure(err, hintProject)
	}
	out, err := t.entryView(ctx, userID, entry.ID, c)
	return nil, entryOutput{Status: statusCreated, Entry: out}, err
}

// interval reads a manual entry's start and its end, given or from a duration.
func (c clock) interval(start, stop string, duration int) (time.Time, time.Time, error) {
	started, _, err := c.parse("started_at", start)
	if err != nil {
		return started, started, err
	}
	switch {
	case stop != "":
		stopped, _, err := c.parse("stopped_at", stop)
		return started, stopped, err
	case duration > 0:
		return started, started.Add(time.Duration(duration) * time.Minute), nil
	}
	return started, started, errStop
}
