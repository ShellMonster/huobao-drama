package response

import (
	"net/http"
	"time"

	"github.com/drama-generator/backend/pkg/i18n"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type PaginationData struct {
	Items      interface{} `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		Message:   i18n.Translate(i18n.GetLang(c, i18n.LangZhCN), message, nil),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func SuccessWithPagination(c *gin.Context, items interface{}, total int64, page int, pageSize int) {
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: PaginationData{
			Items: items,
			Pagination: Pagination{
				Page:       page,
				PageSize:   pageSize,
				Total:      total,
				TotalPages: totalPages,
			},
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func Error(c *gin.Context, statusCode int, errCode string, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errCode,
			Message: i18n.Translate(i18n.GetLang(c, i18n.LangZhCN), message, nil),
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func ErrorWithDetails(c *gin.Context, statusCode int, errCode string, message string, details interface{}) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errCode,
			Message: i18n.Translate(i18n.GetLang(c, i18n.LangZhCN), message, nil),
			Details: details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func ErrorKey(c *gin.Context, statusCode int, errCode string, key string, args map[string]string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errCode,
			Message: i18n.Translate(i18n.GetLang(c, i18n.LangZhCN), key, args),
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func ErrorKeyWithDetails(c *gin.Context, statusCode int, errCode string, key string, args map[string]string, details interface{}) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errCode,
			Message: i18n.Translate(i18n.GetLang(c, i18n.LangZhCN), key, args),
			Details: details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

func BadRequestKey(c *gin.Context, key string, args map[string]string) {
	ErrorKey(c, http.StatusBadRequest, "BAD_REQUEST", key, args)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func UnauthorizedKey(c *gin.Context, key string, args map[string]string) {
	ErrorKey(c, http.StatusUnauthorized, "UNAUTHORIZED", key, args)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

func ForbiddenKey(c *gin.Context, key string, args map[string]string) {
	ErrorKey(c, http.StatusForbidden, "FORBIDDEN", key, args)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

func NotFoundKey(c *gin.Context, key string, args map[string]string) {
	ErrorKey(c, http.StatusNotFound, "NOT_FOUND", key, args)
}

func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func InternalErrorKey(c *gin.Context, key string, args map[string]string) {
	ErrorKey(c, http.StatusInternalServerError, "INTERNAL_ERROR", key, args)
}
