package handlers

import (
	"net/http"
	"strconv"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/cache"
	"github.com/drama-generator/backend/pkg/events"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VideoGenerationHandler struct {
	videoService *services.VideoGenerationService
	log          *logger.Logger
}

func NewVideoGenerationHandler(db *gorm.DB, transferService *services.ResourceTransferService, localStorage *storage.LocalStorage, aiService *services.AIService, log *logger.Logger, hub *events.VideoGenerationHub) *VideoGenerationHandler {
	return &VideoGenerationHandler{
		videoService: services.NewVideoGenerationService(db, transferService, localStorage, aiService, log, hub),
		log:          log,
	}
}

func (h *VideoGenerationHandler) GenerateVideo(c *gin.Context) {

	var req services.GenerateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	videoGen, err := h.videoService.GenerateVideo(&req)
	if err != nil {
		h.log.Errorw("Failed to generate video", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, videoGen)
}

func (h *VideoGenerationHandler) GenerateVideoFromImage(c *gin.Context) {

	imageGenID, err := strconv.ParseUint(c.Param("image_gen_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的图片ID")
		return
	}

	videoGen, err := h.videoService.GenerateVideoFromImage(uint(imageGenID))
	if err != nil {
		h.log.Errorw("Failed to generate video from image", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, videoGen)
}

func (h *VideoGenerationHandler) BatchGenerateForEpisode(c *gin.Context) {

	episodeID := c.Param("episode_id")

	videos, err := h.videoService.BatchGenerateVideosForEpisode(episodeID)
	if err != nil {
		h.log.Errorw("Failed to batch generate videos", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, videos)
}

func (h *VideoGenerationHandler) GetVideoGeneration(c *gin.Context) {

	videoGenID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	cacheKey := cache.NamespaceKey(cache.NamespaceVideoDetail, c.Param("id"))
	if entry, ok := cache.Get(cacheKey); ok {
		c.Header("ETag", entry.ETag)
		c.Header("Cache-Control", "private, max-age=0, must-revalidate")
		if cache.MatchETag(c.GetHeader("If-None-Match"), entry.ETag) {
			c.Status(http.StatusNotModified)
			return
		}
		response.Success(c, entry.Data)
		return
	}

	videoGen, err := h.videoService.GetVideoGeneration(uint(videoGenID))
	if err != nil {
		response.NotFound(c, "视频生成记录不存在")
		return
	}

	if entry, err := cache.Set(cacheKey, videoGen, cacheTTLVideoDetail); err == nil {
		c.Header("ETag", entry.ETag)
		c.Header("Cache-Control", "private, max-age=0, must-revalidate")
	}
	response.Success(c, videoGen)
}

func (h *VideoGenerationHandler) ListVideoGenerations(c *gin.Context) {
	var storyboardID *uint
	// 优先使用storyboard_id参数
	if storyboardIDStr := c.Query("storyboard_id"); storyboardIDStr != "" {
		id, err := strconv.ParseUint(storyboardIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			storyboardID = &uid
		}
	}
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var dramaIDUint *uint
	if dramaIDStr := c.Query("drama_id"); dramaIDStr != "" {
		did, err := strconv.ParseUint(dramaIDStr, 10, 32)
		if err == nil {
			didUint := uint(did)
			dramaIDUint = &didUint
		}
	}

	normalizedQuery := c.Request.URL.Query()
	normalizedQuery.Set("page", strconv.Itoa(page))
	normalizedQuery.Set("page_size", strconv.Itoa(pageSize))
	if storyboardID != nil {
		normalizedQuery.Set("storyboard_id", strconv.FormatUint(uint64(*storyboardID), 10))
	} else {
		normalizedQuery.Del("storyboard_id")
	}
	if dramaIDUint != nil {
		normalizedQuery.Set("drama_id", strconv.FormatUint(uint64(*dramaIDUint), 10))
	} else {
		normalizedQuery.Del("drama_id")
	}

	cacheKey := cache.NamespaceKeyWithQuery(cache.NamespaceVideoList, normalizedQuery)
	if entry, ok := cache.Get(cacheKey); ok {
		c.Header("ETag", entry.ETag)
		c.Header("Cache-Control", "private, max-age=0, must-revalidate")
		if cache.MatchETag(c.GetHeader("If-None-Match"), entry.ETag) {
			c.Status(http.StatusNotModified)
			return
		}
		response.Success(c, entry.Data)
		return
	}

	// 计算offset：(page - 1) * pageSize
	offset := (page - 1) * pageSize
	videos, total, err := h.videoService.ListVideoGenerations(dramaIDUint, storyboardID, status, pageSize, offset)

	if err != nil {
		h.log.Errorw("Failed to list videos", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	payload := response.PaginationData{
		Items: videos,
		Pagination: response.Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
	if entry, err := cache.Set(cacheKey, payload, cacheTTLVideoList); err == nil {
		c.Header("ETag", entry.ETag)
		c.Header("Cache-Control", "private, max-age=0, must-revalidate")
	}
	response.Success(c, payload)
}

func (h *VideoGenerationHandler) DeleteVideoGeneration(c *gin.Context) {

	videoGenID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.videoService.DeleteVideoGeneration(uint(videoGenID)); err != nil {
		h.log.Errorw("Failed to delete video", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
