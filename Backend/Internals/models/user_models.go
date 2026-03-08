package models

import (
	"time"

	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// Request / Response types
// ─────────────────────────────────────────────────────────────────────────────

// RegisterRequest is the input for creating a new account.
type RegisterRequest struct {
	Username string
	Email    string
	Password string // plain-text; hashed before storage
	FullName string
}

// LoginRequest is the input for authenticating an existing user.
type LoginRequest struct {
	Email    string
	Password string // plain-text; compared against stored hash
}

// UpdateProfileRequest carries mutable fields the user may change.
type UpdateProfileRequest struct {
	UserID   uuid.UUID
	Username string
	Email    string
	FullName string
}

// ChangePasswordRequest carries the old and new plain-text passwords.
type ChangePasswordRequest struct {
	UserID      uuid.UUID
	OldPassword string
	NewPassword string
}

// UserResponse is the safe, outward-facing user representation — it never
// carries the password hash.
type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	FullName  *string    `json:"full_name,omitempty"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Token     string  	 `json:"token,omitempty"`
}

