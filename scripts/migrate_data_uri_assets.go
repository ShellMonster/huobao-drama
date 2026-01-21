package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/drama-generator/backend/infrastructure/database"
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/utils"
	"gorm.io/gorm"
)

type migrateStats struct {
	Scanned  int
	Migrated int
	Failed   int
}

func main() {
	dryRun := flag.Bool("dry-run", false, "report only, do not write changes")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		panic(fmt.Errorf("open database: %w", err))
	}

	localStorage, err := storage.NewLocalStorage(cfg.Storage.LocalPath, cfg.Storage.BaseURL)
	if err != nil {
		panic(fmt.Errorf("init local storage: %w", err))
	}

	stats := migrateStats{}

	migrateTable(db, localStorage, "image_generations", "image_url", "images", dryRun, &stats)
	migrateTable(db, localStorage, "storyboards", "composed_image", "storyboards", dryRun, &stats)
	migrateTable(db, localStorage, "scenes", "image_url", "scenes", dryRun, &stats)
	migrateTable(db, localStorage, "characters", "image_url", "characters", dryRun, &stats)
	migrateTable(db, localStorage, "character_libraries", "image_url", "character_library", dryRun, &stats)
	migrateTable(db, localStorage, "dramas", "thumbnail", "thumbnails", dryRun, &stats)
	migrateTable(db, localStorage, "episodes", "thumbnail", "thumbnails", dryRun, &stats)

	migrateTable(db, localStorage, "video_generations", "video_url", "videos", dryRun, &stats)
	migrateTable(db, localStorage, "storyboards", "video_url", "videos", dryRun, &stats)
	migrateTable(db, localStorage, "episodes", "video_url", "videos", dryRun, &stats)
	migrateTable(db, localStorage, "video_merges", "merged_url", "videos", dryRun, &stats)

	migrateAssetURLs(db, localStorage, dryRun, &stats)

	fmt.Printf("Scan completed. scanned=%d migrated=%d failed=%d\n", stats.Scanned, stats.Migrated, stats.Failed)
	if *dryRun {
		fmt.Println("Dry-run mode enabled: no changes were written.")
	}
}

func migrateTable(db *gorm.DB, localStorage *storage.LocalStorage, table, column, category string, dryRun *bool, stats *migrateStats) {
	rows := []struct {
		ID  uint   `gorm:"column:id"`
		URL string `gorm:"column:url"`
	}{}

	query := fmt.Sprintf("%s LIKE ?", column)
	if err := db.Table(table).
		Select(fmt.Sprintf("id, %s as url", column)).
		Where(query, "data:%").
		Find(&rows).Error; err != nil {
		fmt.Printf("[WARN] query %s.%s failed: %v\n", table, column, err)
		stats.Failed++
		return
	}

	for _, row := range rows {
		stats.Scanned++
		newURL, err := storeDataURI(localStorage, row.URL, category)
		if err != nil {
			fmt.Printf("[WARN] %s.%s id=%d migrate failed: %v\n", table, column, row.ID, err)
			stats.Failed++
			continue
		}
		if *dryRun {
			stats.Migrated++
			continue
		}
		if err := db.Table(table).Where("id = ?", row.ID).Update(column, newURL).Error; err != nil {
			fmt.Printf("[WARN] %s.%s id=%d update failed: %v\n", table, column, row.ID, err)
			stats.Failed++
			continue
		}
		stats.Migrated++
	}
}

func migrateAssetURLs(db *gorm.DB, localStorage *storage.LocalStorage, dryRun *bool, stats *migrateStats) {
	migrateAssetColumn(db, localStorage, "url", dryRun, stats)
	migrateAssetColumn(db, localStorage, "thumbnail_url", dryRun, stats)
}

func migrateAssetColumn(db *gorm.DB, localStorage *storage.LocalStorage, column string, dryRun *bool, stats *migrateStats) {
	rows := []struct {
		ID  uint   `gorm:"column:id"`
		URL string `gorm:"column:url"`
	}{}

	query := fmt.Sprintf("%s LIKE ?", column)
	if err := db.Table("assets").
		Select(fmt.Sprintf("id, %s as url", column)).
		Where(query, "data:%").
		Find(&rows).Error; err != nil {
		fmt.Printf("[WARN] query assets.%s failed: %v\n", column, err)
		stats.Failed++
		return
	}

	for _, row := range rows {
		stats.Scanned++
		newURL, err := storeDataURIWithCategory(localStorage, row.URL)
		if err != nil {
			fmt.Printf("[WARN] assets.%s id=%d migrate failed: %v\n", column, row.ID, err)
			stats.Failed++
			continue
		}
		if *dryRun {
			stats.Migrated++
			continue
		}
		if err := db.Table("assets").Where("id = ?", row.ID).Update(column, newURL).Error; err != nil {
			fmt.Printf("[WARN] assets.%s id=%d update failed: %v\n", column, row.ID, err)
			stats.Failed++
			continue
		}
		stats.Migrated++
	}
}

func storeDataURI(localStorage *storage.LocalStorage, dataURI string, category string) (string, error) {
	data, mimeType, err := utils.ParseDataURI(dataURI)
	if err != nil {
		return "", err
	}
	return localStorage.UploadBytes(data, mimeType, category)
}

func storeDataURIWithCategory(localStorage *storage.LocalStorage, dataURI string) (string, error) {
	data, mimeType, err := utils.ParseDataURI(dataURI)
	if err != nil {
		return "", err
	}
	category := "assets"
	if strings.HasPrefix(mimeType, "image/") {
		category = "images"
	} else if strings.HasPrefix(mimeType, "video/") {
		category = "videos"
	}
	return localStorage.UploadBytes(data, mimeType, category)
}
