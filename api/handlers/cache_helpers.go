package handlers

import (
	"net/http"
	"time"

	"github.com/drama-generator/backend/pkg/cache"
	"github.com/gin-gonic/gin"
)

func tryServeCached(c *gin.Context, cacheKey string) (*cache.Entry, bool) {
	entry, ok := cache.Get(cacheKey)
	if !ok {
		return nil, false
	}
	writeCacheHeaders(c, entry.ETag)
	if cache.MatchETag(c.GetHeader("If-None-Match"), entry.ETag) {
		c.Status(http.StatusNotModified)
		return nil, true
	}
	return &entry, true
}

func saveCachedResponse(c *gin.Context, cacheKey string, data interface{}, ttl time.Duration) {
	if entry, err := cache.Set(cacheKey, data, ttl); err == nil {
		writeCacheHeaders(c, entry.ETag)
	}
}

func writeCacheHeaders(c *gin.Context, etag string) {
	c.Header("ETag", etag)
	c.Header("Cache-Control", "private, max-age=0, must-revalidate")
}
