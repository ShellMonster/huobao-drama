package image

import (
	"strings"

	"github.com/drama-generator/backend/pkg/config"
)

func defaultImageSize(fallback string) string {
	value := strings.TrimSpace(config.GetTuning().Defaults.Image.Size)
	if value != "" {
		return value
	}
	return fallback
}

func defaultImageQuality(fallback string) string {
	value := strings.TrimSpace(config.GetTuning().Defaults.Image.Quality)
	if value != "" {
		return value
	}
	return fallback
}
