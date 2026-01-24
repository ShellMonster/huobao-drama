package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Brand struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string         `gorm:"type:varchar(200);not null" json:"name"`
	DisplayName  *string        `gorm:"type:varchar(200)" json:"display_name,omitempty"`
	LogoURL      *string        `gorm:"type:varchar(1000)" json:"logo_url,omitempty"`
	LogoDarkURL  *string        `gorm:"type:varchar(1000)" json:"logo_dark_url,omitempty"`
	LogoLightURL *string        `gorm:"type:varchar(1000)" json:"logo_light_url,omitempty"`
	Description  *string        `gorm:"type:text" json:"description,omitempty"`
	Tags         datatypes.JSON `gorm:"type:json" json:"tags,omitempty"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Specs []BrandSpec `gorm:"foreignKey:BrandID" json:"specs,omitempty"`
}

func (Brand) TableName() string {
	return "brands"
}

type BrandSpec struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	BrandID      uint           `gorm:"index;not null" json:"brand_id"`
	Name         string         `gorm:"type:varchar(200);not null" json:"name"`
	Description  *string        `gorm:"type:text" json:"description,omitempty"`
	AllowedSizes datatypes.JSON `gorm:"type:json" json:"allowed_sizes,omitempty"`
	AspectRatios datatypes.JSON `gorm:"type:json" json:"aspect_ratios,omitempty"`
	SafeArea     datatypes.JSON `gorm:"type:json" json:"safe_area,omitempty"`
	LogoRules    datatypes.JSON `gorm:"type:json" json:"logo_rules,omitempty"`
	TextRules    datatypes.JSON `gorm:"type:json" json:"text_rules,omitempty"`
	SortOrder    int            `gorm:"default:0" json:"sort_order"`
	IsDefault    bool           `gorm:"default:false" json:"is_default"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Brand Brand `gorm:"foreignKey:BrandID" json:"brand,omitempty"`
}

func (BrandSpec) TableName() string {
	return "brand_specs"
}
