## 4. Updated Schema Summary
```
users                             image_metadata
──────────────────────────        ──────────────────────────────
id          UUID (PK)  ◄───────── user_id     UUID (FK, nullable)
username    VARCHAR               id           UUID (PK)
email       VARCHAR               serial_id    BIGSERIAL
password_hash TEXT                title        TEXT
full_name   TEXT                  mime_type    VARCHAR
is_active   BOOLEAN               width_px     INTEGER
created_at  TIMESTAMPTZ           height_px    INTEGER
updated_at  TIMESTAMPTZ           is_ai_generated BOOLEAN
                                  captured_at  TIMESTAMPTZ
                                  created_at   TIMESTAMPTZ
                                  updated_at   TIMESTAMPTZ
                                       │
                                       ▼
                                  image_vectors
                                  ──────────────────────────
                                  id          UUID (PK)
                                  image_id    UUID (FK)
                                  vector      VECTOR(256)