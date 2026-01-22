package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/database"
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/utils"
	"gorm.io/gorm"
)

type migrator struct {
	db            *gorm.DB
	store         *storage.LocalStorage
	baseURL       string
	basePath      string
	dryRun        bool
	includeRemote bool
	episodeCache  map[uint]uint
	sceneCache    map[uint]uint
	dramaCache    map[uint]uint
	forceLocal    bool
}

type stats struct {
	scanned int
	updated int
	skipped int
	failed  int
}

func main() {
	dryRun := flag.Bool("dry-run", false, "report changes without writing")
	includeRemote := flag.Bool("include-remote", false, "also download non-local http(s) urls")
	forceLocal := flag.Bool("force-local", false, "re-copy local urls even if already in target layout")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}

	baseURL := strings.TrimRight(cfg.Storage.BaseURL, "/")
	store, err := storage.NewLocalStorage(cfg.Storage.LocalPath, baseURL)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	m := &migrator{
		db:            db,
		store:         store,
		baseURL:       baseURL,
		basePath:      cfg.Storage.LocalPath,
		dryRun:        *dryRun,
		includeRemote: *includeRemote,
		episodeCache:  make(map[uint]uint),
		sceneCache:    make(map[uint]uint),
		dramaCache:    make(map[uint]uint),
		forceLocal:    *forceLocal,
	}

	if err := m.run(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

func (m *migrator) run() error {
	log.Printf("starting migration (dry_run=%v, include_remote=%v, force_local=%v)", m.dryRun, m.includeRemote, m.forceLocal)

	sections := []struct {
		name string
		fn   func() (stats, error)
	}{
		{"image_generations", m.migrateImageGenerations},
		{"storyboards", m.migrateStoryboards},
		{"scenes", m.migrateScenes},
		{"characters", m.migrateCharacters},
		{"character_libraries", m.migrateCharacterLibraries},
		{"episodes", m.migrateEpisodes},
		{"dramas", m.migrateDramas},
		{"video_generations", m.migrateVideoGenerations},
		{"video_merges", m.migrateVideoMerges},
		{"assets", m.migrateAssets},
	}

	for _, section := range sections {
		sectionStats, err := section.fn()
		if err != nil {
			return fmt.Errorf("%s: %w", section.name, err)
		}
		log.Printf("%s scanned=%d updated=%d skipped=%d failed=%d", section.name, sectionStats.scanned, sectionStats.updated, sectionStats.skipped, sectionStats.failed)
	}

	return nil
}

func (m *migrator) migrateImageGenerations() (stats, error) {
	var rows []models.ImageGeneration
	if err := m.db.Find(&rows).Error; err != nil {
		return stats{}, err
	}

	result := stats{}
	for _, row := range rows {
		result.scanned++
		if row.ImageURL == nil || *row.ImageURL == "" {
			result.skipped++
			continue
		}

		episodeID, storyboardID := m.resolveImageContext(&row)
		category := buildCategory("images", row.DramaID, episodeID, storyboardID)
		newURL, updated, err := m.migrateURL(*row.ImageURL, category)
		if err != nil {
			result.failed++
			log.Printf("image_generations id=%d migrate failed: %v", row.ID, err)
			continue
		}
		if !updated {
			result.skipped++
			continue
		}
		if !m.dryRun {
			if err := m.db.Model(&models.ImageGeneration{}).Where("id = ?", row.ID).Update("image_url", newURL).Error; err != nil {
				result.failed++
				log.Printf("image_generations id=%d update failed: %v", row.ID, err)
				continue
			}
		}
		result.updated++
	}

	return result, nil
}

func (m *migrator) migrateStoryboards() (stats, error) {
	var rows []models.Storyboard
	if err := m.db.Find(&rows).Error; err != nil {
		return stats{}, err
	}

	result := stats{}
	for _, row := range rows {
		result.scanned++
		dramaID := m.dramaIDForEpisode(row.EpisodeID)
		imageCategory := buildCategory("images", dramaID, row.EpisodeID, row.ID)
		videoCategory := buildCategory("videos", dramaID, row.EpisodeID, row.ID)

		updatedAny := false
		if row.ComposedImage != nil && *row.ComposedImage != "" {
			newURL, updated, err := m.migrateURL(*row.ComposedImage, imageCategory)
			if err != nil {
				result.failed++
				log.Printf("storyboards id=%d composed_image migrate failed: %v", row.ID, err)
			} else if updated {
				updatedAny = true
				if !m.dryRun {
					if err := m.db.Model(&models.Storyboard{}).Where("id = ?", row.ID).Update("composed_image", newURL).Error; err != nil {
						result.failed++
						log.Printf("storyboards id=%d composed_image update failed: %v", row.ID, err)
					} else {
						result.updated++
					}
				} else {
					result.updated++
				}
			}
		}

		if row.VideoURL != nil && *row.VideoURL != "" {
			newURL, updated, err := m.migrateURL(*row.VideoURL, videoCategory)
			if err != nil {
				result.failed++
				log.Printf("storyboards id=%d video_url migrate failed: %v", row.ID, err)
			} else if updated {
				updatedAny = true
				if !m.dryRun {
					if err := m.db.Model(&models.Storyboard{}).Where("id = ?", row.ID).Update("video_url", newURL).Error; err != nil {
						result.failed++
						log.Printf("storyboards id=%d video_url update failed: %v", row.ID, err)
					} else {
						result.updated++
					}
				} else {
					result.updated++
				}
			}
		}

		if !updatedAny {
			result.skipped++
		}
	}

	return result, nil
}

func (m *migrator) migrateScenes() (stats, error) {
	var rows []models.Scene
	if err := m.db.Find(&rows).Error; err != nil {
		return stats{}, err
	}

	result := stats{}
	for _, row := range rows {
		result.scanned++
		if row.ImageURL == nil || *row.ImageURL == "" {
			result.skipped++
			continue
		}
		var episodeID uint
		if row.EpisodeID != nil {
			episodeID = *row.EpisodeID
		}
		category := buildCategory("images", row.DramaID, episodeID, 0)
		newURL, updated, err := m.migrateURL(*row.ImageURL, category)
		if err != nil {
			result.failed++
			log.Printf("scenes id=%d migrate failed: %v", row.ID, err)
			continue
		}
		if !updated {
			result.skipped++
			continue
		}
		if !m.dryRun {
			if err := m.db.Model(&models.Scene{}).Where("id = ?", row.ID).Update("image_url", newURL).Error; err != nil {
				result.failed++
				log.Printf("scenes id=%d update failed: %v", row.ID, err)
				continue
			}
		}
		result.updated++
	}

	return result, nil
}

func (m *migrator) migrateCharacters() (stats, error) {
	var rows []models.Character
	if err := m.db.Find(&rows).Error; err != nil {
		return stats{}, err
	}

	result := stats{}
	for _, row := range rows {
		result.scanned++
		if row.ImageURL == nil || *row.ImageURL == "" {
			result.skipped++
			continue
		}
		category := buildCategory("images", row.DramaID, 0, 0)
		newURL, updated, err := m.migrateURL(*row.ImageURL, category)
		if err != nil {
			result.failed++
			log.Printf("characters id=%d migrate failed: %v", row.ID, err)
			continue
		}
		if !updated {
			result.skipped++
			continue
		}
		if !m.dryRun {
			if err := m.db.Model(&models.Character{}).Where("id = ?", row.ID).Update("image_url", newURL).Error; err != nil {
				result.failed++
				log.Printf("characters id=%d update failed: %v", row.ID, err)
				continue
			}
		}
		result.updated++
	}

	return result, nil
}

func (m *migrator) migrateCharacterLibraries() (stats, error) {
	var rows []models.CharacterLibrary
	if err := m.db.Find(&rows).Error; err != nil {
		return stats{}, err
	}

	result := stats{}
	for _, row := range rows {
		result.scanned++
		if row.ImageURL == "" {
			result.skipped++
			continue
		}
		category := buildCategory("images", 0, 0, 0)
		newURL, updated, err := m.migrateURL(row.ImageURL, category)
		if err != nil {
			result.failed++
			log.Printf("character_libraries id=%d migrate failed: %v", row.ID, err)
			continue
		}
		if !updated {
			result.skipped++
			continue
		}
		if !m.dryRun {
			if err := m.db.Model(&models.CharacterLibrary{}).Where("id = ?", row.ID).Update("image_url", newURL).Error; err != nil {
				result.failed++
				log.Printf("character_libraries id=%d update failed: %v", row.ID, err)
				continue
			}
		}
		result.updated++
	}

	return result, nil
}

func (m *migrator) migrateEpisodes() (stats, error) {
	var rows []models.Episode
	if err := m.db.Find(&rows).Error; err != nil {
		return stats{}, err
	}

	result := stats{}
	for _, row := range rows {
		result.scanned++
		updatedAny := false
		imageCategory := buildCategory("images", row.DramaID, row.ID, 0)
		videoCategory := buildCategory("videos", row.DramaID, row.ID, 0)

		if row.Thumbnail != nil && *row.Thumbnail != "" {
			newURL, updated, err := m.migrateURL(*row.Thumbnail, imageCategory)
			if err != nil {
				result.failed++
				log.Printf("episodes id=%d thumbnail migrate failed: %v", row.ID, err)
			} else if updated {
				updatedAny = true
				if !m.dryRun {
					if err := m.db.Model(&models.Episode{}).Where("id = ?", row.ID).Update("thumbnail", newURL).Error; err != nil {
						result.failed++
						log.Printf("episodes id=%d thumbnail update failed: %v", row.ID, err)
					} else {
						result.updated++
					}
				} else {
					result.updated++
				}
			}
		}

		if row.VideoURL != nil && *row.VideoURL != "" {
			newURL, updated, err := m.migrateURL(*row.VideoURL, videoCategory)
			if err != nil {
				result.failed++
				log.Printf("episodes id=%d video_url migrate failed: %v", row.ID, err)
			} else if updated {
				updatedAny = true
				if !m.dryRun {
					if err := m.db.Model(&models.Episode{}).Where("id = ?", row.ID).Update("video_url", newURL).Error; err != nil {
						result.failed++
						log.Printf("episodes id=%d video_url update failed: %v", row.ID, err)
					} else {
						result.updated++
					}
				} else {
					result.updated++
				}
			}
		}

		if !updatedAny {
			result.skipped++
		}
	}

	return result, nil
}

func (m *migrator) migrateDramas() (stats, error) {
	var rows []models.Drama
	if err := m.db.Find(&rows).Error; err != nil {
		return stats{}, err
	}

	result := stats{}
	for _, row := range rows {
		result.scanned++
		if row.Thumbnail == nil || *row.Thumbnail == "" {
			result.skipped++
			continue
		}
		category := buildCategory("images", row.ID, 0, 0)
		newURL, updated, err := m.migrateURL(*row.Thumbnail, category)
		if err != nil {
			result.failed++
			log.Printf("dramas id=%d migrate failed: %v", row.ID, err)
			continue
		}
		if !updated {
			result.skipped++
			continue
		}
		if !m.dryRun {
			if err := m.db.Model(&models.Drama{}).Where("id = ?", row.ID).Update("thumbnail", newURL).Error; err != nil {
				result.failed++
				log.Printf("dramas id=%d update failed: %v", row.ID, err)
				continue
			}
		}
		result.updated++
	}

	return result, nil
}

func (m *migrator) migrateVideoGenerations() (stats, error) {
	var rows []models.VideoGeneration
	if err := m.db.Find(&rows).Error; err != nil {
		return stats{}, err
	}

	result := stats{}
	for _, row := range rows {
		result.scanned++
		updatedAny := false
		episodeID := m.episodeIDForStoryboardPtr(row.StoryboardID)
		videoCategory := buildCategory("videos", row.DramaID, episodeID, derefUint(row.StoryboardID))
		frameCategory := buildCategory("video_frames", row.DramaID, episodeID, derefUint(row.StoryboardID))

		if row.VideoURL != nil && *row.VideoURL != "" {
			newURL, updated, err := m.migrateURL(*row.VideoURL, videoCategory)
			if err != nil {
				result.failed++
				log.Printf("video_generations id=%d video_url migrate failed: %v", row.ID, err)
			} else if updated {
				updatedAny = true
				if !m.dryRun {
					if err := m.db.Model(&models.VideoGeneration{}).Where("id = ?", row.ID).Update("video_url", newURL).Error; err != nil {
						result.failed++
						log.Printf("video_generations id=%d video_url update failed: %v", row.ID, err)
					} else {
						result.updated++
					}
				} else {
					result.updated++
				}
			}
		}

		if row.FirstFrameURL != nil && *row.FirstFrameURL != "" {
			newURL, updated, err := m.migrateURL(*row.FirstFrameURL, frameCategory)
			if err != nil {
				result.failed++
				log.Printf("video_generations id=%d first_frame_url migrate failed: %v", row.ID, err)
			} else if updated {
				updatedAny = true
				if !m.dryRun {
					if err := m.db.Model(&models.VideoGeneration{}).Where("id = ?", row.ID).Update("first_frame_url", newURL).Error; err != nil {
						result.failed++
						log.Printf("video_generations id=%d first_frame_url update failed: %v", row.ID, err)
					} else {
						result.updated++
					}
				} else {
					result.updated++
				}
			}
		}

		if row.LastFrameURL != nil && *row.LastFrameURL != "" {
			newURL, updated, err := m.migrateURL(*row.LastFrameURL, frameCategory)
			if err != nil {
				result.failed++
				log.Printf("video_generations id=%d last_frame_url migrate failed: %v", row.ID, err)
			} else if updated {
				updatedAny = true
				if !m.dryRun {
					if err := m.db.Model(&models.VideoGeneration{}).Where("id = ?", row.ID).Update("last_frame_url", newURL).Error; err != nil {
						result.failed++
						log.Printf("video_generations id=%d last_frame_url update failed: %v", row.ID, err)
					} else {
						result.updated++
					}
				} else {
					result.updated++
				}
			}
		}

		if !updatedAny {
			result.skipped++
		}
	}

	return result, nil
}

func (m *migrator) migrateVideoMerges() (stats, error) {
	var rows []models.VideoMerge
	if err := m.db.Find(&rows).Error; err != nil {
		return stats{}, err
	}

	result := stats{}
	for _, row := range rows {
		result.scanned++
		if row.MergedURL == nil || *row.MergedURL == "" {
			result.skipped++
			continue
		}
		category := buildCategory("videos", row.DramaID, row.EpisodeID, 0)
		newURL, updated, err := m.migrateURL(*row.MergedURL, category)
		if err != nil {
			result.failed++
			log.Printf("video_merges id=%d migrate failed: %v", row.ID, err)
			continue
		}
		if !updated {
			result.skipped++
			continue
		}
		if !m.dryRun {
			if err := m.db.Model(&models.VideoMerge{}).Where("id = ?", row.ID).Update("merged_url", newURL).Error; err != nil {
				result.failed++
				log.Printf("video_merges id=%d update failed: %v", row.ID, err)
				continue
			}
		}
		result.updated++
	}

	return result, nil
}

func (m *migrator) migrateAssets() (stats, error) {
	var rows []models.Asset
	if err := m.db.Find(&rows).Error; err != nil {
		return stats{}, err
	}

	result := stats{}
	for _, row := range rows {
		result.scanned++

		base := ""
		switch row.Type {
		case models.AssetTypeImage:
			base = "images"
		case models.AssetTypeVideo:
			base = "videos"
		default:
			result.skipped++
			continue
		}

		dramaID := derefUint(row.DramaID)
		episodeID := derefUint(row.EpisodeID)
		storyboardID := derefUint(row.StoryboardID)
		category := buildCategory(base, dramaID, episodeID, storyboardID)

		updatedAny := false
		if row.URL != "" {
			newURL, updated, err := m.migrateURL(row.URL, category)
			if err != nil {
				result.failed++
				log.Printf("assets id=%d url migrate failed: %v", row.ID, err)
			} else if updated {
				updatedAny = true
				if !m.dryRun {
					if err := m.db.Model(&models.Asset{}).Where("id = ?", row.ID).Update("url", newURL).Error; err != nil {
						result.failed++
						log.Printf("assets id=%d url update failed: %v", row.ID, err)
					} else {
						result.updated++
					}
				} else {
					result.updated++
				}
			}
		}

		if row.ThumbnailURL != nil && *row.ThumbnailURL != "" {
			thumbCategory := buildCategory("images", dramaID, episodeID, storyboardID)
			newURL, updated, err := m.migrateURL(*row.ThumbnailURL, thumbCategory)
			if err != nil {
				result.failed++
				log.Printf("assets id=%d thumbnail migrate failed: %v", row.ID, err)
			} else if updated {
				updatedAny = true
				if !m.dryRun {
					if err := m.db.Model(&models.Asset{}).Where("id = ?", row.ID).Update("thumbnail_url", newURL).Error; err != nil {
						result.failed++
						log.Printf("assets id=%d thumbnail update failed: %v", row.ID, err)
					} else {
						result.updated++
					}
				} else {
					result.updated++
				}
			}
		}

		if !updatedAny {
			result.skipped++
		}
	}

	return result, nil
}

func (m *migrator) resolveImageContext(imageGen *models.ImageGeneration) (uint, uint) {
	if imageGen == nil {
		return 0, 0
	}
	if imageGen.StoryboardID != nil {
		storyboardID := *imageGen.StoryboardID
		return m.episodeIDForStoryboard(storyboardID), storyboardID
	}
	if imageGen.SceneID != nil {
		return m.episodeIDForScene(*imageGen.SceneID), 0
	}
	return 0, 0
}

func (m *migrator) episodeIDForStoryboardPtr(storyboardID *uint) uint {
	if storyboardID == nil {
		return 0
	}
	return m.episodeIDForStoryboard(*storyboardID)
}

func (m *migrator) episodeIDForStoryboard(storyboardID uint) uint {
	if storyboardID == 0 {
		return 0
	}
	if cached, ok := m.episodeCache[storyboardID]; ok {
		return cached
	}
	var row models.Storyboard
	if err := m.db.Select("episode_id").Where("id = ?", storyboardID).First(&row).Error; err != nil {
		m.episodeCache[storyboardID] = 0
		return 0
	}
	m.episodeCache[storyboardID] = row.EpisodeID
	return row.EpisodeID
}

func (m *migrator) episodeIDForScene(sceneID uint) uint {
	if sceneID == 0 {
		return 0
	}
	if cached, ok := m.sceneCache[sceneID]; ok {
		return cached
	}
	var row models.Scene
	if err := m.db.Select("episode_id").Where("id = ?", sceneID).First(&row).Error; err != nil {
		m.sceneCache[sceneID] = 0
		return 0
	}
	if row.EpisodeID == nil {
		m.sceneCache[sceneID] = 0
		return 0
	}
	m.sceneCache[sceneID] = *row.EpisodeID
	return *row.EpisodeID
}

func (m *migrator) dramaIDForEpisode(episodeID uint) uint {
	if episodeID == 0 {
		return 0
	}
	if cached, ok := m.dramaCache[episodeID]; ok {
		return cached
	}
	var row models.Episode
	if err := m.db.Select("drama_id").Where("id = ?", episodeID).First(&row).Error; err != nil {
		m.dramaCache[episodeID] = 0
		return 0
	}
	m.dramaCache[episodeID] = row.DramaID
	return row.DramaID
}

func (m *migrator) migrateURL(value string, category string) (string, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false, nil
	}

	if utils.IsDataURI(value) {
		data, mimeType, err := utils.ParseDataURI(value)
		if err != nil {
			return "", false, err
		}
		if m.dryRun {
			return m.fakeURL(category, ""), true, nil
		}
		newURL, err := m.store.UploadBytes(data, mimeType, category)
		if err != nil {
			return "", false, err
		}
		return newURL, true, nil
	}

	if strings.HasPrefix(value, m.baseURL+"/") {
		relPath := strings.TrimPrefix(value, m.baseURL+"/")
		if strings.HasPrefix(relPath, category+"/") && !m.forceLocal {
			return value, false, nil
		}
		if m.dryRun {
			return m.fakeURL(category, filepath.Ext(relPath)), true, nil
		}
		newURL, err := m.copyLocalFile(relPath, category)
		if err != nil {
			return "", false, err
		}
		return newURL, true, nil
	}

	if (strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")) && m.includeRemote {
		if m.dryRun {
			return m.fakeURL(category, ""), true, nil
		}
		newURL, err := m.store.DownloadFromURL(value, category)
		if err != nil {
			return "", false, err
		}
		return newURL, true, nil
	}

	return value, false, nil
}

func (m *migrator) copyLocalFile(relPath string, category string) (string, error) {
	srcPath := filepath.Join(m.basePath, filepath.FromSlash(relPath))
	file, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("open source file: %w", err)
	}
	defer file.Close()

	ext := filepath.Ext(srcPath)
	if ext == "" {
		ext = ".bin"
	}

	filename, err := newFilename(ext)
	if err != nil {
		return "", err
	}

	return m.store.Upload(file, filename, category)
}

func (m *migrator) fakeURL(category string, ext string) string {
	if ext == "" {
		ext = ".bin"
	}
	filename := "dry_run" + ext
	return fmt.Sprintf("%s/%s/%s", m.baseURL, category, filename)
}

func buildCategory(base string, dramaID uint, episodeID uint, storyboardID uint) string {
	return fmt.Sprintf("%s/dramas/%d/episodes/%d/storyboards/%d", base, dramaID, episodeID, storyboardID)
}

func newFilename(ext string) (string, error) {
	id, err := utils.NewRandomID()
	if err != nil {
		return "", err
	}
	if ext == "" {
		ext = ".bin"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return id + ext, nil
}

func derefUint(value *uint) uint {
	if value == nil {
		return 0
	}
	return *value
}
