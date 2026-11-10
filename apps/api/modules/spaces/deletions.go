package spaces

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/FacileStudio/porte/directory"

	"gorm.io/gorm"
)

// Deletions erases the people Registre erased. It polls Registre's deletion
// list with the same client credentials as the snapshot sync and accepts
// Registre's signed push on directory.DeletionsPath.
type Deletions struct {
	orm          *gorm.DB
	storageDir   string
	eraseAccount func(context.Context, string, string) error
	receiver     *directory.Deletions
}

// NewDeletions builds the receiver over Sablier's tables. eraseAccount is the
// account store's EraseAccount, which deletes the person's rows and account
// in one transaction; erase is porte's own rows eraser, pg.Store.EraseIdentity.
func NewDeletions(
	orm *gorm.DB, client directory.Client, eraseAccount, erase func(context.Context, string, string) error,
	storageDir string, logger *slog.Logger,
) *Deletions {
	d := &Deletions{orm: orm, storageDir: storageDir, eraseAccount: eraseAccount}
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

// ErasePerson deletes the account behind a Registre identity and what it owns,
// through the account store, then the legacy avatar file it uploaded. A
// subject that is already gone is a no-op, so a replayed notice changes
// nothing.
func (d *Deletions) ErasePerson(ctx context.Context, provider, subject string) error {
	var uploads []string
	err := d.orm.WithContext(ctx).Raw(
		`SELECT coalesce(u.avatar_upload_path, '') FROM users u
		 JOIN porte_identities i ON i.user_id = u.id
		 WHERE i.provider IN (?, ?) AND i.subject = ?`,
		provider, toggleSlash(provider), subject).Scan(&uploads).Error
	if err != nil {
		return err
	}
	if err := d.eraseAccount(ctx, provider, subject); err != nil {
		return err
	}
	for _, upload := range uploads {
		if upload != "" {
			removeUpload(d.storageDir, upload)
		}
	}
	return nil
}

func toggleSlash(provider string) string {
	if trimmed, ok := strings.CutSuffix(provider, "/"); ok {
		return trimmed
	}
	return provider + "/"
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
