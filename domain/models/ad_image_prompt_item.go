package models

import "time"

type AdImagePromptItem struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PromptID   uint      `gorm:"index;not null" json:"prompt_id"`
	DramaID    uint      `gorm:"index;not null" json:"drama_id"`
	BrandID    *uint     `gorm:"index" json:"brand_id,omitempty"`
	SpecID     *uint     `gorm:"index" json:"spec_id,omitempty"`
	PromptType string    `gorm:"size:20;index;not null" json:"prompt_type"`
	Prompt     string    `gorm:"type:text;not null" json:"prompt"`
	SortOrder  int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt  time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (AdImagePromptItem) TableName() string {
	return "ad_image_prompt_items"
}
