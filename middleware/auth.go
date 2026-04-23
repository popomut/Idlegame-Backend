package middleware

import (
	"github.com/gofiber/fiber/v2"
	"idlegame-backend/utils"
	"strings"
)

// AuthMiddleware validates JWT token and extracts user ID
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try to get token from Authorization header first
		auth := c.Get("Authorization")
		var token string
		
		if auth != "" {
			// Extract token from "Bearer <token>"
			parts := strings.Split(auth, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}
		
		// If no Authorization header, try to get token from httpOnly cookie
		if token == "" {
			token = c.Cookies("auth_token")
		}
		
		// Token is required
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization token",
			})
		}

		// Verify JWT token
		userID, err := utils.VerifyJWT(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		// Store user ID in context for use in handlers
		c.Locals("user_id", userID)

		return c.Next()
	}
}
