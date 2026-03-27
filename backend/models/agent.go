package models

import (
	"time"
)

type Agent struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	Role          string    `gorm:"size:255" json:"role"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	ContentCN     string    `gorm:"type:text" json:"content_cn"`
	Capabilities  string    `gorm:"type:text" json:"capabilities"`
	Category      string    `gorm:"size:100" json:"category"`
	Source        string    `gorm:"size:20;default:'custom'" json:"source"` // builtin/custom
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AgentResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	Content      string `json:"content"`
	ContentCN    string `json:"content_cn"`
	Capabilities string `json:"capabilities"`
	Category     string `json:"category"`
	Source       string `json:"source"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
