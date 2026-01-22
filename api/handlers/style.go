package handlers

import (
	"errors"
	"strconv"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/cache"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StyleHandler struct {
	styleService *services.StyleService
	log          *logger.Logger
}

func NewStyleHandler(db *gorm.DB, log *logger.Logger) *StyleHandler {
	return &StyleHandler{
		styleService: services.NewStyleService(db, log),
		log:          log,
	}
}

func (h *StyleHandler) ListStyles(c *gin.Context) {
	includeInactive := c.Query("include_inactive") == "true" || c.Query("all") == "true"

	var (
		styles []models.Style
		err    error
	)

	if includeInactive {
		styles, err = h.styleService.ListStyles(true)
	} else {
		styles, err = h.styleService.ListActiveStyles()
	}
	if err != nil {
		h.log.Errorw("Failed to list styles", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, styles)
}

func (h *StyleHandler) CreateStyle(c *gin.Context) {
	var req services.CreateStyleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	style, err := h.styleService.CreateStyle(&req)
	if err != nil {
		if errors.Is(err, services.ErrStyleKeyExists) {
			response.BadRequest(c, "风格Key已存在")
			return
		}
		response.InternalError(c, "创建失败")
		return
	}

	cache.BumpNamespace(cache.NamespaceStyles)
	if err := h.styleService.RefreshStyleCatalog(); err != nil {
		h.log.Warnw("Failed to refresh style catalog", "error", err)
	}
	response.Created(c, style)
}

func (h *StyleHandler) UpdateStyle(c *gin.Context) {
	styleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的风格ID")
		return
	}

	var req services.UpdateStyleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	style, err := h.styleService.UpdateStyle(uint(styleID), &req)
	if err != nil {
		if errors.Is(err, services.ErrStyleNotFound) {
			response.NotFound(c, "风格不存在")
			return
		}
		response.InternalError(c, "更新失败")
		return
	}

	cache.BumpNamespace(cache.NamespaceStyles)
	if err := h.styleService.RefreshStyleCatalog(); err != nil {
		h.log.Warnw("Failed to refresh style catalog", "error", err)
	}
	response.Success(c, style)
}

func (h *StyleHandler) DeleteStyle(c *gin.Context) {
	styleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的风格ID")
		return
	}

	if err := h.styleService.DeleteStyle(uint(styleID)); err != nil {
		if errors.Is(err, services.ErrStyleNotFound) {
			response.NotFound(c, "风格不存在")
			return
		}
		if errors.Is(err, services.ErrStyleSystemDelete) {
			response.BadRequest(c, "系统风格不可删除")
			return
		}
		response.InternalError(c, "删除失败")
		return
	}

	cache.BumpNamespace(cache.NamespaceStyles)
	if err := h.styleService.RefreshStyleCatalog(); err != nil {
		h.log.Warnw("Failed to refresh style catalog", "error", err)
	}
	response.Success(c, gin.H{"message": "删除成功"})
}
