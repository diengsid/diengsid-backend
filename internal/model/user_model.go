package model

import "id.diengs.backend/internal/entity"

type UserResponse struct {
	ID            string  `json:"id,omitempty"`
	Name          string  `json:"name,omitempty"`
	Email         string  `json:"email,omitempty"`
	PhoneNumber   string  `json:"phone_number,omitempty"`
	EmailVerified bool    `json:"email_verified,omitempty"`
	Picture       *string `json:"picture,omitempty"`
	Role          string  `json:"role,omitempty"`
	CreatedAt     int64   `json:"created_at,omitempty"`
	UpdatedAt     int64   `json:"updated_at,omitempty"`
}

func UserToResponse(user *entity.User) *UserResponse {
	return &UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		PhoneNumber:   user.PhoneNumber,
		Picture:       user.Picture,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
