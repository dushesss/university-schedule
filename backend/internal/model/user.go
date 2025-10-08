package model

import "time"

// User модель пользователя
type User struct {
	ID           uint32     `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Не показываем в JSON
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// UserRole роли пользователей
type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleTeacher   UserRole = "teacher"
	RoleStudent   UserRole = "student"
	RoleModerator UserRole = "moderator"
)

// LoginRequest запрос на вход
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=teacher student"`
}

// AuthResponse ответ с токеном
type AuthResponse struct {
	User  *User  `json:"user"`
	Token string `json:"token"`
}

// JWTClaims данные в JWT токене
type JWTClaims struct {
	UserID uint32   `json:"user_id"`
	Email  string   `json:"email"`
	Role   UserRole `json:"role"`
	Exp    int64    `json:"exp"`
}
