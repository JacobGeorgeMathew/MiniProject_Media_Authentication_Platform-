package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/services"
)

// ─────────────────────────────────────────────────────────────────────────────
// Enterprise JWT
// Uses a separate secret and claims struct from the user JWT so the two token
// spaces are cryptographically isolated.
// ─────────────────────────────────────────────────────────────────────────────

// enterpriseJWTSecret should be loaded from an environment variable in production.
var enterpriseJWTSecret = []byte("enterprise-jwt-secret-change-in-production")

type enterpriseClaims struct {
	EnterpriseID string `json:"enterprise_id"`
	Plan         string `json:"plan"`
	jwt.RegisteredClaims
}

// GenerateEnterpriseToken signs a 24-hour JWT for a logged-in enterprise.
// The plan tier is embedded so middleware can enforce quotas without a DB call.
func GenerateEnterpriseToken(enterpriseID uuid.UUID, plan string) (string, error) {
	claims := enterpriseClaims{
		EnterpriseID: enterpriseID.String(),
		Plan:         plan,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(enterpriseJWTSecret)
}

// ─────────────────────────────────────────────────────────────────────────────
// EnterpriseJWTMiddleware
// ─────────────────────────────────────────────────────────────────────────────

// EnterpriseJWTMiddleware validates the Bearer JWT issued at enterprise login.
//
// Injects into Fiber locals on success:
//
//	"enterpriseID" → uuid.UUID
//	"plan"         → string  ("free" | "starter" | "pro" | "enterprise")
//
// Use this on the enterprise dashboard / management route group
// (/api/v1/enterprise/me, /api/v1/enterprise/keys, /api/v1/enterprise/usage).
func EnterpriseJWTMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header is required",
		})
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header format must be 'Bearer <token>'",
		})
	}

	token, err := jwt.ParseWithClaims(parts[1], &enterpriseClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return enterpriseJWTSecret, nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or expired enterprise token",
		})
	}

	claims, ok := token.Claims.(*enterpriseClaims)
	if !ok || claims.EnterpriseID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "malformed enterprise token claims",
		})
	}

	id, err := uuid.Parse(claims.EnterpriseID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid enterprise ID in token",
		})
	}

	c.Locals("enterpriseID", id)
	c.Locals("plan", claims.Plan)
	return c.Next()
}

// ─────────────────────────────────────────────────────────────────────────────
// EnterpriseAPIKeyMiddleware
// ─────────────────────────────────────────────────────────────────────────────

// EnterpriseAPIKeyMiddleware authenticates incoming requests from enterprise
// clients using the X-API-Key header.
//
// Injects into Fiber locals on success:
//
//	"enterpriseID" → uuid.UUID
//	"apiKeyID"     → uuid.UUID
//	"plan"         → string
//
// Use this on the Watermarking-as-a-Service route group
// (/api/v1/enterprise/watermark, /api/v1/enterprise/authenticate).
func EnterpriseAPIKeyMiddleware(svc *services.EnterpriseService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawKey := c.Get("X-API-Key")
		if rawKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "X-API-Key header is required",
			})
		}

		result, err := svc.ValidateAPIKey(c.Context(), rawKey)
		if err != nil {
			switch err {
			case services.ErrAPIKeyNotFound, services.ErrAPIKeyInvalid:
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "invalid or revoked API key",
				})
			case services.ErrEnterpriseInactive:
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "enterprise account is deactivated",
				})
			default:
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "API key validation failed",
				})
			}
		}

		c.Locals("enterpriseID", result.EnterpriseID)
		c.Locals("apiKeyID", result.KeyID)
		c.Locals("plan", result.Plan)
		return c.Next()
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// EnterpriseIDFromLocals — shared helper used by all enterprise handlers
// ─────────────────────────────────────────────────────────────────────────────

// EnterpriseIDFromLocals reads the enterprise UUID injected by either
// EnterpriseJWTMiddleware or EnterpriseAPIKeyMiddleware.
func EnterpriseIDFromLocals(c *fiber.Ctx) (uuid.UUID, error) {
	raw := c.Locals("enterpriseID")
	if raw == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "missing enterprise authentication")
	}
	switch v := raw.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "invalid enterprise ID in context")
	}
}
