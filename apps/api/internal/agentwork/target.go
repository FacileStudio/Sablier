package agentwork

import (
	"errors"
	"time"

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

// ErrProjectNotVisible is returned when an explicit project id is not one the owner can see.
var ErrProjectNotVisible = errors.New("project is not visible to the owner")

// Target is where an agent session is booked.
type Target struct {
	Project schemas.Project
	Task    schemas.Task
}

// ResolveTarget picks the project and task for an owner's agent session, in order: the
// explicit project id, a project whose name matches the hint, the project of the
// owner's own entry running at the session start, then the shared "Agent work"
// catch-all under a task named after the hint. Only projects the owner can see count:
// those outside any space, and those in a space the owner belongs to.
func ResolveTarget(tx *gorm.DB, ownerID, projectID int64, hint string, at time.Time) (*Target, error) {
	if projectID > 0 {
		return targetInProject(tx, ownerID, projectID)
	}
	if hint != "" {
		var project schemas.Project
		err := visibleProject(tx, ownerID).Where("lower(name) = lower(?)", hint).Order("id asc").First(&project).Error
		if err == nil {
			return withTask(tx, project, SessionsTaskName)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if target, err := runningTarget(tx, ownerID, at); target != nil || err != nil {
		return target, err
	}
	return catchAllTarget(tx, hint)
}

func visibleProject(tx *gorm.DB, ownerID int64) *gorm.DB {
	return tx.Model(&schemas.Project{}).Where(
		"space_id IS NULL OR CAST(space_id AS TEXT) IN (SELECT space_id FROM space_members WHERE user_id = ?)", ownerID)
}

func targetInProject(tx *gorm.DB, ownerID, projectID int64) (*Target, error) {
	var project schemas.Project
	err := visibleProject(tx, ownerID).Where("id = ?", projectID).First(&project).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProjectNotVisible
	}
	if err != nil {
		return nil, err
	}
	return withTask(tx, project, SessionsTaskName)
}

func runningTarget(tx *gorm.DB, ownerID int64, at time.Time) (*Target, error) {
	var entry schemas.TimeEntry
	err := tx.Where("user_id = ? AND started_at <= ? AND (stopped_at IS NULL OR stopped_at >= ?)", ownerID, at, at).
		Where("project_id IN (?)", visibleProject(tx, ownerID).Select("id")).
		Order("started_at desc").First(&entry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var project schemas.Project
	if err := tx.First(&project, entry.ProjectID).Error; err != nil {
		return nil, err
	}
	var task schemas.Task
	if entry.TaskID > 0 && tx.Where("id = ? AND project_id = ?", entry.TaskID, project.ID).First(&task).Error == nil {
		return &Target{Project: project, Task: task}, nil
	}
	return withTask(tx, project, SessionsTaskName)
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
