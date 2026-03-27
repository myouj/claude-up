package models

import (
	"time"
)

type Translation struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EntityType string    `gorm:"size:20;not null" json:"entity_type"` // prompt/skill/agent
	EntityID   uint      `gorm:"not null" json:"entity_id"`
	SourceLang string    `gorm:"size:10;not null" json:"source_lang"` // en/zh
	TargetLang string    `gorm:"size:10;not null" json:"target_lang"`
	SourceText string    `gorm:"type:text" json:"source_text"`
	TargetText string    `gorm:"type:text" json:"target_text"`
	CreatedAt  time.Time `json:"created_at"`
}

type TranslateRequest struct {
	Text       string `json:"text" binding:"required"`
	SourceLang string `json:"source_lang"` // default: "en"
	TargetLang string `json:"target_lang"` // default: "zh"
}

type TranslateResponse struct {
	SourceText string `json:"source_text"`
	TargetText string `json:"target_text"`
}
