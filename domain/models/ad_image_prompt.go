package models

import (
	"time"

	"gorm.io/datatypes"
)

type AdImagePrompt struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	DramaID        uint           `gorm:"index;not null" json:"drama_id"`
	BrandID        *uint          `gorm:"index" json:"brand_id,omitempty"`
	SpecID         *uint          `gorm:"index" json:"spec_id,omitempty"`
	PromptType     string         `gorm:"size:20;index;not null" json:"prompt_type"`
	SourceText     *string        `gorm:"type:text" json:"source_text,omitempty"`
	SourceImageURL *string        `gorm:"type:text" json:"source_image_url,omitempty"`
	Prompts        datatypes.JSON `gorm:"type:json;not null" json:"prompts"`
	CreatedAt      time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (AdImagePrompt) TableName() string {
	return "ad_image_prompts"
}

const (
	AdPromptTypeText  = "text"
	AdPromptTypeImage = "image"
)
