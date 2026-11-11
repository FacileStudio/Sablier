package users

// User is the serialized shape of an account.
type User struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	Name         string  `json:"name"`
	AvatarURL    string  `json:"avatar_url"`
	AvatarSource string  `json:"avatar_source"`
	Color        string  `json:"color"`
	Locale       string  `json:"locale"`
	Rate         float64 `json:"rate"`
	RateType     string  `json:"rate_type"`
	WorkdayHours float64 `json:"workday_hours"`
	CreatedAt    string  `json:"created_at"`
	Kind         string  `json:"kind"`
	OwnerUserID  *int64  `json:"owner_user_id,omitempty"`
	IsAdmin      bool    `json:"is_admin,omitempty"`
}

// Peer is another account as the members list, a member's page and the cost
// totals read it. It carries no locale and no avatar origin, which only the
// account's own profile needs.
type Peer struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	Name         string  `json:"name"`
	AvatarURL    string  `json:"avatar_url"`
	Color        string  `json:"color"`
	Rate         float64 `json:"rate"`
	RateType     string  `json:"rate_type"`
	WorkdayHours float64 `json:"workday_hours"`
	CreatedAt    string  `json:"created_at"`
	Kind         string  `json:"kind"`
	OwnerUserID  *int64  `json:"owner_user_id,omitempty"`
}

// PeerResponse wraps one visible account.
type PeerResponse struct {
	User Peer `json:"user"`
}

// MeResponse wraps the requesting user's profile.
type MeResponse struct {
	User User `json:"user"`
}

// ListResponse wraps the accounts visible to the caller.
type ListResponse struct {
	Users []Peer `json:"users"`
}

// UpdateRequest is the body for PATCH /users/me; any subset of fields applies.
type UpdateRequest struct {
	Name         *string  `json:"name"`
	Color        *string  `json:"color"`
	Rate         *float64 `json:"rate"`
	RateType     *string  `json:"rate_type"`
	WorkdayHours *float64 `json:"workday_hours"`
}

// ApiTokenResponse is the shape returned when a token is minted, carrying the
// raw token exactly once.
type ApiTokenResponse struct {
	Token     string `json:"token"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// ApiTokenStatusResponse reports whether a named token exists and its label.
type ApiTokenStatusResponse struct {
	HasToken  bool   `json:"has_token"`
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// CreateApiTokenRequest is the body for minting a named API token.
type CreateApiTokenRequest struct {
	Name string `json:"name"`
}

// DeleteTokenResponse reports whether an API token was deleted.
type DeleteTokenResponse struct {
	Deleted bool `json:"deleted"`
}
