package video

import (
	"strings"

	"github.com/drama-generator/backend/pkg/config"
)

func defaultVideoDuration(fallback int) int {
	value := config.GetTuning().Defaults.Video.DurationSeconds
	if value > 0 {
		return value
	}
	return fallback
}

func defaultVideoAspectRatio(fallback string) string {
	value := strings.TrimSpace(config.GetTuning().Defaults.Video.AspectRatio)
	if value != "" {
		return value
	}
	return fallback
}

func defaultVideoResolution(fallback string) string {
	value := strings.TrimSpace(config.GetTuning().Defaults.Video.Resolution)
	if value != "" {
		return value
	}
	return fallback
}

func defaultVideoMotionLevel(fallback int) int {
	value := config.GetTuning().Defaults.Video.MotionLevel
	if value > 0 {
		return value
	}
	return fallback
}
