package services

import (
	"context"
	"errors"
	"fmt"
	"image"

	"github.com/google/uuid"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/models"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/repository"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/watermark/engine"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/watermark/fingerprint"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/watermark/payload"
)

// ─────────────────────────────────────────────────────────────────────────────
// Request / Response types
// ─────────────────────────────────────────────────────────────────────────────

// EmbedRequest carries all caller-supplied metadata needed to embed a watermark.
type EmbedRequest struct {
	UserID        uuid.UUID
	Title         string
	Description   string // ← new: stored in image_metadata.description
	MimeType      string
	IsAIGenerated bool
}

// EmbedResult is returned by EmbedWatermarkInImage on success.
type EmbedResult struct {
	WatermarkedImage *image.YCbCr
	Fingerprint      []float64
	ImageID          uuid.UUID
	SerialID         int64
}

// AuthResult is returned by ImageAuth and VerifyImage.
type AuthResult struct {
	// Watermark channel — resolved from the extracted watermark payload.
	ExtractedMetadata *models.ImageMetadata

	// Fingerprint channel — resolved from pgvector similarity search.
	SimilarImages    []*models.ImageMetadata
	SimilarityScores []float64

	// TamperScore is 0 for a fully intact image; higher values indicate damage
	// or manipulation.
	TamperScore float64
}

// ─────────────────────────────────────────────────────────────────────────────
// ImageService
// ─────────────────────────────────────────────────────────────────────────────

type ImageService struct {
	repo repository.ImageRepository
}

func NewImageService(repo repository.ImageRepository) *ImageService {
	return &ImageService{repo: repo}
}

// ─────────────────────────────────────────────────────────────────────────────
// EmbedWatermarkInImage  (registered users only)
// ─────────────────────────────────────────────────────────────────────────────

// EmbedWatermarkInImage embeds an invisible watermark and persists both the
// metadata (including description) and the 256-D fingerprint vector.
//
// Pipeline:
//  1. Check for existing watermark.
//  2. Insert metadata + vector in one transaction → get serial_id.
//  3. Build binary payload from serial_id.
//  4. Embed watermark in frequency domain.
//  5. Generate fingerprint from watermarked image.
//  6. Update vector row with real fingerprint.
func (s *ImageService) EmbedWatermarkInImage(
	ctx context.Context,
	img image.Image,
	req EmbedRequest,
) (*EmbedResult, error) {

	////////////////////////////////////////////////////////////
	// 1️⃣  Check for existing watermark
	////////////////////////////////////////////////////////////

	coeffMatrices := defaultCoeffMatrices()

	_, _, alreadyWatermarked := engine.Identify(img, coeffMatrices)
	if alreadyWatermarked {
		return nil, errors.New("image is already watermarked")
	}

	////////////////////////////////////////////////////////////
	// 2️⃣  Persist metadata → get serial_id
	//      Pass a zero vector now; we update it after fingerprint generation.
	////////////////////////////////////////////////////////////

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Placeholder zero vector so the transaction succeeds immediately.
	zeroVec := make([]float64, 256)

	imageID, serialID, err := s.repo.InsertImageWithVector(
		ctx,
		req.UserID,
		req.Title,
		req.Description, // ← description forwarded to repo
		req.MimeType,
		w, h,
		req.IsAIGenerated,
		zeroVec,
	)
	if err != nil {
		return nil, fmt.Errorf("EmbedWatermarkInImage: initial insert: %w", err)
	}

	////////////////////////////////////////////////////////////
	// 3️⃣  Build payload from serial_id
	////////////////////////////////////////////////////////////

	payloadFields := payload.PayloadFields{
		Version:    1,
		IsAI:       req.IsAIGenerated,
		Reserved:   0,
		MetadataID: uint64(serialID),
	}

	payloadBits, err := payload.PayloadGenerate(payloadFields)
	if err != nil {
		return nil, fmt.Errorf("EmbedWatermarkInImage: payload generation: %w", err)
	}

	////////////////////////////////////////////////////////////
	// 4️⃣  Embed watermark in frequency domain
	////////////////////////////////////////////////////////////

	watermarkedImg, ok := engine.EmbedWatermark(img, payloadBits, coeffMatrices)
	if !ok {
		return nil, errors.New("EmbedWatermarkInImage: frequency-domain embedding failed")
	}

	////////////////////////////////////////////////////////////
	// 5️⃣  Generate 256-D perceptual fingerprint
	////////////////////////////////////////////////////////////

	fp := fingerprint.Createfingerprint(watermarkedImg)

	////////////////////////////////////////////////////////////
	// 6️⃣  Update the vector row with the real fingerprint
	//      We do this as a separate UPDATE rather than a second INSERT to avoid
	//      a duplicate metadata row.
	////////////////////////////////////////////////////////////

	if err := s.repo.UpdateImageVector(ctx, imageID, fp); err != nil {
		return nil, fmt.Errorf("EmbedWatermarkInImage: vector update: %w", err)
	}

	return &EmbedResult{
		WatermarkedImage: watermarkedImg,
		Fingerprint:      fp,
		ImageID:          imageID,
		SerialID:         serialID,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ImageAuth  (registered users — full dual-channel result)
// ─────────────────────────────────────────────────────────────────────────────

// ImageAuth authenticates a query image using the dual-layer strategy:
//   - Watermark channel: extract payload → CRC verify → serial_id → PostgreSQL.
//   - Fingerprint channel: generate 256-D vector → pgvector ANN search.
//
// k controls how many similar images are returned.
func (s *ImageService) ImageAuth(
	ctx context.Context,
	img image.Image,
	k int,
) (*AuthResult, error) {
	return s.runAuth(ctx, img, k)
}

// ─────────────────────────────────────────────────────────────────────────────
// VerifyImage  (public / unauthenticated users)
// ─────────────────────────────────────────────────────────────────────────────

// VerifyImage runs the same dual-channel authentication pipeline as ImageAuth
// but is designed for unauthenticated callers who want to look up ownership or
// provenance information about an image.
//
// The result is identical to AuthResult — callers receive:
//   - ExtractedMetadata : owner info, timestamp, title from the watermark payload.
//   - SimilarImages     : visually similar indexed images and their similarity scores.
//   - TamperScore       : composite integrity score.
func (s *ImageService) VerifyImage(
	ctx context.Context,
	img image.Image,
	k int,
) (*AuthResult, error) {
	return s.runAuth(ctx, img, k)
}

// ─────────────────────────────────────────────────────────────────────────────
// runAuth — shared authentication logic
// ─────────────────────────────────────────────────────────────────────────────

func (s *ImageService) runAuth(
	ctx context.Context,
	img image.Image,
	k int,
) (*AuthResult, error) {

	result := &AuthResult{}
	coeffMatrices := defaultCoeffMatrices()

	////////////////////////////////////////////////////////////
	// 1️⃣  Locate tile origin & verify watermark presence
	////////////////////////////////////////////////////////////

	_, _, hasWatermark := engine.Identify(img, coeffMatrices)
	if !hasWatermark {
		return nil, errors.New("no watermark detected in image")
	}

	////////////////////////////////////////////////////////////
	// 2️⃣  Extract watermark bits from all tiles
	////////////////////////////////////////////////////////////

	payloadCopies, ok := engine.ExtractWatermark(img, coeffMatrices)
	if !ok {
		return nil, errors.New("watermark extraction failed")
	}

	////////////////////////////////////////////////////////////
	// 3️⃣  Verify payload via majority vote + CRC
	////////////////////////////////////////////////////////////

	fields, err := payload.PayloadVerify(payloadCopies)
	if err != nil {
		return nil, fmt.Errorf("payload verification failed: %w", err)
	}

	////////////////////////////////////////////////////////////
	// 4️⃣  serial_id is embedded directly — no UUID conversion needed
	////////////////////////////////////////////////////////////

	serialID := int64(fields.MetadataID)

	////////////////////////////////////////////////////////////
	// 5️⃣  Fetch metadata from PostgreSQL via watermark channel
	////////////////////////////////////////////////////////////

	meta, err := s.repo.GetImageMetadataBySerialID(ctx, serialID)
	if err != nil {
		return nil, fmt.Errorf("metadata lookup (watermark channel): %w", err)
	}
	if meta == nil {
		return nil, errors.New("no metadata found for extracted serial_id")
	}

	result.ExtractedMetadata = meta

	////////////////////////////////////////////////////////////
	// 6️⃣  Generate fingerprint & find similar images via pgvector
	////////////////////////////////////////////////////////////

	fp := fingerprint.Createfingerprint(img)

	similarResults, err := s.repo.FindSimilarImages(ctx, fp)
	if err != nil {
		return nil, fmt.Errorf("vector similarity search: %w", err)
	}

	// Cap to k results.
	if k > 0 && len(similarResults) > k {
		similarResults = similarResults[:k]
	}

	////////////////////////////////////////////////////////////
	// 7️⃣  Unpack similarity results
	////////////////////////////////////////////////////////////

	var scores []float64
	var similarMetas []*models.ImageMetadata
	for i := range similarResults {
		scores = append(scores, similarResults[i].Similarity)
		m := similarResults[i].Metadata // copy to heap
		similarMetas = append(similarMetas, &m)
	}

	result.SimilarImages = similarMetas
	result.SimilarityScores = scores

	////////////////////////////////////////////////////////////
	// 8️⃣  Compute composite tamper score
	////////////////////////////////////////////////////////////

	result.TamperScore = computeTamperScore(meta, similarMetas, scores)

	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// computeTamperScore
// ─────────────────────────────────────────────────────────────────────────────

func computeTamperScore(
	watermarkMeta *models.ImageMetadata,
	similarMetas []*models.ImageMetadata,
	scores []float64,
) float64 {
	const (
		weightChannelMismatch  = 0.50
		weightLowSimilarity    = 0.25
		weightNoMatch          = 0.25
		weightDuplicateMatches = 0.10
	)

	if len(similarMetas) == 0 || len(scores) == 0 {
		return weightNoMatch + weightChannelMismatch
	}

	score := 0.0

	if watermarkMeta != nil && similarMetas[0] != nil {
		if watermarkMeta.ID != similarMetas[0].ID {
			score += weightChannelMismatch
		}
	}

	topSim := scores[0]
	if topSim < 1.0 {
		lowSimContrib := weightLowSimilarity * (1.0 - topSim) / 0.05
		if lowSimContrib > weightLowSimilarity {
			lowSimContrib = weightLowSimilarity
		}
		score += lowSimContrib
	}

	highMatchCount := 0
	for _, s := range scores {
		if s > 0.99 {
			highMatchCount++
		}
	}
	if highMatchCount > 1 {
		score += weightDuplicateMatches
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

// ─────────────────────────────────────────────────────────────────────────────
// defaultCoeffMatrices
// ─────────────────────────────────────────────────────────────────────────────

func defaultCoeffMatrices() []engine.Constants {
	return []engine.Constants{
		*engine.GetConstant(2, 3),
		*engine.GetConstant(3, 2),
	}
}
