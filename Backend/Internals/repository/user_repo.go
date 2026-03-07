package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// UserRepo wraps DB and exposes user-related repository methods.
// It intentionally re-uses the same *DB wrapper so callers only manage one
// pool.
// ─────────────────────────────────────────────────────────────────────────────

// CreateUser inserts a new user row and returns the generated UUID.
// password_hash must already be bcrypt/argon2-hashed by the caller — this
// layer never touches raw passwords.
func (db *DB) CreateUser(
	ctx context.Context,
	u models.User,
) (uuid.UUID, error) {
	const q = `
		INSERT INTO users (
			username, email, password_hash, full_name,
			is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id`

	var id uuid.UUID
	err := db.pool.QueryRow(ctx, q,
		u.Username,
		u.Email,
		u.PasswordHash,
		u.FullName,
		u.IsActive,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("CreateUser: %w", err)
	}
	return id, nil
}

// GetUserByID fetches a user by primary key UUID.
func (db *DB) GetUserByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.User, error) {
	const q = `
		SELECT id, username, email, password_hash, full_name,
		       is_active, created_at, updated_at
		FROM users
		WHERE id = $1`

	row := db.pool.QueryRow(ctx, q, id)
	u, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("GetUserByID(%s): %w", id, err)
	}
	return u, nil
}

// GetUserByEmail fetches a user by email address.
// Useful for login / existence checks before password verification.
func (db *DB) GetUserByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {
	const q = `
		SELECT id, username, email, password_hash, full_name,
		       is_active, created_at, updated_at
		FROM users
		WHERE email = $1`

	row := db.pool.QueryRow(ctx, q, email)
	u, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("GetUserByEmail(%s): %w", email, err)
	}
	return u, nil
}

// GetUserByUsername fetches a user by username.
func (db *DB) GetUserByUsername(
	ctx context.Context,
	username string,
) (*models.User, error) {
	const q = `
		SELECT id, username, email, password_hash, full_name,
		       is_active, created_at, updated_at
		FROM users
		WHERE username = $1`

	row := db.pool.QueryRow(ctx, q, username)
	u, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("GetUserByUsername(%s): %w", username, err)
	}
	return u, nil
}

// UpdateUser applies a full overwrite of mutable user fields.
// Increment updated_at server-side to avoid clock-skew issues.
func (db *DB) UpdateUser(
	ctx context.Context,
	u models.User,
) error {
	const q = `
		UPDATE users SET
			username      = $1,
			email         = $2,
			full_name     = $3,
			is_active     = $4,
			updated_at    = NOW()
		WHERE id = $5`

	tag, err := db.pool.Exec(ctx, q,
		u.Username,
		u.Email,
		u.FullName,
		u.IsActive,
		u.ID,
	)
	if err != nil {
		return fmt.Errorf("UpdateUser(%s): %w", u.ID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateUser(%s): no row found", u.ID)
	}
	return nil
}

// UpdatePasswordHash replaces a user's stored password hash.
// Only call this after the caller has already validated the old password
// and produced a new hash.
func (db *DB) UpdatePasswordHash(
	ctx context.Context,
	id uuid.UUID,
	newHash string,
) error {
	const q = `
		UPDATE users SET
			password_hash = $1,
			updated_at    = NOW()
		WHERE id = $2`

	tag, err := db.pool.Exec(ctx, q, newHash, id)
	if err != nil {
		return fmt.Errorf("UpdatePasswordHash(%s): %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("UpdatePasswordHash(%s): no row found", id)
	}
	return nil
}

// DeactivateUser sets is_active = false without deleting the row, preserving
// referential integrity with image_metadata.
func (db *DB) DeactivateUser(
	ctx context.Context,
	id uuid.UUID,
) error {
	const q = `
		UPDATE users SET
			is_active  = FALSE,
			updated_at = NOW()
		WHERE id = $1`

	tag, err := db.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("DeactivateUser(%s): %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("DeactivateUser(%s): no row found", id)
	}
	return nil
}

// UserExists returns true when a user with the given email already exists.
// Cheaper than a full SELECT when you only need a boolean check (e.g. during
// registration).
func (db *DB) UserExists(
	ctx context.Context,
	email string,
) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	if err := db.pool.QueryRow(ctx, q, email).Scan(&exists); err != nil {
		return false, fmt.Errorf("UserExists: %w", err)
	}
	return exists, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// helper
// ─────────────────────────────────────────────────────────────────────────────

func scanUser(s scanner) (*models.User, error) {
	var u models.User
	err := s.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}