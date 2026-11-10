package mcp

import (
	"context"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// week is the Monday-to-Monday range a summary covers, and its days.
type week struct {
	start time.Time
	days  [7]time.Time
}

// weekOf finds the week holding date, this week when empty.
func (c clock) weekOf(date string) (week, error) {
	day := c.now
	if date != "" {
		at, _, err := c.parse("week_of", date)
		if err != nil {
			return week{}, err
		}
		day = at
	}
	midnight := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, c.loc)
	w := week{start: midnight.AddDate(0, 0, -((int(midnight.Weekday()) + 6) % 7))}
	for i := range w.days {
		w.days[i] = w.start.AddDate(0, 0, i)
	}
	return w, nil
}

// end is the Monday after the week.
func (w week) end() time.Time {
	return w.start.AddDate(0, 0, 7)
}

// summary is sablier_weekly_summary.
func (t *toolset) summary(ctx context.Context, req *sdk.CallToolRequest, in summaryInput) (*sdk.CallToolResult, summaryOutput, error) {
	userID, c, err := t.callerClock(ctx, req, in.Timezone)
	if err != nil {
		return nil, summaryOutput{}, err
	}
	w, err := c.weekOf(in.WeekOf)
	if err != nil {
		return nil, summaryOutput{}, err
	}
	entries, err := t.inScope(ctx, userID, in.scopeInput, w.start, w.end())
	if err != nil {
		return nil, summaryOutput{}, err
	}
	out := summarize(c, w, entries, in.Everyone || in.UserID > 0)
	out.Timezone = c.loc.String()
	return nil, fit(out), nil
}

// summarize adds up a week of entries by day, project and task, and person when asked.
func summarize(c clock, w week, entries []timeentries.TimeEntryResponse, people bool) summaryOutput {
	projects, persons, sum := newTally(), newTally(), time.Duration(0)
	var days [7]time.Duration
	out := summaryOutput{WeekStart: w.start.Format(layoutDate), WeekEnd: w.days[6].Format(layoutDate), Days: make([]line, 7)}
	for i, day := range w.days {
		out.Days[i] = line{Name: day.Format(layoutDate), Weekday: day.Weekday().String()}
	}
	for _, e := range entries {
		d := within(e, w.start, w.end(), c.now)
		sum += d
		projects.addTask(e.ProjectID, e.ProjectName, e.TaskID, e.TaskName, d)
		persons.add(e.UserID, personOf(e), d)
		for i, day := range w.days {
			days[i] += within(e, day, day.AddDate(0, 0, 1), c.now)
		}
		out.Running = out.Running || e.StoppedAt == nil
		if e.NeedsReview {
			out.NeedsReview++
		}
	}
	for i, d := range days {
		out.Days[i].Minutes = minutes(d)
		out.Days[i].Hours = hoursOf(out.Days[i].Minutes)
	}
	out.Total, out.ByProject = totalOf(sum, len(entries)), projects.projects(true)
	if people {
		out.ByPerson = persons.lines()
	}
	return out
}

// fit drops the smallest projects when the summary would pass the result budget.
func fit(out summaryOutput) summaryOutput {
	all := out.ByProject
	n := mcpkit.FitCount(len(all), func(n int) int { out.ByProject = all[:n]; return mcpkit.EncodedSize(out) })
	out.ByProject, out.Truncated = all[:n], n < len(all)
	return out
}
