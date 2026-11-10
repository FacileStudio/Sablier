package mcp

import (
	"context"
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	tasksPerProject = 50
	tasksOneProject = 500
	projectsPerPage = 50
)

// catalog is what sablier_list_projects filters: the projects in scope, their tasks and
// the spaces they live in.
type catalog struct {
	in     projectsInput
	tasks  map[int64][]schemas.Task
	spaces map[string]string
}

// listProjects is sablier_list_projects.
func (t *toolset) listProjects(ctx context.Context, req *sdk.CallToolRequest, in projectsInput) (*sdk.CallToolResult, projectsOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, projectsOutput{}, err
	}
	var at cursor
	if err := mcpkit.DecodeCursor(in.Cursor, &at); err != nil {
		return nil, projectsOutput{}, err
	}
	spaces, err := t.Spaces.Of(ctx, userID)
	if err != nil {
		return nil, projectsOutput{}, mcpkit.Failure(err, "")
	}
	projects, tasks, err := t.Projects.Catalog(ctx, userID, scopeOf(in.Space))
	if err != nil {
		return nil, projectsOutput{}, mcpkit.Failure(err, hintSpace)
	}
	cat := catalog{in: in, tasks: map[int64][]schemas.Task{}, spaces: map[string]string{}}
	out := projectsOutput{Spaces: make([]spaceOut, len(spaces))}
	for i, space := range spaces {
		out.Spaces[i], cat.spaces[space.ID] = spaceOut{ID: space.ID, Name: space.Name}, space.Name
	}
	for _, task := range tasks {
		cat.tasks[task.ProjectID] = append(cat.tasks[task.ProjectID], task)
	}
	return nil, cat.page(out, cat.matching(projects), at.Offset), nil
}

// matching renders the projects the filters keep, in the REST list's order.
func (c catalog) matching(projects []schemas.Project) []projectOut {
	query := strings.ToLower(strings.TrimSpace(c.in.Query))
	limit := tasksPerProject
	if c.in.ProjectID > 0 {
		limit = tasksOneProject
	}
	kept := []projectOut{}
	for _, project := range projects {
		if c.in.ProjectID > 0 && project.ID != c.in.ProjectID {
			continue
		}
		named := query == "" || strings.Contains(strings.ToLower(project.Name), query)
		tasks := c.tasksOf(project.ID, query, named)
		if !named && len(tasks) == 0 {
			continue
		}
		out := projectOut{ID: project.ID, Name: project.Name, Tasks: tasks[:min(len(tasks), limit)], MoreTasks: max(len(tasks)-limit, 0)}
		if project.SpaceID != nil {
			out.SpaceID, out.Space = *project.SpaceID, c.spaces[*project.SpaceID]
		}
		kept = append(kept, out)
	}
	return kept
}

// tasksOf lists a project's tasks: open ones unless include_done, and only those whose
// name matches query unless the project's own name did.
func (c catalog) tasksOf(projectID int64, query string, named bool) []taskOut {
	out := []taskOut{}
	for _, task := range c.tasks[projectID] {
		if task.Status == schemas.StatusDone && !c.in.IncludeDone {
			continue
		}
		if !named && !strings.Contains(strings.ToLower(task.Name), query) {
			continue
		}
		out = append(out, taskOut{ID: task.ID, Name: task.Name, Status: task.Status})
	}
	return out
}

// page keeps one page of projects from offset, under the result budget.
func (c catalog) page(out projectsOutput, projects []projectOut, offset int) projectsOutput {
	rest := projects[min(offset, len(projects)):]
	rest = rest[:min(len(rest), projectsPerPage)]
	n := mcpkit.FitCount(len(rest), func(n int) int { out.Projects = rest[:n]; return mcpkit.EncodedSize(out) })
	out.Projects = rest[:max(n, min(1, len(rest)))]
	if next := offset + len(out.Projects); next < len(projects) {
		out.NextCursor = mcpkit.EncodeCursor(cursor{Offset: next})
	}
	return out
}
