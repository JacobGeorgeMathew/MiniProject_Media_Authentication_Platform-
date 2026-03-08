package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// EnterpriseRepository interface
// Program against this interface — not the concrete *EnterpriseDB — in every
// service and test.
// ─────────────────────────────────────────────────────────────────────────────

type EnterpriseRepository interface {
	// Enterprise account CRUD
	CreateEnterprise(ctx context.Context, e models.Enterprise) (uuid.UUID, error)
	GetEnterpriseByID(ctx context.Context, id uuid.UUID) (*models.Enterprise, error)
	GetEnterpriseByEmail(ctx context.Context, email string) (*models.Enterprise, error)
	UpdateEnterprise(ctx context.Context, e models.Enterprise) error
	UpdateEnterprisePasswordHash(ctx context.Context, id uuid.UUID, newHash string) error
	DeactivateEnterprise(ctx context.Context, id uuid.UUID) error
	EnterpriseExists(ctx context.Context, email string) (bool, error)

	// API key management
	CreateAPIKey(ctx context.Context, k models.EnterpriseAPIKey) (uuid.UUID, error)
	GetAPIKeysByEnterpriseID(ctx context.Context, enterpriseID uuid.UUID) ([]models.EnterpriseAPIKey, error)
	GetActiveAPIKeyByPrefix(ctx context.Context, prefix string) (*models.EnterpriseAPIKey, error)
	RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error
	TouchAPIKeyLastUsed(ctx context.Context, keyID uuid.UUID) error

	// Usage / billing
	RecordAPIUsage(ctx context.Context, u models.EnterpriseAPIUsage) error
	GetUsageSummary(ctx context.Context, enterpriseID uuid.UUID, from, to time.Time) ([]UsageSummaryRow, error)
}

// UsageSummaryRow is one row in the aggregated usage report used by the
// billing dashboard.
type UsageSummaryRow struct {
	Endpoint   string `json:"endpoint"`
	Method     string `json:"method"`
	TotalCalls int64  `json:"total_calls"`
	Errors     int64  `json:"errors"` // calls where status_code >= 400
}

// ─────────────────────────────────────────────────────────────────────────────
// EnterpriseDB — concrete pgxpool-backed repository
// ─────────────────────────────────────────────────────────────────────────────

type EnterpriseDB struct {
	pool *pgxpool.Pool
}

func NewEnterpriseRepo(pool *pgxpool.Pool) *EnterpriseDB {
	return &EnterpriseDB{pool: pool}
}

// ─────────────────────────────────────────────────────────────────────────────
// Enterprise account CRUD
// ─────────────────────────────────────────────────────────────────────────────

func (db *EnterpriseDB) CreateEnterprise(
	ctx context.Context,
	e models.Enterprise,
) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.pool.QueryRow(ctx, `
		INSERT INTO enterprises
			(company_name, email, password_hash, website, plan, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`, e.CompanyName, e.Email, e.PasswordHash, e.Website, e.Plan, e.IsActive).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("CreateEnterprise: %w", err)
	}
	return id, nil
}

func (db *EnterpriseDB) GetEnterpriseByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Enterprise, error) {
	row := db.pool.QueryRow(ctx, `
		SELECT id, company_name, email, password_hash, website, plan, is_active, created_at, updated_at
		FROM enterprises WHERE id = $1
	`, id)
	e, err := scanEnterprise(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetEnterpriseByID: %w", err)
	}
	return e, nil
}

func (db *EnterpriseDB) GetEnterpriseByEmail(
	ctx context.Context,
	email string,
) (*models.Enterprise, error) {
	row := db.pool.QueryRow(ctx, `
		SELECT id, company_name, email, password_hash, website, plan, is_active, created_at, updated_at
		FROM enterprises WHERE email = $1
	`, email)
	e, err := scanEnterprise(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetEnterpriseByEmail: %w", err)
	}
	return e, nil
}

func (db *EnterpriseDB) UpdateEnterprise(
	ctx context.Context,
	e models.Enterprise,
) error {
	_, err := db.pool.Exec(ctx, `
		UPDATE enterprises
		SET company_name = $1, email = $2, website = $3, plan = $4, updated_at = NOW()
		WHERE id = $5
	`, e.CompanyName, e.Email, e.Website, e.Plan, e.ID)
	if err != nil {
		return fmt.Errorf("UpdateEnterprise: %w", err)
	}
	return nil
}

func (db *EnterpriseDB) UpdateEnterprisePasswordHash(
	ctx context.Context,
	id uuid.UUID,
	newHash string,
) error {
	_, err := db.pool.Exec(ctx, `
		UPDATE enterprises SET password_hash = $1, updated_at = NOW() WHERE id = $2
	`, newHash, id)
	if err != nil {
		return fmt.Errorf("UpdateEnterprisePasswordHash: %w", err)
	}
	return nil
}

func (db *EnterpriseDB) DeactivateEnterprise(
	ctx context.Context,
	id uuid.UUID,
) error {
	_, err := db.pool.Exec(ctx, `
		UPDATE enterprises SET is_active = FALSE, updated_at = NOW() WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("DeactivateEnterprise: %w", err)
	}
	return nil
}

func (db *EnterpriseDB) EnterpriseExists(
	ctx context.Context,
	email string,
) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM enterprises WHERE email = $1)
	`, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("EnterpriseExists: %w", err)
	}
	return exists, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// API key management
// ─────────────────────────────────────────────────────────────────────────────

func (db *EnterpriseDB) CreateAPIKey(
	ctx context.Context,
	k models.EnterpriseAPIKey,
) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.pool.QueryRow(ctx, `
		INSERT INTO enterprise_api_keys
			(enterprise_id, key_hash, key_prefix, label, is_active, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id
	`, k.EnterpriseID, k.KeyHash, k.KeyPrefix, k.Label, k.IsActive, k.ExpiresAt).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("CreateAPIKey: %w", err)
	}
	return id, nil
}

// GetAPIKeysByEnterpriseID returns all keys (active and revoked) for the
// enterprise dashboard.
func (db *EnterpriseDB) GetAPIKeysByEnterpriseID(
	ctx context.Context,
	enterpriseID uuid.UUID,
) ([]models.EnterpriseAPIKey, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, enterprise_id, key_hash, key_prefix, label, is_active,
		       last_used_at, expires_at, created_at
		FROM enterprise_api_keys
		WHERE enterprise_id = $1
		ORDER BY created_at DESC
	`, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("GetAPIKeysByEnterpriseID: %w", err)
	}
	defer rows.Close()

	var keys []models.EnterpriseAPIKey
	for rows.Next() {
		k, err := scanAPIKey(rows)
		if err != nil {
			return nil, fmt.Errorf("GetAPIKeysByEnterpriseID scan: %w", err)
		}
		keys = append(keys, *k)
	}
	return keys, rows.Err()
}

// GetActiveAPIKeyByPrefix returns the first active, non-expired key whose
// key_prefix matches the first 12 characters of the submitted raw key.
// The caller MUST still bcrypt.Compare the full raw key against key_hash.
func (db *EnterpriseDB) GetActiveAPIKeyByPrefix(
	ctx context.Context,
	prefix string,
) (*models.EnterpriseAPIKey, error) {
	row := db.pool.QueryRow(ctx, `
		SELECT id, enterprise_id, key_hash, key_prefix, label, is_active,
		       last_used_at, expires_at, created_at
		FROM enterprise_api_keys
		WHERE key_prefix = $1
		  AND is_active   = TRUE
		  AND (expires_at IS NULL OR expires_at > NOW())
		LIMIT 1
	`, prefix)
	k, err := scanAPIKey(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetActiveAPIKeyByPrefix(%q): %w", prefix, err)
	}
	return k, nil
}

func (db *EnterpriseDB) RevokeAPIKey(
	ctx context.Context,
	keyID uuid.UUID,
) error {
	_, err := db.pool.Exec(ctx, `
		UPDATE enterprise_api_keys SET is_active = FALSE WHERE id = $1
	`, keyID)
	if err != nil {
		return fmt.Errorf("RevokeAPIKey: %w", err)
	}
	return nil
}

// TouchAPIKeyLastUsed is called on every authenticated request.
// It runs asynchronously in the service layer so it never slows down responses.
func (db *EnterpriseDB) TouchAPIKeyLastUsed(
	ctx context.Context,
	keyID uuid.UUID,
) error {
	_, err := db.pool.Exec(ctx, `
		UPDATE enterprise_api_keys SET last_used_at = NOW() WHERE id = $1
	`, keyID)
	if err != nil {
		return fmt.Errorf("TouchAPIKeyLastUsed: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Usage / billing
// ─────────────────────────────────────────────────────────────────────────────

// RecordAPIUsage appends one row to the audit log.
// Call it as  go repo.RecordAPIUsage(...)  in the handler so it never blocks.
func (db *EnterpriseDB) RecordAPIUsage(
	ctx context.Context,
	u models.EnterpriseAPIUsage,
) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO enterprise_api_usage
			(enterprise_id, api_key_id, endpoint, method, status_code, latency_ms, request_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`, u.EnterpriseID, u.APIKeyID, u.Endpoint, u.Method, u.StatusCode, u.LatencyMs)
	if err != nil {
		return fmt.Errorf("RecordAPIUsage: %w", err)
	}
	return nil
}

// GetUsageSummary returns per-endpoint call counts grouped within the window
// [from, to].  Used by the billing dashboard.
func (db *EnterpriseDB) GetUsageSummary(
	ctx context.Context,
	enterpriseID uuid.UUID,
	from, to time.Time,
) ([]UsageSummaryRow, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT
			endpoint,
			method,
			COUNT(*)                                         AS total_calls,
			COUNT(*) FILTER (WHERE status_code >= 400)       AS errors
		FROM enterprise_api_usage
		WHERE enterprise_id = $1
		  AND request_at BETWEEN $2 AND $3
		GROUP BY endpoint, method
		ORDER BY total_calls DESC
	`, enterpriseID, from, to)
	if err != nil {
		return nil, fmt.Errorf("GetUsageSummary: %w", err)
	}
	defer rows.Close()

	var result []UsageSummaryRow
	for rows.Next() {
		var r UsageSummaryRow
		if err := rows.Scan(&r.Endpoint, &r.Method, &r.TotalCalls, &r.Errors); err != nil {
			return nil, fmt.Errorf("GetUsageSummary scan: %w", err)
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

type enterpriseScanner interface {
	Scan(dest ...any) error
}

func scanEnterprise(s enterpriseScanner) (*models.Enterprise, error) {
	var e models.Enterprise
	err := s.Scan(
		&e.ID, &e.CompanyName, &e.Email, &e.PasswordHash,
		&e.Website, &e.Plan, &e.IsActive,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func scanAPIKey(s enterpriseScanner) (*models.EnterpriseAPIKey, error) {
	var k models.EnterpriseAPIKey
	err := s.Scan(
		&k.ID, &k.EnterpriseID, &k.KeyHash, &k.KeyPrefix,
		&k.Label, &k.IsActive, &k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &k, nil
}
