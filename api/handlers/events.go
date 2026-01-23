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

type SSEEnvelope struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
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

func (h *EventHandler) StreamUnifiedEvents(c *gin.Context) {
	eventTypes := parseStringSetNoError(c.Query("types"))
	if len(eventTypes) == 0 {
		c.JSON(400, gin.H{"success": false, "error": gin.H{"message": "types is required"}})
		return
	}

	taskIDs, err := parseStringSet(c.Query("task_ids"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	taskTypes := parseStringSetNoError(c.Query("task_types"))
	resourceID := strings.TrimSpace(c.Query("resource_id"))

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

	videoIDs, err := parseUintSet(c.Query("video_ids"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}

	framePromptTaskIDs, err := parseUintSet(c.Query("frame_prompt_task_ids"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	framePromptStoryboardID := parseUintPointer(c.Query("frame_prompt_storyboard_id"))
	framePromptFrameType := strings.TrimSpace(c.Query("frame_prompt_frame_type"))

	mergeIDs, err := parseUintSet(c.Query("merge_ids"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	mergeEpisodeID := parseUintPointer(c.Query("merge_episode_id"))
	mergeDramaID := parseUintPointer(c.Query("merge_drama_id"))

	ctx := c.Request.Context()
	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	type envelope = SSEEnvelope
	type eventPayload struct {
		eventType string
		payload   interface{}
	}
	eventsCh := make(chan eventPayload, 64)
	send := func(eventType string, payload interface{}) bool {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		c.SSEvent("event", envelope{Type: eventType, Data: payload})
		c.Writer.Flush()
		return true
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.Flush()

	if _, ok := eventTypes["task"]; ok {
		filter := func(task *models.AsyncTask) bool {
			if len(taskIDs) > 0 {
				if _, ok := taskIDs[task.ID]; !ok {
					return false
				}
			}
			if len(taskTypes) > 0 {
				if _, ok := taskTypes[task.Type]; !ok {
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
		go func() {
			for payload := range ch {
				select {
				case <-ctx.Done():
					return
				case eventsCh <- eventPayload{eventType: "task", payload: payload}:
				default:
				}
			}
		}()
	}

	if _, ok := eventTypes["image_generation"]; ok {
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
		go func() {
			for payload := range ch {
				select {
				case <-ctx.Done():
					return
				case eventsCh <- eventPayload{eventType: "image_generation", payload: payload}:
				default:
				}
			}
		}()
	}

	if _, ok := eventTypes["video_generation"]; ok {
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
		go func() {
			for payload := range ch {
				select {
				case <-ctx.Done():
					return
				case eventsCh <- eventPayload{eventType: "video_generation", payload: payload}:
				default:
				}
			}
		}()
	}

	if _, ok := eventTypes["frame_prompt_task"]; ok {
		filter := func(task *models.FramePromptTask) bool {
			if len(framePromptTaskIDs) > 0 {
				if _, ok := framePromptTaskIDs[task.ID]; !ok {
					return false
				}
			}
			if framePromptStoryboardID != nil && task.StoryboardID != *framePromptStoryboardID {
				return false
			}
			if framePromptFrameType != "" && task.FrameType != framePromptFrameType {
				return false
			}
			return true
		}
		_, ch, unsubscribe := h.hub.FramePromptTasks.Subscribe(filter)
		defer unsubscribe()
		go func() {
			for payload := range ch {
				select {
				case <-ctx.Done():
					return
				case eventsCh <- eventPayload{eventType: "frame_prompt_task", payload: payload}:
				default:
				}
			}
		}()
	}

	if _, ok := eventTypes["video_merge"]; ok {
		filter := func(merge *models.VideoMerge) bool {
			if len(mergeIDs) > 0 {
				if _, ok := mergeIDs[merge.ID]; !ok {
					return false
				}
			}
			if mergeEpisodeID != nil && merge.EpisodeID != *mergeEpisodeID {
				return false
			}
			if mergeDramaID != nil && merge.DramaID != *mergeDramaID {
				return false
			}
			return true
		}
		_, ch, unsubscribe := h.hub.VideoMerges.Subscribe(filter)
		defer unsubscribe()
		go func() {
			for payload := range ch {
				select {
				case <-ctx.Done():
					return
				case eventsCh <- eventPayload{eventType: "video_merge", payload: payload}:
				default:
				}
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-eventsCh:
			if !send(evt.eventType, evt.payload) {
				return
			}
		case <-heartbeat.C:
			_, _ = c.Writer.WriteString(": ping\n\n")
			c.Writer.Flush()
		}
	}
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

func parseStringSetNoError(raw string) map[string]struct{} {
	parts := parseStringList(raw)
	if len(parts) == 0 {
		return map[string]struct{}{}
	}
	result := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		result[p] = struct{}{}
	}
	return result
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
