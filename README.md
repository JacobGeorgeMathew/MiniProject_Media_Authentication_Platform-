# Media Authentication Platform (MAP)

> **Dual-layer invisible watermarking and perceptual fingerprinting for robust image provenance, tamper detection, and Watermarking-as-a-Service (WaaS).**

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
  - [Core Technologies](#core-technologies)
  - [Embedding Pipeline](#1-watermark-embedding-pipeline)
  - [Fingerprint Pipeline](#2-fingerprint-generation-pipeline)
  - [Authentication Pipeline](#3-authentication-pipeline)
  - [Dual-Layer Verification](#4-dual-layer-identity-verification)
  - [Tamper Score](#5-tamper-score-computation)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Environment Variables](#environment-variables)
  - [Database Setup](#database-setup)
  - [Running the Server](#running-the-server)
- [API Reference](#api-reference)
  - [Health](#health)
  - [Users (Public)](#users--public)
  - [Users (Protected)](#users--protected)
  - [Images (Protected)](#images--protected)
  - [Images (Public)](#images--public)
- [Enterprise WaaS Layer](#enterprise-waas-layer)
  - [Enterprise Management (JWT)](#enterprise-management-endpoints--jwt)
  - [Enterprise API (X-API-Key)](#enterprise-api-endpoints--x-api-key)
  - [API Key Format](#api-key-format)
  - [Auth Guards](#authentication-guards)
- [Postman Testing](#postman-testing)
- [HTTP Status Codes](#http-status-codes)
- [Technical Deep Dives](#technical-deep-dives)
  - [DWT Coordinate Contract](#dwt-coordinate-contract)
  - [PartialDWT_HL Optimisation](#partialdwt_hl-optimisation)
  - [Spatial Update Formula](#spatial-update-formula)
  - [FindValueOptimized](#findvalueoptimized)
- [Data Stores](#data-stores)
- [Key Design Decisions](#key-design-decisions)

---

## Overview

The Media Authentication Platform is a production-grade backend system built in Go that embeds invisible digital watermarks into images and later extracts them to verify ownership, detect tampering, and identify visual duplicates. It is deployed as a REST API and exposes a **Watermarking-as-a-Service (WaaS)** tier for commercial enterprise clients.

The platform combines two independent authentication channels:

| Channel | Mechanism | Strength |
|---|---|---|
| **Watermark** | Binary payload in DWT HL sub-band via DCT basis injection | Carries explicit owner metadata; survives cropping |
| **Fingerprint** | 256-D perceptual vector via 2-level db4 DWT | Robust to compression, mild edits, format conversion |

Both channels must independently resolve to the same record for a high-confidence authentication result. Divergence between the channels is itself a tamper signal.

---

## Architecture

### Core Technologies

- **Haar DWT (critically sampled)** — Standard one-level forward/inverse Haar Discrete Wavelet Transform. A 256×256 spatial tile produces four 128×128 sub-bands (LL, LH, HL, HH). The HL sub-band is the embedding target.
- **DCT basis injection (QIM)** — Imperceptible payload encoding in the HL sub-band using Quantisation Index Modulation applied to 16×16 blocks.
- **Daubechies (db4) DWT fingerprinting** — Multi-level wavelet decomposition producing a 256-D perceptual feature vector for content-based identity lookup.
- **Vector similarity search** — ANN search (Qdrant) for duplicate and tampered-image detection using cosine similarity with a 95% threshold.
- **Pre-computed constant matrices** — DCT basis values baked into the binary at build time via code generation (`gen_constants/main.go` → `constants_data.go`), yielding O(1) runtime lookup.

---

### 1. Watermark Embedding Pipeline

```
Input Image
    │
    ▼
Format-Agnostic Ingestion (JPEG, PNG, TIFF, BMP, WebP)
    │
    ▼
RGB → YCbCr Conversion  (watermark embedded in Y channel only)
    │
    ▼
PartialDWT_HL Pre-Check  ──► 409 Conflict if already watermarked
    │
    ▼
Fingerprint Collision Check (Vector DB, >95% similarity) ──► 409 if duplicate
    │
    ▼
Metadata Storage in PostgreSQL  →  retrieves serial_id (BIGSERIAL)
    │
    ▼
Payload Construction: [ serial_id | CRC checksum | flags ]
    │
    ▼
Tile Decomposition (256×256 non-overlapping tiles)
    │
    ▼
For each tile:
    ├── One-level Haar DWT → 4× 128×128 sub-bands (LL, LH, HL, HH)
    ├── HL Block Grid: 8×8 grid of 16×16 blocks (64 blocks, offsets 0–112)
    │   └── First row (by=0) and first column (bx=0) reserved as sync markers
    └── Per eligible block:
        A′(x,y) = A(x,y) + k · B_uv(x,y)   ← spatial update formula
    │
    ▼
Inverse DWT (IDWT) per tile  →  reconstruct 256×256 spatial tile
    │
    ▼
Y Channel Reconstruction + Normalisation [0, 255]
    │
    ▼
Merge Y + Cb + Cr  →  Output Watermarked Image
```

**Response headers returned:**
- `X-Image-ID` — UUID of the `image_metadata` row
- `X-Serial-ID` — BIGSERIAL value embedded in the watermark payload

---

### 2. Fingerprint Generation Pipeline

```
Input Image
    │
    ▼
Resize to 256×256
    │
    ▼
RGB → YCbCr → Y channel (256×256 luminance plane)
    │
    ▼
2-Level db4 DWT decomposition
    │
    ▼
Extract LL2 sub-band (64×64 — global low-frequency structure)
    │
    ▼
Partition into 64 non-overlapping 8×8 blocks
    │
    ▼
Per block: max(|coeff|) across LL, LH, HL, HH orientations
    │
    ▼
Assemble 256-D vector  (64 blocks × 4 orientations)
    │
    ▼
L2 Normalisation (unit vector)
    │
    ▼
Store in Qdrant vector DB keyed to metadata_id
```

---

### 3. Authentication Pipeline

```
Query Image
    │
    ▼
RGB → YCbCr → Y channel
    │
    ▼
Identify()  →  scan 256-aligned offsets with PartialDWT_HL
              to locate tile origin even after cropping
    │
    ▼
Tile Decomposition at identified origin
    │
    ▼
For each tile:
    ├── One-level Haar DWT → 128×128 HL sub-band
    └── Per 16×16 HL block:
        FindValueOptimized()  →  dot product against DCT basis
        qimExtract()          →  convert coefficient delta → bits
    │
    ▼
Bitstream Reconstruction
    │
    ▼
CRC Verification  ──► 422 Unprocessable if CRC fails
    │
    ▼
Parse serial_id from payload
    │
    ▼
PostgreSQL Lookup  →  owner, timestamp, rights metadata
    │
    ▼
Generate fresh 256-D fingerprint of query image
    │
    ▼
Qdrant ANN Search (cosine similarity, threshold 95%)
    │
    ▼
Tamper Score Computation (CRC + spatial consistency +
    coefficient delta variance + channel agreement)
    │
    ▼
Return Authentication Report
```

---

### 4. Dual-Layer Identity Verification

| | Watermark Channel | Fingerprint Channel |
|---|---|---|
| **Carrier** | HL sub-band coefficients (DWT) | 256-D perceptual vector (db4 DWT) |
| **Payload** | Explicit `serial_id` + CRC + flags | Cosine similarity match |
| **Verification** | CRC checksum | >95% similarity threshold |
| **Crop robustness** | ✅ Identify() scans aligned offsets | ✅ Content-based, layout-independent |
| **Sensitivity** | Heavy modification or overwriting | Significant visual content changes |

When both channels independently resolve to the same metadata record, authentication confidence is highest. Divergence is itself a tamper signal:

- **Matching fingerprint + corrupted watermark** → incidental damage (compression, format conversion)
- **Matching watermark + non-matching fingerprint** → content substitution (metadata header copied to a different image)

---

### 5. Tamper Score Computation

The tamper score is a `float64` in `[0.0, 1.0]` returned with every authentication response.

| Signal | Contribution |
|---|---|
| CRC pass/fail | Fail = high contribution |
| Spatial bit consistency across tiles | High variance → localised modification |
| Coefficient delta magnitude variance | Large deviation from expected `k` → coefficient manipulation |
| Channel agreement | Watermark `serial_id` ≠ fingerprint `metadata_id` → raised score |
| Vector similarity | Score just above 95% vs. 99%+ reflects different confidence |

`tamper_score = 0.0` means a fully intact, unmodified watermarked image. Scores above a configurable threshold are flagged in the authentication report.

---

## Project Structure

```
.
├── main.go                     # Server entry point, route wiring
├── embed.go                    # Watermark embedding logic
├── extract.go                  # Watermark extraction logic
├── partial_dwt.go              # PartialDWT_HL optimised block transform
├── performwatermark.go         # High-level embed/extract orchestration
├── constants_data.go           # Pre-computed DCT basis matrices (generated)
├── gen_constants/
│   └── main.go                 # Build-time code generator for DCT constants
├── image_service.go            # Image business logic (embed, authenticate)
├── image_repo.go               # PostgreSQL queries for image_metadata
├── enterprise_models.go        # Enterprise, EnterpriseAPIKey, EnterpriseAPIUsage structs
├── enterprise_repo.go          # EnterpriseRepository interface + pgxpool implementation
├── enterprise_service.go       # Enterprise business logic, API key generation/validation
├── enterprise_middleware.go    # EnterpriseJWTMiddleware + EnterpriseAPIKeyMiddleware
├── enterprise_handler.go       # 12 enterprise endpoint handlers
├── enterprise_schema.sql       # DB migration: enterprises, enterprise_api_keys, enterprise_api_usage
└── user_service.go             # User authentication and profile management
```

---

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Qdrant (vector database)
- `goose` or `psql` for database migrations

### Environment Variables

```env
# Server
PORT=5000

# PostgreSQL
DATABASE_URL=postgres://user:password@localhost:5432/map_db

# Qdrant
QDRANT_HOST=localhost
QDRANT_PORT=6333
QDRANT_COLLECTION=image_fingerprints

# JWT
JWT_SECRET=your-user-jwt-secret
ENTERPRISE_JWT_SECRET=your-enterprise-jwt-secret   # cryptographically isolated

# Watermarking
WATERMARK_STRENGTH=0.1                             # embedding strength scalar k
```

> ⚠️ The user JWT secret and the enterprise JWT secret **must** be different values. The two token spaces are cryptographically isolated by design.

### Database Setup

```bash
# Apply base schema
psql $DATABASE_URL -f schema.sql

# Apply enterprise schema extension
psql $DATABASE_URL -f enterprise_schema.sql
```

The enterprise schema adds:
- `enterprises` — company accounts
- `enterprise_api_keys` — bcrypt-hashed keys with 12-char plain-text prefix for fast indexed lookup (the raw key is never stored)
- `enterprise_api_usage` — append-only audit log for billing
- `enterprise_id` (nullable FK) on `image_metadata` — traces each watermarked image back to the enterprise that registered it

### Running the Server

```bash
# Generate pre-computed DCT constants (only needed once or after algorithm changes)
go run gen_constants/main.go

# Run the server
go run main.go
```

The server starts on `http://localhost:5000`.

---

## API Reference

**Base URL:** `http://localhost:5000/api/v1`

All protected routes require the header:
```
Authorization: Bearer <JWT>
```

All image endpoints use `multipart/form-data`, **not** JSON.

---

### Health

#### `GET /health`
Smoke test. No auth required.

```json
{ "status": "ok" }
```

---

### Users — Public

#### `POST /users/register`

| Field | Type | Required | Description |
|---|---|---|---|
| `username` | string | ✅ | Unique username |
| `email` | string | ✅ | Unique email address |
| `password` | string | ✅ | Minimum 8 characters |
| `full_name` | string | ❌ | Display name |

**Response `201`:**
```json
{
  "id": "a1b2c3d4-e5f6-...",
  "username": "alice",
  "email": "alice@example.com",
  "full_name": "Alice Smith",
  "is_active": true,
  "created_at": "2025-03-08T10:00:00Z"
}
```

---

#### `POST /users/login`

| Field | Type | Required |
|---|---|---|
| `email` | string | ✅ |
| `password` | string | ✅ |

**Response `200`:**
```json
{
  "id": "a1b2c3d4-...",
  "username": "alice",
  "email": "alice@example.com",
  "token": "<JWT string>"
}
```

---

### Users — Protected

All routes below require `Authorization: Bearer <token>`.

#### `GET /users/me`
Returns the authenticated user's own profile.

#### `PUT /users/me`
Updates `username`, `email`, or `full_name`. At least one field required.

#### `PUT /users/me/password`

| Field | Type | Required |
|---|---|---|
| `old_password` | string | ✅ |
| `new_password` | string | ✅ (min 8 chars) |

#### `DELETE /users/me`
Soft-deletes the account (`is_active = false`). All image provenance records are preserved. Cannot be undone via the API.

#### `GET /users/:id`
Fetches any user's public profile by UUID.

---

### Images — Protected

> ⚠️ These endpoints use `form-data`. Set the `image` field type to **File** in Postman.

#### `POST /images/watermark`

Embeds an invisible DWT+DCT watermark, stores metadata in PostgreSQL, and indexes a perceptual fingerprint in Qdrant. Returns the watermarked image as binary.

| Field | Type | Required | Description |
|---|---|---|---|
| `image` | File | ✅ | JPEG, PNG, TIFF, BMP, or WebP |
| `title` | string | ❌ | Human-readable label |
| `description` | string | ❌ | Free-text description |
| `is_ai_generated` | string | ❌ | `"true"` or `"false"` |

**Response:** Binary image body (use *Send and Download* in Postman to save it).

**Response Headers:**
```
X-Image-ID: f3a7b901-1234-...     ← UUID of image_metadata row
X-Serial-ID: 42                    ← BIGSERIAL embedded in watermark payload
Content-Type: image/jpeg           ← mirrors uploaded MIME type
```

---

#### `POST /images/authenticate`

Full dual-channel authentication. Extracts watermark payload, verifies CRC, queries PostgreSQL for ownership, and runs Qdrant ANN search.

| Field | Type | Required | Description |
|---|---|---|---|
| `image` | File | ✅ | The watermarked image to authenticate |
| `k` | string | ❌ | Max similar images to return (default: 5) |

**Response `200`:**
```json
{
  "watermark_channel": {
    "image_id": "f3a7b901-...",
    "serial_id": 42,
    "title": "My Photo",
    "description": "Landscape shot",
    "mime_type": "image/jpeg",
    "width_px": 1920,
    "height_px": 1080,
    "is_ai_generated": false
  },
  "similar_images": [
    { "image_id": "f3a7b901-...", "similarity": 0.998, "title": "My Photo" }
  ],
  "tamper_score": 0.0
}
```

---

### Images — Public

#### `POST /verify`

Identical to `/images/authenticate` but requires no JWT. Designed for unregistered users verifying the provenance of any image they encounter.

| Field | Type | Required |
|---|---|---|
| `image` | File | ✅ |
| `k` | string | ❌ |

---

## Enterprise WaaS Layer

The enterprise layer exposes the full watermarking and authentication pipeline as a **Watermarking-as-a-Service** offering for commercial clients. It is completely independent of the user tier — separate account records, separate JWT signing secret, and API key authentication via `X-API-Key`.

### Complete Route Tree

```
POST   /api/v1/enterprise/register          — public
POST   /api/v1/enterprise/login             — public
GET    /api/v1/enterprise/me                — Enterprise JWT
PUT    /api/v1/enterprise/me                — Enterprise JWT
PUT    /api/v1/enterprise/me/password       — Enterprise JWT
DELETE /api/v1/enterprise/me                — Enterprise JWT
POST   /api/v1/enterprise/keys              — Enterprise JWT
GET    /api/v1/enterprise/keys              — Enterprise JWT
DELETE /api/v1/enterprise/keys/:keyId       — Enterprise JWT
GET    /api/v1/enterprise/usage             — Enterprise JWT
POST   /api/v1/enterprise/api/watermark     — X-API-Key
POST   /api/v1/enterprise/api/authenticate  — X-API-Key
```

---

### Enterprise Management Endpoints — JWT

These endpoints are used by enterprise account administrators through their dashboard.

#### `POST /enterprise/register` — Public
Registers a new enterprise account.

#### `POST /enterprise/login` — Public
Authenticates an enterprise and returns a JWT signed with `ENTERPRISE_JWT_SECRET`.

#### `GET /enterprise/me` — Enterprise JWT
Returns the authenticated enterprise's profile.

#### `PUT /enterprise/me` — Enterprise JWT
Updates mutable enterprise profile fields.

#### `PUT /enterprise/me/password` — Enterprise JWT
Rotates the enterprise account password.

#### `DELETE /enterprise/me` — Enterprise JWT
Deactivates the enterprise account.

#### `POST /enterprise/keys` — Enterprise JWT
Generates a new API key. The raw key is returned **exactly once** and never stored — only the bcrypt hash is persisted.

**Response includes:**
```json
{
  "key_id": "...",
  "raw_key": "map_<43-char-base64url>",
  "prefix": "map_<first-12-chars>",
  "created_at": "..."
}
```

#### `GET /enterprise/keys` — Enterprise JWT
Lists all active API keys for the enterprise (prefix and metadata only — raw keys are never returned after creation).

#### `DELETE /enterprise/keys/:keyId` — Enterprise JWT
Revokes a specific API key by ID.

#### `GET /enterprise/usage` — Enterprise JWT
Returns usage records from `enterprise_api_usage` for billing and audit purposes.

---

### Enterprise API Endpoints — X-API-Key

These are the programmatic WaaS endpoints called by enterprise clients in their own applications.

```
X-API-Key: map_<your-api-key>
```

#### `POST /enterprise/api/watermark`
Identical behaviour to `POST /images/watermark`. Uses the same `ImageService` pipeline — no code duplication. Every watermarked image has its `enterprise_id` recorded in `image_metadata`.

#### `POST /enterprise/api/authenticate`
Identical behaviour to `POST /images/authenticate`. Returns watermark channel data, similar images, and tamper score.

---

### API Key Format

```
map_<43-character base64url string>
```

**Key lifecycle:**
1. `GenerateAPIKey` produces the full key, stores only the bcrypt hash, and returns the raw key exactly once.
2. On each request, `ValidateAPIKey` extracts the first 12 characters as a prefix, performs a fast indexed DB lookup (`WHERE prefix = $1 AND is_active = TRUE AND (expires_at IS NULL OR expires_at > NOW())`), then bcrypt-compares.
3. `TouchAPIKeyLastUsed` is called asynchronously (goroutine) so it never adds latency to the authentication path.

---

### Authentication Guards

Two independent Fiber middleware guards are provided:

| Middleware | Header | Secret | Injects into locals |
|---|---|---|---|
| `EnterpriseJWTMiddleware` | `Authorization: Bearer <token>` | `ENTERPRISE_JWT_SECRET` | `enterpriseID`, `plan` |
| `EnterpriseAPIKeyMiddleware` | `X-API-Key: map_<key>` | bcrypt hash in DB | `enterpriseID`, `plan`, `apiKeyID` |

The enterprise JWT secret and the user JWT secret are cryptographically isolated — a user JWT cannot be used on enterprise routes and vice versa.

---

## Postman Testing

### Environment Setup

Create a Postman Environment called **MAP Backend** with these variables:

| Variable | Initial Value | Description |
|---|---|---|
| `base_url` | `http://localhost:5000/api/v1` | API root |
| `token` | *(empty)* | User JWT — auto-filled after Login |
| `user_id` | *(empty)* | Copy from Register response |
| `image_id` | *(empty)* | Copy from `X-Image-ID` header after Watermark |
| `serial_id` | *(empty)* | Copy from `X-Serial-ID` header after Watermark |

### Auto-capture JWT

Add this to the **Tests** tab of `POST /users/login`:

```javascript
const json = pm.response.json();
if (json.token) {
  pm.environment.set('token', json.token);
  console.log('Token saved:', json.token);
}
```

### Recommended End-to-End Test Flow

| # | Request | What to Verify |
|---|---|---|
| 1 | `GET /health` | Status 200 · `{ status: 'ok' }` |
| 2 | `POST /users/register` | Status 201 · copy `id` → `user_id` |
| 3 | `POST /users/login` | Status 200 · token auto-saved |
| 4 | `GET /users/me` | Status 200 · profile matches registration |
| 5 | `PUT /users/me` | Status 200 · updated fields reflected |
| 6 | `PUT /users/me/password` | Status 200 · confirm old password now returns 401 |
| 7 | `POST /images/watermark` | Status 200 binary · save `X-Image-ID` & `X-Serial-ID` |
| 8 | `POST /images/authenticate` | `tamper_score: 0.0` · `serial_id` matches step 7 |
| 9 | `POST /verify` (no token) | Identical result to step 8 |
| 10 | `POST /verify` (edited image) | `tamper_score > 0` or 404 if watermark destroyed |
| 11 | `DELETE /users/me` | Status 200 · subsequent login returns 403 |

> **Step 10 tip:** Open the watermarked image in any editor, apply a heavy crop or colour filter, save it, then upload. A moderate edit raises `tamper_score`; destroying the tile grid entirely causes 404.

### Common Mistakes

| Mistake | Fix |
|---|---|
| Sending image endpoints as raw JSON | Set Body to `form-data`; change `image` field type dropdown to **File** |
| Token not saved after login | Check the Tests script is present; re-run Login and inspect Postman Console |
| Uploading original file to Authenticate instead of watermarked output | Use *Send and Download* on the Watermark request first |
| 404 on `/verify` for a heavily edited image | The watermark tile may be destroyed — try a lightly compressed version |
| 409 on Embed Watermark | Image is already watermarked — use a fresh, never-watermarked source file |
| `{{token}}` not resolving | Ensure **MAP Backend** is selected in the environment dropdown (top-right) |

---

## HTTP Status Codes

| Status | Scenario | Typical Message |
|---|---|---|
| 200 | Successful request | Response body contains data |
| 201 | Register succeeded | User object returned |
| 400 | Missing field / invalid JSON | `"email and password are required"` |
| 401 | Wrong password / expired token | `"invalid email or password"` |
| 403 | Deactivated account login | `"account is deactivated"` |
| 404 | No watermark detected | `"no watermark detected in image"` |
| 409 | Email taken / image already watermarked | `"an account with this email already exists"` |
| 422 | CRC failure / corrupt payload | `"payload verification failed"` |
| 500 | Unexpected server or DB error | `"watermark embedding failed: ..."` |

---

## Technical Deep Dives

### DWT Coordinate Contract

Because DWT is critically sampled, coordinates differ from any RDWT-based system:

| Property | Value |
|---|---|
| HL sub-band size | 128×128 (half the 256×256 tile per axis) |
| HL block offsets | 0, 16, 32, …, 112 (8 positions per axis → 64 blocks) |
| Spatial window per block | 32×32 pixels at tile coordinates `(2·bx, 2·by)` |
| GetBlock call (embed) | `GetBlock(&tile, 2·bx, 2·by, 32)` → DWT → modify HL → IDWT → PutBlock |
| GetBlock call (extract) | `PartialDWT_HL(tile, bx, by)` — reads 32×32 pixels via closed form |
| Last valid block | `bx=112, by=112` → spatial `tile[224:256][224:256]` ✓ |

---

### PartialDWT_HL Optimisation

Rather than performing a full DWT on every 256×256 tile to read one 16×16 HL block during the duplicate-watermark pre-check and `Identify()` scan, `PartialDWT_HL` computes only the required block using a closed-form expression:

```
Row pass:  H[i][j]  = (tile[i][2j]   − tile[i][2j+1])   / √2
Col pass:  HL[i][j] = (H[2i][j]     + H[2i+1][j])       / √2
         = (tile[2i][2j] − tile[2i][2j+1] + tile[2i+1][2j] − tile[2i+1][2j+1]) × 0.5
```

Each HL coefficient depends on exactly **4 spatially adjacent pixels** in a 2×2 neighbourhood. The 16×16 HL block at `(bx, by)` is computed directly from the 32×32 spatial window `tile[2·by : 2·by+32][2·bx : 2·bx+32]`. **This is exact, not an approximation** — results are identical to a full `FastDWT2D` followed by reading `HL[by:by+16][bx:bx+16]`.

**Speedup:** ~34× per candidate tile vs. full DWT decomposition.

---

### Spatial Update Formula

The core embedding operation applies a pre-computed DCT basis image to modify the HL sub-band:

```
A′(x,y) = A(x,y) + k · B_uv(x,y)
```

| Symbol | Meaning |
|---|---|
| `A(x,y)` | Original HL coefficient at position (x,y) within the 16×16 block |
| `A′(x,y)` | Modified coefficient after watermark injection |
| `k` | Embedding strength scalar — controls robustness vs. imperceptibility |
| `B_uv(x,y)` | Pre-computed DCT basis image for frequency index (u,v) |

Using pre-computed basis values (baked into the binary via code generation) eliminates all trigonometric calculations at runtime. Per-block cost is O(N²) multiply-add operations — no full DCT decomposition needed.

**Coefficient pairs used:** `(1,3), (3,1), (2,3), (3,2), (3,4), (4,3)`

---

### FindValueOptimized

During extraction, the DCT coefficient value at frequency `(u,v)` for each 16×16 HL block is recovered via a dot product — the same basis used for embedding:

```
coeff_uv = Σ_{x,y} A(x,y) · B_uv(x,y)
```

| Property | Value |
|---|---|
| Input | 16×16 block of HL sub-band values (from `FastDWT2D` or `PartialDWT_HL`) |
| Output | Scalar coefficient at DCT frequency (u,v) |
| Complexity | O(N²) — 256 multiply-adds for a 16×16 block |
| Advantage | No full DCT decomposition needed; reads a single frequency component directly |

---

## Data Stores

### PostgreSQL — Structured Metadata

Stores the authoritative record of every watermarked image.

**Key tables:**

- `image_metadata` — `serial_id` (BIGSERIAL PK for watermark payload), `id` (UUID for Qdrant lookups), `owner_id`, `enterprise_id` (nullable FK), timestamp, dimensions, content description, embedding parameters (`k`, `u`, `v`)
- `enterprises` — company accounts
- `enterprise_api_keys` — bcrypt-hashed keys with 12-char prefix index; raw key never stored
- `enterprise_api_usage` — append-only billing audit log

> **Design note:** `serial_id` (BIGSERIAL) is used exclusively for watermark payload identification. The UUID `id` continues to serve Qdrant vector fingerprint lookups. This eliminates the lossy UUID↔uint64 conversion bug present in earlier versions.

### Qdrant — Perceptual Fingerprint Index

High-performance ANN index storing 256-D L2-normalised fingerprint vectors keyed to `metadata_id`. Supports sub-millisecond similarity search across millions of indexed images.

- **Similarity metric:** Cosine similarity
- **Match threshold:** >95%

---

## Key Design Decisions

**Why standard DWT instead of RDWT?**
RDWT's 4:1 overcomplete redundancy causes injected HL deltas to redistribute across neighbouring positions during inversion, yielding ~41% extraction failure on the round-trip. Standard DWT is critically sampled, so IDWT is an exact inverse and the HL modification is perfectly preserved.

**Why BIGSERIAL surrogate keys instead of UUID subsets?**
Attempting to fit a 128-bit UUID into a 64-bit integer payload requires lossy truncation. Using a dedicated `BIGSERIAL serial_id` as the watermark payload identifier eliminates this class of round-trip corruption bugs entirely.

**Why embed in the spatial domain rather than modifying RDWT coefficients directly?**
Embedding by modifying RDWT coefficients and inverting does not allow re-extraction via forward RDWT — the delta redistributes. Embedding must happen in the spatial domain using the DCT basis delta formula, with extraction via a forward DWT-only transform.

**Why pre-compute DCT basis matrices?**
Trigonometric functions are expensive and the same basis values are reused for every block of every tile. Code generation at build time bakes these as Go literals, giving O(1) runtime lookup with zero recomputation cost.

**Why isolate enterprise and user JWT secrets?**
A compromised user token must not grant access to enterprise management or billing routes. Separate signing secrets ensure the two token spaces are cryptographically independent.

---

*Media Authentication Platform — Built with Go, Fiber, PostgreSQL, and Qdrant.*
