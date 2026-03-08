-- ─────────────────────────────────────────────────────────────────────────────
-- Extensions
-- ─────────────────────────────────────────────────────────────────────────────

-- pgvector must be installed on the server first:
--   CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS vector;

-- pgcrypto gives us gen_random_uuid() for UUID primary keys
CREATE EXTENSION IF NOT EXISTS pgcrypto;


-- ─────────────────────────────────────────────────────────────────────────────
-- users
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE users (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(50)   NOT NULL UNIQUE,
    email         VARCHAR(255)  NOT NULL UNIQUE,
    password_hash TEXT          NOT NULL,
    full_name     TEXT,
    is_active     BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- Fast lookup by email (used by Login)
CREATE INDEX idx_users_email    ON users (email);
-- Fast lookup by username
CREATE INDEX idx_users_username ON users (username);


-- ─────────────────────────────────────────────────────────────────────────────
-- image_metadata
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE image_metadata (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    serial_id       BIGSERIAL     NOT NULL UNIQUE,   -- embedded in the watermark payload
    user_id         UUID          REFERENCES users (id) ON DELETE SET NULL,
    title           TEXT,
    description     TEXT,
    mime_type       VARCHAR(100)  NOT NULL,
    width_px        INTEGER       NOT NULL CHECK (width_px  > 0),
    height_px       INTEGER       NOT NULL CHECK (height_px > 0),
    is_ai_generated BOOLEAN       NOT NULL DEFAULT FALSE,
    captured_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- Watermark extraction path: serial_id → metadata row
CREATE INDEX idx_image_metadata_serial_id ON image_metadata (serial_id);
-- Filter all images belonging to a specific user
CREATE INDEX idx_image_metadata_user_id   ON image_metadata (user_id);


-- ─────────────────────────────────────────────────────────────────────────────
-- image_vectors
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE image_vectors (
    id        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    image_id  UUID        NOT NULL UNIQUE REFERENCES image_metadata (id) ON DELETE CASCADE,
    vector    VECTOR(256) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- IVFFlat ANN index for cosine similarity search used by FindSimilarImages.
-- lists = 100 is a good starting point; tune upward as the table grows.
-- Requires at least one row to exist before CREATE INDEX will succeed,
-- so run this after your first batch of inserts if the table starts empty.
CREATE INDEX idx_image_vectors_ivfflat
    ON image_vectors
    USING ivfflat (vector vector_cosine_ops)
    WITH (lists = 100);


-- ─────────────────────────────────────────────────────────────────────────────
-- updated_at auto-maintenance trigger
-- ─────────────────────────────────────────────────────────────────────────────

-- Single trigger function reused by both tables.
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

CREATE TRIGGER trg_image_metadata_updated_at
    BEFORE UPDATE ON image_metadata
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();