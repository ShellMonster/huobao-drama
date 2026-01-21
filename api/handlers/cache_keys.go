package handlers

import "time"

const (
	cacheTTLDramaList   = 30 * time.Second
	cacheTTLDramaDetail = 15 * time.Second
	cacheTTLDramaStats  = 30 * time.Second
	cacheTTLAIConfig    = 30 * time.Second
	cacheTTLImageList   = 20 * time.Second
	cacheTTLVideoList   = 20 * time.Second
	cacheTTLImageDetail = 15 * time.Second
	cacheTTLVideoDetail = 15 * time.Second
	cacheTTLStoryboards = 15 * time.Second
)
