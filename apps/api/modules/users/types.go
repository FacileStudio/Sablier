package users

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

type MeResponse struct {
	User User `json:"user"`
}

type ListResponse struct {
	Users []User `json:"users"`
}

type UpdateRequest struct {
	Name         *string  `json:"name"`
	Email        *string  `json:"email"`
	Password     *string  `json:"password"`
	Color        *string  `json:"color"`
	Rate         *float64 `json:"rate"`
	RateType     *string  `json:"rate_type"`
	WorkdayHours *float64 `json:"workday_hours"`
}

type ApiTokenResponse struct {
	Token     string `json:"token"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type ApiTokenStatusResponse struct {
	HasToken  bool   `json:"has_token"`
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type CreateApiTokenRequest struct {
	Name string `json:"name"`
}
