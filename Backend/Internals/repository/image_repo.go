package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/models"
)

const (
	SIMILARITY_CUTOFF = 0.95
	TOP_K             = 10
)

// ─────────────────────────────────────────────────────────────────────────────
// ImageRepository interface
// ─────────────────────────────────────────────────────────────────────────────

type ImageRepository interface {
	InsertImageMetadata(ctx context.Context, m models.ImageMetadata) (uuid.UUID, int64, error)
	InsertImageWithVector(ctx context.Context, userID uuid.UUID, title, description, mimeType string, w, h int, isAIGenerated bool, vec []float64) (uuid.UUID, int64, error)
	UpdateImageVector(ctx context.Context, imageID uuid.UUID, vec []float64) error
	GetImageMetadataBySerialID(ctx context.Context, serialID int64) (*models.ImageMetadata, error)
	GetImageMetadata(ctx context.Context, id uuid.UUID) (*models.ImageMetadata, error)
	GetImageMetadataBatch(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*models.ImageMetadata, error) // ← fixed signature
	FindSimilarImages(ctx context.Context, vec []float64) ([]SearchResult, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// SearchResult — returned by FindSimilarImages
// ─────────────────────────────────────────────────────────────────────────────

type SearchResult struct {
	Metadata   models.ImageMetadata
	Similarity float64
}

// ─────────────────────────────────────────────────────────────────────────────
// DB — concrete pgxpool-backed repository
// ─────────────────────────────────────────────────────────────────────────────

type DB struct {
	pool *pgxpool.Pool
}

func NewImageRepo(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}

// ─────────────────────────────────────────────────────────────────────────────
// InsertImageMetadata
// ─────────────────────────────────────────────────────────────────────────────

func (db *DB) InsertImageMetadata(
	ctx context.Context,
	m models.ImageMetadata,
) (uuid.UUID, int64, error) {
	const q = `
		INSERT INTO image_metadata (
			user_id, title, description, mime_type, width_px, height_px,
			is_ai_generated, captured_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, serial_id`

	var id uuid.UUID
	var serialID int64

	err := db.pool.QueryRow(ctx, q,
		m.UserID,
		m.Title,
		m.Description,
		m.MimeType,
		m.WidthPx,
		m.HeightPx,
		m.IsAIGenerated,
		m.CapturedAt,
	).Scan(&id, &serialID)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("InsertImageMetadata: %w", err)
	}
	return id, serialID, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// InsertImageWithVector
// ─────────────────────────────────────────────────────────────────────────────

// InsertImageWithVector inserts image_metadata and image_vectors atomically.
func (db *DB) InsertImageWithVector(
	ctx context.Context,
	userID uuid.UUID,
	title, description, mimeType string,
	w, h int,
	isAIGenerated bool,
	vec []float64,
) (uuid.UUID, int64, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("InsertImageWithVector begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var imageID uuid.UUID
	var serialID int64

	err = tx.QueryRow(ctx, `
		INSERT INTO image_metadata (
			user_id, title, description, mime_type, width_px, height_px,
			is_ai_generated, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, serial_id
	`, userID, title, description, mimeType, w, h, isAIGenerated).Scan(&imageID, &serialID)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("metadata insert: %w", err)
	}

	pgVec := pgvector.NewVector(float32Slice(vec))
	_, err = tx.Exec(ctx, `
		INSERT INTO image_vectors (image_id, vector, created_at)
		VALUES ($1, $2, NOW())
	`, imageID, pgVec)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("vector insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, 0, fmt.Errorf("InsertImageWithVector commit: %w", err)
	}
	return imageID, serialID, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetImageMetadataBySerialID
// ─────────────────────────────────────────────────────────────────────────────

func (db *DB) GetImageMetadataBySerialID(
	ctx context.Context,
	serialID int64,
) (*models.ImageMetadata, error) {
	const q = `
		SELECT id, serial_id, user_id, title, description, mime_type,
		       width_px, height_px, is_ai_generated, captured_at, created_at, updated_at
		FROM image_metadata
		WHERE serial_id = $1`

	m, err := scanImageMetadata(db.pool.QueryRow(ctx, q, serialID))
	if err != nil {
		return nil, fmt.Errorf("GetImageMetadataBySerialID(%d): %w", serialID, err)
	}
	return m, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetImageMetadata
// ─────────────────────────────────────────────────────────────────────────────

func (db *DB) GetImageMetadata(
	ctx context.Context,
	id uuid.UUID,
) (*models.ImageMetadata, error) {
	const q = `
		SELECT id, serial_id, user_id, title, description, mime_type,
		       width_px, height_px, is_ai_generated, captured_at, created_at, updated_at
		FROM image_metadata
		WHERE id = $1`

	m, err := scanImageMetadata(db.pool.QueryRow(ctx, q, id))
	if err != nil {
		return nil, fmt.Errorf("GetImageMetadata(%s): %w", id, err)
	}
	return m, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetImageMetadataBatch  (fixed: []uuid.UUID, not []models.ImageMetadata)
// ─────────────────────────────────────────────────────────────────────────────

func (db *DB) GetImageMetadataBatch(
	ctx context.Context,
	ids []uuid.UUID,
) (map[uuid.UUID]*models.ImageMetadata, error) {
	if len(ids) == 0 {
		return make(map[uuid.UUID]*models.ImageMetadata), nil
	}

	const q = `
		SELECT id, serial_id, user_id, title, description, mime_type,
		       width_px, height_px, is_ai_generated, captured_at, created_at, updated_at
		FROM image_metadata
		WHERE id = ANY($1)`

	rows, err := db.pool.Query(ctx, q, ids)
	if err != nil {
		return nil, fmt.Errorf("GetImageMetadataBatch query: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*models.ImageMetadata, len(ids))
	for rows.Next() {
		m, err := scanImageMetadata(rows)
		if err != nil {
			return nil, fmt.Errorf("GetImageMetadataBatch scan: %w", err)
		}
		result[m.ID] = m
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdateImageVector
// ─────────────────────────────────────────────────────────────────────────────

// UpdateImageVector replaces the stored vector for an existing image row.
// Called after watermark embedding to persist the real fingerprint instead of
// the placeholder zero vector written during the initial transaction.
func (db *DB) UpdateImageVector(
	ctx context.Context,
	imageID uuid.UUID,
	vec []float64,
) error {
	pgVec := pgvector.NewVector(float32Slice(vec))
	_, err := db.pool.Exec(ctx, `
		UPDATE image_vectors SET vector = $1 WHERE image_id = $2
	`, pgVec, imageID)
	if err != nil {
		return fmt.Errorf("UpdateImageVector(%s): %w", imageID, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// FindSimilarImages — pgvector cosine ANN search (no Qdrant)
// ─────────────────────────────────────────────────────────────────────────────

func (db *DB) FindSimilarImages(
	ctx context.Context,
	vec []float64,
) ([]SearchResult, error) {
	pgVec := pgvector.NewVector(float32Slice(vec))
	distanceCutoff := 1.0 - SIMILARITY_CUTOFF

	rows, err := db.pool.Query(ctx, `
		SELECT
			m.id, m.user_id, m.serial_id, m.title, m.description,
			m.mime_type, m.width_px, m.height_px, m.is_ai_generated,
			m.captured_at, m.created_at, m.updated_at,
			1 - (v.vector <=> $1::vector) AS similarity
		FROM image_vectors v
		JOIN image_metadata m ON m.id = v.image_id
		WHERE (v.vector <=> $1::vector) <= $2
		ORDER BY v.vector <=> $1::vector
		LIMIT $3
	`, pgVec, distanceCutoff, TOP_K)
	if err != nil {
		return nil, fmt.Errorf("FindSimilarImages query: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r models.ImageMetadata
		var sim float64
		err := rows.Scan(
			&r.ID, &r.UserID, &r.SerialID, &r.Title, &r.Description,
			&r.MimeType, &r.WidthPx, &r.HeightPx, &r.IsAIGenerated,
			&r.CapturedAt, &r.CreatedAt, &r.UpdatedAt,
			&sim,
		)
		if err != nil {
			return nil, fmt.Errorf("FindSimilarImages scan: %w", err)
		}
		results = append(results, SearchResult{Metadata: r, Similarity: sim})
	}
	return results, rows.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanImageMetadata(s scanner) (*models.ImageMetadata, error) {
	var m models.ImageMetadata
	err := s.Scan(
		&m.ID, &m.SerialID, &m.UserID, &m.Title, &m.Description,
		&m.MimeType, &m.WidthPx, &m.HeightPx, &m.IsAIGenerated,
		&m.CapturedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func float32Slice(in []float64) []float32 {
	out := make([]float32, len(in))
	for i, v := range in {
		out[i] = float32(v)
	}
	return out
}
