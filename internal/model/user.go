// Package model defines the data structures and domain types used across the application.
package model

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID        int64     `json:"id" gorm:"column:id;primaryKey"`
	Email     string    `json:"email" gorm:"column:email;uniqueIndex"`
	Password  string    `json:"-" gorm:"column:password"`
	Name      string    `json:"name" gorm:"column:name"`
	Role      Role      `json:"role" gorm:"column:role"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName overrides the default table name for User.
func (User) TableName() string {
	return "users"
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Name     string `json:"name" binding:"required,min=1,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=1"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateProfileRequest struct {
	Name  string `json:"name" binding:"omitempty,min=1,max=100"`
	Email string `json:"email" binding:"omitempty,email"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
