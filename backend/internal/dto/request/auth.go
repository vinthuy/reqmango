package request

// RegisterRequest is the request body for user registration.
type RegisterRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Username     string `json:"username" binding:"required,min=3,max=128"`
	Password     string `json:"password" binding:"required,min=8"`
	DisplayName  string `json:"display_name"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	UserTimezone string `json:"user_timezone"`
}

// LoginRequest is the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
