package spaces

import (
	"context"
	stderrors "errors"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	porteSpaces "github.com/FacileStudio/porte/spaces"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Service implements space and membership persistence.
type Service struct {
	orm        *gorm.DB
	guard      porteSpaces.Guard
	controller *Controller
}

// NewService wires the spaces service and the membership guard over it.
func NewService(orm *gorm.DB) *Service {
	service := &Service{orm: orm, guard: porteSpaces.Guard{Store: NewStore(orm)}}
	service.controller = newController(service)
	return service
}

func (s *Service) createSpace(ctx context.Context, userID string, name, description string) (*schemas.Space, string, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, "", errors.Invalid("invalid user id")
	}

	space := &schemas.Space{
		Name:        name,
		Description: description,
	}
	err = s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(space).Error; err != nil {
			return err
		}
		member := &schemas.SpaceMember{
			SpaceID: space.ID,
			UserID:  uid,
			Role:    schemas.RoleOwner,
		}
		return tx.Create(member).Error
	})
	if err != nil {
		return nil, "", errors.Internal("failed to create space", err)
	}
	return space, schemas.RoleOwner, nil
}

func (s *Service) listSpaces(ctx context.Context, userID string) ([]spaceWithRole, error) {
	if _, err := strconv.ParseInt(userID, 10, 64); err != nil {
		return nil, errors.Invalid("invalid user id")
	}

	members, err := s.guard.Spaces(ctx, userID)
	if err != nil {
		return nil, guardError(err)
	}
	roles := make(map[string]string, len(members))
	ids := make([]string, 0, len(members))
	for _, member := range members {
		roles[member.SpaceID] = string(member.Role)
		ids = append(ids, member.SpaceID)
	}

	var rows []schemas.Space
	if len(ids) > 0 {
		err = s.orm.WithContext(ctx).
			Where("id IN ?", ids).
			Order("created_at desc").
			Find(&rows).Error
		if err != nil {
			return nil, errors.Internal("failed to list spaces", err)
		}
	}

	result := make([]spaceWithRole, len(rows))
	for i, row := range rows {
		result[i] = spaceWithRole{Space: row, Role: roles[row.ID]}
	}
	return result, nil
}

type spaceWithRole struct {
	schemas.Space
	Role string
}

func (s *Service) getSpace(ctx context.Context, spaceID string) (*schemas.Space, error) {
	var space schemas.Space
	err := s.orm.WithContext(ctx).Where("id = ?", spaceID).First(&space).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("space not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to get space", err)
	}
	return &space, nil
}

func (s *Service) updateSpace(ctx context.Context, spaceID string, name, description string) (*schemas.Space, error) {
	space, err := s.getSpace(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	space.Name = name
	space.Description = description
	if err := s.orm.WithContext(ctx).Save(space).Error; err != nil {
		return nil, errors.Internal("failed to update space", err)
	}
	return space, nil
}

func (s *Service) deleteSpace(ctx context.Context, spaceID string) error {
	return s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("space_id = ?", spaceID).Delete(&schemas.SpaceMember{}).Error; err != nil {
			return errors.Internal("failed to delete space members", err)
		}
		tx.Model(&schemas.Project{}).Where("space_id = ?", spaceID).Update("space_id", nil)
		tx.Model(&schemas.Task{}).Where("space_id = ?", spaceID).Update("space_id", nil)
		tx.Model(&schemas.TimeEntry{}).Where("space_id = ?", spaceID).Update("space_id", nil)
		result := tx.Where("id = ?", spaceID).Delete(&schemas.Space{})
		if result.Error != nil {
			return errors.Internal("failed to delete space", result.Error)
		}
		if result.RowsAffected == 0 {
			return errors.NotFound("space not found")
		}
		return nil
	})
}

func (s *Service) listMembers(ctx context.Context, spaceID string) ([]memberRow, error) {
	type row struct {
		schemas.SpaceMember
		UserEmail string
		UserName  string
	}
	var rows []row
	err := s.orm.WithContext(ctx).
		Model(&schemas.SpaceMember{}).
		Select("space_members.*, users.email as user_email, users.name as user_name").
		Joins("JOIN users ON users.id = space_members.user_id").
		Where("space_members.space_id = ?", spaceID).
		Order("space_members.joined_at asc").
		Find(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to list members", err)
	}
	result := make([]memberRow, len(rows))
	for i, r := range rows {
		result[i] = memberRow{SpaceMember: r.SpaceMember, UserEmail: r.UserEmail, UserName: r.UserName}
	}
	return result, nil
}

type memberRow struct {
	schemas.SpaceMember
	UserEmail string
	UserName  string
}

func (s *Service) addMember(ctx context.Context, spaceID string, targetUserID int64, role string) (*schemas.SpaceMember, error) {
	var existingUser schemas.User
	err := s.orm.WithContext(ctx).Where("id = ?", targetUserID).First(&existingUser).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("user not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to check user", err)
	}

	var existing schemas.SpaceMember
	err = s.orm.WithContext(ctx).Where("space_id = ? AND user_id = ?", spaceID, targetUserID).First(&existing).Error
	if err == nil {
		return nil, errors.Conflict("user is already a member")
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Internal("failed to check membership", err)
	}

	member := &schemas.SpaceMember{
		SpaceID: spaceID,
		UserID:  targetUserID,
		Role:    role,
	}
	if err := s.orm.WithContext(ctx).Create(member).Error; err != nil {
		return nil, errors.Internal("failed to add member", err)
	}
	return member, nil
}

func loadMember(tx *gorm.DB, spaceID, memberID string) (schemas.SpaceMember, error) {
	var member schemas.SpaceMember
	err := tx.Where("id = ? AND space_id = ?", memberID, spaceID).First(&member).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return member, errors.NotFound("member not found")
	}
	if err != nil {
		return member, errors.Internal("failed to get member", err)
	}
	return member, nil
}

func (s *Service) mayActOn(actor porteSpaces.Scope, member schemas.SpaceMember, denied string) error {
	if !s.guard.AssignableBy(actor, porteSpaces.Role(member.Role)) {
		return errors.Forbidden(denied)
	}
	return nil
}

func keepsAnOwner(ctx context.Context, tx *gorm.DB, spaceID string, member schemas.SpaceMember) error {
	guard := porteSpaces.Guard{Store: NewStore(tx)}
	err := guard.CanLeave(ctx, strconv.FormatInt(member.UserID, 10), spaceID)
	if stderrors.Is(err, porteSpaces.ErrSoleOwner) {
		return errors.New("failed_precondition",
			"cannot remove the last owner; promote another owner first", err)
	}
	if err != nil {
		return guardError(err)
	}
	return nil
}

func (s *Service) removeMember(
	ctx context.Context, actor porteSpaces.Scope, spaceID string, memberID string,
) error {
	return s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockMembers(tx, spaceID); err != nil {
			return errors.Internal("failed to check membership", err)
		}

		member, err := loadMember(tx, spaceID, memberID)
		if err != nil {
			return err
		}
		if err := s.mayActOn(actor, member, "cannot remove a member who outranks you"); err != nil {
			return err
		}
		if err := keepsAnOwner(ctx, tx, spaceID, member); err != nil {
			return err
		}

		if err := tx.Delete(&member).Error; err != nil {
			return errors.Internal("failed to remove member", err)
		}
		return nil
	})
}

func (s *Service) updateMemberRole(
	ctx context.Context, actor porteSpaces.Scope, spaceID string, memberID string, role string,
) (*schemas.SpaceMember, error) {
	var member schemas.SpaceMember
	err := s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockMembers(tx, spaceID); err != nil {
			return errors.Internal("failed to check membership", err)
		}

		loaded, err := loadMember(tx, spaceID, memberID)
		if err != nil {
			return err
		}
		if err := s.mayActOn(actor, loaded, "cannot change the role of a member who outranks you"); err != nil {
			return err
		}
		if loaded.Role != role {
			if err := keepsAnOwner(ctx, tx, spaceID, loaded); err != nil {
				return err
			}
		}

		loaded.Role = role
		if err := tx.Save(&loaded).Error; err != nil {
			return errors.Internal("failed to update member role", err)
		}
		member = loaded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// lockMembers takes the space's membership rows for update, so that the count
// CanLeave reads and the delete that follows it see the same set. Without it,
// two owners leaving at the same instant both count two owners, both pass, and
// the space ends with none. SQLite has no row locks and serializes writes
// anyway, so the clause is Postgres-only.
func lockMembers(tx *gorm.DB, spaceID string) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	var ids []string
	return tx.Model(&schemas.SpaceMember{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("space_id = ?", spaceID).
		Pluck("id", &ids).Error
}

func (s *Service) leaveSpace(ctx context.Context, spaceID string, userID string) error {
	return s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockMembers(tx, spaceID); err != nil {
			return errors.Internal("failed to check membership", err)
		}

		guard := porteSpaces.Guard{Store: NewStore(tx)}
		if err := guard.CanLeave(ctx, userID, spaceID); err != nil {
			return guardError(err)
		}

		uid, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			return errors.Invalid("invalid user id")
		}
		result := tx.Where("space_id = ? AND user_id = ?", spaceID, uid).Delete(&schemas.SpaceMember{})
		if result.Error != nil {
			return errors.Internal("failed to leave space", result.Error)
		}
		return nil
	})
}
