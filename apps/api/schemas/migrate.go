package schemas

import (
	"context"
	"log/slog"
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/internal/usercolor"
	"github.com/FacileStudio/Sablier/apps/api/migrations"

	"gorm.io/gorm"
)

// Migrate applies the ordered migrations under migrations/ and then
// Reconcile. issuer is the configured OIDC_ISSUER, which the picture backfill
// derives Registre's avatar URL from.
func Migrate(ctx context.Context, db *gorm.DB, issuer string, logger *slog.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if err := migrations.Up(ctx, migrations.Config(sqlDB, logger)); err != nil {
		return err
	}
	return Reconcile(ctx, db, issuer)
}

// Reconcile runs the idempotent data steps that ran at every boot before the
// schema moved to migrations, in the same order. It changes rows, never the
// schema.
func Reconcile(ctx context.Context, db *gorm.DB, issuer string) error {
	steps := []func() error{
		func() error { return backfillOIDCPictures(db, issuer) },
		func() error { return backfillAvatarUploadPath(db) },
		func() error { return usercolor.BackfillMissing(ctx, db) },
		func() error { return backfillTimeEntryTasks(db) },
		func() error { return BackfillFacileIDs(db) },
		func() error { return backfillAgentEntries(db) },
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

// backfillOIDCPictures stores the picture URL Registre would hand over at the next
// sign-in, for every user holding an identity under the configured issuer.
//
// Registre names one stable URL per human, <issuer>/avatars/<sub>, whether or not a
// picture exists, and answers 404 when none does. A user who last signed in before that
// claim existed holds no URL, or an old per-upload one, and would show initials until
// signing in again. Deriving it here from the identity row fixes them on the next boot.
// It is idempotent: a row already holding the derived URL is not touched.
//
// The issuer is read from configuration and never written into a migration. A non-https
// issuer is skipped, for the same reason PhotoURL refuses a non-https claim.
func backfillOIDCPictures(db *gorm.DB, issuer string) error {
	if !strings.HasPrefix(strings.ToLower(issuer), "https://") {
		return nil
	}
	return db.Exec(
		`UPDATE users SET oidc_picture_url = rtrim(?, '/') || '/avatars/' || i.subject
		 FROM porte_identities i
		 WHERE i.user_id = users.id AND i.provider = ?
		   AND users.oidc_picture_url IS DISTINCT FROM rtrim(?, '/') || '/avatars/' || i.subject`,
		issuer, issuer, issuer).Error
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
//
// The empty string, not NULL, is written in the final pass: a NULL would fail to scan
// into the plain string the model declares.
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
	return db.Exec(`UPDATE users SET avatar_upload_path = '' WHERE avatar_upload_path IS NULL`).Error
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
		if err := tx.Table("time_entries").Where("(task_id IS NULL OR task_id = 0) AND coalesce(description, '') <> '' AND project_id IN (SELECT id FROM projects)").Order("id asc").Find(&entries).Error; err != nil {
			return err
		}

		taskIDs := map[int64]map[string]int64{}
		for _, entry := range entries {
			taskName := strings.TrimSpace(entry.Description)
			if taskName == "" {
				continue
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
