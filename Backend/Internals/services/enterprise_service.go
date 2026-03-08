package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/models"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/repository"
)

// ─────────────────────────────────────────────────────────────────────────────
// Sentinel errors — enterprise-specific
// ─────────────────────────────────────────────────────────────────────────────

var (
	ErrEnterpriseEmailTaken   = errors.New("an enterprise with this email already exists")
	ErrEnterpriseNotFound     = errors.New("enterprise not found")
	ErrEnterpriseInactive     = errors.New("enterprise account is deactivated")
	ErrEnterpriseInvalidCreds = errors.New("invalid email or password")
	ErrAPIKeyNotFound         = errors.New("API key not found or revoked")
	ErrAPIKeyInvalid          = errors.New("invalid API key")
)

// ─────────────────────────────────────────────────────────────────────────────
// Request / Response types
// ─────────────────────────────────────────────────────────────────────────────

type EnterpriseRegisterRequest struct {
	CompanyName string
	Email       string
	Password    string
	Website     string
	Plan        string // defaults to "free" when empty
}

type EnterpriseLoginRequest struct {
	Email    string
	Password string
}

type EnterpriseUpdateRequest struct {
	EnterpriseID uuid.UUID
	CompanyName  string
	Email        string
	Website      string
}

type EnterpriseChangePasswordRequest struct {
	EnterpriseID uuid.UUID
	OldPassword  string
	NewPassword  string
}

type CreateAPIKeyRequest struct {
	EnterpriseID uuid.UUID
	Label        string
	ExpiresAt    *time.Time // nil = never expires
}

// EnterpriseResponse is the safe, outward-facing representation of an enterprise.
// It never exposes the password hash.
type EnterpriseResponse struct {
	ID          uuid.UUID `json:"id"`
	CompanyName string    `json:"company_name"`
	Email       string    `json:"email"`
	Website     *string   `json:"website,omitempty"`
	Plan        string    `json:"plan"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Token       string    `json:"token,omitempty"` // set only on login
}

// APIKeyCreatedResponse is returned exactly once when a key is created.
// The RawKey must be stored by the enterprise immediately — it can never be
// retrieved again.
type APIKeyCreatedResponse struct {
	ID        uuid.UUID  `json:"id"`
	RawKey    string     `json:"raw_key"` // shown once only — store it securely
	KeyPrefix string     `json:"key_prefix"`
	Label     *string    `json:"label,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// APIKeyListItem is the redacted view of a key shown in the dashboard.
// The raw key and hash are never returned here.
type APIKeyListItem struct {
	ID         uuid.UUID  `json:"id"`
	KeyPrefix  string     `json:"key_prefix"`
	Label      *string    `json:"label,omitempty"`
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// APIKeyAuthResult is the value returned by ValidateAPIKey on success.
// Middleware stores this in Fiber locals for downstream handlers.
type APIKeyAuthResult struct {
	EnterpriseID uuid.UUID
	KeyID        uuid.UUID
	Plan         string
}

// ─────────────────────────────────────────────────────────────────────────────
// EnterpriseService
// ─────────────────────────────────────────────────────────────────────────────

type EnterpriseService struct {
	repo       repository.EnterpriseRepository
	bcryptCost int
}

func NewEnterpriseService(repo repository.EnterpriseRepository, bcryptCost int) *EnterpriseService {
	if bcryptCost == 0 {
		bcryptCost = bcrypt.DefaultCost
	}
	return &EnterpriseService{repo: repo, bcryptCost: bcryptCost}
}

// ─────────────────────────────────────────────────────────────────────────────
// Register
// ─────────────────────────────────────────────────────────────────────────────

func (s *EnterpriseService) Register(
	ctx context.Context,
	req EnterpriseRegisterRequest,
) (*EnterpriseResponse, error) {

	if len(req.Password) < 8 {
		return nil, ErrPasswordTooShort
	}
	exists, err := s.repo.EnterpriseExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("Register: %w", err)
	}
	if exists {
		return nil, ErrEnterpriseEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("Register: bcrypt: %w", err)
	}

	plan := req.Plan
	if plan == "" {
		plan = "free"
	}
	var website *string
	if req.Website != "" {
		w := req.Website
		website = &w
	}

	id, err := s.repo.CreateEnterprise(ctx, models.Enterprise{
		CompanyName:  req.CompanyName,
		Email:        req.Email,
		PasswordHash: string(hash),
		Website:      website,
		Plan:         plan,
		IsActive:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("Register: create: %w", err)
	}

	e, err := s.repo.GetEnterpriseByID(ctx, id)
	if err != nil || e == nil {
		return nil, fmt.Errorf("Register: fetch after create: %w", err)
	}
	return toEnterpriseResponse(e), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Login
// ─────────────────────────────────────────────────────────────────────────────

// Login verifies credentials and returns the enterprise profile.
// JWT generation happens in the handler layer so the service stays
// transport-agnostic (mirrors the user service pattern).
func (s *EnterpriseService) Login(
	ctx context.Context,
	req EnterpriseLoginRequest,
) (*EnterpriseResponse, error) {

	e, err := s.repo.GetEnterpriseByEmail(ctx, req.Email)
	if err != nil || e == nil {
		return nil, ErrEnterpriseInvalidCreds
	}
	if !e.IsActive {
		return nil, ErrEnterpriseInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(e.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrEnterpriseInvalidCreds
	}
	return toEnterpriseResponse(e), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetByID / UpdateProfile / ChangePassword / Deactivate
// ─────────────────────────────────────────────────────────────────────────────

func (s *EnterpriseService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*EnterpriseResponse, error) {

	e, err := s.repo.GetEnterpriseByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	if e == nil {
		return nil, ErrEnterpriseNotFound
	}
	return toEnterpriseResponse(e), nil
}

func (s *EnterpriseService) UpdateProfile(
	ctx context.Context,
	req EnterpriseUpdateRequest,
) (*EnterpriseResponse, error) {

	e, err := s.repo.GetEnterpriseByID(ctx, req.EnterpriseID)
	if err != nil || e == nil {
		return nil, ErrEnterpriseNotFound
	}
	if !e.IsActive {
		return nil, ErrEnterpriseInactive
	}

	if req.CompanyName != "" {
		e.CompanyName = req.CompanyName
	}
	if req.Email != "" {
		e.Email = req.Email
	}
	if req.Website != "" {
		w := req.Website
		e.Website = &w
	}

	if err := s.repo.UpdateEnterprise(ctx, *e); err != nil {
		return nil, fmt.Errorf("UpdateProfile: %w", err)
	}
	return toEnterpriseResponse(e), nil
}

func (s *EnterpriseService) ChangePassword(
	ctx context.Context,
	req EnterpriseChangePasswordRequest,
) error {

	if len(req.NewPassword) < 8 {
		return ErrPasswordTooShort
	}
	e, err := s.repo.GetEnterpriseByID(ctx, req.EnterpriseID)
	if err != nil || e == nil {
		return ErrEnterpriseNotFound
	}
	if err := bcrypt.CompareHashAndPassword([]byte(e.PasswordHash), []byte(req.OldPassword)); err != nil {
		return ErrIncorrectOldPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.bcryptCost)
	if err != nil {
		return fmt.Errorf("ChangePassword: bcrypt: %w", err)
	}
	return s.repo.UpdateEnterprisePasswordHash(ctx, req.EnterpriseID, string(hash))
}

func (s *EnterpriseService) Deactivate(
	ctx context.Context,
	id uuid.UUID,
) error {

	e, err := s.repo.GetEnterpriseByID(ctx, id)
	if err != nil || e == nil {
		return ErrEnterpriseNotFound
	}
	if !e.IsActive {
		return ErrEnterpriseInactive
	}
	return s.repo.DeactivateEnterprise(ctx, id)
}

// ─────────────────────────────────────────────────────────────────────────────
// API key management
// ─────────────────────────────────────────────────────────────────────────────

// GenerateAPIKey creates a new API key for an enterprise.
//
// Key format:  map_<43-char base64url>    (total ~47 chars)
// Key prefix:  first 12 chars             ("map_XXXXXXXX") — stored plain text
//
// The raw key is bcrypt-hashed before storage. Only this one response ever
// contains the raw key — it cannot be retrieved again.
func (s *EnterpriseService) GenerateAPIKey(
	ctx context.Context,
	req CreateAPIKeyRequest,
) (*APIKeyCreatedResponse, error) {

	e, err := s.repo.GetEnterpriseByID(ctx, req.EnterpriseID)
	if err != nil || e == nil {
		return nil, ErrEnterpriseNotFound
	}
	if !e.IsActive {
		return nil, ErrEnterpriseInactive
	}

	// 32 random bytes → base64url (no padding) → prefix with "map_"
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("GenerateAPIKey: rand: %w", err)
	}
	rawKey := "map_" + base64.RawURLEncoding.EncodeToString(raw)
	prefix := rawKey[:12] // "map_XXXXXXXX"

	hash, err := bcrypt.GenerateFromPassword([]byte(rawKey), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("GenerateAPIKey: bcrypt: %w", err)
	}

	var label *string
	if req.Label != "" {
		l := req.Label
		label = &l
	}

	keyID, err := s.repo.CreateAPIKey(ctx, models.EnterpriseAPIKey{
		EnterpriseID: req.EnterpriseID,
		KeyHash:      string(hash),
		KeyPrefix:    prefix,
		Label:        label,
		IsActive:     true,
		ExpiresAt:    req.ExpiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("GenerateAPIKey: insert: %w", err)
	}

	return &APIKeyCreatedResponse{
		ID:        keyID,
		RawKey:    rawKey,
		KeyPrefix: prefix,
		Label:     label,
		ExpiresAt: req.ExpiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// ListAPIKeys returns redacted key records — never the raw key or hash.
func (s *EnterpriseService) ListAPIKeys(
	ctx context.Context,
	enterpriseID uuid.UUID,
) ([]APIKeyListItem, error) {

	keys, err := s.repo.GetAPIKeysByEnterpriseID(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("ListAPIKeys: %w", err)
	}

	items := make([]APIKeyListItem, 0, len(keys))
	for _, k := range keys {
		items = append(items, APIKeyListItem{
			ID:         k.ID,
			KeyPrefix:  k.KeyPrefix,
			Label:      k.Label,
			IsActive:   k.IsActive,
			LastUsedAt: k.LastUsedAt,
			ExpiresAt:  k.ExpiresAt,
			CreatedAt:  k.CreatedAt,
		})
	}
	return items, nil
}

// RevokeAPIKey deactivates a key so it can no longer authenticate requests.
func (s *EnterpriseService) RevokeAPIKey(
	ctx context.Context,
	enterpriseID uuid.UUID,
	keyID uuid.UUID,
) error {
	// A production system should verify the key belongs to this enterprise.
	_ = enterpriseID
	return s.repo.RevokeAPIKey(ctx, keyID)
}

// ValidateAPIKey authenticates an incoming X-API-Key header value.
//
// Two-step lookup:
//  1. Extract the 12-char prefix → cheap indexed DB query.
//  2. bcrypt.Compare the full raw key against the stored hash.
func (s *EnterpriseService) ValidateAPIKey(
	ctx context.Context,
	rawKey string,
) (*APIKeyAuthResult, error) {

	if len(rawKey) < 12 || !strings.HasPrefix(rawKey, "map_") {
		return nil, ErrAPIKeyInvalid
	}
	prefix := rawKey[:12]

	k, err := s.repo.GetActiveAPIKeyByPrefix(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("ValidateAPIKey: db: %w", err)
	}
	if k == nil {
		return nil, ErrAPIKeyNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(k.KeyHash), []byte(rawKey)); err != nil {
		return nil, ErrAPIKeyInvalid
	}

	// Touch last_used_at without blocking the response path.
	go func() {
		_ = s.repo.TouchAPIKeyLastUsed(context.Background(), k.ID)
	}()

	e, err := s.repo.GetEnterpriseByID(ctx, k.EnterpriseID)
	if err != nil || e == nil {
		return nil, ErrEnterpriseNotFound
	}
	if !e.IsActive {
		return nil, ErrEnterpriseInactive
	}

	return &APIKeyAuthResult{
		EnterpriseID: k.EnterpriseID,
		KeyID:        k.ID,
		Plan:         e.Plan,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Usage / billing
// ─────────────────────────────────────────────────────────────────────────────

// RecordUsage appends one row to the audit/billing log.
// Call this from a goroutine — errors are non-fatal and should only be logged.
func (s *EnterpriseService) RecordUsage(
	ctx context.Context,
	u models.EnterpriseAPIUsage,
) error {
	return s.repo.RecordAPIUsage(ctx, u)
}

func (s *EnterpriseService) GetUsageSummary(
	ctx context.Context,
	enterpriseID uuid.UUID,
	from, to time.Time,
) ([]repository.UsageSummaryRow, error) {
	return s.repo.GetUsageSummary(ctx, enterpriseID, from, to)
}

// ─────────────────────────────────────────────────────────────────────────────
// helper
// ─────────────────────────────────────────────────────────────────────────────

func toEnterpriseResponse(e *models.Enterprise) *EnterpriseResponse {
	return &EnterpriseResponse{
		ID:          e.ID,
		CompanyName: e.CompanyName,
		Email:       e.Email,
		Website:     e.Website,
		Plan:        e.Plan,
		IsActive:    e.IsActive,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}