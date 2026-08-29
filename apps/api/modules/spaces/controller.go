package spaces

import (
	"context"
	stderrors "errors"
	"strings"

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
	case stderrors.Is(err, porteSpaces.ErrSoleOwner):
		return errors.New("failed_precondition",
			"cannot leave space as the only owner; transfer ownership or delete the space", err)
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

func validRole(role string) bool {
	return role == schemas.RoleOwner || role == schemas.RoleAdmin || role == schemas.RoleMember
}

func toSpaceResponse(s *schemas.Space, role string) SpaceResponse {
	return SpaceResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Role:        role,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func toMemberResponse(m *memberRow) MemberResponse {
	return MemberResponse{
		ID:        m.ID,
		SpaceID:   m.SpaceID,
		UserID:    m.UserID,
		UserEmail: m.UserEmail,
		UserName:  m.UserName,
		Role:      m.Role,
		JoinedAt:  m.JoinedAt,
	}
}

func (c *Controller) create(ctx context.Context, userID string, req *CreateSpaceRequest) (*SpaceResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.Invalid("space name is required")
	}
	space, role, err := c.service.createSpace(ctx, userID, name, strings.TrimSpace(req.Description))
	if err != nil {
		return nil, err
	}
	resp := toSpaceResponse(space, role)
	return &resp, nil
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

func (c *Controller) update(ctx context.Context, spaceID string, userID string, req *UpdateSpaceRequest) (*SpaceResponse, error) {
	scope, err := c.require(ctx, spaceID, userID, porteSpaces.RoleAdmin, "only owners and admins can update a space")
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.Invalid("space name is required")
	}
	space, err := c.service.updateSpace(ctx, spaceID, name, strings.TrimSpace(req.Description))
	if err != nil {
		return nil, err
	}
	resp := toSpaceResponse(space, string(scope.Role))
	return &resp, nil
}

func (c *Controller) delete(ctx context.Context, spaceID string, userID string) error {
	if _, err := c.require(ctx, spaceID, userID, porteSpaces.RoleOwner, "only owners can delete a space"); err != nil {
		return err
	}
	return c.service.deleteSpace(ctx, spaceID)
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

func (c *Controller) addMember(ctx context.Context, spaceID string, userID string, req *AddMemberRequest) (*MemberResponse, error) {
	scope, err := c.require(ctx, spaceID, userID, porteSpaces.RoleAdmin, "only owners and admins can add members")
	if err != nil {
		return nil, err
	}
	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = schemas.RoleMember
	}
	if !validRole(role) {
		return nil, errors.Invalid("role must be owner, admin, or member")
	}
	if !c.service.guard.AssignableBy(scope, porteSpaces.Role(role)) {
		return nil, errors.Forbidden("only owners can assign the owner role")
	}
	member, err := c.service.addMember(ctx, spaceID, req.UserID, role)
	if err != nil {
		return nil, err
	}
	resp := MemberResponse{
		ID:      member.ID,
		SpaceID: member.SpaceID,
		UserID:  member.UserID,
		Role:    member.Role,
	}
	return &resp, nil
}

func (c *Controller) removeMember(ctx context.Context, spaceID string, userID string, memberID string) error {
	scope, err := c.require(ctx, spaceID, userID, porteSpaces.RoleAdmin, "only owners and admins can remove members")
	if err != nil {
		return err
	}
	return c.service.removeMember(ctx, scope, spaceID, memberID)
}

func (c *Controller) updateMemberRole(ctx context.Context, spaceID string, userID string, memberID string, req *UpdateMemberRoleRequest) (*MemberResponse, error) {
	scope, err := c.require(ctx, spaceID, userID, porteSpaces.RoleOwner, "only owners can change member roles")
	if err != nil {
		return nil, err
	}
	role := strings.TrimSpace(req.Role)
	if !validRole(role) {
		return nil, errors.Invalid("role must be owner, admin, or member")
	}
	member, err := c.service.updateMemberRole(ctx, scope, spaceID, memberID, role)
	if err != nil {
		return nil, err
	}
	resp := MemberResponse{
		ID:      member.ID,
		SpaceID: member.SpaceID,
		UserID:  member.UserID,
		Role:    member.Role,
	}
	return &resp, nil
}

func (c *Controller) leave(ctx context.Context, spaceID string, userID string) error {
	return c.service.leaveSpace(ctx, spaceID, userID)
}
