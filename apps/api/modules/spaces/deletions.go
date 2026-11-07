package spaces

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FacileStudio/porte/directory"

	"gorm.io/gorm"
)

// Deletions erases the people Registre erased. It polls Registre's deletion
// list with the same client credentials as the snapshot sync and accepts
// Registre's signed push on directory.DeletionsPath.
type Deletions struct {
	orm        *gorm.DB
	storageDir string
	receiver   *directory.Deletions
}

// NewDeletions builds the receiver over Sablier's tables. erase is porte's own
// rows eraser, pg.Store.EraseIdentity.
func NewDeletions(
	orm *gorm.DB, client directory.Client, erase func(context.Context, string, string) error,
	storageDir string, logger *slog.Logger,
) *Deletions {
	d := &Deletions{orm: orm, storageDir: storageDir}
	d.receiver = &directory.Deletions{
		Client:          client,
		OnPersonDeleted: d.ErasePerson,
		Erase:           erase,
		State:           deletionState{orm: orm},
		Logger:          logger,
	}
	return d
}

// Run polls until ctx ends. Start it in exactly one replica.
func (d *Deletions) Run(ctx context.Context) {
	d.receiver.Run(ctx)
}

// Handler verifies and applies Registre's signed user.deleted push.
func (d *Deletions) Handler() http.Handler {
	return d.receiver.Handler()
}

const (
	personalProjects = `SELECT id FROM projects WHERE space_id IS NULL AND owner_id IN @ids`

	deletePersonalEntries  = `DELETE FROM time_entries WHERE project_id IN (` + personalProjects + `)`
	deletePersonalTasks    = `DELETE FROM tasks WHERE project_id IN (` + personalProjects + `)`
	deletePersonalProjects = `DELETE FROM projects WHERE space_id IS NULL AND owner_id IN @ids`
	deleteOwnEntries       = `DELETE FROM time_entries WHERE user_id IN @ids`
	anonymiseTaskActor     = `UPDATE tasks SET actor_id = NULL WHERE actor_id IN @ids`
	anonymiseProjectOwner  = `UPDATE projects SET owner_id = 0 WHERE owner_id IN @ids`
	deleteHeartbeats       = `DELETE FROM heartbeats WHERE user_id IN @ids`
	deleteDismissals       = `DELETE FROM draft_dismissals WHERE user_id IN @ids`
	deletePushSubscription = `DELETE FROM push_subscriptions WHERE user_id IN @ids`
	deleteMemberships      = `DELETE FROM space_members WHERE user_id IN @ids`
	deleteAccounts         = `DELETE FROM users WHERE id IN @ids`
)

// ErasePerson deletes the account behind a Registre identity and what it owns:
// its personal projects with their tasks and time entries, every time entry it
// logged, its agent users, heartbeats, dismissed drafts, push subscription and
// memberships. A space project it owned stays with no owner, and a task it was
// assigned stays with no assignee. A subject that is already gone is a no-op,
// so a replayed notice changes nothing.
func (d *Deletions) ErasePerson(ctx context.Context, provider, subject string) error {
	var account struct {
		ID         int64
		UploadPath string
	}
	err := d.orm.WithContext(ctx).Raw(
		`SELECT u.id, coalesce(u.avatar_upload_path, '') AS upload_path FROM users u
		 JOIN porte_identities i ON i.user_id = u.id WHERE i.provider = ? AND i.subject = ?`,
		provider, subject).Scan(&account).Error
	if err != nil || account.ID == 0 {
		return err
	}
	err = d.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ids := []int64{account.ID}
		var agents []int64
		if err := tx.Raw(`SELECT id FROM users WHERE owner_user_id = ?`, account.ID).Scan(&agents).Error; err != nil {
			return err
		}
		args := map[string]any{"ids": append(ids, agents...)}
		for _, statement := range []string{
			deletePersonalEntries, deletePersonalTasks, deletePersonalProjects, deleteOwnEntries,
			anonymiseTaskActor, anonymiseProjectOwner, deleteHeartbeats, deleteDismissals,
			deletePushSubscription, deleteMemberships, deleteAccounts,
		} {
			if err := tx.Exec(statement, args).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil && account.UploadPath != "" {
		removeUpload(d.storageDir, account.UploadPath)
	}
	return err
}

func removeUpload(storageDir, relative string) {
	path := filepath.Join(storageDir, filepath.Clean("/"+relative))
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Warn("failed to remove an erased person's avatar", slog.Any("error", err))
	}
}

// deletionState keeps the receiver's cursor and handled subjects in one row.
type deletionState struct {
	orm *gorm.DB
}

// Load returns the saved state, or nil before the first save.
func (s deletionState) Load(ctx context.Context) ([]byte, error) {
	var rows []struct{ State []byte }
	err := s.orm.WithContext(ctx).Raw(`SELECT state FROM directory_deletion_state WHERE id = 1`).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0].State, nil
}

// Save replaces the saved state.
func (s deletionState) Save(ctx context.Context, state []byte) error {
	return s.orm.WithContext(ctx).Exec(
		`INSERT INTO directory_deletion_state (id, state, updated_at) VALUES (1, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT (id) DO UPDATE SET state = EXCLUDED.state, updated_at = CURRENT_TIMESTAMP`, state).Error
}
