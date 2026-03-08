package models

import (
	"time"

	"github.com/google/uuid"
)

// Enterprise is the database model for a registered commercial company.
type Enterprise struct {
	ID           uuid.UUID `db:"id"`
	CompanyName  string    `db:"company_name"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Website      *string   `db:"website"`
	Plan         string    `db:"plan"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// EnterpriseAPIKey is the database model for one API key owned by an enterprise.
// The raw key is never stored — only its bcrypt hash.
type EnterpriseAPIKey struct {
	ID           uuid.UUID  `db:"id"`
	EnterpriseID uuid.UUID  `db:"enterprise_id"`
	KeyHash      string     `db:"key_hash"`   // bcrypt hash of the raw key
	KeyPrefix    string     `db:"key_prefix"` // first 12 chars — shown in dashboard
	Label        *string    `db:"label"`
	IsActive     bool       `db:"is_active"`
	LastUsedAt   *time.Time `db:"last_used_at"`
	ExpiresAt    *time.Time `db:"expires_at"`
	CreatedAt    time.Time  `db:"created_at"`
}

// EnterpriseAPIUsage is one row in the per-request audit / billing log.
type EnterpriseAPIUsage struct {
	ID           int64      `db:"id"`
	EnterpriseID uuid.UUID  `db:"enterprise_id"`
	APIKeyID     *uuid.UUID `db:"api_key_id"`
	Endpoint     string     `db:"endpoint"`
	Method       string     `db:"method"`
	StatusCode   int        `db:"status_code"`
	LatencyMs    *int       `db:"latency_ms"`
	RequestAt    time.Time  `db:"request_at"`
}
