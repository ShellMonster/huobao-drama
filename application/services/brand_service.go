package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BrandService struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewBrandService(db *gorm.DB, log *logger.Logger) *BrandService {
	return &BrandService{db: db, log: log}
}

type BrandListQuery struct {
	IncludeInactive bool `form:"include_inactive"`
	WithSpecs       bool `form:"with_specs"`
}

type CreateBrandRequest struct {
	Name         string   `json:"name" binding:"required,min=1,max=200"`
	DisplayName  string   `json:"display_name"`
	LogoURL      string   `json:"logo_url"`
	LogoDarkURL  string   `json:"logo_dark_url"`
	LogoLightURL string   `json:"logo_light_url"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	IsActive     *bool    `json:"is_active"`
}

type UpdateBrandRequest struct {
	Name         string    `json:"name"`
	DisplayName  *string   `json:"display_name"`
	LogoURL      *string   `json:"logo_url"`
	LogoDarkURL  *string   `json:"logo_dark_url"`
	LogoLightURL *string   `json:"logo_light_url"`
	Description  *string   `json:"description"`
	Tags         *[]string `json:"tags"`
	IsActive     *bool     `json:"is_active"`
}

type CreateBrandSpecRequest struct {
	Name         string                 `json:"name" binding:"required,min=1,max=200"`
	Description  string                 `json:"description"`
	AllowedSizes []string               `json:"allowed_sizes"`
	AspectRatios []string               `json:"aspect_ratios"`
	SafeArea     map[string]interface{} `json:"safe_area"`
	LogoRules    map[string]interface{} `json:"logo_rules"`
	TextRules    map[string]interface{} `json:"text_rules"`
	SortOrder    *int                   `json:"sort_order"`
	IsDefault    bool                   `json:"is_default"`
	IsActive     *bool                  `json:"is_active"`
}

type UpdateBrandSpecRequest struct {
	Name         *string                 `json:"name"`
	Description  *string                 `json:"description"`
	AllowedSizes *[]string               `json:"allowed_sizes"`
	AspectRatios *[]string               `json:"aspect_ratios"`
	SafeArea     *map[string]interface{} `json:"safe_area"`
	LogoRules    *map[string]interface{} `json:"logo_rules"`
	TextRules    *map[string]interface{} `json:"text_rules"`
	SortOrder    *int                    `json:"sort_order"`
	IsDefault    *bool                   `json:"is_default"`
	IsActive     *bool                   `json:"is_active"`
}

func toJSON(value interface{}) datatypes.JSON {
	if value == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return datatypes.JSON(data)
}

func (s *BrandService) ListBrands(query *BrandListQuery) ([]models.Brand, error) {
	var brands []models.Brand
	db := s.db.Model(&models.Brand{})
	if query != nil && !query.IncludeInactive {
		db = db.Where("is_active = ?", true)
	}
	if query != nil && query.WithSpecs {
		db = db.Preload("Specs", func(db *gorm.DB) *gorm.DB {
			if query != nil && !query.IncludeInactive {
				db = db.Where("brand_specs.is_active = ?", true)
			}
			return db.Order("brand_specs.sort_order ASC")
		})
	}
	if err := db.Order("updated_at DESC").Find(&brands).Error; err != nil {
		s.log.Errorw("Failed to list brands", "error", err)
		return nil, err
	}
	return brands, nil
}

func (s *BrandService) GetBrand(brandID string) (*models.Brand, error) {
	var brand models.Brand
	if err := s.db.Where("id = ?", brandID).
		Preload("Specs", func(db *gorm.DB) *gorm.DB {
			return db.Order("brand_specs.sort_order ASC")
		}).
		First(&brand).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("brand not found")
		}
		return nil, err
	}
	return &brand, nil
}

func (s *BrandService) CreateBrand(req *CreateBrandRequest) (*models.Brand, error) {
	brand := &models.Brand{
		Name:     req.Name,
		IsActive: true,
	}
	if req.DisplayName != "" {
		brand.DisplayName = &req.DisplayName
	}
	if req.LogoURL != "" {
		brand.LogoURL = &req.LogoURL
	}
	if req.LogoDarkURL != "" {
		brand.LogoDarkURL = &req.LogoDarkURL
	}
	if req.LogoLightURL != "" {
		brand.LogoLightURL = &req.LogoLightURL
	}
	if req.Description != "" {
		brand.Description = &req.Description
	}
	if len(req.Tags) > 0 {
		brand.Tags = toJSON(req.Tags)
	}
	if req.IsActive != nil {
		brand.IsActive = *req.IsActive
	}

	if err := s.db.Create(brand).Error; err != nil {
		s.log.Errorw("Failed to create brand", "error", err)
		return nil, err
	}
	return brand, nil
}

func (s *BrandService) UpdateBrand(brandID string, req *UpdateBrandRequest) (*models.Brand, error) {
	var brand models.Brand
	if err := s.db.Where("id = ?", brandID).First(&brand).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("brand not found")
		}
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.DisplayName != nil {
		if *req.DisplayName == "" {
			updates["display_name"] = nil
		} else {
			updates["display_name"] = *req.DisplayName
		}
	}
	if req.LogoURL != nil {
		if *req.LogoURL == "" {
			updates["logo_url"] = nil
		} else {
			updates["logo_url"] = *req.LogoURL
		}
	}
	if req.LogoDarkURL != nil {
		if *req.LogoDarkURL == "" {
			updates["logo_dark_url"] = nil
		} else {
			updates["logo_dark_url"] = *req.LogoDarkURL
		}
	}
	if req.LogoLightURL != nil {
		if *req.LogoLightURL == "" {
			updates["logo_light_url"] = nil
		} else {
			updates["logo_light_url"] = *req.LogoLightURL
		}
	}
	if req.Description != nil {
		if *req.Description == "" {
			updates["description"] = nil
		} else {
			updates["description"] = *req.Description
		}
	}
	if req.Tags != nil {
		if len(*req.Tags) == 0 {
			updates["tags"] = nil
		} else {
			updates["tags"] = toJSON(*req.Tags)
		}
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	updates["updated_at"] = time.Now()

	if err := s.db.Model(&brand).Updates(updates).Error; err != nil {
		s.log.Errorw("Failed to update brand", "error", err)
		return nil, err
	}

	return &brand, nil
}

func (s *BrandService) DeleteBrand(brandID string) error {
	result := s.db.Where("id = ?", brandID).Delete(&models.Brand{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("brand not found")
	}
	return nil
}

func (s *BrandService) ListBrandSpecs(brandID string, includeInactive bool) ([]models.BrandSpec, error) {
	var specs []models.BrandSpec
	db := s.db.Model(&models.BrandSpec{}).Where("brand_id = ?", brandID)
	if !includeInactive {
		db = db.Where("is_active = ?", true)
	}
	if err := db.Order("sort_order ASC, id ASC").Find(&specs).Error; err != nil {
		return nil, err
	}
	return specs, nil
}

func (s *BrandService) CreateBrandSpec(brandID string, req *CreateBrandSpecRequest) (*models.BrandSpec, error) {
	var brand models.Brand
	if err := s.db.Where("id = ?", brandID).First(&brand).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("brand not found")
		}
		return nil, err
	}

	spec := &models.BrandSpec{
		BrandID:      brand.ID,
		Name:         req.Name,
		AllowedSizes: toJSON(req.AllowedSizes),
		AspectRatios: toJSON(req.AspectRatios),
		SafeArea:     toJSON(req.SafeArea),
		LogoRules:    toJSON(req.LogoRules),
		TextRules:    toJSON(req.TextRules),
		IsDefault:    req.IsDefault,
		IsActive:     true,
	}
	if req.Description != "" {
		spec.Description = &req.Description
	}
	if req.SortOrder != nil {
		spec.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		spec.IsActive = *req.IsActive
	}

	tx := s.db.Begin()
	if req.IsDefault {
		if err := tx.Model(&models.BrandSpec{}).Where("brand_id = ?", brand.ID).Update("is_default", false).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Create(spec).Error; err != nil {
		tx.Rollback()
		s.log.Errorw("Failed to create brand spec", "error", err)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return spec, nil
}

func (s *BrandService) UpdateBrandSpec(specID string, req *UpdateBrandSpecRequest) (*models.BrandSpec, error) {
	var spec models.BrandSpec
	if err := s.db.Where("id = ?", specID).First(&spec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("brand spec not found")
		}
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != nil && *req.Name != "" {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		if *req.Description == "" {
			updates["description"] = nil
		} else {
			updates["description"] = *req.Description
		}
	}
	if req.AllowedSizes != nil {
		if len(*req.AllowedSizes) == 0 {
			updates["allowed_sizes"] = nil
		} else {
			updates["allowed_sizes"] = toJSON(*req.AllowedSizes)
		}
	}
	if req.AspectRatios != nil {
		if len(*req.AspectRatios) == 0 {
			updates["aspect_ratios"] = nil
		} else {
			updates["aspect_ratios"] = toJSON(*req.AspectRatios)
		}
	}
	if req.SafeArea != nil {
		updates["safe_area"] = toJSON(*req.SafeArea)
	}
	if req.LogoRules != nil {
		updates["logo_rules"] = toJSON(*req.LogoRules)
	}
	if req.TextRules != nil {
		updates["text_rules"] = toJSON(*req.TextRules)
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}
	updates["updated_at"] = time.Now()

	tx := s.db.Begin()
	if req.IsDefault != nil && *req.IsDefault {
		if err := tx.Model(&models.BrandSpec{}).Where("brand_id = ?", spec.BrandID).Update("is_default", false).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Model(&spec).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &spec, nil
}

func (s *BrandService) DeleteBrandSpec(specID string) error {
	result := s.db.Where("id = ?", specID).Delete(&models.BrandSpec{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("brand spec not found")
	}
	return nil
}

func (s *BrandService) GetDefaultSpec(brandID string) (*models.BrandSpec, error) {
	var spec models.BrandSpec
	err := s.db.Where("brand_id = ? AND is_default = ?", brandID, true).First(&spec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("default spec not found")
		}
		return nil, err
	}
	return &spec, nil
}

func (s *BrandService) ResolveSpec(brandID uint, specID *uint) (*models.BrandSpec, error) {
	if specID != nil {
		var spec models.BrandSpec
		if err := s.db.Where("id = ? AND brand_id = ?", *specID, brandID).First(&spec).Error; err != nil {
			return nil, fmt.Errorf("brand spec not found")
		}
		return &spec, nil
	}
	return s.GetDefaultSpec(fmt.Sprintf("%d", brandID))
}
