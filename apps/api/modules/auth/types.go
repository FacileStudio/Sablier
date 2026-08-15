package auth

// RegisterRequest is the body for a local registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the body for a local login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is the {user_id, token} shape returned by register and login.
type AuthResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

// Data carries an account's email address where no fuller profile is needed.
type Data struct {
	Email string `json:"email"`
}

func (d *Data) GetEmail() string {
	if d == nil {
		return ""
	}
	return d.Email
}
