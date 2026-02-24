-- ============================================================
-- Image Fingerprint Database Schema
-- Vector storage is handled externally by Qdrant.
-- This schema covers only relational data.
-- ============================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;  -- gen_random_uuid()

-- ============================================================
-- USERS TABLE
-- ============================================================
CREATE TABLE users (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(50)  NOT NULL UNIQUE,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   TEXT         NOT NULL,    -- Store bcrypt hash, NEVER plaintext
    full_name       VARCHAR(150),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ
);

-- ============================================================
-- ADMINS TABLE
-- ============================================================
CREATE TABLE admins (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(50)  NOT NULL UNIQUE,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   TEXT         NOT NULL,
    full_name       VARCHAR(150),
    role            VARCHAR(30)  NOT NULL DEFAULT 'admin'
                        CHECK (role IN ('admin', 'superadmin')),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ
);

-- ============================================================
-- IMAGE METADATA TABLE
-- We do NOT store the image itself, only metadata about it.
-- The fingerprint vector is stored in Qdrant, linked by this
-- table's UUID (id), which is used as the Qdrant point ID.
-- ============================================================
CREATE TABLE image_metadata (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    submitted_by    UUID         REFERENCES users(id) ON DELETE SET NULL,

    -- ----- Identity / Source -----
    title           VARCHAR(255),
    description     TEXT,
    source_url      TEXT,                      -- URL where the image was found/sourced
    external_ref_id VARCHAR(255),              -- ID in an external system, if any
    checksum_sha256 CHAR(64),                  -- SHA-256 of the image bytes (for dedup)

    -- ----- Image Properties -----
    mime_type       VARCHAR(100),              -- e.g. 'image/jpeg', 'image/png'
    width_px        INT,
    height_px       INT,

    -- ----- AI / Content Analysis -----
    is_ai_generated BOOLEAN      NOT NULL DEFAULT FALSE,
    ai_confidence   NUMERIC(5,4),              -- Detector confidence score [0.0 – 1.0]
    ai_model_used   VARCHAR(100),              -- e.g. 'DALL-E 3', 'Stable Diffusion XL'
    content_flags   TEXT[]       DEFAULT '{}', -- e.g. ARRAY['nsfw','violence','watermarked']

    -- ----- Location (all optional) -----
    location_label  VARCHAR(255),              -- Human-readable, e.g. 'Paris, France'
    latitude        NUMERIC(9,6),              -- e.g. 48.858844  (NULL if unknown)
    longitude       NUMERIC(9,6),              -- e.g. 2.294351   (NULL if unknown)

    -- ----- Classification -----
    category        VARCHAR(100),              -- e.g. 'artwork', 'photograph', 'screenshot'
    tags            TEXT[]       DEFAULT '{}',

    -- ----- Qdrant Sync Status -----
    -- Tracks whether this image's fingerprint vector has been
    -- successfully stored in Qdrant. The UUID in this table (id)
    -- is used directly as the Qdrant point ID to keep them in sync.
    is_indexed      BOOLEAN      NOT NULL DEFAULT FALSE, -- TRUE once stored in Qdrant
    indexed_at      TIMESTAMPTZ,                         -- When it was last indexed
    index_version   VARCHAR(50),                         -- Algorithm version, e.g. 'v1'

    -- ----- Timestamps / Soft Delete -----
    captured_at     TIMESTAMPTZ,               -- When the original image was taken/created
    is_deleted      BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================
-- INDEXES
-- ============================================================

-- Users
CREATE INDEX idx_users_email        ON users(email);
CREATE INDEX idx_users_username     ON users(username);

-- Admins
CREATE INDEX idx_admins_email       ON admins(email);
CREATE INDEX idx_admins_username    ON admins(username);

-- Image metadata — common lookup patterns
CREATE INDEX idx_images_submitted_by  ON image_metadata(submitted_by);
CREATE INDEX idx_images_checksum      ON image_metadata(checksum_sha256);
CREATE INDEX idx_images_created_at    ON image_metadata(created_at DESC);
CREATE INDEX idx_images_is_ai         ON image_metadata(is_ai_generated);
CREATE INDEX idx_images_category      ON image_metadata(category);
CREATE INDEX idx_images_tags          ON image_metadata USING GIN(tags);
CREATE INDEX idx_images_content_flags ON image_metadata USING GIN(content_flags);

-- Partial index: quickly find images not yet sent to Qdrant
-- Useful for background re-indexing jobs after a crash or failure
CREATE INDEX idx_images_not_indexed
    ON image_metadata(created_at ASC)
    WHERE is_deleted = FALSE AND is_indexed = FALSE;

-- Partial index: quickly find active AI-generated images
CREATE INDEX idx_images_ai_active
    ON image_metadata(is_ai_generated, created_at DESC)
    WHERE is_deleted = FALSE AND is_ai_generated = TRUE;

-- ============================================================
-- HELPER: auto-update updated_at timestamps
-- ============================================================
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_admins_updated_at
    BEFORE UPDATE ON admins
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_image_metadata_updated_at
    BEFORE UPDATE ON image_metadata
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- USEFUL VIEWS
-- ============================================================

-- Full image info, safe for API/public queries
CREATE VIEW v_images AS
SELECT
    m.id,
    m.submitted_by,
    m.title,
    m.description,
    m.source_url,
    m.mime_type,
    m.width_px,
    m.height_px,
    m.is_ai_generated,
    m.ai_confidence,
    m.ai_model_used,
    m.content_flags,
    m.location_label,
    m.latitude,
    m.longitude,
    m.category,
    m.tags,
    m.captured_at,
    m.created_at,
    m.is_indexed,
    m.indexed_at,
    m.index_version
FROM image_metadata m
WHERE m.is_deleted = FALSE;