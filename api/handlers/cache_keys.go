package handlers

import (
	"time"

	"github.com/drama-generator/backend/pkg/config"
)

var (
	cacheTTLDramaList   = 30 * time.Second
	cacheTTLDramaDetail = 15 * time.Second
	cacheTTLDramaStats  = 30 * time.Second
	cacheTTLAIConfig    = 30 * time.Second
	cacheTTLImageList   = 20 * time.Second
	cacheTTLVideoList   = 20 * time.Second
	cacheTTLImageDetail = 15 * time.Second
	cacheTTLVideoDetail = 15 * time.Second
	cacheTTLStoryboards = 15 * time.Second
	cacheTTLAdPrompts   = 20 * time.Second
	cacheTTLBrandList   = 30 * time.Second
	cacheTTLBrandDetail = 20 * time.Second
	cacheTTLBrandSpecs  = 30 * time.Second
	cacheTTLAssetList   = 20 * time.Second
	cacheTTLAssetDetail = 15 * time.Second
	cacheTTLCharLibList = 20 * time.Second
	cacheTTLCharLibItem = 15 * time.Second
)

func ApplyCacheConfig() {
	tuning := config.GetTuning()
	cacheTTLDramaList = config.DurationFromSeconds(tuning.Cache.DramaListSeconds, 30*time.Second)
	cacheTTLDramaDetail = config.DurationFromSeconds(tuning.Cache.DramaDetailSeconds, 15*time.Second)
	cacheTTLDramaStats = config.DurationFromSeconds(tuning.Cache.DramaStatsSeconds, 30*time.Second)
	cacheTTLAIConfig = config.DurationFromSeconds(tuning.Cache.AIConfigSeconds, 30*time.Second)
	cacheTTLImageList = config.DurationFromSeconds(tuning.Cache.ImageListSeconds, 20*time.Second)
	cacheTTLVideoList = config.DurationFromSeconds(tuning.Cache.VideoListSeconds, 20*time.Second)
	cacheTTLImageDetail = config.DurationFromSeconds(tuning.Cache.ImageDetailSeconds, 15*time.Second)
	cacheTTLVideoDetail = config.DurationFromSeconds(tuning.Cache.VideoDetailSeconds, 15*time.Second)
	cacheTTLStoryboards = config.DurationFromSeconds(tuning.Cache.StoryboardsSeconds, 15*time.Second)
	cacheTTLAdPrompts = config.DurationFromSeconds(tuning.Cache.AdPromptLatestSeconds, 20*time.Second)
	cacheTTLBrandList = config.DurationFromSeconds(tuning.Cache.BrandListSeconds, 30*time.Second)
	cacheTTLBrandDetail = config.DurationFromSeconds(tuning.Cache.BrandDetailSeconds, 20*time.Second)
	cacheTTLBrandSpecs = config.DurationFromSeconds(tuning.Cache.BrandSpecsSeconds, 30*time.Second)
	cacheTTLAssetList = config.DurationFromSeconds(tuning.Cache.AssetListSeconds, 20*time.Second)
	cacheTTLAssetDetail = config.DurationFromSeconds(tuning.Cache.AssetDetailSeconds, 15*time.Second)
	cacheTTLCharLibList = config.DurationFromSeconds(tuning.Cache.CharLibListSeconds, 20*time.Second)
	cacheTTLCharLibItem = config.DurationFromSeconds(tuning.Cache.CharLibDetailSeconds, 15*time.Second)
}
