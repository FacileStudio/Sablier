package spaces

import (
	"context"
	stderrors "errors"
	"log/slog"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Registre keeps Sablier's spaces equal to the ones Registre owns. issuer is the configured
// OIDC_ISSUER, the provider column of every porte identity Registre's subjects live under.
type Registre struct {
	orm    *gorm.DB
	issuer string
	logger *slog.Logger
	retry  retryPolicy
}

// NewRegistre wires the directory sync over the database, for the given issuer.
func NewRegistre(orm *gorm.DB, issuer string, logger *slog.Logger) *Registre {
	return &Registre{orm: orm, issuer: issuer, logger: logger, retry: retryPolicy{first: retryFirst, max: retryMax}}
}

// Apply makes Sablier's linked spaces match the snapshot, in one transaction. It is
// idempotent: a second call with the same snapshot changes nothing.
//
// It never touches an unlinked space, nor a member with no identity under the issuer, which
// is what a local password account is. A linked space missing from the snapshot is left
// alone and logged, since Registre has no space deletion yet.
func (r *Registre) Apply(ctx context.Context, snapshot directory.Snapshot) error {
	return r.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seen := make([]string, 0, len(snapshot.Spaces))
		for _, remote := range snapshot.Spaces {
			if err := r.applySpace(ctx, tx, remote); err != nil {
				return err
			}
			seen = append(seen, remote.ID)
		}
		return r.reportOrphans(tx, seen)
	})
}

// applySpace links or creates the local row for one Registre space, then sets its members.
func (r *Registre) applySpace(ctx context.Context, tx *gorm.DB, remote directory.Space) error {
	spaceID, err := linkSpace(tx, remote)
	if err != nil {
		return err
	}
	kept := make([]int64, 0, len(remote.Members))
	for _, member := range remote.Members {
		userID, err := r.resolveMember(ctx, tx, member)
		if err != nil {
			return err
		}
		if userID == 0 {
			continue
		}
		if err := upsertMembership(tx, spaceID, userID, localRole(member.Role)); err != nil {
			return err
		}
		kept = append(kept, userID)
	}
	return r.removeAbsent(tx, spaceID, kept)
}

// linkSpace returns the local space for a Registre space: the one already linked to it, else
// the oldest unlinked one with the same name, case-insensitively, else a new row. The name is
// Registre's either way.
func linkSpace(tx *gorm.DB, remote directory.Space) (string, error) {
	var space schemas.Space
	err := tx.Where("registre_space_id = ?", remote.ID).First(&space).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		err = tx.Where("registre_space_id IS NULL AND lower(name) = lower(?)", remote.Name).
			Order("created_at asc").First(&space).Error
	}
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		space = schemas.Space{Name: remote.Name, RegistreSpaceID: &remote.ID}
		if err := tx.Create(&space).Error; err != nil {
			return "", err
		}
		return space.ID, nil
	}
	if err != nil {
		return "", err
	}
	if space.Name == remote.Name && space.RegistreSpaceID != nil {
		return space.ID, nil
	}
	return space.ID, tx.Model(&schemas.Space{}).Where("id = ?", space.ID).
		Updates(map[string]any{"name": remote.Name, "registre_space_id": remote.ID}).Error
}

// upsertMembership gives the user the role in the space, adding the row when there is none.
func upsertMembership(tx *gorm.DB, spaceID string, userID int64, role string) error {
	member := schemas.SpaceMember{SpaceID: spaceID, UserID: userID, Role: role}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "space_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role"}),
	}).Create(&member).Error
}

// removeAbsent drops the space's members who hold an identity under the issuer but are not
// in kept. Members without one are local accounts Registre does not know, and stay.
func (r *Registre) removeAbsent(tx *gorm.DB, spaceID string, kept []int64) error {
	query := tx.Where("space_id = ? AND user_id IN (?)", spaceID,
		tx.Table("porte_identities").Select("user_id").Where("provider = ?", r.issuer))
	if len(kept) > 0 {
		query = query.Where("user_id NOT IN ?", kept)
	}
	return query.Delete(&schemas.SpaceMember{}).Error
}

// reportOrphans logs the linked spaces the snapshot no longer carries.
func (r *Registre) reportOrphans(tx *gorm.DB, seen []string) error {
	query := tx.Model(&schemas.Space{}).Where("registre_space_id IS NOT NULL")
	if len(seen) > 0 {
		query = query.Where("registre_space_id NOT IN ?", seen)
	}
	var orphans []string
	if err := query.Pluck("id", &orphans).Error; err != nil {
		return err
	}
	if len(orphans) > 0 {
		r.logger.Warn("directory: linked spaces missing from the registre snapshot, left alone",
			slog.Any("space_ids", orphans))
	}
	return nil
}

// localRole maps Registre's rank onto Sablier's, which uses the same three words. Anything
// else is the ordinary member role, never a rank Registre did not grant.
func localRole(role string) string {
	switch role {
	case directory.RoleOwner:
		return schemas.RoleOwner
	case directory.RoleAdmin:
		return schemas.RoleAdmin
	}
	return schemas.RoleMember
}
