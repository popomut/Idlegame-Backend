package handlers

import (
	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// GetUser retrieves current user profile
func GetUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var user database.User
	result := database.DB.First(&user, userID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(user)
}

// UpdateUserRequest for profile updates
type UpdateUserRequest struct {
	PlayerName string `json:"player_name"`
	PlayerClass string `json:"player_class"`
}

// UpdateUser updates user profile
func UpdateUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	req := new(UpdateUserRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	updates := map[string]interface{}{}
	if req.PlayerName != "" {
		updates["player_name"] = req.PlayerName
	}
	if req.PlayerClass != "" {
		updates["player_class"] = req.PlayerClass
	}

	result := database.DB.Model(&database.User{ID: userID}).Updates(updates)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update user"})
	}

	var user database.User
	database.DB.First(&user, userID)

	return c.JSON(user)
}
