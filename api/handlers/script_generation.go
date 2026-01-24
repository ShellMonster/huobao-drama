package handlers

import (
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/events"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ScriptGenerationHandler struct {
	scriptService *services.ScriptGenerationService
	taskService   *services.TaskService
	log           *logger.Logger
}

func NewScriptGenerationHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger, hub *events.TaskHub) *ScriptGenerationHandler {
	return &ScriptGenerationHandler{
		scriptService: services.NewScriptGenerationService(db, cfg, log),
		taskService:   services.NewTaskService(db, log, hub),
		log:           log,
	}
}

func (h *ScriptGenerationHandler) GenerateCharacters(c *gin.Context) {
	var req services.GenerateCharactersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 创建异步任务并统一执行入口
	reqCopy := req
	task, err := h.taskService.RunAsync("character_generation", req.DramaID, "开始生成角色...", func(update services.TaskUpdater) (interface{}, error) {
		characters, err := h.scriptService.GenerateCharacters(&reqCopy)
		if err != nil {
			return nil, err
		}
		return gin.H{
			"characters": characters,
			"total":      len(characters),
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
		"message": "角色生成任务已创建，正在后台处理...",
	})
}
