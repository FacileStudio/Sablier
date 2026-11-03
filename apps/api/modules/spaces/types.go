package spaces

import "time"

// SpaceResponse is the serialized shape of a space, including the caller's role.
// RegistreSpaceID is empty until the space is linked to Registre, which is where the
// client sends anyone who wants to change it.
type SpaceResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Role            string    `json:"role"`
	RegistreSpaceID string    `json:"registre_space_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ListSpacesResponse wraps a space list.
type ListSpacesResponse struct {
	Spaces []SpaceResponse `json:"spaces"`
}

// MemberResponse is the serialized shape of a space member.
type MemberResponse struct {
	ID            string    `json:"id"`
	SpaceID       string    `json:"space_id"`
	UserID        int64     `json:"user_id"`
	UserEmail     string    `json:"user_email"`
	UserName      string    `json:"user_name"`
	UserColor     string    `json:"user_color"`
	UserAvatarURL string    `json:"user_avatar_url"`
	UserKind      string    `json:"user_kind"`
	Role          string    `json:"role"`
	JoinedAt      time.Time `json:"joined_at"`
}

// ListMembersResponse wraps a member list.
type ListMembersResponse struct {
	Members []MemberResponse `json:"members"`
}
