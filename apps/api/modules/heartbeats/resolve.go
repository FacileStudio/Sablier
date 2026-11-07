package heartbeats

import (
	"context"
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

type projectIndex struct {
	visible map[int64]bool
	byName  map[string]int64
}

// loadProjects indexes the projects userID can see: every project outside a space,
// plus the projects of the spaces the user belongs to.
func (service *Service) loadProjects(ctx context.Context, userID int64) (projectIndex, error) {
	var projects []schemas.Project
	err := schemas.VisibleSpaces(service.orm.WithContext(ctx).Model(&schemas.Project{}), schemas.Projects, userID).
		Order("id asc").Find(&projects).Error
	if err != nil {
		return projectIndex{}, errors.Internal("failed to list projects", err)
	}
	index := projectIndex{visible: map[int64]bool{}, byName: map[string]int64{}}
	for _, p := range projects {
		index.visible[p.ID] = true
		name := strings.ToLower(strings.TrimSpace(p.Name))
		if _, taken := index.byName[name]; !taken {
			index.byName[name] = p.ID
		}
	}
	return index, nil
}

// resolve picks the heartbeat's project: an explicit visible id first, then a
// case-insensitive name match of the hint, else none with the hint kept.
func (index projectIndex) resolve(h schemas.Heartbeat) beat {
	b := beat{At: h.At.UTC(), Source: h.Source, Hint: h.ProjectHint, Facts: h.HeartbeatFacts, Entity: h.Entity}
	if h.ProjectID != nil && index.visible[*h.ProjectID] {
		id := *h.ProjectID
		b.ProjectID, b.Explicit = &id, true
		return b
	}
	if id, ok := index.byName[strings.ToLower(strings.TrimSpace(h.ProjectHint))]; ok && h.ProjectHint != "" {
		b.ProjectID = &id
	}
	return b
}
