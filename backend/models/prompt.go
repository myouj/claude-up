package models

import (
	"time"
)

type Prompt struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:255;not null" json:"title"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	ContentCN   string    `gorm:"type:text" json:"content_cn"`
	Description string    `gorm:"type:text" json:"description"`
	Category    string    `gorm:"size:100" json:"category"`
	Tags        string    `gorm:"size:500" json:"tags"`   // JSON 存储标签数组
	Variables   string    `gorm:"type:text" json:"variables"` // JSON 存储变量定义
	IsFavorite  bool      `gorm:"default:false" json:"is_favorite"`
	IsPinned    bool      `gorm:"default:false" json:"is_pinned"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PromptResponse struct {
	ID           uint     `json:"id"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	ContentCN    string   `json:"content_cn"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	Variables    []Variable `json:"variables"`
	IsFavorite   bool     `json:"is_favorite"`
	IsPinned     bool     `json:"is_pinned"`
	VersionCount int      `json:"version_count"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

type Variable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     string `json:"default"`
}
