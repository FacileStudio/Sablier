package schemas

import (
	"context"
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/internal/usercolor"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := renameNookPoolColumns(db); err != nil {
		return err
	}
	if err := renamePoolTables(db); err != nil {
		return err
	}
	if err := db.AutoMigrate(&User{}, &Session{}, &Project{}, &Task{}, &TimeEntry{}, &AppSetting{}, &ApiToken{}, &PushSubscription{}, &Space{}, &SpaceMember{}, &AntenneOutbox{}, &AntenneProcessedEvent{}); err != nil {
		return err
	}
	if err := backfillAvatarUploadPath(db); err != nil {
		return err
	}
	if err := usercolor.BackfillMissing(context.Background(), db); err != nil {
		return err
	}
	return backfillTimeEntryTasks(db)
}

// backfillAvatarUploadPath moves the uploaded avatars onto the column that now owns them.
//
// The filename decides, not avatar_source. That column was added after the upload feature,
// so the oldest uploaded avatars have it empty — on the production database two of the four
// rows were exactly that, and keying on avatar_source = 'upload' would have quietly dropped
// their picture. persistAvatarFile has always named uploads "user-<id>-<nanos>" and the old
// OIDC download named its copies "oidc-<id>-<nanos>", so anything that is not an oidc- copy
// is somebody's upload and is kept.
//
// The oidc- copies are the ones with nothing to preserve: oidc_picture_url already holds
// the URL that replaces them. They are left on the volume rather than deleted here — a
// migration that removes files has to be right the first time, and they are a few hundred
// kilobytes that a later sweep can take once this has proven itself.
//
// avatar_url and avatar_source stay in the table, unread, until the next release drops
// them. Expanding first means a rollback is redeploying the old binary, not restoring a
// backup.
func backfillAvatarUploadPath(db *gorm.DB) error {
	if !db.Migrator().HasColumn(&User{}, "avatar_url") {
		return nil
	}
	if err := db.Exec(
		`UPDATE users SET avatar_upload_path = replace(avatar_url, '/files/', '')
		 WHERE coalesce(avatar_url, '') <> ''
		   AND avatar_url NOT LIKE '/files/avatars/oidc-%'
		   AND coalesce(avatar_upload_path, '') = ''`).Error; err != nil {
		return err
	}
	// A NULL here would fail to scan into the plain string the model declares.
	return db.Exec(`UPDATE users SET avatar_upload_path = '' WHERE avatar_upload_path IS NULL`).Error
}

func renamePoolTables(db *gorm.DB) error {
	migrator := db.Migrator()

	renames := [][2]string{
		{"pool_outbox", "antenne_outbox"},
		{"pool_processed_events", "antenne_processed_events"},
	}

	for _, rename := range renames {
		from, to := rename[0], rename[1]
		if !migrator.HasTable(from) || migrator.HasTable(to) {
			continue
		}
		if err := migrator.RenameTable(from, to); err != nil {
			return err
		}
	}

	return nil
}

func renameNookPoolColumns(db *gorm.DB) error {
	migrator := db.Migrator()
	if !migrator.HasTable(&AppSetting{}) {
		return nil
	}

	renames := [][2]string{
		{"nook_pool_url", "antenne_url"},
		{"nook_pool_secret", "antenne_secret"},
		{"nook_pool_enabled", "antenne_enabled"},
	}

	for _, rename := range renames {
		from, to := rename[0], rename[1]
		if !migrator.HasColumn(&AppSetting{}, from) || migrator.HasColumn(&AppSetting{}, to) {
			continue
		}
		if err := migrator.RenameColumn(&AppSetting{}, from, to); err != nil {
			return err
		}
	}

	return nil
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
