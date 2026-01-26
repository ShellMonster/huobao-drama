package config

import "time"

// Current 保存当前加载的全局配置（用于在深层包中读取调参项）。
var Current *Config

type TuningConfig struct {
	AIRetry     AIRetryConfig     `mapstructure:"ai_retry"`
	Polling     PollingConfig     `mapstructure:"polling"`
	Cache       CacheConfig       `mapstructure:"cache"`
	RateLimit   RateLimitConfig   `mapstructure:"rate_limit"`
	SSE         SSEConfig         `mapstructure:"sse"`
	HTTPTimeout HTTPTimeoutConfig `mapstructure:"http_timeout"`
	Scheduler   SchedulerConfig   `mapstructure:"scheduler"`
	Defaults    DefaultsConfig    `mapstructure:"defaults"`
}

type AIRetryConfig struct {
	Attempts        int `mapstructure:"attempts"`
	DelaySeconds    int `mapstructure:"delay_seconds"`
	MaxDelaySeconds int `mapstructure:"max_delay_seconds"`
	JitterMs        int `mapstructure:"jitter_ms"`
}

type PollingConfig struct {
	Image          PollingSetting `mapstructure:"image"`
	Video          PollingSetting `mapstructure:"video"`
	VideoMerge     PollingSetting `mapstructure:"video_merge"`
	CharacterImage PollingSetting `mapstructure:"character_image"`
}

type PollingSetting struct {
	MaxAttempts     int `mapstructure:"max_attempts"`
	IntervalSeconds int `mapstructure:"interval_seconds"`
}

type CacheConfig struct {
	DramaListSeconds      int `mapstructure:"drama_list_seconds"`
	DramaDetailSeconds    int `mapstructure:"drama_detail_seconds"`
	DramaStatsSeconds     int `mapstructure:"drama_stats_seconds"`
	AIConfigSeconds       int `mapstructure:"ai_config_seconds"`
	ImageListSeconds      int `mapstructure:"image_list_seconds"`
	VideoListSeconds      int `mapstructure:"video_list_seconds"`
	ImageDetailSeconds    int `mapstructure:"image_detail_seconds"`
	VideoDetailSeconds    int `mapstructure:"video_detail_seconds"`
	StoryboardsSeconds    int `mapstructure:"storyboards_seconds"`
	StyleCatalogSeconds   int `mapstructure:"style_catalog_seconds"`
	AdPromptLatestSeconds int `mapstructure:"ad_prompt_latest_seconds"`
	BrandListSeconds      int `mapstructure:"brand_list_seconds"`
	BrandDetailSeconds    int `mapstructure:"brand_detail_seconds"`
	BrandSpecsSeconds     int `mapstructure:"brand_specs_seconds"`
	AssetListSeconds      int `mapstructure:"asset_list_seconds"`
	AssetDetailSeconds    int `mapstructure:"asset_detail_seconds"`
	CharLibListSeconds    int `mapstructure:"character_library_list_seconds"`
	CharLibDetailSeconds  int `mapstructure:"character_library_detail_seconds"`
}

type RateLimitConfig struct {
	Limit         int `mapstructure:"limit"`
	WindowSeconds int `mapstructure:"window_seconds"`
}

type SSEConfig struct {
	HeartbeatSeconds int `mapstructure:"heartbeat_seconds"`
	ChannelBuffer    int `mapstructure:"channel_buffer"`
}

type HTTPTimeoutConfig struct {
	AISeconds         int `mapstructure:"ai_seconds"`
	ImageSeconds      int `mapstructure:"image_seconds"`
	VideoSeconds      int `mapstructure:"video_seconds"`
	StyleFetchSeconds int `mapstructure:"style_fetch_seconds"`
}

type SchedulerConfig struct {
	ResourceTransfer ResourceTransferConfig `mapstructure:"resource_transfer"`
}

type ResourceTransferConfig struct {
	HourlyCron  string `mapstructure:"hourly_cron"`
	DailyCron   string `mapstructure:"daily_cron"`
	RecentHours int    `mapstructure:"recent_hours"`
	BatchLimit  int    `mapstructure:"batch_limit"`
}

type DefaultsConfig struct {
	CharacterImage CharacterImageDefaults `mapstructure:"character_image"`
	Image          ImageDefaults          `mapstructure:"image"`
	Video          VideoDefaults          `mapstructure:"video"`
	Models         ModelDefaults          `mapstructure:"models"`
}

type CharacterImageDefaults struct {
	Provider string `mapstructure:"provider"`
	Size     string `mapstructure:"size"`
	Quality  string `mapstructure:"quality"`
}

type ImageDefaults struct {
	Size    string `mapstructure:"size"`
	Quality string `mapstructure:"quality"`
}

type VideoDefaults struct {
	DurationSeconds int    `mapstructure:"duration_seconds"`
	AspectRatio     string `mapstructure:"aspect_ratio"`
	Resolution      string `mapstructure:"resolution"`
	MotionLevel     int    `mapstructure:"motion_level"`
}

type ModelDefaults struct {
	OpenAIText  string `mapstructure:"openai_text"`
	GeminiText  string `mapstructure:"gemini_text"`
	OpenAIImage string `mapstructure:"openai_image"`
	GeminiImage string `mapstructure:"gemini_image"`
}

func GetTuning() TuningConfig {
	if Current == nil {
		return TuningConfig{}
	}
	return Current.Tuning
}

func DurationFromSeconds(seconds int, fallback time.Duration) time.Duration {
	if seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func DurationFromMilliseconds(ms int, fallback time.Duration) time.Duration {
	if ms > 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return fallback
}
