package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

var DB *gorm.DB

func Init() error {
	// Open SQLite database using pure Go driver (no CGO needed)
	db, err := gorm.Open(sqlite.Open("idlegame.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	DB = db

	// Run migrations
	err = migrate()
	if err != nil {
		return err
	}

	// Seed ore types if they don't exist
	err = seedOreTypes()
	if err != nil {
		return err
	}

	return nil
}

func migrate() error {
	// Auto migrate all models
	return DB.AutoMigrate(
		&User{},
		&OreInventory{},
		&OreType{},
		&MiningSession{},
		&ActivityLog{},
	)
}

func seedOreTypes() error {
	// Check if ores already exist
	count := int64(0)
	DB.Model(&OreType{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	ores := []OreType{
		{
			OreKey:       "copper_ore",
			OreName:      "Copper Ore",
			Icon:         "🪨",
			Color:        "#b87333",
			Difficulty:   "Easy",
			MiningTimeMS: 2000,
			XPPerOre:     10,
			LevelRequired: 1,
		},
		{
			OreKey:       "iron_ore",
			OreName:      "Iron Ore",
			Icon:         "⚫",
			Color:        "#5a5a5a",
			Difficulty:   "Normal",
			MiningTimeMS: 2000,
			XPPerOre:     15,
			LevelRequired: 5,
		},
		{
			OreKey:       "gold_ore",
			OreName:      "Gold Ore",
			Icon:         "✨",
			Color:        "#ffd700",
			Difficulty:   "Hard",
			MiningTimeMS: 2000,
			XPPerOre:     25,
			LevelRequired: 15,
		},
		{
			OreKey:       "mithril_ore",
			OreName:      "Mithril Ore",
			Icon:         "💎",
			Color:        "#00bfff",
			Difficulty:   "Very Hard",
			MiningTimeMS: 2000,
			XPPerOre:     40,
			LevelRequired: 30,
		},
		{
			OreKey:       "diamond_ore",
			OreName:      "Diamond Ore",
			Icon:         "💠",
			Color:        "#00ffff",
			Difficulty:   "Impossible",
			MiningTimeMS: 2000,
			XPPerOre:     60,
			LevelRequired: 50,
		},
	}

	return DB.CreateInBatches(ores, 100).Error
}

func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Log helper for debugging
func LogError(msg string, err error) {
	if err != nil {
		log.Printf("[ERROR] %s: %v\n", msg, err)
	}
}
