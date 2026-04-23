package database

import (
	"time"
)

// User represents a player account
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"uniqueIndex;not null"`
	Email     string `gorm:"uniqueIndex;not null"`
	Password  string `gorm:"not null"`
	IsGuest   bool   `gorm:"default:false"`
	
	// Player stats
	PlayerName    string `gorm:"default:'Hero'"`
	PlayerClass   string `gorm:"default:'Apprentice Knight'"`
	Level         int    `gorm:"default:1"`
	XP            int64  `gorm:"default:0"`
	XPToNextLevel int    `gorm:"default:100"`
	
	Gold   int64 `gorm:"default:150"`
	HP     int   `gorm:"default:100"`
	MaxHP  int   `gorm:"default:100"`
	Mana   int   `gorm:"default:50"`
	MaxMana int  `gorm:"default:50"`
	
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OreInventory stores player's ore counts
type OreInventory struct {
	ID       uint   `gorm:"primaryKey"`
	UserID   uint   `gorm:"uniqueIndex;not null"`
	User     User   `gorm:"foreignKey:UserID"`
	
	CopperOre   int `gorm:"default:5"`
	IronOre     int `gorm:"default:2"`
	GoldOre     int `gorm:"default:0"`
	MithrilOre  int `gorm:"default:0"`
	DiamondOre  int `gorm:"default:0"`
	
	UpdatedAt time.Time
}

// OreType defines ore properties (static data)
type OreType struct {
	ID               uint   `gorm:"primaryKey"`
	OreKey           string `gorm:"uniqueIndex;not null"`
	OreName          string `gorm:"not null"`
	Icon             string
	Color            string
	Difficulty       string
	MiningTimeMS     int `gorm:"default:2000"` // milliseconds per ore
	XPPerOre         int `gorm:"default:10"`
	LevelRequired    int `gorm:"default:1"`
	
	CreatedAt time.Time
}

// MiningSession tracks mining progress
type MiningSession struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index:idx_user_active,unique,where:status='active'"`
	User      User      `gorm:"foreignKey:UserID"`
	OreID     uint      `gorm:"not null"`
	OreType   OreType   `gorm:"foreignKey:OreID"`
	
	// Server-side timestamps (cannot be hacked)
	StartedAt   time.Time `gorm:"not null"`
	EndedAt     *time.Time
	
	// How many ores earned in this session
	OresMined   int    `gorm:"default:0"`
	
	// Session status: 'active', 'paused', 'completed'
	Status      string `gorm:"default:'active';index"`
	
	CreatedAt   time.Time
}

// ActivityLog stores player actions
type ActivityLog struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;index"`
	User      User   `gorm:"foreignKey:UserID"`
	Message   string
	
	CreatedAt time.Time
}
