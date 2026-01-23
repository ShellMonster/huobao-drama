package handlers

import (
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// BatchGenerateCharacterImages 批量生成角色图片
func (h *CharacterLibraryHandler) BatchGenerateCharacterImages(c *gin.Context) {

	var req struct {
		CharacterIDs []string `json:"character_ids" binding:"required,min=1"`
		Model        string   `json:"model"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 限制批量生成数量
	if len(req.CharacterIDs) > 10 {
		response.BadRequest(c, "单次最多生成10个角色")
		return
	}

	results := h.libraryService.BatchGenerateCharacterImages(req.CharacterIDs, h.imageService, req.Model)
	successCount := 0
	failCount := 0
	for _, item := range results {
		if item.ImageGenerationID != nil {
			successCount++
		} else {
			failCount++
		}
	}

	response.Success(c, gin.H{
		"message":       "批量生成任务已提交",
		"count":         len(req.CharacterIDs),
		"success_count": successCount,
		"fail_count":    failCount,
		"items":         results,
	})
}
