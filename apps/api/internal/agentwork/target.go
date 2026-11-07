package agentwork

import (
	"errors"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/gorm"
)

const (
	// SessionsTaskName is the task an agent session lands on in a project it matched.
	SessionsTaskName = "Agent sessions"
	// CatchAllProject parks sessions that matched no project, under a task named after the hint.
	CatchAllProject      = "Agent work"
	catchAllOwnerID      = 1
	catchAllProjectNotes = "Agent sessions without a matching Sablier project"
)

// ErrProjectNotVisible is returned when an explicit project or task id is not one the owner can see.
var ErrProjectNotVisible = attribution.ErrNotVisible

// Target is where an agent session is booked, and why it landed there.
type Target struct {
	Project     schemas.Project
	Task        schemas.Task
	Attribution schemas.Attribution
}

// bookTarget turns the resolver's answer into rows to book on: the task it found, else
// the "Agent sessions" task of its project, else the shared "Agent work" catch-all under
// a task named after the hint.
func bookTarget(tx *gorm.DB, result attribution.Result, hint string) (*Target, error) {
	source, confidence := result.Source, result.Confidence
	booked := schemas.Attribution{AttributionSource: &source, AttributionConfidence: &confidence, NeedsReview: result.NeedsReview}
	var target *Target
	var err error
	switch {
	case result.Task != nil:
		target = &Target{Project: *result.Project, Task: *result.Task}
	case result.Project != nil:
		target, err = withTask(tx, *result.Project, SessionsTaskName)
	default:
		target, err = catchAllTarget(tx, hint)
	}
	if err != nil {
		return nil, err
	}
	target.Attribution = booked
	return target, nil
}

func catchAllTarget(tx *gorm.DB, hint string) (*Target, error) {
	taskName := hint
	if taskName == "" {
		taskName = SessionsTaskName
	}
	var project schemas.Project
	err := tx.Where("lower(name) = lower(?)", CatchAllProject).Order("id asc").First(&project).Error
	if err == nil {
		return withTask(tx, project, taskName)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	facileID := schemas.NewFacileID()
	project = schemas.Project{
		Name:        CatchAllProject,
		Description: catchAllProjectNotes,
		OwnerID:     catchAllOwnerID,
		FacileID:    &facileID,
	}
	if err := tx.Create(&project).Error; err != nil {
		return nil, err
	}
	return withTask(tx, project, taskName)
}

func withTask(tx *gorm.DB, project schemas.Project, name string) (*Target, error) {
	var task schemas.Task
	err := tx.Where("project_id = ? AND lower(name) = lower(?)", project.ID, name).First(&task).Error
	if err == nil {
		return &Target{Project: project, Task: task}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	facileID := schemas.NewFacileID()
	task = schemas.Task{ProjectID: project.ID, Name: name, FacileID: &facileID}
	if err := tx.Create(&task).Error; err != nil {
		return nil, err
	}
	return &Target{Project: project, Task: task}, nil
}
