package handlers

import (
	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
)

// GetOreInventoryResponse returns ore counts
type GetOreInventoryResponse struct {
	CopperOre  int `json:"copper_ore"`
	IronOre    int `json:"iron_ore"`
	GoldOre    int `json:"gold_ore"`
	MithrilOre int `json:"mithril_ore"`
	DiamondOre int `json:"diamond_ore"`
}

// GetOreInventory returns player's ore inventory
func GetOreInventory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var inventory database.OreInventory
	result := database.DB.Where("user_id = ?", userID).First(&inventory)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "inventory not found"})
	}

	return c.JSON(GetOreInventoryResponse{
		CopperOre:  inventory.CopperOre,
		IronOre:    inventory.IronOre,
		GoldOre:    inventory.GoldOre,
		MithrilOre: inventory.MithrilOre,
		DiamondOre: inventory.DiamondOre,
	})
}
