package timeentries

import (
	"context"
	stderrors "errors"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

const rowSelect = "time_entries.*, users.email as user_email, users.name as user_name, users.color as user_color, " +
	"users.kind as user_kind, " + schemas.AvatarSelectExpr + " as user_avatar_url, tasks.name as task_name, " +
	"projects.name as project_name"

type timeEntryRow struct {
	schemas.TimeEntry
	UserEmail     string
	UserName      string
	UserColor     string
	UserAvatarURL string
	UserKind      string
	TaskName      string
	ProjectName   string
}

// listFilter is what GET /time-entries narrows on. Space follows schemas.ScopeSpace.
type listFilter struct {
	ProjectID int64
	UserID    int64
	Space     string
	Running   bool
}

func (service *Service) rows(ctx context.Context) *gorm.DB {
	return service.orm.WithContext(ctx).
		Model(&schemas.TimeEntry{}).
		Select(rowSelect).
		Joins("JOIN users ON users.id = time_entries.user_id").
		Joins("LEFT JOIN tasks ON tasks.id = time_entries.task_id").
		Joins("LEFT JOIN projects ON projects.id = time_entries.project_id")
}

func spaceError(err error) error {
	if stderrors.Is(err, schemas.ErrSpaceNotVisible) {
		return errors.NotFound("space not found")
	}
	return errors.Internal("failed to check the space", err)
}

func (service *Service) listEntries(ctx context.Context, callerID int64, filter listFilter) ([]timeEntryRow, error) {
	query, err := schemas.ScopeSpace(service.rows(ctx), schemas.Entries, callerID, filter.Space)
	if err != nil {
		return nil, spaceError(err)
	}
	if filter.ProjectID > 0 {
		query = query.Where("time_entries.project_id = ?", filter.ProjectID)
	}
	if filter.UserID > 0 {
		query = query.Where("time_entries.user_id = ?", filter.UserID)
	}
	if filter.Running {
		query = query.Where("time_entries.stopped_at IS NULL").Order("time_entries.started_at asc")
	} else {
		query = query.Order("time_entries.started_at desc").Limit(100)
	}
	var records []timeEntryRow
	if err := query.Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list entries", err)
	}
	return records, nil
}

func (service *Service) projectFor(ctx context.Context, userID, projectID int64) (*schemas.Project, error) {
	project, err := schemas.VisibleProject(service.orm.WithContext(ctx), userID, projectID)
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Invalid("project_id is not a project you can see")
	}
	if err != nil {
		return nil, errors.Internal("failed to get project", err)
	}
	return project, nil
}
