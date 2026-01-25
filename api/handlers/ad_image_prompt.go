package handlers

import (
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type AdImagePromptHandler struct {
	service *services.AdImagePromptService
}

func NewAdImagePromptHandler(service *services.AdImagePromptService) *AdImagePromptHandler {
	return &AdImagePromptHandler{service: service}
}

func (h *AdImagePromptHandler) GenerateTextPrompts(c *gin.Context) {
	var req services.GenerateAdTextPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	prompts, err := h.service.GenerateTextPrompts(&req)
	if err != nil {
		switch err.Error() {
		case "drama not found":
			response.NotFound(c, "项目不存在")
			return
		case "brand not found":
			response.NotFound(c, "品牌不存在")
			return
		case "brand spec not found":
			response.NotFound(c, "规范不存在")
			return
		case "brand_id is required when spec_id is provided":
			response.BadRequest(c, "选择规范时需要同时选择品牌")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"prompts": prompts})
}

func (h *AdImagePromptHandler) GenerateImagePrompts(c *gin.Context) {
	var req services.GenerateAdImagePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	prompts, err := h.service.GeneratePromptsFromImage(&req)
	if err != nil {
		switch err.Error() {
		case "drama not found":
			response.NotFound(c, "项目不存在")
			return
		case "brand not found":
			response.NotFound(c, "品牌不存在")
			return
		case "brand spec not found":
			response.NotFound(c, "规范不存在")
			return
		case "brand_id is required when spec_id is provided":
			response.BadRequest(c, "选择规范时需要同时选择品牌")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"prompts": prompts})
}

func (h *AdImagePromptHandler) GetLatestPrompts(c *gin.Context) {
	var req services.GetAdPromptRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.GetLatestPrompts(&req)
	if err != nil {
		switch err.Error() {
		case "drama not found":
			response.NotFound(c, "项目不存在")
			return
		case "brand not found":
			response.NotFound(c, "品牌不存在")
			return
		case "brand spec not found":
			response.NotFound(c, "规范不存在")
			return
		case "brand_id is required when spec_id is provided":
			response.BadRequest(c, "选择规范时需要同时选择品牌")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}
