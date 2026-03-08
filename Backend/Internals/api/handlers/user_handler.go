package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/api/middleware"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/models"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/services"
)

// ─────────────────────────────────────────────────────────────────────────────
// UserHandler
// ─────────────────────────────────────────────────────────────────────────────

type UserHandler struct {
	svc *services.UserService
}

func NewUserHandler(svc *services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// ─────────────────────────────────────────────────────────────────────────────
// Wire types  (request bodies decoded from JSON)
// ─────────────────────────────────────────────────────────────────────────────

type registerBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateProfileBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type changePasswordBody struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/users/register
// ─────────────────────────────────────────────────────────────────────────────

// Register creates a new user account.
// Public endpoint — no auth middleware required.
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var body registerBody
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
	if body.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username is required",
		})
	}

	resp, err := h.svc.Register(c.Context(), models.RegisterRequest{
		Username: body.Username,
		Email:    body.Email,
		Password: body.Password,
		FullName: body.FullName,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailAlreadyExists):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrPasswordTooShort):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "registration failed"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /api/v1/users/login
// ─────────────────────────────────────────────────────────────────────────────

// Login verifies credentials and returns the user profile.
// Token issuance (JWT / session cookie) should be added here or in a
// dedicated auth middleware once you wire one in.
// Public endpoint — no auth middleware required.
func (h *UserHandler) Login(c *fiber.Ctx) error {
    var body loginBody
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

    resp, err := h.svc.Login(c.Context(), models.LoginRequest{
        Email:    body.Email,
        Password: body.Password,
    })
    if err != nil {
        switch {
        case errors.Is(err, services.ErrInvalidCredentials):
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
        case errors.Is(err, services.ErrUserInactive):
            return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
        default:
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "login failed"})
        }
    }

    // Generate JWT and attach it to the response struct before returning.
    token, err := middleware.GenerateToken(resp.ID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to generate token",
        })
    }
    resp.Token = token

    return c.Status(fiber.StatusOK).JSON(resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/users/me
// ─────────────────────────────────────────────────────────────────────────────

// GetMe returns the authenticated user's profile.
// Requires auth middleware to set "userID" in Fiber locals.
func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	userID, err := userIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing or invalid authentication",
		})
	}

	resp, err := h.svc.GetByID(c.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch profile"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /api/v1/users/me
// ─────────────────────────────────────────────────────────────────────────────

// UpdateProfile applies mutable field changes (username, email, full_name).
// Requires auth middleware.
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, err := userIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing or invalid authentication",
		})
	}

	var body updateProfileBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON: " + err.Error(),
		})
	}

	if body.Username == "" && body.Email == "" && body.FullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "at least one field (username, email, full_name) must be provided",
		})
	}

	resp, err := h.svc.UpdateProfile(c.Context(), models.UpdateProfileRequest{
		UserID:   userID,
		Username: body.Username,
		Email:    body.Email,
		FullName: body.FullName,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrUserInactive):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "profile update failed"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /api/v1/users/me/password
// ─────────────────────────────────────────────────────────────────────────────

// ChangePassword verifies the old password and replaces it with the new one.
// Requires auth middleware.
func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	userID, err := userIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing or invalid authentication",
		})
	}

	var body changePasswordBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON: " + err.Error(),
		})
	}

	if body.OldPassword == "" || body.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "old_password and new_password are required",
		})
	}

	if err := h.svc.ChangePassword(c.Context(), models.ChangePasswordRequest{
		UserID:      userID,
		OldPassword: body.OldPassword,
		NewPassword: body.NewPassword,
	}); err != nil {
		switch {
		case errors.Is(err, services.ErrIncorrectOldPassword):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrPasswordTooShort):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "password change failed"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "password updated successfully"})
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /api/v1/users/me
// ─────────────────────────────────────────────────────────────────────────────

// DeactivateAccount soft-deletes the authenticated user's account.
// All image_metadata rows referencing this user are preserved for provenance.
// Requires auth middleware.
func (h *UserHandler) DeactivateAccount(c *fiber.Ctx) error {
	userID, err := userIDFromLocals(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing or invalid authentication",
		})
	}

	if err := h.svc.DeactivateAccount(c.Context(), userID); err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrUserInactive):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "account is already deactivated"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "account deactivation failed"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "account deactivated successfully"})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /api/v1/users/:id   (admin / lookup use-case)
// ─────────────────────────────────────────────────────────────────────────────

// GetUserByID fetches any user's public profile by UUID path param.
// Protect this with an admin-role middleware in production.
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user ID format",
		})
	}

	resp, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch user"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}