package handlers

import (
	"strings"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetStoryboardFramePromptTasks 查询镜头的帧提示词任务状态
// GET /api/v1/storyboards/:id/frame-prompt-tasks
func GetStoryboardFramePromptTasks(db *gorm.DB, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		storyboardID := c.Param("id")
		statusRaw := strings.TrimSpace(c.Query("status"))

		query := db.Model(&models.FramePromptTask{}).
			Where("storyboard_id = ?", storyboardID)

		if statusRaw != "" {
			statusList := strings.Split(statusRaw, ",")
			query = query.Where("status IN ?", statusList)
		}

		var tasks []models.FramePromptTask
		if err := query.Order("created_at DESC").Find(&tasks).Error; err != nil {
			log.Errorw("Failed to query frame prompt tasks", "error", err)
			response.InternalError(c, err.Error())
			return
		}

		response.Success(c, gin.H{
			"tasks": tasks,
		})
	}
}
