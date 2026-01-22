package models

import (
	"time"

	"gorm.io/gorm"
)

type Style struct {
	ID         uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Key        string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"key"`
	Name       string         `gorm:"type:varchar(100);not null" json:"name"`
	PromptZh   string         `gorm:"type:text" json:"prompt_zh"`
	PromptEn   string         `gorm:"type:text" json:"prompt_en"`
	PreviewURL string         `gorm:"type:text" json:"preview_url"`
	SortOrder  int            `gorm:"default:0" json:"sort_order"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	IsDefault  bool           `gorm:"default:false" json:"is_default"`
	CreatedAt  time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Style) TableName() string {
	return "styles"
}
