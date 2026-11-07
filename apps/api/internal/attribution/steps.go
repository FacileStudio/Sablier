package attribution

import (
	"errors"
	"slices"

	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/gorm"
)

const (
	nameConfidence    = 0.9
	runningConfidence = 0.85
)

type step func(db *gorm.DB, ownerID int64, item Facts, space *string) (*Result, error)

func visibleTasks(db *gorm.DB, ownerID int64) *gorm.DB {
	query := db.Model(&schemas.Task{}).Select("tasks.*").Joins("JOIN projects ON projects.id = tasks.project_id")
	return schemas.VisibleSpaces(query, "projects.space_id", ownerID)
}

func visibleProjects(db *gorm.DB, ownerID int64) *gorm.DB {
	return schemas.VisibleSpaces(db.Model(&schemas.Project{}), "projects.space_id", ownerID)
}

func found(source string, confidence float64, project *schemas.Project, task *schemas.Task) *Result {
	return &Result{Project: project, Task: task, Source: source, Confidence: confidence}
}

func withProject(db *gorm.DB, source string, confidence float64, task schemas.Task) (*Result, error) {
	var project schemas.Project
	if err := db.First(&project, task.ProjectID).Error; err != nil {
		return nil, err
	}
	return found(source, confidence, &project, &task), nil
}

func explicit(db *gorm.DB, ownerID int64, item Facts, _ *string) (*Result, error) {
	if item.TaskID > 0 {
		var task schemas.Task
		err := visibleTasks(db, ownerID).Where("tasks.id = ?", item.TaskID).First(&task).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotVisible
		}
		if err != nil {
			return nil, err
		}
		return withProject(db, SourceExplicit, 1, task)
	}
	if item.ProjectID <= 0 {
		return nil, nil
	}
	project, err := schemas.VisibleProject(db, ownerID, item.ProjectID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotVisible
	}
	if err != nil {
		return nil, err
	}
	return found(SourceExplicit, 1, project, nil), nil
}

func opus(db *gorm.DB, ownerID int64, item Facts, _ *string) (*Result, error) {
	ids := item.OpusTaskIDs
	if item.OpusUpdatedTaskID != "" {
		ids = append([]string{item.OpusUpdatedTaskID}, ids...)
	}
	if len(ids) > 0 {
		var tasks []schemas.Task
		if err := visibleTasks(db, ownerID).Where("tasks.facile_id IN ?", ids).Find(&tasks).Error; err != nil {
			return nil, err
		}
		slices.SortStableFunc(tasks, func(a, b schemas.Task) int {
			return slices.Index(ids, *a.FacileID) - slices.Index(ids, *b.FacileID)
		})
		if len(tasks) > 0 {
			return withProject(db, SourceOpus, 1, tasks[0])
		}
	}
	if item.OpusProjectID == "" {
		return nil, nil
	}
	return firstProject(visibleProjects(db, ownerID).Where("projects.facile_id = ?", item.OpusProjectID), SourceOpus, 1)
}

func firstProject(query *gorm.DB, source string, confidence float64) (*Result, error) {
	var project schemas.Project
	err := query.Order("projects.id asc").First(&project).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return found(source, confidence, &project, nil), nil
}

func byName(db *gorm.DB, ownerID int64, item Facts, space *string) (*Result, error) {
	if item.ProjectHint == "" {
		return nil, nil
	}
	named := func() *gorm.DB {
		return visibleProjects(db, ownerID).Where("lower(projects.name) = lower(?)", item.ProjectHint)
	}
	if space != nil {
		result, err := firstProject(named().Where("CAST(projects.space_id AS TEXT) = ?", *space), SourceName, nameConfidence)
		if result != nil || err != nil {
			return result, err
		}
	}
	return firstProject(named(), SourceName, nameConfidence)
}

func running(db *gorm.DB, ownerID int64, item Facts, _ *string) (*Result, error) {
	if item.At.IsZero() {
		return nil, nil
	}
	var entry schemas.TimeEntry
	err := db.Where("user_id = ? AND started_at <= ? AND (stopped_at IS NULL OR stopped_at >= ?)", ownerID, item.At, item.At).
		Where("project_id IN (?)", visibleProjects(db.Session(&gorm.Session{NewDB: true}), ownerID).Select("projects.id")).
		Order("started_at desc").First(&entry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var task schemas.Task
	if entry.TaskID > 0 && db.Where("id = ? AND project_id = ?", entry.TaskID, entry.ProjectID).First(&task).Error == nil {
		return withProject(db, SourceRunning, runningConfidence, task)
	}
	return firstProject(db.Model(&schemas.Project{}).Where("projects.id = ?", entry.ProjectID), SourceRunning, runningConfidence)
}

func memberSpace(db *gorm.DB, ownerID int64, item Facts) (*string, error) {
	query := db.Model(&schemas.Space{}).Where("spaces.id IN (?)", schemas.MemberSpaces(db, ownerID))
	switch {
	case item.RegistreSpaceID != "":
		query = query.Where("spaces.registre_space_id = ?", item.RegistreSpaceID)
	case item.OpusWorkspace != "":
		query = query.Where("lower(spaces.name) = lower(?)", item.OpusWorkspace)
	default:
		return nil, nil
	}
	var ids []string
	if err := query.Order("spaces.id asc").Limit(1).Pluck("spaces.id", &ids).Error; err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	return &ids[0], nil
}
