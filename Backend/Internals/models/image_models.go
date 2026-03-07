package models

import (
	"time"

	"github.com/google/uuid"
)

// User maps directly to the `users` table.
type User struct {
	ID           uuid.UUID `db:"id"`
	Username     string    `db:"username"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	FullName     *string   `db:"full_name"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// ImageMetadata maps directly to the `image_metadata` table.
type ImageMetadata struct {
	ID            uuid.UUID  `db:"id"`
	SerialID      int64      `db:"serial_id"`
	UserID        *uuid.UUID `db:"user_id"`   // nullable FK
	Title         *string    `db:"title"`
	Description   *string    `db:"description"`
	MimeType      string     `db:"mime_type"`
	WidthPx       int        `db:"width_px"`
	HeightPx      int        `db:"height_px"`
	IsAIGenerated bool       `db:"is_ai_generated"`
	CapturedAt    *time.Time `db:"captured_at"` // nullable — not all images have EXIF
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

// ImageVector maps directly to the `image_vectors` table.
// The raw []float32 slice is what you pass to pgvector.NewVector().
type ImageVector struct {
	ID        uuid.UUID `db:"id"`
	ImageID   uuid.UUID `db:"image_id"`
	Vector    []float32 `db:"vector"`
	CreatedAt time.Time `db:"created_at"`
}