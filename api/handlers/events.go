package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/events"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	hub *events.EventHub
	log *logger.Logger
}

func NewEventHandler(hub *events.EventHub, log *logger.Logger) *EventHandler {
	return &EventHandler{
		hub: hub,
		log: log,
	}
}

func (h *EventHandler) StreamTasks(c *gin.Context) {
	taskIDs, err := parseStringSet(c.Query("task_ids"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	taskTypes := parseStringList(c.Query("types"))
	taskTypeSet := make(map[string]struct{})
	for _, t := range taskTypes {
		taskTypeSet[t] = struct{}{}
	}
	resourceID := strings.TrimSpace(c.Query("resource_id"))

	filter := func(task *models.AsyncTask) bool {
		if len(taskIDs) > 0 {
			if _, ok := taskIDs[task.ID]; !ok {
				return false
			}
		}
		if len(taskTypeSet) > 0 {
			if _, ok := taskTypeSet[task.Type]; !ok {
				return false
			}
		}
		if resourceID != "" && task.ResourceID != resourceID {
			return false
		}
		return true
	}

	_, ch, unsubscribe := h.hub.Tasks.Subscribe(filter)
	defer unsubscribe()

	streamSSE(c, "task", ch)
}

func (h *EventHandler) StreamImageGenerations(c *gin.Context) {
	imageIDs, err := parseUintSet(c.Query("image_ids"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	dramaID := parseUintPointer(c.Query("drama_id"))
	sceneID := parseUintPointer(c.Query("scene_id"))
	characterID := parseUintPointer(c.Query("character_id"))
	storyboardID := parseUintPointer(c.Query("storyboard_id"))
	frameType := strings.TrimSpace(c.Query("frame_type"))
	imageType := strings.TrimSpace(c.Query("image_type"))

	filter := func(imageGen *models.ImageGeneration) bool {
		if len(imageIDs) > 0 {
			if _, ok := imageIDs[imageGen.ID]; !ok {
				return false
			}
		}
		if dramaID != nil && imageGen.DramaID != *dramaID {
			return false
		}
		if sceneID != nil && (imageGen.SceneID == nil || *imageGen.SceneID != *sceneID) {
			return false
		}
		if characterID != nil && (imageGen.CharacterID == nil || *imageGen.CharacterID != *characterID) {
			return false
		}
		if storyboardID != nil && (imageGen.StoryboardID == nil || *imageGen.StoryboardID != *storyboardID) {
			return false
		}
		if frameType != "" {
			if imageGen.FrameType == nil || *imageGen.FrameType != frameType {
				return false
			}
		}
		if imageType != "" && imageGen.ImageType != imageType {
			return false
		}
		return true
	}

	_, ch, unsubscribe := h.hub.ImageGenerations.Subscribe(filter)
	defer unsubscribe()

	streamSSE(c, "image_generation", ch)
}

func (h *EventHandler) StreamVideoGenerations(c *gin.Context) {
	videoIDs, err := parseUintSet(c.Query("video_ids"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	dramaID := parseUintPointer(c.Query("drama_id"))
	storyboardID := parseUintPointer(c.Query("storyboard_id"))

	filter := func(videoGen *models.VideoGeneration) bool {
		if len(videoIDs) > 0 {
			if _, ok := videoIDs[videoGen.ID]; !ok {
				return false
			}
		}
		if dramaID != nil && videoGen.DramaID != *dramaID {
			return false
		}
		if storyboardID != nil && (videoGen.StoryboardID == nil || *videoGen.StoryboardID != *storyboardID) {
			return false
		}
		return true
	}

	_, ch, unsubscribe := h.hub.VideoGenerations.Subscribe(filter)
	defer unsubscribe()

	streamSSE(c, "video_generation", ch)
}

func (h *EventHandler) StreamFramePromptTasks(c *gin.Context) {
	taskIDs, err := parseUintSet(c.Query("task_ids"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	storyboardID := parseUintPointer(c.Query("storyboard_id"))
	frameType := strings.TrimSpace(c.Query("frame_type"))

	filter := func(task *models.FramePromptTask) bool {
		if len(taskIDs) > 0 {
			if _, ok := taskIDs[task.ID]; !ok {
				return false
			}
		}
		if storyboardID != nil && task.StoryboardID != *storyboardID {
			return false
		}
		if frameType != "" && task.FrameType != frameType {
			return false
		}
		return true
	}

	_, ch, unsubscribe := h.hub.FramePromptTasks.Subscribe(filter)
	defer unsubscribe()

	streamSSE(c, "frame_prompt_task", ch)
}

func (h *EventHandler) StreamVideoMerges(c *gin.Context) {
	mergeIDs, err := parseUintSet(c.Query("merge_ids"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	episodeID := parseUintPointer(c.Query("episode_id"))
	dramaID := parseUintPointer(c.Query("drama_id"))

	filter := func(merge *models.VideoMerge) bool {
		if len(mergeIDs) > 0 {
			if _, ok := mergeIDs[merge.ID]; !ok {
				return false
			}
		}
		if episodeID != nil && merge.EpisodeID != *episodeID {
			return false
		}
		if dramaID != nil && merge.DramaID != *dramaID {
			return false
		}
		return true
	}

	_, ch, unsubscribe := h.hub.VideoMerges.Subscribe(filter)
	defer unsubscribe()

	streamSSE(c, "video_merge", ch)
}

func streamSSE[T any](c *gin.Context, eventName string, ch <-chan T) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	c.Writer.Flush()

	ctx := c.Request.Context()
	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case payload, ok := <-ch:
			if !ok {
				return
			}
			c.SSEvent(eventName, payload)
			c.Writer.Flush()
		case <-heartbeat.C:
			_, _ = c.Writer.WriteString(": ping\n\n")
			c.Writer.Flush()
		}
	}
}

func parseStringList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseStringSet(raw string) (map[string]struct{}, error) {
	parts := parseStringList(raw)
	if len(parts) == 0 {
		return map[string]struct{}{}, nil
	}
	result := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		result[p] = struct{}{}
	}
	return result, nil
}

func parseUintSet(raw string) (map[uint]struct{}, error) {
	parts := parseStringList(raw)
	if len(parts) == 0 {
		return map[uint]struct{}{}, nil
	}
	result := make(map[uint]struct{}, len(parts))
	for _, p := range parts {
		val, err := strconv.ParseUint(p, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid id: %s", p)
		}
		result[uint(val)] = struct{}{}
	}
	return result, nil
}

func parseUintPointer(raw string) *uint {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	val, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return nil
	}
	parsed := uint(val)
	return &parsed
}
