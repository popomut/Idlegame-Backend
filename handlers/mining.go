package handlers

import (
	"github.com/gofiber/fiber/v2"
	"idlegame-backend/database"
	"time"
)

// StartMiningRequest contains ore selection
type StartMiningRequest struct {
	OreID uint `json:"ore_id"`
}

// MiningStatusResponse returns mining progress and offline gains
type MiningStatusResponse struct {
	IsActive      bool              `json:"is_active"`
	CurrentOre    *OreTypeResponse  `json:"current_ore,omitempty"`
	StartedAt     time.Time         `json:"started_at,omitempty"`
	OfflineGains  OfflineGainsInfo  `json:"offline_gains,omitempty"`
	CurrentOres   map[string]int    `json:"current_ores"`
}

type OreTypeResponse struct {
	ID             uint   `json:"id"`
	OreKey         string `json:"ore_key"`
	OreName        string `json:"ore_name"`
	Icon           string `json:"icon"`
	Difficulty     string `json:"difficulty"`
	MiningTimeMS   int    `json:"mining_time_ms"`
}

type OfflineGainsInfo struct {
	WasOffline     bool      `json:"was_offline"`
	OfflineTime    int64     `json:"offline_time_ms"`
	OresGained     int       `json:"ores_gained"`
	OreName        string    `json:"ore_name"`
	LastActiveTime time.Time `json:"last_active_time"`
}

// StartMining begins a mining session
func StartMining(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	req := new(StartMiningRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	// Check if ore exists
	var ore database.OreType
	if err := database.DB.First(&ore, req.OreID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ore not found"})
	}

	// Stop any existing mining session
	var existingSession database.MiningSession
	database.DB.Where("user_id = ? AND status = ?", userID, "active").First(&existingSession)
	if existingSession.ID != 0 {
		// Calculate and save offline gains before stopping
		CalculateAndSaveOreGains(userID, existingSession)
		database.DB.Model(&existingSession).Update("status", "completed")
	}

	// Start new mining session (server-side timestamp!)
	session := database.MiningSession{
		UserID:    userID,
		OreID:     ore.ID,
		StartedAt: time.Now().UTC(), // SERVER TIMESTAMP - Cannot be spoofed
		Status:    "active",
	}

	if err := database.DB.Create(&session).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to start mining"})
	}

	database.LogActivity(userID, "Started mining "+ore.OreName)

	return c.JSON(fiber.Map{
		"status":      "mining started",
		"session_id":  session.ID,
		"ore_name":    ore.OreName,
		"started_at":  session.StartedAt,
	})
}

// StopMining stops the current mining session
func StopMining(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	// Find active session
	var session database.MiningSession
	result := database.DB.Where("user_id = ? AND status = ?", userID, "active").First(&session)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no active mining session"})
	}

	// Calculate and save ore gains
	oredGained := CalculateAndSaveOreGains(userID, session)

	// End session (server-side timestamp!)
	now := time.Now().UTC()
	database.DB.Model(&session).Updates(map[string]interface{}{
		"status":   "completed",
		"ended_at": now,
	})

	var ore database.OreType
	database.DB.First(&ore, session.OreID)

	database.LogActivity(userID, "Stopped mining "+ore.OreName+". Gained "+string(rune(oredGained))+" ores")

	return c.JSON(fiber.Map{
		"status":    "mining stopped",
		"ores_gained": oredGained,
	})
}

// GetMiningStatus returns current mining status and offline gains
func GetMiningStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	// Get current mining session
	var session database.MiningSession
	isActive := false
	var ore database.OreType

	result := database.DB.Where("user_id = ? AND status = ?", userID, "active").Preload("OreType").First(&session)
	if result.Error == nil {
		isActive = true
		ore = session.OreType
	}

	// Get current ore inventory
	var inventory database.OreInventory
	database.DB.Where("user_id = ?", userID).First(&inventory)

	currentOres := map[string]int{
		"copper_ore":   inventory.CopperOre,
		"iron_ore":     inventory.IronOre,
		"gold_ore":     inventory.GoldOre,
		"mithril_ore":  inventory.MithrilOre,
		"diamond_ore":  inventory.DiamondOre,
	}

	// If actively mining, calculate pending earnings and add to display
	if isActive {
		now := time.Now().UTC()
		elapsed := now.Sub(session.StartedAt)
		pendingOres := int(elapsed.Milliseconds()) / ore.MiningTimeMS
		
		// Add pending ores to the current display (but NOT saved to DB yet)
		switch ore.OreKey {
		case "copper_ore":
			currentOres["copper_ore"] += pendingOres
		case "iron_ore":
			currentOres["iron_ore"] += pendingOres
		case "gold_ore":
			currentOres["gold_ore"] += pendingOres
		case "mithril_ore":
			currentOres["mithril_ore"] += pendingOres
		case "diamond_ore":
			currentOres["diamond_ore"] += pendingOres
		}
	}

	response := MiningStatusResponse{
		IsActive:    isActive,
		CurrentOres: currentOres,
	}

	if isActive {
		response.CurrentOre = &OreTypeResponse{
			ID:           ore.ID,
			OreKey:       ore.OreKey,
			OreName:      ore.OreName,
			Icon:         ore.Icon,
			Difficulty:   ore.Difficulty,
			MiningTimeMS: ore.MiningTimeMS,
		}
		response.StartedAt = session.StartedAt

		// Calculate offline gains if user had previous session
		// (This happens when they close browser and come back)
		offlineGains := CalculateOfflineGains(userID, session)
		response.OfflineGains = offlineGains
	}

	return c.JSON(response)
}

// CalculateAndSaveOreGains calculates earned ores and updates inventory
// This is SERVER-SIDE calculation - prevents cheating!
func CalculateAndSaveOreGains(userID uint, session database.MiningSession) int {
	// Get time elapsed using SERVER timestamps (not client)
	now := time.Now().UTC()
	elapsed := now.Sub(session.StartedAt)
	
	var ore database.OreType
	database.DB.First(&ore, session.OreID)

	// Calculate ores earned: elapsed_time / mining_time_per_ore
	oresEarned := int(elapsed.Milliseconds()) / ore.MiningTimeMS
	if oresEarned == 0 {
		return 0
	}

	// Update inventory with earned ores (SERVER-SIDE, not client)
	var inventory database.OreInventory
	database.DB.Where("user_id = ?", userID).First(&inventory)

	// Use switch to prevent injection attacks
	switch ore.OreKey {
	case "copper_ore":
		inventory.CopperOre += oresEarned
	case "iron_ore":
		inventory.IronOre += oresEarned
	case "gold_ore":
		inventory.GoldOre += oresEarned
	case "mithril_ore":
		inventory.MithrilOre += oresEarned
	case "diamond_ore":
		inventory.DiamondOre += oresEarned
	}

	database.DB.Save(&inventory)

	// Log activity
	database.LogActivity(userID, "Mined "+string(rune(oresEarned))+" "+ore.OreName)

	return oresEarned
}

// CalculateOfflineGains determines what player earned while offline
func CalculateOfflineGains(userID uint, session database.MiningSession) OfflineGainsInfo {
	gains := OfflineGainsInfo{
		WasOffline: false,
	}

	var ore database.OreType
	database.DB.First(&ore, session.OreID)
	gains.OreName = ore.OreName

	now := time.Now().UTC()
	elapsed := now.Sub(session.StartedAt)

	// If more than 1 ore worth of time passed, consider it offline gains
	miningTimePerOre := time.Duration(ore.MiningTimeMS) * time.Millisecond
	if elapsed > miningTimePerOre {
		gains.WasOffline = true
		gains.OfflineTime = elapsed.Milliseconds()
		gains.OresGained = int(elapsed.Milliseconds()) / ore.MiningTimeMS
		gains.LastActiveTime = session.StartedAt
	}

	return gains
}
