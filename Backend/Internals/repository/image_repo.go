// package repository

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"

// 	//"time"
// 	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/models"

// 	"github.com/google/uuid"
// 	//"github.com/jackc/pgx/v5/pgxpool"
// )

// // ---------------------------------------------------------------------------
// // Connection pool
// // ---------------------------------------------------------------------------

// // DB wraps the standard sql.DB (already connected via your database.go)
// type DB struct {
//     pool *sql.DB
// }

// // NewDB wraps an already-connected *sql.DB from your existing database.go.
// // You pass in the result of your Connect() function directly.
// func NewDB(sqlDB *sql.DB) *DB {
//     return &DB{pool: sqlDB}
// }

// // Close shuts down the connection pool.
// func (db *DB) Close() { db.pool.Close() }

// // ---------------------------------------------------------------------------
// // Domain types
// // ---------------------------------------------------------------------------

// // ---------------------------------------------------------------------------
// // Image Metadata — Write operations
// // ---------------------------------------------------------------------------

// // InsertImageMetadata inserts a new metadata record and returns its generated UUID.
// // is_indexed defaults to FALSE — call MarkAsIndexed after storing in Qdrant.
// func (db *DB) InsertImageMetadata(ctx context.Context, m models.ImageMetadata) (uuid.UUID, error) {
// 	const q = `
// 		INSERT INTO image_metadata (
// 			submitted_by, title, description, source_url, external_ref_id,
// 			checksum_sha256, mime_type, width_px, height_px,
// 			is_ai_generated, ai_confidence, ai_model_used, content_flags,
// 			location_label, latitude, longitude,
// 			category, tags, captured_at
// 		) VALUES (
// 			$1,$2,$3,$4,$5,
// 			$6,$7,$8,$9,
// 			$10,$11,$12,$13,
// 			$14,$15,$16,
// 			$17,$18,$19
// 		)
// 		RETURNING id`

// 	var id uuid.UUID
// 	err := db.pool.QueryRowContext(ctx, q,          // ← QueryRowContext, not QueryRow
// 		m.SubmittedBy, m.Title, m.Description, m.SourceURL, m.ExternalRefID,
// 		m.ChecksumSHA256, m.MimeType, m.WidthPx, m.HeightPx,
// 		m.IsAIGenerated, m.AIConfidence, m.AIModelUsed, m.ContentFlags,
// 		m.LocationLabel, m.Latitude, m.Longitude,
// 		m.Category, m.Tags, m.CapturedAt,
// 	).Scan(&id)
// 	if err != nil {
// 		return uuid.Nil, fmt.Errorf("insert image_metadata: %w", err)
// 	}
// 	return id, nil
// }

// // MarkAsIndexed updates the Qdrant sync status after the fingerprint vector
// // has been successfully stored in Qdrant.
// // Call this immediately after a successful QdrantDB.StoreFingerprint.
// func (db *DB) MarkAsIndexed(ctx context.Context, id uuid.UUID, version string) error {
// 	_, err := db.pool.ExecContext(ctx, `   -- ← ExecContext, not Exec
// 		UPDATE image_metadata
// 		SET is_indexed    = TRUE,
// 		    indexed_at    = NOW(),
// 		    index_version = $2
// 		WHERE id = $1`,
// 		id, version,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("mark as indexed: %w", err)
// 	}
// 	return nil
// }

// // MarkAsUnindexed resets the Qdrant sync flag, e.g. after a Qdrant delete or failure.
// func (db *DB) MarkAsUnindexed(ctx context.Context, id uuid.UUID) error {
// 	_, err := db.pool.ExecContext(ctx, `   -- ← ExecContext, not Exec
// 		UPDATE image_metadata
// 		SET is_indexed    = FALSE,
// 		    indexed_at    = NULL,
// 		    index_version = NULL
// 		WHERE id = $1`,
// 		id,
// 	)
// 	return err
// }

// // SetAIFlag updates the AI-generation analysis fields for an existing image.
// func (db *DB) SetAIFlag(ctx context.Context, id uuid.UUID, isAI bool, confidence float64, model string) error {
// 	_, err := db.pool.ExecContext(ctx, `   -- ← ExecContext, not Exec
// 		UPDATE image_metadata
// 		SET is_ai_generated = $2,
// 		    ai_confidence   = $3,
// 		    ai_model_used   = $4
// 		WHERE id = $1`,
// 		id, isAI, confidence, model,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("set ai flag: %w", err)
// 	}
// 	return nil
// }

// // SoftDeleteImage marks an image as deleted without removing the row.
// // Also call QdrantDB.DeleteFingerprint for the same UUID to remove the vector.
// func (db *DB) SoftDeleteImage(ctx context.Context, id uuid.UUID) error {
// 	_, err := db.pool.ExecContext(ctx,     // ← ExecContext, not Exec
// 		`UPDATE image_metadata SET is_deleted = TRUE WHERE id = $1`, id)
// 	if err != nil {
// 		return fmt.Errorf("soft delete image: %w", err)
// 	}
// 	return nil
// }

// // ---------------------------------------------------------------------------
// // Image Metadata — Read operations
// // ---------------------------------------------------------------------------

// // GetImageMetadata fetches a single metadata row by its UUID.
// func (db *DB) GetImageMetadata(ctx context.Context, id uuid.UUID) (*models.ImageMetadata, error) {
// 	const q = `
// 		SELECT id, submitted_by, title, description, source_url, external_ref_id,
// 		       checksum_sha256, mime_type, width_px, height_px,
// 		       is_ai_generated, ai_confidence, ai_model_used, content_flags,
// 		       location_label, latitude, longitude,
// 		       category, tags,
// 		       is_indexed, indexed_at, index_version,
// 		       captured_at, created_at, updated_at
// 		FROM image_metadata
// 		WHERE id = $1 AND is_deleted = FALSE`

// 	m := &models.ImageMetadata{}
// 	err := db.pool.QueryRowContext(ctx, q, id).Scan(  // ← QueryRowContext, not QueryRow
// 		&m.ID, &m.SubmittedBy, &m.Title, &m.Description, &m.SourceURL, &m.ExternalRefID,
// 		&m.ChecksumSHA256, &m.MimeType, &m.WidthPx, &m.HeightPx,
// 		&m.IsAIGenerated, &m.AIConfidence, &m.AIModelUsed, &m.ContentFlags,
// 		&m.LocationLabel, &m.Latitude, &m.Longitude,
// 		&m.Category, &m.Tags,
// 		&m.IsIndexed, &m.IndexedAt, &m.IndexVersion,
// 		&m.CapturedAt, &m.CreatedAt, &m.UpdatedAt,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("get image_metadata %s: %w", id, err)
// 	}
// 	return m, nil
// }

// // GetImageMetadataBatch fetches multiple metadata rows by a slice of UUIDs in one query.
// // Used after a Qdrant similarity search returns a list of matching IDs.
// // The returned map is keyed by UUID for easy lookup.
// func (db *DB) GetImageMetadataBatch(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*models.ImageMetadata, error) {
// 	if len(ids) == 0 {
// 		return map[uuid.UUID]*models.ImageMetadata{}, nil
// 	}

// 	// Convert []uuid.UUID to []string for the ANY($1::uuid[]) binding
// 	idStrs := make([]string, len(ids))
// 	for i, id := range ids {
// 		idStrs[i] = id.String()
// 	}

// 	const q = `
// 		SELECT id, submitted_by, title, description, source_url, external_ref_id,
// 		       checksum_sha256, mime_type, width_px, height_px,
// 		       is_ai_generated, ai_confidence, ai_model_used, content_flags,
// 		       location_label, latitude, longitude,
// 		       category, tags,
// 		       is_indexed, indexed_at, index_version,
// 		       captured_at, created_at, updated_at
// 		FROM image_metadata
// 		WHERE id = ANY($1::uuid[]) AND is_deleted = FALSE`

// 	rows, err := db.pool.QueryContext(ctx, q, idStrs)  // ← QueryContext, not Query
// 	if err != nil {
// 		return nil, fmt.Errorf("batch get image_metadata: %w", err)
// 	}
// 	defer rows.Close()

// 	result := make(map[uuid.UUID]*models.ImageMetadata, len(ids))
// 	for rows.Next() {
// 		m := &models.ImageMetadata{}
// 		err := rows.Scan(
// 			&m.ID, &m.SubmittedBy, &m.Title, &m.Description, &m.SourceURL, &m.ExternalRefID,
// 			&m.ChecksumSHA256, &m.MimeType, &m.WidthPx, &m.HeightPx,
// 			&m.IsAIGenerated, &m.AIConfidence, &m.AIModelUsed, &m.ContentFlags,
// 			&m.LocationLabel, &m.Latitude, &m.Longitude,
// 			&m.Category, &m.Tags,
// 			&m.IsIndexed, &m.IndexedAt, &m.IndexVersion,
// 			&m.CapturedAt, &m.CreatedAt, &m.UpdatedAt,
// 		)
// 		if err != nil {
// 			return nil, fmt.Errorf("scan image_metadata row: %w", err)
// 		}
// 		result[m.ID] = m
// 	}
// 	return result, rows.Err()
// }

// // GetUnindexedImages returns images not yet sent to Qdrant.
// // Useful for a background job that re-indexes failed or missing entries.
// func (db *DB) GetUnindexedImages(ctx context.Context, limit int) ([]models.ImageMetadata, error) {
// 	const q = `
// 		SELECT id, submitted_by, title, description, source_url, external_ref_id,
// 		       checksum_sha256, mime_type, width_px, height_px,
// 		       is_ai_generated, ai_confidence, ai_model_used, content_flags,
// 		       location_label, latitude, longitude,
// 		       category, tags,
// 		       is_indexed, indexed_at, index_version,
// 		       captured_at, created_at, updated_at
// 		FROM image_metadata
// 		WHERE is_deleted = FALSE AND is_indexed = FALSE
// 		ORDER BY created_at ASC
// 		LIMIT $1`

// 	rows, err := db.pool.QueryContext(ctx, q, limit)  // ← QueryContext, not Query
// 	if err != nil {
// 		return nil, fmt.Errorf("get unindexed images: %w", err)
// 	}
// 	defer rows.Close()

// 	var results []models.ImageMetadata
// 	for rows.Next() {
// 		m := models.ImageMetadata{}
// 		err := rows.Scan(
// 			&m.ID, &m.SubmittedBy, &m.Title, &m.Description, &m.SourceURL, &m.ExternalRefID,
// 			&m.ChecksumSHA256, &m.MimeType, &m.WidthPx, &m.HeightPx,
// 			&m.IsAIGenerated, &m.AIConfidence, &m.AIModelUsed, &m.ContentFlags,
// 			&m.LocationLabel, &m.Latitude, &m.Longitude,
// 			&m.Category, &m.Tags,
// 			&m.IsIndexed, &m.IndexedAt, &m.IndexVersion,
// 			&m.CapturedAt, &m.CreatedAt, &m.UpdatedAt,
// 		)
// 		if err != nil {
// 			return nil, fmt.Errorf("scan unindexed image: %w", err)
// 		}
// 		results = append(results, m)
// 	}
// 	return results, rows.Err()
// }

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/models"
	"github.com/google/uuid"
)

type DB struct {
	pool *sql.DB
}

func NewDB(sqlDB *sql.DB) *DB {
	return &DB{pool: sqlDB}
}

// InsertImageMetadata inserts a new record and returns both the UUID and the
// serial_id (BIGSERIAL). The serial_id is what gets embedded in the watermark.
func (db *DB) InsertImageMetadata(
	ctx context.Context,
	m models.ImageMetadata,
) (uuid.UUID, int64, error) {

	query := `
    INSERT INTO image_metadata (
        title,
        description,
        mime_type,
        width_px,
        height_px,
        is_ai_generated,
        captured_at
    )
    VALUES ($1,$2,$3,$4,$5,$6,$7)
    RETURNING id, serial_id;
    `

	var id uuid.UUID
	var serialID int64

	err := db.pool.QueryRowContext(
		ctx, query,
		m.Title,
		m.Description,
		m.MimeType,
		m.WidthPx,
		m.HeightPx,
		m.IsAIGenerated,
		m.CapturedAt,
	).Scan(&id, &serialID)

	if err != nil {
		return uuid.Nil, 0, err
	}

	return id, serialID, nil
}

// GetImageMetadataBySerialID looks up a row using the watermark-embedded serial_id.
func (db *DB) GetImageMetadataBySerialID(
	ctx context.Context,
	serialID int64,
) (*models.ImageMetadata, error) {

	query := `
    SELECT
        id,
        serial_id,
        title,
        description,
        mime_type,
        width_px,
        height_px,
        is_ai_generated,
        captured_at,
        created_at,
        updated_at
    FROM image_metadata
    WHERE serial_id = $1;
    `

	var m models.ImageMetadata
	err := db.pool.QueryRowContext(ctx, query, serialID).Scan(
		&m.ID,
		&m.SerialID,
		&m.Title,
		&m.Description,
		&m.MimeType,
		&m.WidthPx,
		&m.HeightPx,
		&m.IsAIGenerated,
		&m.CapturedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &m, nil
}

// GetImageMetadata fetches by UUID (still useful for Qdrant results).
func (db *DB) GetImageMetadata(
	ctx context.Context,
	id uuid.UUID,
) (*models.ImageMetadata, error) {

	query := `
    SELECT
        id,
        serial_id,
        title,
        description,
        mime_type,
        width_px,
        height_px,
        is_ai_generated,
        captured_at,
        created_at,
        updated_at
    FROM image_metadata
    WHERE id = $1;
    `

	var m models.ImageMetadata
	err := db.pool.QueryRowContext(ctx, query, id).Scan(
		&m.ID,
		&m.SerialID,
		&m.Title,
		&m.Description,
		&m.MimeType,
		&m.WidthPx,
		&m.HeightPx,
		&m.IsAIGenerated,
		&m.CapturedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &m, nil
}

// GetImageMetadataBatch fetches multiple rows by UUID slice (for Qdrant results).
func (db *DB) GetImageMetadataBatch(
	ctx context.Context,
	ids []uuid.UUID,
) (map[uuid.UUID]*models.ImageMetadata, error) {

	if len(ids) == 0 {
		return map[uuid.UUID]*models.ImageMetadata{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
    SELECT
        id,
        serial_id,
        title,
        description,
        mime_type,
        width_px,
        height_px,
        is_ai_generated,
        captured_at,
        created_at,
        updated_at
    FROM image_metadata
    WHERE id IN (%s);
    `, strings.Join(placeholders, ","))

	rows, err := db.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*models.ImageMetadata)
	for rows.Next() {
		var m models.ImageMetadata
		err := rows.Scan(
			&m.ID,
			&m.SerialID,
			&m.Title,
			&m.Description,
			&m.MimeType,
			&m.WidthPx,
			&m.HeightPx,
			&m.IsAIGenerated,
			&m.CapturedAt,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result[m.ID] = &m
	}

	return result, rows.Err()
}
