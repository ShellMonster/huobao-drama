package services

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func (s *StyleService) EnsureDefaults(storagePath, baseURL string) error {
	seeds := DefaultStyleSeeds()
	if len(seeds) == 0 {
		return nil
	}

	styles, err := buildStylesFromSeeds(seeds, storagePath, baseURL, s.log)
	if err != nil {
		return err
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("1 = 1").Delete(&models.Style{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&styles).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	cache.BumpNamespace(cache.NamespaceStyles)
	LoadStyleCatalog(styles)
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
		key := hashStyleKey(seed.ImageURL)
		filename := key + ".webp"
		localPath := filepath.Join(styleDir, filename)
		if err := downloadStyleImage(seed.ImageURL, localPath); err != nil {
			log.Warnw("Failed to download style image", "name", seed.Name, "url", seed.ImageURL, "error", err)
		}

		styles = append(styles, models.Style{
			Key:        key,
			Name:       seed.Name,
			PromptZh:   seed.Name,
			PromptEn:   seed.Name,
			PreviewURL: fmt.Sprintf("%s/styles/%s", base, filename),
			SortOrder:  idx,
			IsActive:   true,
			IsDefault:  idx == 0,
		})
	}

	return styles, nil
}

func hashStyleKey(source string) string {
	sum := md5.Sum([]byte(source))
	return hex.EncodeToString(sum[:])
}

func downloadStyleImage(url, path string) error {
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		return nil
	}

	tmpPath := path + ".tmp"
	client := &http.Client{Timeout: 30 * time.Second}
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
