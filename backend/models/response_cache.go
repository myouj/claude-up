package models

import "time"

type ResponseCache struct {
	Hash        string    `gorm:"primaryKey" json:"hash"`
	Provider    string    `gorm:"size:50;not null" json:"provider"`
	Model       string    `gorm:"size:100" json:"model"`
	RequestHash string    `gorm:"size:64;not null" json:"request_hash"`
	Response    string    `gorm:"type:text" json:"response"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (ResponseCache) TableName() string {
	return "response_cache"
}
