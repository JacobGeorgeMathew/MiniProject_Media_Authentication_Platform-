package handlers

import (
	"bytes"
	"context"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/api/middleware"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/api/utils"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/models"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/services"
)

// ─────────────────────────────────────────────────────────────────────────────
// EnterpriseHandler
//
// Two distinct concerns, same struct:
//
//   (a) Dashboard / management routes  — authenticated with enterprise JWT
//       POST   /api/v1/enterprise/register
//       POST   /api/v1/enterprise/login
//       GET    /api/v1/enterprise/me
//       PUT    /api/v1/enterprise/me
//       PUT    /api/v1/enterprise/me/password
//       DELETE /api/v1/enterprise/me
//       POST   /api/v1/enterprise/keys
//       GET    /api/v1/enterprise/keys
//       DELETE /api/v1/enterprise/keys/:keyId
//       GET    /api/v1/enterprise/usage
//
//   (b) Watermarking-as-a-Service routes — authenticated with X-API-Key
//       POST   /api/v1/enterprise/watermark
//       POST   /api/v1/enterprise/authenticate
// ─────────────────────────────────────────────────────────────────────────────

type EnterpriseHandler struct {
	entSvc *services.EnterpriseService
	imgSvc *services.ImageService
}

func NewEnterpriseHandler(
	entSvc *services.EnterpriseService,
	imgSvc *services.ImageService,
) *EnterpriseHandler {
	return &EnterpriseHandler{entSvc: entSvc, imgSvc: imgSvc}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/enterprise/register   [public]
// ─────────────────────────────────────────────────────────────────────────────

func (h *EnterpriseHandler) Register(c *fiber.Ctx) error {
	var body struct {
		CompanyName string `json:"company_name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		Website     string `json:"website"`
		Plan        string `json:"plan"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON: " + err.Error(),
		})
	}
	if body.CompanyName == "" || body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "company_name, email, and password are required",
		})
	}

	resp, err := h.entSvc.Register(c.Context(), services.EnterpriseRegisterRequest{
		CompanyName: body.CompanyName,
		Email:       body.Email,
		Password:    body.Password,
		Website:     body.Website,
		Plan:        body.Plan,
	})
	if err != nil {
		return enterpriseErrResponse(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/enterprise/login   [public]
// ─────────────────────────────────────────────────────────────────────────────

func (h *EnterpriseHandler) Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON: " + err.Error(),
		})
	}
	if body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email and password are required",
		})
	}

	resp, err := h.entSvc.Login(c.Context(), services.EnterpriseLoginRequest{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		return enterpriseErrResponse(c, err)
	}

	token, err := middleware.GenerateEnterpriseToken(resp.ID, resp.Plan)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate token",
		})
	}
	resp.Token = token

	return c.Status(fiber.StatusOK).JSON(resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/enterprise/me   [enterprise JWT]
// ─────────────────────────────────────────────────────────────────────────────

func (h *EnterpriseHandler) GetMe(c *fiber.Ctx) error {
	id, err := middleware.EnterpriseIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	resp, err := h.entSvc.GetByID(c.Context(), id)
	if err != nil {
		return enterpriseErrResponse(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /api/v1/enterprise/me   [enterprise JWT]
// ─────────────────────────────────────────────────────────────────────────────

func (h *EnterpriseHandler) UpdateProfile(c *fiber.Ctx) error {
	id, err := middleware.EnterpriseIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var body struct {
		CompanyName string `json:"company_name"`
		Email       string `json:"email"`
		Website     string `json:"website"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON: " + err.Error()})
	}
	if body.CompanyName == "" && body.Email == "" && body.Website == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "at least one field (company_name, email, website) must be provided",
		})
	}

	resp, err := h.entSvc.UpdateProfile(c.Context(), services.EnterpriseUpdateRequest{
		EnterpriseID: id,
		CompanyName:  body.CompanyName,
		Email:        body.Email,
		Website:      body.Website,
	})
	if err != nil {
		return enterpriseErrResponse(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /api/v1/enterprise/me/password   [enterprise JWT]
// ─────────────────────────────────────────────────────────────────────────────

func (h *EnterpriseHandler) ChangePassword(c *fiber.Ctx) error {
	id, err := middleware.EnterpriseIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON: " + err.Error()})
	}
	if body.OldPassword == "" || body.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "old_password and new_password are required",
		})
	}

	if err := h.entSvc.ChangePassword(c.Context(), services.EnterpriseChangePasswordRequest{
		EnterpriseID: id,
		OldPassword:  body.OldPassword,
		NewPassword:  body.NewPassword,
	}); err != nil {
		return enterpriseErrResponse(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "password updated successfully"})
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /api/v1/enterprise/me   [enterprise JWT]
// ─────────────────────────────────────────────────────────────────────────────

func (h *EnterpriseHandler) Deactivate(c *fiber.Ctx) error {
	id, err := middleware.EnterpriseIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.entSvc.Deactivate(c.Context(), id); err != nil {
		return enterpriseErrResponse(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "enterprise account deactivated successfully"})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/enterprise/keys   [enterprise JWT]
// ─────────────────────────────────────────────────────────────────────────────

// CreateAPIKey generates a new API key.
// The raw_key in the response is shown ONCE and never again — the enterprise
// must save it immediately.
func (h *EnterpriseHandler) CreateAPIKey(c *fiber.Ctx) error {
	id, err := middleware.EnterpriseIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var body struct {
		Label     string `json:"label"`
		ExpiresAt string `json:"expires_at"` // RFC3339, optional
	}
	_ = c.BodyParser(&body) // body is fully optional

	var expiresAt *time.Time
	if body.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, body.ExpiresAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "expires_at must be RFC3339 format, e.g. 2026-12-31T23:59:59Z",
			})
		}
		expiresAt = &t
	}

	result, err := h.entSvc.GenerateAPIKey(c.Context(), services.CreateAPIKeyRequest{
		EnterpriseID: id,
		Label:        body.Label,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return enterpriseErrResponse(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/enterprise/keys   [enterprise JWT]
// ─────────────────────────────────────────────────────────────────────────────

// ListAPIKeys returns redacted key records (no raw key, no hash).
func (h *EnterpriseHandler) ListAPIKeys(c *fiber.Ctx) error {
	id, err := middleware.EnterpriseIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	keys, err := h.entSvc.ListAPIKeys(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list API keys"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"keys": keys})
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /api/v1/enterprise/keys/:keyId   [enterprise JWT]
// ─────────────────────────────────────────────────────────────────────────────

// RevokeAPIKey deactivates a key by its UUID.
func (h *EnterpriseHandler) RevokeAPIKey(c *fiber.Ctx) error {
	id, err := middleware.EnterpriseIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	keyID, err := uuid.Parse(c.Params("keyId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid key ID format"})
	}

	if err := h.entSvc.RevokeAPIKey(c.Context(), id, keyID); err != nil {
		return enterpriseErrResponse(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "API key revoked successfully"})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/enterprise/usage?from=2025-01-01&to=2025-12-31   [enterprise JWT]
// ─────────────────────────────────────────────────────────────────────────────

// GetUsage returns per-endpoint call counts for the requested date window.
func (h *EnterpriseHandler) GetUsage(c *fiber.Ctx) error {
	id, err := middleware.EnterpriseIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	from, err := time.Parse("2006-01-02", c.Query("from", ""))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "from query param is required — use YYYY-MM-DD format",
		})
	}
	to, err := time.Parse("2006-01-02", c.Query("to", ""))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "to query param is required — use YYYY-MM-DD format",
		})
	}
	// Make 'to' inclusive by bumping to end of that day.
	to = to.Add(24*time.Hour - time.Nanosecond)

	rows, err := h.entSvc.GetUsageSummary(c.Context(), id, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch usage data"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"from":  from.Format("2006-01-02"),
		"to":    c.Query("to"),
		"usage": rows,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/enterprise/watermark   [X-API-Key]
// ─────────────────────────────────────────────────────────────────────────────

// WatermarkImage is the Watermarking-as-a-Service endpoint.
// Enterprises POST a multipart form with their image; the platform embeds a
// watermark, records metadata, and returns the watermarked binary.
//
// Form fields:
//
//	image           — image file (JPEG / PNG / TIFF / BMP / WebP)   REQUIRED
//	title           — label stored in image_metadata                 optional
//	description     — free-text stored in image_metadata             optional
//	is_ai_generated — "true" / "false"                               optional
//
// Response:
//
//	Body              — binary watermarked image
//	X-Image-ID        — UUID of the new image_metadata row
//	X-Serial-ID       — BIGSERIAL embedded in the watermark payload
//	Content-Type      — mirrors the uploaded MIME type
func (h *EnterpriseHandler) WatermarkImage(c *fiber.Ctx) error {
	// Capture start time immediately so latency includes all processing below.
	startTime := time.Now()

	enterpriseID, err := middleware.EnterpriseIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing 'image' field in form"})
	}
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open uploaded file"})
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "could not decode image: " + err.Error(),
		})
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	isAI := strings.EqualFold(c.FormValue("is_ai_generated", "false"), "true")

	result, err := h.imgSvc.EmbedWatermarkInImage(c.Context(), img, services.EmbedRequest{
		UserID:        enterpriseID, // enterprise owns the image
		Title:         c.FormValue("title", fileHeader.Filename),
		Description:   c.FormValue("description", ""),
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

	// Snapshot the values needed for the usage record before the goroutine,
	// since the Fiber context (c) must not be accessed after the handler returns.
	latencyMs := int(time.Since(startTime).Milliseconds())
	apiKeyID, _ := c.Locals("apiKeyID").(uuid.UUID)
	statusCode := fiber.StatusOK

	// Record usage asynchronously — never block the response.
	go func() {
		usageRecord := models.EnterpriseAPIUsage{
			EnterpriseID: enterpriseID,
			APIKeyID:     &apiKeyID,
			Endpoint:     "/api/v1/enterprise/api/watermark",
			Method:       "POST",
			StatusCode:   statusCode,
			LatencyMs:    &latencyMs,
		}
		if err := h.entSvc.RecordUsage(context.Background(), usageRecord); err != nil {
			// Non-fatal — log and continue. Never crash the server over a usage record.
			log.Printf("WARN: failed to record API usage: %v", err)
		}
	}()

	c.Set("X-Image-ID", result.ImageID.String())
	c.Set("X-Serial-ID", strconv.FormatInt(result.SerialID, 10))
	c.Set("Content-Type", mimeType)
	return c.Status(fiber.StatusOK).Send(buf.Bytes())
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/enterprise/authenticate   [X-API-Key]
// ─────────────────────────────────────────────────────────────────────────────

// AuthenticateImage is the dual-channel authentication endpoint for enterprise
// clients. Runs the identical pipeline to the user-facing /images/authenticate.
//
// Form fields:
//
//	image — image file to authenticate   REQUIRED
//	k     — max similar images to return (optional, default 5)
func (h *EnterpriseHandler) AuthenticateImage(c *fiber.Ctx) error {
	img, k, err := parseImageAndK(c)
	if err != nil {
		return err // parseImageAndK already wrote the error response
	}

	result, err := h.imgSvc.ImageAuth(c.Context(), img, k)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "authentication failed: " + err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(buildAuthResponse(result))
}

// ─────────────────────────────────────────────────────────────────────────────
// error helper
// ─────────────────────────────────────────────────────────────────────────────

func enterpriseErrResponse(c *fiber.Ctx, err error) error {
	switch err {
	case services.ErrEnterpriseEmailTaken:
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case services.ErrEnterpriseNotFound:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case services.ErrEnterpriseInactive:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case services.ErrEnterpriseInvalidCreds, services.ErrAPIKeyInvalid, services.ErrAPIKeyNotFound:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	case services.ErrPasswordTooShort, services.ErrIncorrectOldPassword:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}