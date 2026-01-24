package handlers

import (
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/cache"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/events"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StoryboardHandler struct {
	storyboardService *services.StoryboardService
	taskService       *services.TaskService
	log               *logger.Logger
}

func NewStoryboardHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger, hub *events.TaskHub) *StoryboardHandler {
	return &StoryboardHandler{
		storyboardService: services.NewStoryboardService(db, cfg, log),
		taskService:       services.NewTaskService(db, log, hub),
		log:               log,
	}
}

// GenerateStoryboard 生成分镜头（异步）
func (h *StoryboardHandler) GenerateStoryboard(c *gin.Context) {
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
	task, err := h.taskService.RunAsync("storyboard_generation", episodeID, "开始生成分镜...", func(update services.TaskUpdater) (interface{}, error) {
		result, err := h.storyboardService.GenerateStoryboard(episodeID, req.Model)
		if err != nil {
			return nil, err
		}
		cache.BumpNamespace(cache.NamespaceDramaList)
		cache.BumpNamespace(cache.NamespaceDramaDetail)
		cache.BumpNamespace(cache.NamespaceStoryboards)
		return result, nil
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
		"message": "分镜头生成任务已创建，正在后台处理...",
	})
}

// UpdateStoryboard 更新分镜
func (h *StoryboardHandler) UpdateStoryboard(c *gin.Context) {
	storyboardID := c.Param("id")

	var req map[string]interface{}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	err := h.storyboardService.UpdateStoryboard(storyboardID, req)
	if err != nil {
		h.log.Errorw("Failed to update storyboard", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	cache.BumpNamespace(cache.NamespaceDramaList)
	cache.BumpNamespace(cache.NamespaceDramaDetail)
	cache.BumpNamespace(cache.NamespaceStoryboards)
	response.Success(c, gin.H{"message": "Storyboard updated successfully"})
}
