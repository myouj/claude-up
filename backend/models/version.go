package models

import (
	"time"
)

type PromptVersion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PromptID  uint      `gorm:"index;not null" json:"prompt_id"`
	Version   int       `gorm:"not null" json:"version"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Comment   string    `gorm:"size:500" json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type VersionResponse struct {
	ID        uint   `json:"id"`
	PromptID  uint   `json:"prompt_id"`
	Version   int    `json:"version"`
	Content   string `json:"content"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
}
