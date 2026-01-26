package handlers

import (
	"strconv"
	"time"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/cache"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/events"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type imageListItem struct {
	models.ImageGeneration
	ReusePrev          bool  `json:"reuse_prev,omitempty"`
	SourceStoryboardID *uint `json:"source_storyboard_id,omitempty"`
}

type ImageGenerationHandler struct {
	imageService *services.ImageGenerationService
	taskService  *services.TaskService
	log          *logger.Logger
}

func NewImageGenerationHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger, transferService *services.ResourceTransferService, localStorage storage.Storage, taskHub *events.TaskHub, imageHub *events.ImageGenerationHub) *ImageGenerationHandler {
	return &ImageGenerationHandler{
		imageService: services.NewImageGenerationService(db, cfg, transferService, localStorage, log, imageHub),
		taskService:  services.NewTaskService(db, log, taskHub),
		log:          log,
	}
}

func (h *ImageGenerationHandler) GenerateImage(c *gin.Context) {

	var req services.GenerateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	imageGen, err := h.imageService.GenerateImage(&req)
	if err != nil {
		h.log.Errorw("Failed to generate image", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, imageGen)
}

func (h *ImageGenerationHandler) GenerateImagesForScene(c *gin.Context) {

	sceneID := c.Param("scene_id")

	images, err := h.imageService.GenerateImagesForScene(sceneID)
	if err != nil {
		h.log.Errorw("Failed to generate images for scene", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, images)
}

func (h *ImageGenerationHandler) GetBackgroundsForEpisode(c *gin.Context) {

	episodeID := c.Param("episode_id")

	backgrounds, err := h.imageService.GetScencesForEpisode(episodeID)
	if err != nil {
		h.log.Errorw("Failed to get backgrounds", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, backgrounds)
}

func (h *ImageGenerationHandler) ExtractBackgroundsForEpisode(c *gin.Context) {
	episodeID := c.Param("episode_id")

	// 接收可选的 model 参数
	var req struct {
		Model string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有提供body或者解析失败，使用空字符串（使用默认模型）
		req.Model = ""
	}

	// 创建异步任务并统一执行入口
	task, err := h.taskService.RunAsync("background_extraction", episodeID, "开始提取场景...", func(update services.TaskUpdater) (interface{}, error) {
		backgrounds, err := h.imageService.ExtractBackgroundsForEpisode(episodeID, req.Model)
		if err != nil {
			return nil, err
		}
		return gin.H{
			"backgrounds": backgrounds,
			"total":       len(backgrounds),
		}, nil
	})
	if err != nil {
		h.log.Errorw("Failed to create task", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	// 立即返回任务ID
	response.Success(c, gin.H{
		"task_id": task.ID,
		"status":  models.TaskStatusPending,
		"message": "场景提取任务已创建，正在后台处理...",
	})
}

func (h *ImageGenerationHandler) BatchGenerateForEpisode(c *gin.Context) {

	episodeID := c.Param("episode_id")

	images, err := h.imageService.BatchGenerateImagesForEpisode(episodeID)
	if err != nil {
		h.log.Errorw("Failed to batch generate images", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, images)
}

func (h *ImageGenerationHandler) GetImageGeneration(c *gin.Context) {

	imageGenID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	cacheKey := cache.NamespaceKey(cache.NamespaceImageDetail, c.Param("id"))
	if entry, ok := tryServeCached(c, cacheKey); ok {
		if entry != nil {
			response.Success(c, entry.Data)
		}
		return
	}

	imageGen, err := h.imageService.GetImageGeneration(uint(imageGenID))
	if err != nil {
		response.NotFound(c, "图片生成记录不存在")
		return
	}

	saveCachedResponse(c, cacheKey, imageGen, cacheTTLImageDetail)
	response.Success(c, imageGen)
}

func (h *ImageGenerationHandler) ListImageGenerations(c *gin.Context) {
	start := time.Now()
	var sceneID *uint
	if sceneIDStr := c.Query("scene_id"); sceneIDStr != "" {
		id, err := strconv.ParseUint(sceneIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			sceneID = &uid
		}
	}

	var storyboardID *uint
	if storyboardIDStr := c.Query("storyboard_id"); storyboardIDStr != "" {
		id, err := strconv.ParseUint(storyboardIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			storyboardID = &uid
		}
	}

	frameType := c.Query("frame_type")
	imageType := c.Query("image_type")
	status := c.Query("status")
	adPromptType := c.Query("ad_prompt_type")
	includeReusePrevLast := c.Query("include_reuse_prev_last") == "true"
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
	if sceneID != nil {
		normalizedQuery.Set("scene_id", strconv.FormatUint(uint64(*sceneID), 10))
	} else {
		normalizedQuery.Del("scene_id")
	}
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
	if includeReusePrevLast {
		normalizedQuery.Set("include_reuse_prev_last", "true")
	} else {
		normalizedQuery.Del("include_reuse_prev_last")
	}
	if adPromptType != "" {
		normalizedQuery.Set("ad_prompt_type", adPromptType)
	} else {
		normalizedQuery.Del("ad_prompt_type")
	}

	cacheKey := cache.NamespaceKeyWithQuery(cache.NamespaceImageList, normalizedQuery)
	sceneIDVal := interface{}(nil)
	if sceneID != nil {
		sceneIDVal = *sceneID
	}
	storyboardIDVal := interface{}(nil)
	if storyboardID != nil {
		storyboardIDVal = *storyboardID
	}
	dramaIDVal := interface{}(nil)
	if dramaIDUint != nil {
		dramaIDVal = *dramaIDUint
	}
	logFields := []interface{}{
		"scene_id", sceneIDVal,
		"storyboard_id", storyboardIDVal,
		"drama_id", dramaIDVal,
		"frame_type", frameType,
		"image_type", imageType,
		"ad_prompt_type", adPromptType,
		"status", status,
		"page", page,
		"page_size", pageSize,
		"include_reuse_prev_last", includeReusePrevLast,
	}
	withFields := func(extra ...interface{}) []interface{} {
		fields := make([]interface{}, 0, len(logFields)+len(extra))
		fields = append(fields, logFields...)
		fields = append(fields, extra...)
		return fields
	}
	h.log.Infow("List images start", withFields("cache_key", cacheKey)...)
	if entry, ok := tryServeCached(c, cacheKey); ok {
		h.log.Infow("List images cache hit", withFields("cache_key", cacheKey, "duration_ms", time.Since(start).Milliseconds())...)
		if entry != nil {
			response.Success(c, entry.Data)
		}
		return
	}

	h.log.Infow("List images db query start", withFields("cache_key", cacheKey)...)
	dbStart := time.Now()
	images, total, err := h.imageService.ListImageGenerations(dramaIDUint, sceneID, storyboardID, frameType, imageType, status, adPromptType, page, pageSize)

	if err != nil {
		h.log.Errorw("Failed to list images", withFields(
			"cache_key", cacheKey,
			"db_duration_ms", time.Since(dbStart).Milliseconds(),
			"duration_ms", time.Since(start).Milliseconds(),
			"error", err,
		)...)
		response.InternalError(c, err.Error())
		return
	}
	var items interface{} = images
	totalWithReuse := total
	itemsCount := len(images)
	if includeReusePrevLast && storyboardID != nil && (frameType == "" || frameType == models.FrameTypeFirst) {
		reuseImages, sourceID, reuseErr := h.imageService.ListReusePrevLastImages(*storyboardID)
		if reuseErr != nil {
			h.log.Warnw("Failed to load reuse previous last images", withFields("error", reuseErr)...)
		} else if len(reuseImages) > 0 {
			merged := make([]imageListItem, 0, len(images)+len(reuseImages))
			for _, img := range images {
				merged = append(merged, imageListItem{ImageGeneration: img})
			}
			for _, img := range reuseImages {
				copyImg := img
				frameTypeFirst := models.FrameTypeFirst
				copyImg.FrameType = &frameTypeFirst
				copyImg.StoryboardID = storyboardID
				merged = append(merged, imageListItem{
					ImageGeneration:    copyImg,
					ReusePrev:          true,
					SourceStoryboardID: sourceID,
				})
			}
			items = merged
			totalWithReuse = total + int64(len(reuseImages))
			itemsCount = len(merged)
		}
	}
	h.log.Infow("List images db query done", withFields(
		"cache_key", cacheKey,
		"db_duration_ms", time.Since(dbStart).Milliseconds(),
		"duration_ms", time.Since(start).Milliseconds(),
		"total", totalWithReuse,
		"items", itemsCount,
	)...)

	totalPages := (totalWithReuse + int64(pageSize) - 1) / int64(pageSize)
	payload := response.PaginationData{
		Items: items,
		Pagination: response.Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      totalWithReuse,
			TotalPages: totalPages,
		},
	}
	saveCachedResponse(c, cacheKey, payload, cacheTTLImageList)
	response.Success(c, payload)
}

func (h *ImageGenerationHandler) DeleteImageGeneration(c *gin.Context) {

	imageGenID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.imageService.DeleteImageGeneration(uint(imageGenID)); err != nil {
		h.log.Errorw("Failed to delete image", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
