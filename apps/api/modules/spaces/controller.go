package spaces

import (
	"context"
	stderrors "errors"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	porteSpaces "github.com/FacileStudio/porte/spaces"
	"github.com/FacileStudio/tronc/errors"
)

// Controller serves the space and membership routes.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func guardError(err error) error {
	switch {
	case stderrors.Is(err, porteSpaces.ErrNotMember):
		return errors.New("permission_denied", "not a member of this space", err)
	case stderrors.Is(err, porteSpaces.ErrForbidden):
		return errors.New("permission_denied", "insufficient role in this space", err)
	case stderrors.Is(err, porteSpaces.ErrUnknownRole):
		return errors.Internal("unknown space role", err)
	}
	return errors.Internal("failed to check membership", err)
}

func (c *Controller) require(
	ctx context.Context, spaceID, userID string, min porteSpaces.Role, denied string,
) (porteSpaces.Scope, error) {
	scope, err := c.service.guard.Require(ctx, userID, spaceID, min)
	if err == nil {
		return scope, nil
	}
	if stderrors.Is(err, porteSpaces.ErrForbidden) {
		return porteSpaces.Scope{}, errors.New("permission_denied", denied, err)
	}
	return porteSpaces.Scope{}, guardError(err)
}

func toSpaceResponse(s *schemas.Space, role string) SpaceResponse {
	response := SpaceResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Role:        role,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
	if s.RegistreSpaceID != nil {
		response.RegistreSpaceID = *s.RegistreSpaceID
	}
	return response
}

func toMemberResponse(m *memberRow) MemberResponse {
	return MemberResponse{
		ID:            m.ID,
		SpaceID:       m.SpaceID,
		UserID:        m.UserID,
		UserEmail:     m.UserEmail,
		UserName:      m.UserName,
		UserColor:     m.UserColor,
		UserAvatarURL: m.UserAvatarURL,
		UserKind:      m.UserKind,
		Role:          m.Role,
		JoinedAt:      m.JoinedAt,
	}
}

func (c *Controller) list(ctx context.Context, userID string) (*ListSpacesResponse, error) {
	rows, err := c.service.listSpaces(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]SpaceResponse, len(rows))
	for i, r := range rows {
		items[i] = toSpaceResponse(&r.Space, r.Role)
	}
	return &ListSpacesResponse{Spaces: items}, nil
}

func (c *Controller) get(ctx context.Context, spaceID string, userID string) (*SpaceResponse, error) {
	scope, err := c.require(ctx, spaceID, userID, porteSpaces.RoleMember, "not a member of this space")
	if err != nil {
		return nil, err
	}
	space, err := c.service.getSpace(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	resp := toSpaceResponse(space, string(scope.Role))
	return &resp, nil
}

func (c *Controller) listMembers(ctx context.Context, spaceID string, userID string) (*ListMembersResponse, error) {
	_, err := c.require(ctx, spaceID, userID, porteSpaces.RoleMember, "not a member of this space")
	if err != nil {
		return nil, err
	}
	rows, err := c.service.listMembers(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	items := make([]MemberResponse, len(rows))
	for i := range rows {
		items[i] = toMemberResponse(&rows[i])
	}
	return &ListMembersResponse{Members: items}, nil
}
