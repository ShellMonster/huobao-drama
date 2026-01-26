package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func NewDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()

	if cfg.Type == "sqlite" {
		dbDir := filepath.Dir(dsn)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	gormConfig := &gorm.Config{
		Logger: NewCustomLogger(),
	}

	var db *gorm.DB
	var err error

	if cfg.Type == "sqlite" {
		// 使用 modernc.org/sqlite 纯 Go 驱动（无需 CGO）
		// 添加并发优化参数：WAL 模式、busy_timeout、cache
		dsnWithParams := buildSQLiteDSN(dsn, cfg)
		db, err = gorm.Open(sqlite.Dialector{
			DriverName: "sqlite",
			DSN:        dsnWithParams,
		}, gormConfig)
	} else {
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// SQLite 连接池配置（限制并发连接数）
	if cfg.Type == "sqlite" {
		maxOpen := cfg.MaxOpen
		if maxOpen < 2 {
			maxOpen = 2
		} else if maxOpen > 8 {
			maxOpen = 8
		}

		maxIdle := cfg.MaxIdle
		if maxIdle < 1 {
			maxIdle = 1
		} else if maxIdle > maxOpen {
			maxIdle = maxOpen
		}

		sqlDB.SetMaxIdleConns(maxIdle)
		sqlDB.SetMaxOpenConns(maxOpen)
	} else {
		sqlDB.SetMaxIdleConns(cfg.MaxIdle)
		sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	}
	connMaxLifetime := config.DurationFromSeconds(cfg.ConnMaxLifetimeSeconds, time.Hour)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func buildSQLiteDSN(dsn string, cfg config.DatabaseConfig) string {
	base := dsn
	if !strings.HasPrefix(base, "file:") {
		base = "file:" + base
	}

	busyTimeout := cfg.SQLiteBusyTimeoutMs
	if busyTimeout <= 0 {
		busyTimeout = 5000
	}

	params := []string{
		"_pragma=journal_mode(WAL)",
		fmt.Sprintf("_pragma=busy_timeout(%d)", busyTimeout),
		"_pragma=synchronous(NORMAL)",
		"cache=shared",
	}

	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}

	return base + sep + strings.Join(params, "&")
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// 核心模型
		&models.Drama{},
		&models.Brand{},
		&models.BrandSpec{},
		&models.Episode{},
		&models.Character{},
		&models.Scene{},
		&models.Storyboard{},
		&models.FramePrompt{},
		&models.FramePromptTask{},

		// 生成相关
		&models.ImageGeneration{},
		&models.VideoGeneration{},
		&models.VideoMerge{},
		&models.AdImagePrompt{},
		&models.AdImagePromptItem{},

		// 风格配置
		&models.Style{},

		// AI配置
		&models.AIServiceConfig{},
		&models.AIServiceProvider{},

		// 资源管理
		&models.Asset{},
		&models.CharacterLibrary{},

		// 任务管理
		&models.AsyncTask{},
	)
}
