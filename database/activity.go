package database

import (
	"time"
)

// LogActivity adds an activity log entry
func LogActivity(userID uint, message string) {
	log := ActivityLog{
		UserID:    userID,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	}
	DB.Create(&log)
}

// GetActivityLogs retrieves recent activity for a user
func GetActivityLogs(userID uint, limit int) []ActivityLog {
	var logs []ActivityLog
	DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs)
	return logs
}
