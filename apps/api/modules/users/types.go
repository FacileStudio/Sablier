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

// MeResponse wraps the requesting user's profile.
type MeResponse struct {
	User User `json:"user"`
}

// ListResponse wraps a user list.
type ListResponse struct {
	Users []User `json:"users"`
}

// UpdateRequest is the body for PATCH /users/me; any subset of fields applies.
type UpdateRequest struct {
	Name         *string  `json:"name"`
	Email        *string  `json:"email"`
	Password     *string  `json:"password"`
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
