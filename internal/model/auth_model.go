package model

type AuthSendOtpReq struct {
	Email string `json:"email" validate:"required"`
}

type AuthVerifyOtpRequest struct {
	Email string `json:"email" validate:"required"`
	Otp   string `json:"otp" validate:"required"`
}

type AuthGoogleRequest struct {
	IP        string `json:"-"`
	UserAgent string `json:"-"`
	Token     string `json:"token" validate:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type VerifyUserRequest struct {
	Token string `validate:"required"`
}
type AuthResponse struct {
	User  UserResponse `json:"user,omitempty"`
	Token string       `json:"token,omitempty"`
}
