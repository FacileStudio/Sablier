package schemas

import "time"

type Task struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	ProjectID int64     `gorm:"column:project_id;index;uniqueIndex:idx_tasks_project_name,priority:1"`
	Name      string    `gorm:"column:name;uniqueIndex:idx_tasks_project_name,priority:2"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Task) TableName() string { return "tasks" }
