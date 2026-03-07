package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// jwtSecret is loaded from config / environment in production.
// Replace with config.LoadConfig().JWTSecret or an equivalent.
var jwtSecret = []byte("change-me-in-production")

// jwtClaims is the payload we sign into every token.
type jwtClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates the Bearer JWT from the Authorization header and
// injects the caller's UUID into Fiber locals under the key "userID".
//
// Protected route groups should declare this middleware:
//
//	api.Group("/images", handlers.AuthMiddleware)
func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header is required",
		})
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header format must be 'Bearer <token>'",
		})
	}
	tokenStr := parts[1]

	// Parse and validate the token.
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or expired token",
		})
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || claims.UserID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "malformed token claims",
		})
	}

	// Parse the string UUID and store the typed value in locals.
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid user ID in token",
		})
	}

	c.Locals("userID", id)
	return c.Next()
}

// GenerateToken creates a signed JWT for the given user UUID.
// Call this from the Login handler (or a dedicated token endpoint) once the
// user's credentials are verified.
func GenerateToken(userID uuid.UUID) (string, error) {
	claims := jwtClaims{
		UserID: userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}