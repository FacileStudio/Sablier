package schemas

import (
	"context"
	"strings"

	"api/internal/usercolor"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&User{}, &Session{}, &Project{}, &Task{}, &TimeEntry{}, &UserSetting{}); err != nil {
		return err
	}
	if err := usercolor.BackfillMissing(context.Background(), db); err != nil {
		return err
	}
	return backfillTimeEntryTasks(db)
}

func backfillTimeEntryTasks(db *gorm.DB) error {
	type legacyEntry struct {
		ID          int64  `gorm:"column:id"`
		ProjectID   int64  `gorm:"column:project_id"`
		TaskID      int64  `gorm:"column:task_id"`
		Description string `gorm:"column:description"`
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var entries []legacyEntry
		if err := tx.Table("time_entries").Where("task_id IS NULL OR task_id = 0").Order("id asc").Find(&entries).Error; err != nil {
			return err
		}

		taskIDs := map[int64]map[string]int64{}
		for _, entry := range entries {
			taskName := strings.TrimSpace(entry.Description)
			if taskName == "" {
				taskName = "Untitled task"
			}

			projectTasks, ok := taskIDs[entry.ProjectID]
			if !ok {
				projectTasks = map[string]int64{}
				taskIDs[entry.ProjectID] = projectTasks
			}

			taskID, ok := projectTasks[strings.ToLower(taskName)]
			if !ok {
				var task Task
				err := tx.Where("project_id = ? AND lower(name) = lower(?)", entry.ProjectID, taskName).First(&task).Error
				if err != nil {
					if err != gorm.ErrRecordNotFound {
						return err
					}
					task = Task{ProjectID: entry.ProjectID, Name: taskName}
					if err := tx.Create(&task).Error; err != nil {
						return err
					}
				}
				taskID = task.ID
				projectTasks[strings.ToLower(taskName)] = taskID
			}

			if err := tx.Table("time_entries").Where("id = ?", entry.ID).Update("task_id", taskID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
