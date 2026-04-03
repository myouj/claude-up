package models

import (
	"time"
)

type ActivityLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EntityType string    `gorm:"size:20;not null" json:"entity_type"` // prompt/skill/agent/version/test
	EntityID   uint      `gorm:"not null" json:"entity_id"`
	Action     string    `gorm:"size:50;not null" json:"action"` // created/updated/deleted/cloned/tested/optimized/translated/favorited
	ActionType  string    `gorm:"size:50" json:"action_type"` // task_created | task_completed | ...
	Description string    `gorm:"type:text" json:"description"`  // Human-readable description of the activity
	UserID     uint      `gorm:"default:1" json:"user_id"`
	Details    string    `gorm:"type:text" json:"details"`        // JSON additional context
	CreatedAt  time.Time `json:"created_at"`
}

type ActivityLogResponse struct {
	ID         uint   `json:"id"`
	EntityType string `json:"entity_type"`
	EntityID   uint   `json:"entity_id"`
	Action      string `json:"action"`
	ActionType  string `json:"action_type"`
	Description string `json:"description"`
	UserID     uint   `json:"user_id"`
	Details    string `json:"details"`
	CreatedAt  string `json:"created_at"`
}
