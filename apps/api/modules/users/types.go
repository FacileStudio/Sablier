package users

// User is the serialized shape of an account.
type User struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	Name         string  `json:"name"`
	AvatarURL    string  `json:"avatar_url"`
	AvatarSource string  `json:"avatar_source"`
	Color        string  `json:"color"`
	Rate         float64 `json:"rate"`
	RateType     string  `json:"rate_type"`
	WorkdayHours float64 `json:"workday_hours"`
	CreatedAt    string  `json:"created_at"`
}

// MeResponse wraps the requesting user's profile. Token carries a replacement
// session token and is present only when the call rotated one, which changing
// a password does: porte writes the new cookie itself, and this is the same
// rotation for the clients holding a bearer token instead.
type MeResponse struct {
	User  User   `json:"user"`
	Token string `json:"token,omitempty"`
}

// ListResponse wraps a user list.
type ListResponse struct {
	Users []User `json:"users"`
}

// UpdateRequest is the body for PATCH /users/me; any subset of fields applies.
//
// CurrentPassword accompanies Password when the account already has one.
// Leaving it out is only valid for an account that has none — an SSO user
// adding a first password — and anything else is refused rather than applied,
// so a borrowed session cannot become a permanent takeover.
type UpdateRequest struct {
	Name            *string  `json:"name"`
	Email           *string  `json:"email"`
	Password        *string  `json:"password"`
	CurrentPassword *string  `json:"current_password"`
	Color           *string  `json:"color"`
	Rate            *float64 `json:"rate"`
	RateType        *string  `json:"rate_type"`
	WorkdayHours    *float64 `json:"workday_hours"`
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
