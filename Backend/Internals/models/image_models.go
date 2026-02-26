package models

import (
	"time"

	"github.com/google/uuid"
)

// ImageMetadata mirrors the image_metadata table.
// Pointer fields represent nullable columns.
// type ImageMetadata struct {
// 	ID             uuid.UUID
// 	SubmittedBy    *uuid.UUID // FK → users.id (nullable)
// 	Title          *string
// 	Description    *string
// 	SourceURL      *string
// 	ExternalRefID  *string
// 	ChecksumSHA256 *string
// 	MimeType       *string
// 	WidthPx        *int
// 	HeightPx       *int

// 	// AI analysis
// 	IsAIGenerated bool
// 	AIConfidence  *float64 // 0.0 – 1.0
// 	AIModelUsed   *string
// 	ContentFlags  []string

// 	// Location (all optional)
// 	LocationLabel *string
// 	Latitude      *float64
// 	Longitude     *float64

// 	// Classification
// 	Category *string
// 	Tags     []string

// 	// Qdrant sync status
// 	IsIndexed    bool
// 	IndexedAt    *time.Time
// 	IndexVersion *string

// 	CapturedAt *time.Time
// 	CreatedAt  time.Time
// 	UpdatedAt  time.Time
// }

type ImageMetadata struct {
	ID            uuid.UUID
	SerialID      int64 // ← new: the watermark-embedded key
	Title         *string
	Description   *string
	MimeType      *string
	WidthPx       *int
	HeightPx      *int
	IsAIGenerated bool
	CapturedAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
