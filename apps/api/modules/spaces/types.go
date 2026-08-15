package spaces

import "time"

// CreateSpaceRequest is the body for creating a space.
type CreateSpaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateSpaceRequest is the body for updating a space.
type UpdateSpaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AddMemberRequest is the body for adding a member.
type AddMemberRequest struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

// UpdateMemberRoleRequest is the body for changing a member's role.
type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}

// SpaceResponse is the serialized shape of a space, including the caller's role.
type SpaceResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListSpacesResponse wraps a space list.
type ListSpacesResponse struct {
	Spaces []SpaceResponse `json:"spaces"`
}

// MemberResponse is the serialized shape of a space member.
type MemberResponse struct {
	ID        string    `json:"id"`
	SpaceID   string    `json:"space_id"`
	UserID    int64     `json:"user_id"`
	UserEmail string    `json:"user_email"`
	UserName  string    `json:"user_name"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// ListMembersResponse wraps a member list.
type ListMembersResponse struct {
	Members []MemberResponse `json:"members"`
}
