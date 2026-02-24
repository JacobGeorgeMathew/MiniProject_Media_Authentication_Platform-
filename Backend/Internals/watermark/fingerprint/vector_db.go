package fingerprint

// qdrant.go — Vector similarity search via Qdrant
//
// Architecture:
//   PostgreSQL (db.go)  → users, admins, image_metadata
//   Qdrant (qdrant.go)  → image_fingerprints (1024-D vectors)
//
// The image UUID from PostgreSQL is used as the point ID in Qdrant,
// keeping both databases in sync.

import (
	"context"
	"fmt"
	"image"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

const (
	collectionName = "image_fingerprints"
	vectorSize     = 1024
)

// QdrantDB wraps the Qdrant client.
type QdrantDB struct {
	client *qdrant.Client
}

// NewQdrantDB connects to a running Qdrant instance.
// addr is the host:port of the gRPC endpoint, e.g. "localhost:6334"
func NewQdrantDB(ctx context.Context, addr string) (*QdrantDB, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to qdrant: %w", err)
	}
	return &QdrantDB{client: client}, nil
}

// Close shuts down the Qdrant client connection.
func (q *QdrantDB) Close() { q.client.Close() }

// CreateCollection sets up the vector collection in Qdrant.
// Call this once when setting up the database for the first time.
// It is safe to call multiple times — skips creation if already exists.
func (q *QdrantDB) CreateCollection(ctx context.Context) error {
	exists, err := q.client.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("check collection: %w", err)
	}
	if exists {
		return nil // already set up, nothing to do
	}

	err = q.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant.Distance_Cosine, // matches our DCT fingerprint comparison
		}),
	})
	if err != nil {
		return fmt.Errorf("create collection: %w", err)
	}
	return nil
}

// StoreFingerprint saves a 1024-D fingerprint vector to Qdrant.
// imageID is the UUID from PostgreSQL's image_metadata table — this is
// how we link the vector back to the full metadata.
func (q *QdrantDB) StoreFingerprint(ctx context.Context, imageID uuid.UUID, vec []float64) error {
	if len(vec) != vectorSize {
		return fmt.Errorf("fingerprint must be %d-dimensional, got %d", vectorSize, len(vec))
	}

	// Convert []float64 to []float32 (Qdrant uses float32 internally)
	vec32 := make([]float32, len(vec))
	for i, v := range vec {
		vec32[i] = float32(v)
	}

	_, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points: []*qdrant.PointStruct{
			{
				// Store the PostgreSQL UUID as the Qdrant point ID
				Id:      qdrant.NewIDUUID(imageID.String()),
				Vectors: qdrant.NewVectors(vec32...),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("store fingerprint in qdrant: %w", err)
	}
	return nil
}

// DeleteFingerprint removes a vector from Qdrant by image UUID.
// Call this when soft-deleting or permanently deleting an image.
func (q *QdrantDB) DeleteFingerprint(ctx context.Context, imageID uuid.UUID) error {
	_, err := q.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: collectionName,
		Points:         qdrant.NewPointsSelector(qdrant.NewIDUUID(imageID.String())),
	})
	return err
}

// FindSimilar generates a fingerprint from queryImg and returns the
// imageIDs of the k most similar images, ranked by cosine similarity.
// Use these IDs to fetch full metadata from PostgreSQL via db.go.
func (q *QdrantDB) FindSimilar(ctx context.Context, queryImg image.Image, k int) ([]uuid.UUID, []float32, error) {
	vec := Createfingerprint(queryImg)
	return q.FindSimilarByVector(ctx, vec, k)
}

// FindSimilarByVector performs nearest-neighbour search using a raw vector.
// Returns a slice of imageIDs (to fetch metadata from PostgreSQL) and
// their corresponding similarity scores (1.0 = identical, 0.0 = unrelated).
func (q *QdrantDB) FindSimilarByVector(ctx context.Context, vec []float64, k int) ([]uuid.UUID, []float32, error) {
	if len(vec) != vectorSize {
		return nil, nil, fmt.Errorf("query vector must be %d-dimensional, got %d", vectorSize, len(vec))
	}

	vec32 := make([]float32, len(vec))
	for i, v := range vec {
		vec32[i] = float32(v)
	}

	results, err := q.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQuery(vec32...),
		Limit:          qdrant.PtrOf(uint64(k)),
		WithPayload:    qdrant.NewWithPayload(false), // we only need IDs + scores
	})
	if err != nil {
		return nil, nil, fmt.Errorf("qdrant similarity search: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(results))
	scores := make([]float32, 0, len(results))

	for _, r := range results {
		uid, err := uuid.Parse(r.Id.GetUuid())
		if err != nil {
			continue
		}
		ids = append(ids, uid)
		scores = append(scores, r.Score)
	}

	return ids, scores, nil
}