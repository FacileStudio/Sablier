package spaces

import (
	"context"
	stderrors "errors"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	porteSpaces "github.com/FacileStudio/porte/spaces"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
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

func (s *Service) listMembers(ctx context.Context, spaceID string) ([]memberRow, error) {
	var rows []memberRow
	err := s.orm.WithContext(ctx).
		Model(&schemas.SpaceMember{}).
		Select("space_members.*, users.email as user_email, users.name as user_name, users.color as user_color, users.kind as user_kind, "+schemas.AvatarSelectExpr+" as user_avatar_url").
		Joins("JOIN users ON users.id = space_members.user_id").
		Where("space_members.space_id = ?", spaceID).
		Order("space_members.joined_at asc").
		Find(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to list members", err)
	}
	return rows, nil
}

type memberRow struct {
	schemas.SpaceMember
	UserEmail     string
	UserName      string
	UserColor     string
	UserAvatarURL string
	UserKind      string
}
