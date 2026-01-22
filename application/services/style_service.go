package services

import (
	"errors"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/cache"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
)

type StyleService struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewStyleService(db *gorm.DB, log *logger.Logger) *StyleService {
	return &StyleService{
		db:  db,
		log: log,
	}
}

func (s *StyleService) EnsureDefaults() error {
	defaults := DefaultStyles()
	if len(defaults) == 0 {
		return nil
	}

	var count int64
	if err := s.db.Model(&models.Style{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		if err := s.db.Create(&defaults).Error; err != nil {
			return err
		}
		cache.BumpNamespace(cache.NamespaceStyles)
		return nil
	}

	for _, def := range defaults {
		var existing models.Style
		err := s.db.Unscoped().Where("key = ?", def.Key).First(&existing).Error
		if err == nil {
			if existing.DeletedAt.Valid {
				updates := map[string]interface{}{
					"name":        def.Name,
					"prompt_zh":   def.PromptZh,
					"prompt_en":   def.PromptEn,
					"preview_url": def.PreviewURL,
					"sort_order":  def.SortOrder,
					"is_active":   true,
					"deleted_at":  nil,
				}
				if err := s.db.Model(&models.Style{}).Unscoped().Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
					s.log.Warnw("Failed to restore style", "key", def.Key, "error", err)
				}
			}
			continue
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := s.db.Create(&def).Error; err != nil {
				s.log.Warnw("Failed to seed style", "key", def.Key, "error", err)
			}
		}
	}

	var defaultCount int64
	if err := s.db.Model(&models.Style{}).Where("is_default = ?", true).Count(&defaultCount).Error; err == nil && defaultCount == 0 {
		_ = s.db.Model(&models.Style{}).Where("key = ?", fallbackDefaultStyleKey).Update("is_default", true).Error
	}

	cache.BumpNamespace(cache.NamespaceStyles)
	return nil
}

func (s *StyleService) ListActiveStyles() ([]models.Style, error) {
	cacheKey := cache.NamespaceKey(cache.NamespaceStyles, "active")
	if entry, ok := cache.Get(cacheKey); ok {
		if styles, ok := entry.Data.([]models.Style); ok {
			LoadStyleCatalog(styles)
			return styles, nil
		}
	}

	styles := []models.Style{}
	if err := s.db.Where("is_active = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&styles).Error; err != nil {
		return nil, err
	}

	if _, err := cache.Set(cacheKey, styles, 10*time.Minute); err != nil {
		s.log.Warnw("Failed to cache styles", "error", err)
	}

	LoadStyleCatalog(styles)
	return styles, nil
}

func (s *StyleService) RefreshStyleCatalog() error {
	styles := []models.Style{}
	if err := s.db.Where("is_active = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&styles).Error; err != nil {
		return err
	}
	LoadStyleCatalog(styles)
	return nil
}
