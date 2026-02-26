-- Enable required extension for UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS image_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    title TEXT NULL,
    description TEXT NULL,

    mime_type TEXT NULL,
    width_px INTEGER NULL,
    height_px INTEGER NULL,

    -- AI analysis
    is_ai_generated BOOLEAN NOT NULL DEFAULT FALSE,

    -- Qdrant indexing flag
    is_indexed BOOLEAN NOT NULL DEFAULT FALSE,

    captured_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Useful indexes
CREATE INDEX IF NOT EXISTS idx_image_metadata_created_at
ON image_metadata(created_at);

CREATE INDEX IF NOT EXISTS idx_image_metadata_is_indexed
ON image_metadata(is_indexed);