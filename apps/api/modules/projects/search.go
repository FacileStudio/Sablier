package projects

import (
	"context"
	"database/sql/driver"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultSearchLimit = 20
	maxSearchLimit     = 50
	maxSearchQuery     = 200
	lastUsedExpr       = "(SELECT max(te.started_at) FROM time_entries te WHERE te.task_id = tasks.id AND te.user_id = ?)"
	searchSelect       = "tasks.id AS task_id, tasks.name AS task_name, tasks.facile_id AS task_facile_id, " +
		"tasks.synced_at IS NOT NULL AS opus, projects.id AS project_id, projects.name AS project_name, " +
		"projects.icon AS project_icon, spaces.id AS space_id, spaces.name AS space_name, " + lastUsedExpr + " AS last_used_at"
	searchRank = "CASE WHEN EXISTS (SELECT 1 FROM time_entries te WHERE te.task_id = tasks.id AND te.user_id = ?) THEN 0 " +
		"WHEN tasks.synced_at IS NOT NULL AND tasks.status <> ? THEN 1 " +
		"WHEN lower(tasks.name) LIKE ? ESCAPE '\\' THEN 2 ELSE 3 END"
)

type searchRow struct {
	TaskID       int64
	TaskName     string
	TaskFacileID *string
	Opus         bool
	ProjectID    int64
	ProjectName  string
	ProjectIcon  *string
	SpaceID      *string
	SpaceName    *string
	LastUsedAt   scannedTime
}

var sqliteTimeLayouts = []string{"2006-01-02 15:04:05.999999999-07:00", "2006-01-02 15:04:05.999999999", time.RFC3339Nano}

type scannedTime struct {
	at *time.Time
}

// Value writes the timestamp back, which gorm requires of any type it scans.
func (s scannedTime) Value() (driver.Value, error) {
	if s.at == nil {
		return nil, nil
	}
	return *s.at, nil
}

// Scan reads an aggregate timestamp: PostgreSQL hands over a time, SQLite the text it stores.
func (s *scannedTime) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		s.at = nil
	case time.Time:
		s.at = &v
	case string:
		for _, layout := range sqliteTimeLayouts {
			if parsed, err := time.Parse(layout, v); err == nil {
				s.at = &parsed
				return nil
			}
		}
		return fmt.Errorf("unreadable timestamp %q", v)
	default:
		return fmt.Errorf("unsupported timestamp type %T", value)
	}
	return nil
}

func likePattern(q string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(strings.ToLower(q))
	return "%" + escaped + "%"
}

// searchTasks ranks the tasks userID can see: the ones userID last worked on first, then
// open Opus tasks, then name matches on the task and finally on the project.
func (service *Service) searchTasks(ctx context.Context, userID int64, q string, limit int) ([]searchRow, error) {
	pattern := likePattern(q)
	query := service.orm.WithContext(ctx).Table("tasks").Select(searchSelect, userID).
		Joins("JOIN projects ON projects.id = tasks.project_id").
		Joins("LEFT JOIN spaces ON spaces.id = CAST(projects.space_id AS TEXT)")
	query = schemas.VisibleSpaces(query, "projects.space_id", userID).
		Where("(lower(tasks.name) LIKE ? ESCAPE '\\' OR lower(projects.name) LIKE ? ESCAPE '\\')", pattern, pattern).
		Order(clause.OrderBy{Expression: gorm.Expr(searchRank+", "+lastUsedExpr+" DESC, tasks.name asc, tasks.id asc",
			userID, schemas.StatusDone, pattern, userID)}).
		Limit(limit)
	var rows []searchRow
	if err := query.Find(&rows).Error; err != nil {
		return nil, errors.Internal("failed to search tasks", err)
	}
	return rows, nil
}

func toSearchResult(row searchRow) TaskSearchResult {
	result := TaskSearchResult{
		Task:       TaskRef{ID: row.TaskID, Name: row.TaskName, FacileID: row.TaskFacileID},
		Project:    ProjectRef{ID: row.ProjectID, Name: row.ProjectName, Icon: row.ProjectIcon},
		LastUsedAt: row.LastUsedAt.at,
		Opus:       row.Opus,
	}
	if row.SpaceID != nil && row.SpaceName != nil {
		result.Space = &SpaceRef{ID: *row.SpaceID, Name: *row.SpaceName}
	}
	return result
}

func (service *Service) handleSearch(w http.ResponseWriter, request *http.Request) {
	identity, _ := authcontext.IdentityFromContext(request.Context())
	userID, err := strconv.ParseInt(identity.UserID, 10, 64)
	if err != nil {
		httpjson.WriteError(w, errors.Invalid("invalid user id"))
		return
	}
	q, limit := strings.TrimSpace(request.URL.Query().Get("q")), defaultSearchLimit
	if raw := request.URL.Query().Get("limit"); raw != "" {
		if limit, err = strconv.Atoi(raw); err != nil || limit < 1 || limit > maxSearchLimit {
			httpjson.WriteError(w, errors.Invalid("limit must be between 1 and 50"))
			return
		}
	}
	if len(q) > maxSearchQuery {
		httpjson.WriteError(w, errors.Invalid("q is at most 200 characters"))
		return
	}
	rows, err := service.searchTasks(request.Context(), userID, q, limit)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	results := make([]TaskSearchResult, len(rows))
	for i, row := range rows {
		results[i] = toSearchResult(row)
	}
	httpjson.WriteJSON(w, http.StatusOK, TaskSearchResponse{Results: results})
}
