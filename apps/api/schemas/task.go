package schemas

import (
	"fmt"
	"strings"
	"time"
)

const (
	StatusTodo       = "to-do"
	StatusInProgress = "in-progress"
	StatusInReview   = "in-review"
	StatusDone       = "done"
)

var validStatuses = map[string]bool{
	StatusTodo:       true,
	StatusInProgress: true,
	StatusInReview:   true,
	StatusDone:       true,
}

// IsValidStatus reports whether s is one of the known task statuses.
func IsValidStatus(s string) bool {
	return validStatuses[s]
}

// NormalizeStatus returns s when it is valid, otherwise the default to-do status.
func NormalizeStatus(s string) string {
	if validStatuses[s] {
		return s
	}
	return StatusTodo
}

// ValidateStatus returns an error unless s is a known task status.
func ValidateStatus(s string) error {
	if validStatuses[s] {
		return nil
	}
	return fmt.Errorf("invalid status %q, valid: %s", s, strings.Join(AllStatuses(), ", "))
}

// AllStatuses lists every valid task status.
func AllStatuses() []string {
	return []string{StatusTodo, StatusInProgress, StatusInReview, StatusDone}
}

// Task is a unit of work inside a project.
type Task struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	ProjectID int64     `gorm:"column:project_id;index;uniqueIndex:idx_tasks_project_name,priority:1"`
	Name      string    `gorm:"column:name;uniqueIndex:idx_tasks_project_name,priority:2"`
	Status    string    `gorm:"column:status;default:to-do"`
	ActorID   *int64    `gorm:"column:actor_id;index"`
	SpaceID   *string   `gorm:"column:space_id;type:uuid;index" json:"space_id,omitempty"`
	FacileID  *string   `gorm:"column:facile_id;uniqueIndex" json:"facile_id,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Task) TableName() string { return "tasks" }
