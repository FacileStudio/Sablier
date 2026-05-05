package users

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Color     string `json:"color"`
}

type MeResponse struct {
	User User `json:"user"`
}

type ListResponse struct {
	Users []User `json:"users"`
}

type UpdateRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Color    *string `json:"color"`
}
