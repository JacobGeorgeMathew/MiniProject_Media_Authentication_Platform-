package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"

	//"image/jpeg"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/services"
)

// ImageHandler holds a reference to the service layer
type ImageHandler struct {
	imageService *services.ImageService
}

// NewImageHandler creates a new ImageHandler
func NewImageHandler(imageService *services.ImageService) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
	}
}

// -----------------------------------------------------------------------
// REQUEST STRUCTS
// -----------------------------------------------------------------------

// EmbedMetadata is the JSON structure expected in the "metadata" form field
// when calling the watermark endpoint.
type EmbedMetadata struct {
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	IsAIGenerated bool    `json:"is_ai_generated"`
	// CapturedAt is optional; expected as RFC3339 string e.g. "2024-01-15T10:30:00Z"
	CapturedAt *string `json:"captured_at"`
}

// -----------------------------------------------------------------------
// RESPONSE STRUCTS
// -----------------------------------------------------------------------

// WatermarkResponse is returned when watermarking fails (JSON error body).
// On success the handler streams the image directly with a custom header
// carrying the base64-encoded fingerprint.
type errorResponse struct {
	Error string `json:"error"`
}

// -----------------------------------------------------------------------
// HANDLER 1 — Embed watermark
// -----------------------------------------------------------------------
//
// Expects multipart/form-data with:
//   - "image"    → image file  (JPEG or PNG)
//   - "metadata" → JSON string (EmbedMetadata)
//
// Returns:
//   - The watermarked image as the response body (same format as input)
//   - Header  X-Fingerprint: <comma-separated float64 values>
//   - Header  X-Image-ID:    <uuid of the stored metadata record>

func (h *ImageHandler) ImageWatermarkHandler(c *fiber.Ctx) error {

	// ── 1. Receive image ──────────────────────────────────────────────
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
			Error: "field 'image' is required (multipart/form-data)",
		})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse{
			Error: "could not open uploaded image",
		})
	}
	defer src.Close()

	imgBytes, err := io.ReadAll(src)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse{
			Error: "could not read uploaded image",
		})
	}

	img, format, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
			Error: "invalid image file: " + err.Error(),
		})
	}

	// ── 2. Receive & parse metadata ───────────────────────────────────
	metaStr := c.FormValue("metadata")
	if metaStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
			Error: "field 'metadata' is required (JSON string)",
		})
	}

	var embedMeta EmbedMetadata
	if err := json.Unmarshal([]byte(metaStr), &embedMeta); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
			Error: "invalid metadata JSON: " + err.Error(),
		})
	}

	// ── 3. Resolve MIME type ──────────────────────────────────────────
	mimeType := "image/" + format
	if format == "jpg" {
		mimeType = "image/jpeg"
		format = "jpeg"
	}

	// ── 4. Parse optional CapturedAt timestamp ────────────────────────
	var capturedAt *time.Time
	if embedMeta.CapturedAt != nil && *embedMeta.CapturedAt != "" {
		t, err := time.Parse(time.RFC3339, *embedMeta.CapturedAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
				Error: "invalid captured_at format, expected RFC3339 (e.g. 2024-01-15T10:30:00Z)",
			})
		}
		capturedAt = &t
	}

	// ── 5. Build service request ──────────────────────────────────────
	serviceReq := services.EmbedRequest{
		Title:         embedMeta.Title,
		Description:   embedMeta.Description,
		MimeType:      &mimeType,
		IsAIGenerated: embedMeta.IsAIGenerated,
		CapturedAt:    capturedAt,
	}

	// ── 6. Call service ───────────────────────────────────────────────
	watermarkedImg, fingerprint, err := h.imageService.EmbedWatermarkInImage(c.Context(), img, serviceReq)
	if err != nil {
		// "already watermarked" is a 409 Conflict, everything else is 500
		status := fiber.StatusInternalServerError
		if err.Error() == "image is already watermarked" {
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(errorResponse{Error: err.Error()})
	}

	// ── 7. Encode watermarked image into memory buffer ────────────────
	var buf bytes.Buffer

	// switch format {
	// case "jpeg":
	// 	err = jpeg.Encode(&buf, watermarkedImg, &jpeg.Options{Quality: 92})
	// default: // png and everything else
	// 	err = png.Encode(&buf, watermarkedImg)
	// 	format = "png"
	// }
	if true {
	err = png.Encode(&buf, watermarkedImg)
	format = "png"
	} else {
		err = jpeg.Encode(&buf, watermarkedImg, &jpeg.Options{Quality: 92})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse{
			Error: "failed to encode watermarked image: " + err.Error(),
		})
	}

	// ── 8. Attach fingerprint as response header ──────────────────────
	// Serialise []float64 → JSON array and put it in X-Fingerprint header.
	fpJSON, err := json.Marshal(fingerprint)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse{
			Error: "failed to serialise fingerprint",
		})
	}

	// ── 9. Stream image back to client ────────────────────────────────
	c.Set(fiber.HeaderContentType, "image/"+format)
	c.Set("Content-Disposition", "attachment; filename=watermarked."+format)
	c.Set("X-Fingerprint", string(fpJSON)) // e.g. [0.12,0.98, ...]

	return c.Send(buf.Bytes())
}

// -----------------------------------------------------------------------
// HANDLER 2 — Authenticate image
// -----------------------------------------------------------------------
//
// Expects multipart/form-data with:
//   - "image" → image file (JPEG or PNG)
//   - "k"     → optional integer string, number of similar images to return
//               (defaults to 5 if omitted)
//
// Returns JSON:
//
//	{
//	  "watermark_valid": true,
//	  "extracted_metadata": { ...models.ImageMetadata fields... },
//	  "similar_images": [ { ...models.ImageMetadata... }, ... ],
//	  "similarity_scores": [0.98, 0.94, ...]
//	}

func (h *ImageHandler) ImageAuthHandler(c *fiber.Ctx) error {

	// ── 1. Receive image ──────────────────────────────────────────────
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
			Error: "field 'image' is required (multipart/form-data)",
		})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse{
			Error: "could not open uploaded image",
		})
	}
	defer src.Close()

	imgBytes, err := io.ReadAll(src)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse{
			Error: "could not read uploaded image",
		})
	}

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
			Error: "invalid image file: " + err.Error(),
		})
	}

	// ── 2. Parse optional k (number of similar images) ────────────────
	k := 5 // sensible default
	if kStr := strings.TrimSpace(c.FormValue("k")); kStr != "" {
		if _, err := fmt.Sscanf(kStr, "%d", &k); err != nil || k < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
				Error: "'k' must be a positive integer",
			})
		}
	}

	// ── 3. Call service ───────────────────────────────────────────────
	authResult, err := h.imageService.ImageAuth(c.Context(), img, k)
	if err != nil {
		// Distinguish "no watermark" (404) from other failures (500)
		status := fiber.StatusInternalServerError
		msg := err.Error()

		switch msg {
		case "no watermark detected in image":
			status = fiber.StatusNotFound
		case "failed to extract watermark",
			"metadata not found for extracted watermark ID":
			status = fiber.StatusUnprocessableEntity
		}

		return c.Status(status).JSON(errorResponse{Error: msg})
	}

	// ── 4. Return structured JSON result ──────────────────────────────
	return c.Status(fiber.StatusOK).JSON(authResult)
}
