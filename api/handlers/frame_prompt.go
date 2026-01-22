package handlers

import (
	"strconv"
	"strings"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/cache"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// FramePromptHandler 处理帧提示词生成请求
type FramePromptHandler struct {
	framePromptService *services.FramePromptService
	log                *logger.Logger
}

// NewFramePromptHandler 创建帧提示词处理器
func NewFramePromptHandler(framePromptService *services.FramePromptService, log *logger.Logger) *FramePromptHandler {
	return &FramePromptHandler{
		framePromptService: framePromptService,
		log:                log,
	}
}

// GenerateFramePrompt 生成指定类型的帧提示词
// POST /api/v1/storyboards/:id/frame-prompt
func (h *FramePromptHandler) GenerateFramePrompt(c *gin.Context) {
	storyboardID := c.Param("id")

	var req struct {
		FrameType     string `json:"frame_type"`
		PanelCount    int    `json:"panel_count"`
		Model         string `json:"model"`
		ReusePrevLast *bool  `json:"reuse_prev_last"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.ReusePrevLast != nil && *req.ReusePrevLast {
		if req.FrameType != string(services.FrameTypeFirst) {
			response.BadRequest(c, "reuse_prev_last only supports first frame")
			return
		}
		storyboardIDUint, err := strconv.ParseUint(storyboardID, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid storyboard id")
			return
		}
		preview, err := h.framePromptService.ReusePrevLastFrame(uint(storyboardIDUint))
		if err != nil {
			msg := err.Error()
			if strings.Contains(msg, "previous storyboard") || strings.Contains(msg, "last frame") {
				response.BadRequest(c, msg)
				return
			}
			h.log.Errorw("Failed to reuse previous last frame", "error", err, "storyboard_id", storyboardID)
			response.InternalError(c, msg)
			return
		}
		cache.BumpNamespace(cache.NamespaceDramaList)
		cache.BumpNamespace(cache.NamespaceDramaDetail)
		cache.BumpNamespace(cache.NamespaceStoryboards)
		response.Success(c, gin.H{
			"reuse_preview": preview,
		})
		return
	}

	if req.ReusePrevLast != nil && !*req.ReusePrevLast {
		storyboardIDUint, err := strconv.ParseUint(storyboardID, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid storyboard id")
			return
		}
		if err := h.framePromptService.SetReusePrevLastFrame(uint(storyboardIDUint), false); err != nil {
			h.log.Errorw("Failed to disable reuse previous last frame", "error", err, "storyboard_id", storyboardID)
			response.InternalError(c, err.Error())
			return
		}
	}

	serviceReq := services.GenerateFramePromptRequest{
		StoryboardID: storyboardID,
		FrameType:    services.FrameType(req.FrameType),
		PanelCount:   req.PanelCount,
	}

	result, err := h.framePromptService.CreateFramePromptTask(serviceReq, req.Model)
	if err != nil {
		h.log.Errorw("Failed to generate frame prompt", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"task": result,
	})
}
