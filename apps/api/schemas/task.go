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

func IsValidStatus(s string) bool {
	return validStatuses[s]
}

func NormalizeStatus(s string) string {
	if validStatuses[s] {
		return s
	}
	return StatusTodo
}

func ValidateStatus(s string) error {
	if validStatuses[s] {
		return nil
	}
	return fmt.Errorf("invalid status %q, valid: %s", s, strings.Join(AllStatuses(), ", "))
}

func AllStatuses() []string {
	return []string{StatusTodo, StatusInProgress, StatusInReview, StatusDone}
}

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
