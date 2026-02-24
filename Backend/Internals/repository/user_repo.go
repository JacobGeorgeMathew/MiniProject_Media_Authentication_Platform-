package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// User helpers
// ---------------------------------------------------------------------------

// CreateUser inserts a new user. passwordHash must be a bcrypt hash.
func (db *DB) CreateUser(ctx context.Context, username, email, passwordHash, fullName string) (uuid.UUID, error) {
	const q = `
		INSERT INTO users (username, email, password_hash, full_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	var id uuid.UUID
	err := db.pool.QueryRow(ctx, q, username, email, passwordHash, fullName).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create user: %w", err)
	}
	return id, nil
}

// GetUserPasswordHash returns the stored hash for a given email.
// Use this during login — compare the returned hash with bcrypt.CompareHashAndPassword.
func (db *DB) GetUserPasswordHash(ctx context.Context, email string) (uuid.UUID, string, error) {
	const q = `SELECT id, password_hash FROM users WHERE email = $1 AND is_active = TRUE`
	var id uuid.UUID
	var hash string
	if err := db.pool.QueryRow(ctx, q, email).Scan(&id, &hash); err != nil {
		return uuid.Nil, "", fmt.Errorf("get user by email: %w", err)
	}
	return id, hash, nil
}

// RecordUserLogin stamps last_login_at for the given user.
func (db *DB) RecordUserLogin(ctx context.Context, userID uuid.UUID) error {
	_, err := db.pool.Exec(ctx,
		`UPDATE users SET last_login_at = NOW() WHERE id = $1`, userID)
	return err
}

// ---------------------------------------------------------------------------
// Admin helpers
// ---------------------------------------------------------------------------

// CreateAdmin inserts a new admin. role must be 'admin' or 'superadmin'.
func (db *DB) CreateAdmin(ctx context.Context, username, email, passwordHash, fullName, role string) (uuid.UUID, error) {
	const q = `
		INSERT INTO admins (username, email, password_hash, full_name, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	var id uuid.UUID
	err := db.pool.QueryRow(ctx, q, username, email, passwordHash, fullName, role).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create admin: %w", err)
	}
	return id, nil
}

// GetAdminPasswordHash returns the stored hash and role for a given admin email.
func (db *DB) GetAdminPasswordHash(ctx context.Context, email string) (uuid.UUID, string, string, error) {
	const q = `SELECT id, password_hash, role FROM admins WHERE email = $1 AND is_active = TRUE`
	var id uuid.UUID
	var hash, role string
	if err := db.pool.QueryRow(ctx, q, email).Scan(&id, &hash, &role); err != nil {
		return uuid.Nil, "", "", fmt.Errorf("get admin by email: %w", err)
	}
	return id, hash, role, nil
}
