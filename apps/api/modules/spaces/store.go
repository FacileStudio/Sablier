package spaces

import (
	"context"
	stderrors "errors"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	porteSpaces "github.com/FacileStudio/porte/spaces"

	"gorm.io/gorm"
)

// Store reads space_members for porte/spaces. It is the whole of the boundary
// between the package's opaque string ids and Sablier's own types: a space id
// is already a uuid string, and a user id is an int64 that crosses as its
// decimal form.
//
// A user id Sablier cannot parse is not a member of anything, so it is
// ErrNotMember rather than a malformed-input error: the guard's callers turn
// that into the same refusal a stranger gets, and no lookup runs on a value the
// column could never hold.
type Store struct {
	orm *gorm.DB
}

// NewStore wires the membership store over an ORM handle.
func NewStore(orm *gorm.DB) *Store { return &Store{orm: orm} }

func (s *Store) membership(row schemas.SpaceMember) porteSpaces.Membership {
	return porteSpaces.Membership{
		SpaceID: row.SpaceID,
		UserID:  strconv.FormatInt(row.UserID, 10),
		Role:    porteSpaces.Role(row.Role),
	}
}

// Membership returns the user's row in one space, or ErrNotMember when there
// is none. Every field of the result is read from the row, never echoed back
// from the arguments, so the guard's cross-check has something to check.
func (s *Store) Membership(ctx context.Context, spaceID, userID string) (porteSpaces.Membership, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return porteSpaces.Membership{}, porteSpaces.ErrNotMember
	}

	var row schemas.SpaceMember
	err = s.orm.WithContext(ctx).
		Where("space_id = ? AND user_id = ?", spaceID, uid).
		First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return porteSpaces.Membership{}, porteSpaces.ErrNotMember
	}
	if err != nil {
		return porteSpaces.Membership{}, err
	}
	return s.membership(row), nil
}

// Memberships returns every space the user belongs to. An empty result is not
// an error.
func (s *Store) Memberships(ctx context.Context, userID string) ([]porteSpaces.Membership, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, nil
	}

	var rows []schemas.SpaceMember
	err = s.orm.WithContext(ctx).
		Where("user_id = ?", uid).
		Order("joined_at asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]porteSpaces.Membership, len(rows))
	for i, row := range rows {
		out[i] = s.membership(row)
	}
	return out, nil
}

// CountRole returns how many members of the space hold exactly that role.
func (s *Store) CountRole(ctx context.Context, spaceID string, role porteSpaces.Role) (int, error) {
	var count int64
	err := s.orm.WithContext(ctx).
		Model(&schemas.SpaceMember{}).
		Where("space_id = ? AND role = ?", spaceID, string(role)).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
