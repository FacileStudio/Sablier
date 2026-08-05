package spaces

import (
	"context"
	stderrors "errors"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/internal/errors"
	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/gorm"
)

type Service struct {
	orm        *gorm.DB
	controller *Controller
}

func NewService(orm *gorm.DB) *Service {
	service := &Service{orm: orm}
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
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}

	type row struct {
		schemas.Space
		Role string
	}
	var rows []row
	err = s.orm.WithContext(ctx).
		Model(&schemas.Space{}).
		Select("spaces.*, space_members.role as role").
		Joins("JOIN space_members ON space_members.space_id = spaces.id").
		Where("space_members.user_id = ?", uid).
		Order("spaces.created_at desc").
		Find(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to list spaces", err)
	}

	result := make([]spaceWithRole, len(rows))
	for i, r := range rows {
		result[i] = spaceWithRole{Space: r.Space, Role: r.Role}
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

func (s *Service) getMembership(ctx context.Context, spaceID string, userID string) (*schemas.SpaceMember, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}
	var member schemas.SpaceMember
	err = s.orm.WithContext(ctx).Where("space_id = ? AND user_id = ?", spaceID, uid).First(&member).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Forbidden("not a member of this space")
	}
	if err != nil {
		return nil, errors.Internal("failed to check membership", err)
	}
	return &member, nil
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

func (s *Service) removeMember(ctx context.Context, spaceID string, memberID string) error {
	result := s.orm.WithContext(ctx).Where("id = ? AND space_id = ?", memberID, spaceID).Delete(&schemas.SpaceMember{})
	if result.Error != nil {
		return errors.Internal("failed to remove member", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("member not found")
	}
	return nil
}

func (s *Service) updateMemberRole(ctx context.Context, spaceID string, memberID string, role string) (*schemas.SpaceMember, error) {
	var member schemas.SpaceMember
	err := s.orm.WithContext(ctx).Where("id = ? AND space_id = ?", memberID, spaceID).First(&member).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("member not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to get member", err)
	}
	member.Role = role
	if err := s.orm.WithContext(ctx).Save(&member).Error; err != nil {
		return nil, errors.Internal("failed to update member role", err)
	}
	return &member, nil
}

func (s *Service) leaveSpace(ctx context.Context, spaceID string, userID string) error {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return errors.Invalid("invalid user id")
	}

	var member schemas.SpaceMember
	err = s.orm.WithContext(ctx).Where("space_id = ? AND user_id = ?", spaceID, uid).First(&member).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NotFound("not a member of this space")
	}
	if err != nil {
		return errors.Internal("failed to check membership", err)
	}

	if member.Role == schemas.RoleOwner {
		var ownerCount int64
		s.orm.WithContext(ctx).Model(&schemas.SpaceMember{}).Where("space_id = ? AND role = ?", spaceID, schemas.RoleOwner).Count(&ownerCount)
		if ownerCount <= 1 {
			return errors.Failed("cannot leave space as the only owner; transfer ownership or delete the space")
		}
	}

	result := s.orm.WithContext(ctx).Where("id = ?", member.ID).Delete(&schemas.SpaceMember{})
	if result.Error != nil {
		return errors.Internal("failed to leave space", result.Error)
	}
	return nil
}
