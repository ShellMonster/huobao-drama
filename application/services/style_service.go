package services

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/cache"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/google/uuid"
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

var (
	ErrStyleNotFound     = errors.New("style not found")
	ErrStyleKeyExists    = errors.New("style key already exists")
	ErrStyleSystemDelete = errors.New("system style cannot be deleted")
)

const styleSeedDir = "assets/style-seeds"

type CreateStyleRequest struct {
	Key        string `json:"key"`
	Name       string `json:"name" binding:"required"`
	PromptZh   string `json:"prompt_zh"`
	PromptEn   string `json:"prompt_en"`
	PreviewURL string `json:"preview_url"`
	SortOrder  int    `json:"sort_order"`
	IsActive   *bool  `json:"is_active"`
	IsDefault  bool   `json:"is_default"`
}

type UpdateStyleRequest struct {
	Name       *string `json:"name"`
	PromptZh   *string `json:"prompt_zh"`
	PromptEn   *string `json:"prompt_en"`
	PreviewURL *string `json:"preview_url"`
	SortOrder  *int    `json:"sort_order"`
	IsActive   *bool   `json:"is_active"`
	IsDefault  *bool   `json:"is_default"`
}

func (s *StyleService) EnsureDefaults(storagePath, baseURL string) error {
	seeds := DefaultStyleSeeds()
	if len(seeds) == 0 {
		return nil
	}

	styles, err := buildStylesFromSeeds(seeds, storagePath, baseURL, s.log)
	if err != nil {
		return err
	}

	var total int64
	if err := s.db.Model(&models.Style{}).Count(&total).Error; err != nil {
		return err
	}

	if total == 0 {
		if err := s.db.Create(&styles).Error; err != nil {
			return err
		}
		cache.BumpNamespace(cache.NamespaceStyles)
		LoadStyleCatalog(styles)
		return nil
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, style := range styles {
			var existing models.Style
			if err := tx.Where("key = ?", style.Key).First(&existing).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err := tx.Create(&style).Error; err != nil {
						return err
					}
					continue
				}
				return err
			}

			if !existing.IsSystem {
				if err := tx.Model(&models.Style{}).Where("id = ?", existing.ID).Update("is_system", true).Error; err != nil {
					return err
				}
			}
		}

		return ensureDefaultStyle(tx)
	}); err != nil {
		return err
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

	tuning := config.GetTuning()
	cacheTTL := config.DurationFromSeconds(tuning.Cache.StyleCatalogSeconds, 10*time.Minute)
	if _, err := cache.Set(cacheKey, styles, cacheTTL); err != nil {
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

func (s *StyleService) ListStyles(includeInactive bool) ([]models.Style, error) {
	if !includeInactive {
		return s.ListActiveStyles()
	}

	styles := []models.Style{}
	if err := s.db.Order("sort_order ASC, id ASC").
		Find(&styles).Error; err != nil {
		return nil, err
	}

	LoadStyleCatalog(styles)
	return styles, nil
}

func (s *StyleService) CreateStyle(req *CreateStyleRequest) (*models.Style, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("style name is required")
	}

	key := strings.TrimSpace(req.Key)
	if key == "" {
		key = "style_" + uuid.New().String()
	}
	if len(key) > 50 {
		return nil, fmt.Errorf("style key too long")
	}

	var existingCount int64
	if err := s.db.Model(&models.Style{}).Where("key = ?", key).Count(&existingCount).Error; err != nil {
		return nil, err
	}
	if existingCount > 0 {
		return nil, ErrStyleKeyExists
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.IsDefault {
		isActive = true
	}

	promptZh := strings.TrimSpace(req.PromptZh)
	if promptZh == "" {
		promptZh = name
	}
	promptEn := strings.TrimSpace(req.PromptEn)
	if promptEn == "" {
		promptEn = name
	}

	style := &models.Style{
		Key:        key,
		Name:       name,
		PromptZh:   promptZh,
		PromptEn:   promptEn,
		PreviewURL: strings.TrimSpace(req.PreviewURL),
		SortOrder:  req.SortOrder,
		IsActive:   isActive,
		IsDefault:  req.IsDefault,
		IsSystem:   false,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if style.IsDefault {
			if err := tx.Model(&models.Style{}).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(style).Error; err != nil {
			return err
		}
		return ensureDefaultStyle(tx)
	}); err != nil {
		return nil, err
	}

	return style, nil
}

func (s *StyleService) UpdateStyle(id uint, req *UpdateStyleRequest) (*models.Style, error) {
	var style models.Style
	if err := s.db.First(&style, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStyleNotFound
		}
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return nil, fmt.Errorf("style name is required")
		}
		updates["name"] = trimmed
	}
	if req.PromptZh != nil {
		updates["prompt_zh"] = strings.TrimSpace(*req.PromptZh)
	}
	if req.PromptEn != nil {
		updates["prompt_en"] = strings.TrimSpace(*req.PromptEn)
	}
	if req.PreviewURL != nil {
		updates["preview_url"] = strings.TrimSpace(*req.PreviewURL)
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
		if !*req.IsActive && style.IsDefault {
			updates["is_default"] = false
		}
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
		if *req.IsDefault {
			updates["is_active"] = true
		}
	}

	if len(updates) == 0 {
		return &style, nil
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if req.IsDefault != nil && *req.IsDefault {
			if err := tx.Model(&models.Style{}).Where("id <> ?", style.ID).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&models.Style{}).Where("id = ?", style.ID).Updates(updates).Error; err != nil {
			return err
		}
		if err := ensureDefaultStyle(tx); err != nil {
			return err
		}
		return tx.First(&style, style.ID).Error
	}); err != nil {
		return nil, err
	}

	return &style, nil
}

func (s *StyleService) DeleteStyle(id uint) error {
	var style models.Style
	if err := s.db.First(&style, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStyleNotFound
		}
		return err
	}

	if style.IsSystem {
		return ErrStyleSystemDelete
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&style).Error; err != nil {
			return err
		}
		return ensureDefaultStyle(tx)
	})
}

func buildStylesFromSeeds(seeds []StyleSeed, storagePath, baseURL string, log *logger.Logger) ([]models.Style, error) {
	if storagePath == "" {
		return nil, fmt.Errorf("storage path is empty")
	}
	styleDir := filepath.Join(storagePath, "styles")
	if err := os.MkdirAll(styleDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create style directory: %w", err)
	}
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		return nil, fmt.Errorf("base url is empty")
	}

	styles := make([]models.Style, 0, len(seeds))
	for idx, seed := range seeds {
		if seed.ImageURL == "" || seed.Name == "" {
			continue
		}
		promptZh := strings.TrimSpace(seed.PromptZh)
		if promptZh == "" {
			promptZh = seed.Name
		}
		promptEn := strings.TrimSpace(seed.PromptEn)
		if promptEn == "" {
			promptEn = seed.Name
		}
		key := hashStyleKey(seed.ImageURL)
		filename := key + ".webp"
		localPath := filepath.Join(styleDir, filename)
		seedPath := filepath.Join(styleSeedDir, filename)
		if err := ensureStyleImage(seed.ImageURL, localPath, seedPath); err != nil {
			log.Warnw("Failed to prepare style image", "name", seed.Name, "url", seed.ImageURL, "error", err)
		}

		styles = append(styles, models.Style{
			Key:        key,
			Name:       seed.Name,
			PromptZh:   promptZh,
			PromptEn:   promptEn,
			PreviewURL: fmt.Sprintf("%s/styles/%s", base, filename),
			SortOrder:  idx,
			IsActive:   true,
			IsDefault:  idx == 0,
			IsSystem:   true,
		})
	}

	return styles, nil
}

func hashStyleKey(source string) string {
	sum := md5.Sum([]byte(source))
	return hex.EncodeToString(sum[:])
}

func ensureDefaultStyle(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&models.Style{}).Where("is_default = ?", true).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var fallback models.Style
	if err := tx.Where("is_active = ?", true).
		Order("sort_order ASC, id ASC").
		First(&fallback).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	return tx.Model(&models.Style{}).Where("id = ?", fallback.ID).Update("is_default", true).Error
}

func downloadStyleImage(url, path string) error {
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		return nil
	}

	tmpPath := path + ".tmp"
	timeout := config.DurationFromSeconds(config.GetTuning().HTTPTimeout.StyleFetchSeconds, 30*time.Second)
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, resp.Body); err != nil {
		_ = file.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func ensureStyleImage(url, path, seedPath string) error {
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		return nil
	}
	if seedPath != "" {
		if info, err := os.Stat(seedPath); err == nil && info.Size() > 0 {
			return copyStyleSeedImage(seedPath, path)
		}
	}
	return downloadStyleImage(url, path)
}

func copyStyleSeedImage(src, dest string) error {
	tmpPath := dest + ".tmp"
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, dest); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}
