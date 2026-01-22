package handlers

import (
	"github.com/drama-generator/backend/application/services"
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
	styles, err := h.styleService.ListActiveStyles()
	if err != nil {
		h.log.Errorw("Failed to list styles", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, styles)
}
