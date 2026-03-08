// package handlers

// import (
// 	"bytes"
// 	"fmt"
// 	"image"
// 	_ "image/jpeg"
// 	_ "image/png"
// 	"strconv"
// 	"strings"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/google/uuid"

// 	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/api/utils"
// 	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/services"
// )

// // ─────────────────────────────────────────────────────────────────────────────
// // ImageHandler
// // ─────────────────────────────────────────────────────────────────────────────

// type ImageHandler struct {
// 	svc *services.ImageService
// }

// func NewImageHandler(svc *services.ImageService) *ImageHandler {
// 	return &ImageHandler{svc: svc}
// }

// // ─────────────────────────────────────────────────────────────────────────────
// // Wire types
// // ─────────────────────────────────────────────────────────────────────────────

// type similarImageItem struct {
// 	ImageID       string  `json:"image_id"`
// 	Title         *string `json:"title,omitempty"`
// 	Description   *string `json:"description,omitempty"`
// 	MimeType      string  `json:"mime_type"`
// 	Width         int     `json:"width_px"`
// 	Height        int     `json:"height_px"`
// 	IsAIGenerated bool    `json:"is_ai_generated"`
// 	Similarity    float64 `json:"similarity"`
// }

// type watermarkChannelInfo struct {
// 	ImageID       string  `json:"image_id"`
// 	SerialID      int64   `json:"serial_id"`
// 	Title         *string `json:"title,omitempty"`
// 	Description   *string `json:"description,omitempty"`
// 	MimeType      string  `json:"mime_type"`
// 	Width         int     `json:"width_px"`
// 	Height        int     `json:"height_px"`
// 	IsAIGenerated bool    `json:"is_ai_generated"`
// }

// type authResponse struct {
// 	WatermarkChannel *watermarkChannelInfo `json:"watermark_channel,omitempty"`
// 	SimilarImages    []similarImageItem    `json:"similar_images"`
// 	TamperScore      float64               `json:"tamper_score"`
// }

// // ─────────────────────────────────────────────────────────────────────────────
// // POST /api/v1/images/watermark   [protected — registered users only]
// // ─────────────────────────────────────────────────────────────────────────────

// // ImageWatermarkHandler embeds a watermark into the uploaded image and returns
// // the watermarked binary alongside metadata headers.
// //
// // Multipart form fields:
// //
// //	image           — image file (JPEG / PNG / TIFF / BMP / WebP)
// //	title           — human-readable label           (optional, defaults to filename)
// //	description     — free-text description          (optional)
// //	is_ai_generated — "true" / "false"               (optional, default false)
// func (h *ImageHandler) ImageWatermarkHandler(c *fiber.Ctx) error {

// 	// ── Parse image file ─────────────────────────────────────────────────────
// 	fileHeader, err := c.FormFile("image")
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "missing 'image' field in form",
// 		})
// 	}

// 	file, err := fileHeader.Open()
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "failed to open uploaded file",
// 		})
// 	}
// 	defer file.Close()

// 	img, _, err := image.Decode(file)
// 	if err != nil {
// 		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
// 			"error": "could not decode image: " + err.Error(),
// 		})
// 	}

// 	// ── Form fields ──────────────────────────────────────────────────────────
// 	title := c.FormValue("title", fileHeader.Filename)
// 	description := c.FormValue("description", "") // ← new field
// 	mimeType := fileHeader.Header.Get("Content-Type")
// 	if mimeType == "" {
// 		mimeType = "application/octet-stream"
// 	}
// 	isAI := strings.EqualFold(c.FormValue("is_ai_generated", "false"), "true")

// 	// ── User ID from auth middleware ─────────────────────────────────────────
// 	userID, err := userIDFromLocals(c)
// 	if err != nil {
// 		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 			"error": "missing or invalid authentication",
// 		})
// 	}

// 	// ── Call service ─────────────────────────────────────────────────────────
// 	result, err := h.svc.EmbedWatermarkInImage(c.Context(), img, services.EmbedRequest{
// 		UserID:        userID,
// 		Title:         title,
// 		Description:   description, // ← forwarded
// 		MimeType:      mimeType,
// 		IsAIGenerated: isAI,
// 	})
// 	if err != nil {
// 		if strings.Contains(err.Error(), "already watermarked") {
// 			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
// 		}
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "watermark embedding failed: " + err.Error(),
// 		})
// 	}

// 	// ── Encode watermarked image and stream back ──────────────────────────────
// 	var buf bytes.Buffer
// 	if err := utils.EncodeImageToWriter(&buf, result.WatermarkedImage, mimeType); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "failed to encode watermarked image",
// 		})
// 	}

// 	c.Set("X-Image-ID", result.ImageID.String())
// 	c.Set("X-Serial-ID", strconv.FormatInt(result.SerialID, 10))
// 	c.Set("Content-Type", mimeType)
// 	return c.Status(fiber.StatusOK).Send(buf.Bytes())
// }

// // ─────────────────────────────────────────────────────────────────────────────
// // POST /api/v1/images/authenticate   [protected — registered users only]
// // ─────────────────────────────────────────────────────────────────────────────

// // ImageAuthHandler authenticates an uploaded image using the dual-layer
// // watermark + fingerprint strategy. Returns ownership info, similar images,
// // and a tamper score.
// //
// // Multipart form fields:
// //
// //	image — image file to authenticate
// //	k     — max similar images to return (optional, default 5)
// func (h *ImageHandler) ImageAuthHandler(c *fiber.Ctx) error {
// 	img, k, err := parseImageAndK(c)
// 	if err != nil {
// 		return err // response already written inside parseImageAndK
// 	}

// 	result, err := h.svc.ImageAuth(c.Context(), img, k)
// 	if err != nil {
// 		return authErrorResponse(c, err)
// 	}

// 	return c.Status(fiber.StatusOK).JSON(buildAuthResponse(result))
// }

// // ─────────────────────────────────────────────────────────────────────────────
// // POST /api/v1/verify   [public — no authentication required]
// // ─────────────────────────────────────────────────────────────────────────────

// // ImageVerifyHandler allows anyone — even unregistered users — to verify the
// // provenance of an image. It runs the same dual-channel watermark + fingerprint
// // pipeline and returns ownership/tamper information without requiring a JWT.
// //
// // Multipart form fields:
// //
// //	image — image file to verify
// //	k     — max similar images to return (optional, default 5)
// func (h *ImageHandler) ImageVerifyHandler(c *fiber.Ctx) error {
// 	img, k, err := parseImageAndK(c)
// 	if err != nil {
// 		return err
// 	}

// 	result, err := h.svc.VerifyImage(c.Context(), img, k)
// 	if err != nil {
// 		return authErrorResponse(c, err)
// 	}

// 	return c.Status(fiber.StatusOK).JSON(buildAuthResponse(result))
// }

// // ─────────────────────────────────────────────────────────────────────────────
// // shared helpers
// // ─────────────────────────────────────────────────────────────────────────────

// // parseImageAndK decodes the "image" form-file and optional "k" parameter.
// // On error it writes the error response itself and returns a non-nil error so
// // the caller can return immediately.
// func parseImageAndK(c *fiber.Ctx) (image.Image, int, error) {
// 	fileHeader, err := c.FormFile("image")
// 	if err != nil {
// 		_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "missing 'image' field in form",
// 		})
// 		return nil, 0, err
// 	}

// 	file, err := fileHeader.Open()
// 	if err != nil {
// 		_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "failed to open uploaded file",
// 		})
// 		return nil, 0, err
// 	}
// 	defer file.Close()

// 	img, _, err := image.Decode(file)
// 	if err != nil {
// 		_ = c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
// 			"error": "could not decode image: " + err.Error(),
// 		})
// 		return nil, 0, err
// 	}

// 	k := 5
// 	if raw := c.FormValue("k", ""); raw != "" {
// 		if parsed, parseErr := strconv.Atoi(raw); parseErr == nil && parsed > 0 {
// 			k = parsed
// 		}
// 	}

// 	return img, k, nil
// }

// // authErrorResponse maps service errors to appropriate HTTP statuses.
// func authErrorResponse(c *fiber.Ctx, err error) error {
// 	msg := err.Error()
// 	switch {
// 	case strings.Contains(msg, "no watermark detected"):
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": msg})
// 	case strings.Contains(msg, "payload verification failed"):
// 		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": msg})
// 	default:
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "image verification failed: " + msg,
// 		})
// 	}
// }

// // buildAuthResponse converts an *AuthResult into the JSON wire type.
// func buildAuthResponse(result *services.AuthResult) authResponse {
// 	resp := authResponse{
// 		TamperScore:   result.TamperScore,
// 		SimilarImages: make([]similarImageItem, 0, len(result.SimilarImages)),
// 	}

// 	if result.ExtractedMetadata != nil {
// 		m := result.ExtractedMetadata
// 		resp.WatermarkChannel = &watermarkChannelInfo{
// 			ImageID:       m.ID.String(),
// 			SerialID:      m.SerialID,
// 			Title:         m.Title,
// 			Description:   m.Description,
// 			MimeType:      m.MimeType,
// 			Width:         m.WidthPx,
// 			Height:        m.HeightPx,
// 			IsAIGenerated: m.IsAIGenerated,
// 		}
// 	}

// 	for i, m := range result.SimilarImages {
// 		var sim float64
// 		if i < len(result.SimilarityScores) {
// 			sim = result.SimilarityScores[i]
// 		}
// 		resp.SimilarImages = append(resp.SimilarImages, similarImageItem{
// 			ImageID:       m.ID.String(),
// 			Title:         m.Title,
// 			Description:   m.Description,
// 			MimeType:      m.MimeType,
// 			Width:         m.WidthPx,
// 			Height:        m.HeightPx,
// 			IsAIGenerated: m.IsAIGenerated,
// 			Similarity:    sim,
// 		})
// 	}

// 	return resp
// }

// // ─────────────────────────────────────────────────────────────────────────────
// // Context helper
// // ─────────────────────────────────────────────────────────────────────────────

// // userIDFromLocals reads the UUID injected by the auth middleware under "userID".
// func userIDFromLocals(c *fiber.Ctx) (uuid.UUID, error) {
// 	raw := c.Locals("userID")
// 	if raw == nil {
// 		return uuid.Nil, fmt.Errorf("userID not found in locals")
// 	}
// 	switch v := raw.(type) {
// 	case uuid.UUID:
// 		return v, nil
// 	case string:
// 		return uuid.Parse(v)
// 	default:
// 		return uuid.Nil, fmt.Errorf("unexpected userID type in locals: %T", raw)
// 	}
// }

package handlers

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/api/utils"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/services"
)

// ─────────────────────────────────────────────────────────────────────────────
// ImageHandler
// ─────────────────────────────────────────────────────────────────────────────

type ImageHandler struct {
	svc *services.ImageService
}
//hello
func NewImageHandler(svc *services.ImageService) *ImageHandler {
	return &ImageHandler{svc: svc}
}

// ─────────────────────────────────────────────────────────────────────────────
// JSON wire types
// ─────────────────────────────────────────────────────────────────────────────

type similarImageItem struct {
	ImageID       string  `json:"image_id"`
	Title         *string `json:"title,omitempty"`
	Description   *string `json:"description,omitempty"`
	MimeType      string  `json:"mime_type"`
	Width         int     `json:"width_px"`
	Height        int     `json:"height_px"`
	IsAIGenerated bool    `json:"is_ai_generated"`
	Similarity    float64 `json:"similarity"`
}

type watermarkChannelInfo struct {
	ImageID       string  `json:"image_id"`
	SerialID      int64   `json:"serial_id"`
	Title         *string `json:"title,omitempty"`
	Description   *string `json:"description,omitempty"`
	MimeType      string  `json:"mime_type"`
	Width         int     `json:"width_px"`
	Height        int     `json:"height_px"`
	IsAIGenerated bool    `json:"is_ai_generated"`
}

// authResponse is the unified JSON body for both /authenticate and /verify.
//
//   - has_watermark       — tells the client whether a valid watermark was found.
//   - watermark_message   — human-readable explanation of the watermark channel
//     outcome; always present, regardless of has_watermark.
//   - watermark_channel   — non-null only when has_watermark is true.
//   - similar_images      — always present (may be empty).
//   - tamper_score        — only meaningful when has_watermark is true; 0 otherwise.
type authResponse struct {
	HasWatermark     bool                  `json:"has_watermark"`
	WatermarkMessage string                `json:"watermark_message"`
	WatermarkChannel *watermarkChannelInfo `json:"watermark_channel,omitempty"`
	SimilarImages    []similarImageItem    `json:"similar_images"`
	TamperScore      float64               `json:"tamper_score"`
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/images/watermark   [protected]
// ─────────────────────────────────────────────────────────────────────────────

// ImageWatermarkHandler embeds a watermark into the uploaded image and returns
// the watermarked binary alongside metadata headers.
//
// Multipart form fields:
//
//	image           — image file (JPEG / PNG / TIFF / BMP / WebP)
//	title           — human-readable label           (optional, defaults to filename)
//	description     — free-text description          (optional)
//	is_ai_generated — "true" / "false"               (optional, default false)
func (h *ImageHandler) ImageWatermarkHandler(c *fiber.Ctx) error {

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing 'image' field in form",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to open uploaded file",
		})
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "could not decode image: " + err.Error(),
		})
	}

	title := c.FormValue("title", fileHeader.Filename)
	description := c.FormValue("description", "")
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	isAI := strings.EqualFold(c.FormValue("is_ai_generated", "false"), "true")

	userID, err := userIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing or invalid authentication",
		})
	}

	result, err := h.svc.EmbedWatermarkInImage(c.Context(), img, services.EmbedRequest{
		UserID:        userID,
		Title:         title,
		Description:   description,
		MimeType:      mimeType,
		IsAIGenerated: isAI,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already watermarked") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "watermark embedding failed: " + err.Error(),
		})
	}

	var buf bytes.Buffer
	if err := utils.EncodeImageToWriter(&buf, result.WatermarkedImage, mimeType); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to encode watermarked image",
		})
	}

	c.Set("X-Image-ID", result.ImageID.String())
	c.Set("X-Serial-ID", strconv.FormatInt(result.SerialID, 10))
	c.Set("Content-Type", mimeType)
	return c.Status(fiber.StatusOK).Send(buf.Bytes())
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/images/authenticate   [protected]
// ─────────────────────────────────────────────────────────────────────────────

// ImageAuthHandler authenticates an uploaded image. Always returns HTTP 200
// with a structured body — the has_watermark flag and watermark_message fields
// tell the client what the watermark channel found.
//
// Multipart form fields:
//
//	image — image file to authenticate
//	k     — max similar images to return (optional, default 5)
func (h *ImageHandler) ImageAuthHandler(c *fiber.Ctx) error {
	img, k, err := parseImageAndK(c)
	if err != nil {
		return err
	}

	result, err := h.svc.ImageAuth(c.Context(), img, k)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "authentication failed: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(buildAuthResponse(result))
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/verify   [public — no JWT required]
// ─────────────────────────────────────────────────────────────────────────────

// ImageVerifyHandler allows unauthenticated users to verify image provenance.
// Runs the same dual-channel pipeline as ImageAuthHandler.
//
// Multipart form fields:
//
//	image — image file to verify
//	k     — max similar images to return (optional, default 5)
func (h *ImageHandler) ImageVerifyHandler(c *fiber.Ctx) error {
	img, k, err := parseImageAndK(c)
	if err != nil {
		return err
	}

	result, err := h.svc.VerifyImage(c.Context(), img, k)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "image verification failed: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(buildAuthResponse(result))
}

// ─────────────────────────────────────────────────────────────────────────────
// shared helpers
// ─────────────────────────────────────────────────────────────────────────────

// parseImageAndK decodes the "image" multipart file and optional "k" parameter.
// On parse failure it writes the error response and returns a non-nil error so
// the caller can return immediately.
func parseImageAndK(c *fiber.Ctx) (image.Image, int, error) {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing 'image' field in form",
		})
		return nil, 0, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to open uploaded file",
		})
		return nil, 0, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		_ = c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "could not decode image: " + err.Error(),
		})
		return nil, 0, err
	}

	k := 5
	if raw := c.FormValue("k", ""); raw != "" {
		if parsed, parseErr := strconv.Atoi(raw); parseErr == nil && parsed > 0 {
			k = parsed
		}
	}

	return img, k, nil
}

// buildAuthResponse converts an *AuthResult into the JSON wire type.
// Both channels' outcomes are always represented — the caller never needs
// to handle nil results or inspect error codes for business-logic branches.
func buildAuthResponse(result *services.AuthResult) authResponse {
	resp := authResponse{
		HasWatermark:     result.HasWatermark,
		WatermarkMessage: result.WatermarkMessage,
		TamperScore:      result.TamperScore,
		SimilarImages:    make([]similarImageItem, 0, len(result.SimilarImages)),
	}

	// Watermark channel — only populated when HasWatermark is true.
	if result.HasWatermark && result.ExtractedMetadata != nil {
		m := result.ExtractedMetadata
		resp.WatermarkChannel = &watermarkChannelInfo{
			ImageID:       m.ID.String(),
			SerialID:      m.SerialID,
			Title:         m.Title,
			Description:   m.Description,
			MimeType:      m.MimeType,
			Width:         m.WidthPx,
			Height:        m.HeightPx,
			IsAIGenerated: m.IsAIGenerated,
		}
	}

	// Fingerprint channel — always populated (may be empty slice).
	for i, m := range result.SimilarImages {
		var sim float64
		if i < len(result.SimilarityScores) {
			sim = result.SimilarityScores[i]
		}
		resp.SimilarImages = append(resp.SimilarImages, similarImageItem{
			ImageID:       m.ID.String(),
			Title:         m.Title,
			Description:   m.Description,
			MimeType:      m.MimeType,
			Width:         m.WidthPx,
			Height:        m.HeightPx,
			IsAIGenerated: m.IsAIGenerated,
			Similarity:    sim,
		})
	}

	return resp
}

// ─────────────────────────────────────────────────────────────────────────────
// Context helper
// ─────────────────────────────────────────────────────────────────────────────

func userIDFromLocals(c *fiber.Ctx) (uuid.UUID, error) {
	raw := c.Locals("userID")
	if raw == nil {
		return uuid.Nil, fmt.Errorf("userID not found in locals")
	}
	switch v := raw.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Nil, fmt.Errorf("unexpected userID type in locals: %T", raw)
	}
}
