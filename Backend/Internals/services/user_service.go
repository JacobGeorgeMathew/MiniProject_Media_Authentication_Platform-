package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// Interfaces
// ─────────────────────────────────────────────────────────────────────────────

// UserRepository covers the PostgreSQL operations used by UserService.
type UserRepository interface {
	CreateUser(ctx context.Context, u models.User) (uuid.UUID, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUser(ctx context.Context, u models.User) error
	UpdatePasswordHash(ctx context.Context, id uuid.UUID, newHash string) error
	DeactivateUser(ctx context.Context, id uuid.UUID) error
	UserExists(ctx context.Context, email string) (bool, error)
}

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
}

// ─────────────────────────────────────────────────────────────────────────────
// Sentinel errors
// ─────────────────────────────────────────────────────────────────────────────

var (
	ErrEmailAlreadyExists   = errors.New("an account with this email already exists")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrUserNotFound         = errors.New("user not found")
	ErrUserInactive         = errors.New("account is deactivated")
	ErrPasswordTooShort     = errors.New("password must be at least 8 characters")
	ErrIncorrectOldPassword = errors.New("old password is incorrect")
)

// ─────────────────────────────────────────────────────────────────────────────
// UserService
// ─────────────────────────────────────────────────────────────────────────────

// UserService handles user registration, authentication, and profile
// management.  It is transport-agnostic and has no knowledge of HTTP/gRPC.
type UserService struct {
	repo        UserRepository
	bcryptCost  int
}

// NewUserService constructs a UserService.
// bcryptCost should be bcrypt.DefaultCost (12) in production; lower it in
// tests for speed.
func NewUserService(repo UserRepository, bcryptCost int) *UserService {
	if bcryptCost == 0 {
		bcryptCost = bcrypt.DefaultCost
	}
	return &UserService{repo: repo, bcryptCost: bcryptCost}
}

// ─────────────────────────────────────────────────────────────────────────────
// Register
// ─────────────────────────────────────────────────────────────────────────────

// Register creates a new user account.  It checks for email uniqueness,
// validates the password, hashes it, and persists the record.
func (s *UserService) Register(
	ctx context.Context,
	req RegisterRequest,
) (*UserResponse, error) {

	// ── Validate password length ─────────────────────────────────────────────
	if len(req.Password) < 8 {
		return nil, ErrPasswordTooShort
	}

	// ── Duplicate email check ────────────────────────────────────────────────
	exists, err := s.repo.UserExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("Register: existence check: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// ── Hash password ────────────────────────────────────────────────────────
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("Register: bcrypt: %w", err)
	}

	// ── Build model ──────────────────────────────────────────────────────────
	var fullName *string
	if req.FullName != "" {
		fn := req.FullName
		fullName = &fn
	}

	u := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		FullName:     fullName,
		IsActive:     true,
	}

	id, err := s.repo.CreateUser(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("Register: create user: %w", err)
	}

	// ── Return safe representation ───────────────────────────────────────────
	return &UserResponse{
		ID:       id,
		Username: u.Username,
		Email:    u.Email,
		FullName: u.FullName,
		IsActive: u.IsActive,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Login
// ─────────────────────────────────────────────────────────────────────────────

// Login verifies credentials and returns the user on success.
// The caller is responsible for issuing a session token / JWT.
func (s *UserService) Login(
	ctx context.Context,
	req LoginRequest,
) (*UserResponse, error) {

	u, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Deliberately vague to prevent email enumeration.
		return nil, ErrInvalidCredentials
	}
	if u == nil {
		return nil, ErrInvalidCredentials
	}

	// ── Guard: deactivated account ───────────────────────────────────────────
	if !u.IsActive {
		return nil, ErrUserInactive
	}

	// ── Verify password ──────────────────────────────────────────────────────
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return toUserResponse(u), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetByID
// ─────────────────────────────────────────────────────────────────────────────

// GetByID fetches a user's public profile by UUID.
func (s *UserService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*UserResponse, error) {

	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return toUserResponse(u), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdateProfile
// ─────────────────────────────────────────────────────────────────────────────

// UpdateProfile applies mutable field changes (username, email, full_name).
// It does NOT touch the password hash.
func (s *UserService) UpdateProfile(
	ctx context.Context,
	req UpdateProfileRequest,
) (*UserResponse, error) {

	u, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("UpdateProfile: fetch: %w", err)
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	if !u.IsActive {
		return nil, ErrUserInactive
	}

	// Apply changes.
	if req.Username != "" {
		u.Username = req.Username
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.FullName != "" {
		fn := req.FullName
		u.FullName = &fn
	}

	if err := s.repo.UpdateUser(ctx, *u); err != nil {
		return nil, fmt.Errorf("UpdateProfile: save: %w", err)
	}

	return toUserResponse(u), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ChangePassword
// ─────────────────────────────────────────────────────────────────────────────

// ChangePassword verifies the old password and replaces the hash with a new
// one derived from newPassword.
func (s *UserService) ChangePassword(
	ctx context.Context,
	req ChangePasswordRequest,
) error {

	if len(req.NewPassword) < 8 {
		return ErrPasswordTooShort
	}

	u, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("ChangePassword: fetch: %w", err)
	}
	if u == nil {
		return ErrUserNotFound
	}

	// Verify current password before allowing a change.
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.OldPassword)); err != nil {
		return ErrIncorrectOldPassword
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.bcryptCost)
	if err != nil {
		return fmt.Errorf("ChangePassword: bcrypt: %w", err)
	}

	if err := s.repo.UpdatePasswordHash(ctx, req.UserID, string(newHash)); err != nil {
		return fmt.Errorf("ChangePassword: save: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// DeactivateAccount
// ─────────────────────────────────────────────────────────────────────────────

// DeactivateAccount soft-deletes a user by setting is_active = false.
// All image_metadata rows remain intact for provenance purposes.
func (s *UserService) DeactivateAccount(
	ctx context.Context,
	id uuid.UUID,
) error {

	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("DeactivateAccount: fetch: %w", err)
	}
	if u == nil {
		return ErrUserNotFound
	}
	if !u.IsActive {
		return ErrUserInactive
	}

	if err := s.repo.DeactivateUser(ctx, id); err != nil {
		return fmt.Errorf("DeactivateAccount: save: %w", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// helper
// ─────────────────────────────────────────────────────────────────────────────

func toUserResponse(u *models.User) *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FullName:  u.FullName,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}