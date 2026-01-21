package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	models "github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/external/ffmpeg"
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/cache"
	"github.com/drama-generator/backend/pkg/events"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/utils"
	"github.com/drama-generator/backend/pkg/video"
	"gorm.io/gorm"
)

type VideoGenerationService struct {
	db              *gorm.DB
	transferService *ResourceTransferService
	log             *logger.Logger
	localStorage    *storage.LocalStorage
	aiService       *AIService
	ffmpeg          *ffmpeg.FFmpeg
	events          *events.VideoGenerationHub
}

func NewVideoGenerationService(db *gorm.DB, transferService *ResourceTransferService, localStorage *storage.LocalStorage, aiService *AIService, log *logger.Logger, events *events.VideoGenerationHub) *VideoGenerationService {
	service := &VideoGenerationService{
		db:              db,
		localStorage:    localStorage,
		transferService: transferService,
		aiService:       aiService,
		log:             log,
		ffmpeg:          ffmpeg.NewFFmpeg(log),
		events:          events,
	}

	go service.RecoverPendingTasks()

	return service
}

type GenerateVideoRequest struct {
	StoryboardID *uint  `json:"storyboard_id"`
	DramaID      string `json:"drama_id" binding:"required"`
	ImageGenID   *uint  `json:"image_gen_id"`

	// 参考图模式：single, first_last, multiple, none
	ReferenceMode string `json:"reference_mode"`

	// 单图模式
	ImageURL string `json:"image_url"`

	// 首尾帧模式
	FirstFrameURL *string `json:"first_frame_url"`
	LastFrameURL  *string `json:"last_frame_url"`

	// 多图模式
	ReferenceImageURLs []string `json:"reference_image_urls"`

	Prompt       string  `json:"prompt" binding:"required,min=5,max=2000"`
	Provider     string  `json:"provider"`
	Model        string  `json:"model"`
	Duration     *int    `json:"duration"`
	FPS          *int    `json:"fps"`
	AspectRatio  *string `json:"aspect_ratio"`
	Style        *string `json:"style"`
	MotionLevel  *int    `json:"motion_level"`
	CameraMotion *string `json:"camera_motion"`
	Seed         *int64  `json:"seed"`
}

func (s *VideoGenerationService) GenerateVideo(request *GenerateVideoRequest) (*models.VideoGeneration, error) {
	if request.StoryboardID != nil {
		var storyboard models.Storyboard
		if err := s.db.Preload("Episode").Where("id = ?", *request.StoryboardID).First(&storyboard).Error; err != nil {
			return nil, fmt.Errorf("storyboard not found")
		}
		if fmt.Sprintf("%d", storyboard.Episode.DramaID) != request.DramaID {
			return nil, fmt.Errorf("storyboard does not belong to drama")
		}
	}

	if request.ImageGenID != nil {
		var imageGen models.ImageGeneration
		if err := s.db.Where("id = ?", *request.ImageGenID).First(&imageGen).Error; err != nil {
			return nil, fmt.Errorf("image generation not found")
		}
	}

	provider := request.Provider
	if provider == "" {
		provider = "doubao"
	}

	dramaID, _ := strconv.ParseUint(request.DramaID, 10, 32)

	videoGen := &models.VideoGeneration{
		StoryboardID: request.StoryboardID,
		DramaID:      uint(dramaID),
		ImageGenID:   request.ImageGenID,
		Provider:     provider,
		Prompt:       request.Prompt,
		Model:        request.Model,
		Duration:     request.Duration,
		FPS:          request.FPS,
		AspectRatio:  request.AspectRatio,
		Style:        request.Style,
		MotionLevel:  request.MotionLevel,
		CameraMotion: request.CameraMotion,
		Seed:         request.Seed,
		Status:       models.VideoStatusPending,
	}

	// 根据参考图模式处理不同的参数
	if request.ReferenceMode != "" {
		videoGen.ReferenceMode = &request.ReferenceMode
	}

	switch request.ReferenceMode {
	case "single":
		// 单图模式
		if request.ImageURL != "" {
			videoGen.ImageURL = &request.ImageURL
		}
	case "first_last":
		// 首尾帧模式
		if request.FirstFrameURL != nil {
			videoGen.FirstFrameURL = request.FirstFrameURL
		}
		if request.LastFrameURL != nil {
			videoGen.LastFrameURL = request.LastFrameURL
		}
	case "multiple":
		// 多图模式
		if len(request.ReferenceImageURLs) > 0 {
			referenceImagesJSON, err := json.Marshal(request.ReferenceImageURLs)
			if err == nil {
				referenceImagesStr := string(referenceImagesJSON)
				videoGen.ReferenceImageURLs = &referenceImagesStr
			}
		}
	case "none":
		// 无参考图，纯文本生成
	default:
		// 向后兼容：如果没有指定模式，根据提供的参数自动判断
		if request.ImageURL != "" {
			videoGen.ImageURL = &request.ImageURL
			mode := "single"
			videoGen.ReferenceMode = &mode
		} else if request.FirstFrameURL != nil || request.LastFrameURL != nil {
			videoGen.FirstFrameURL = request.FirstFrameURL
			videoGen.LastFrameURL = request.LastFrameURL
			mode := "first_last"
			videoGen.ReferenceMode = &mode
		} else if len(request.ReferenceImageURLs) > 0 {
			referenceImagesJSON, err := json.Marshal(request.ReferenceImageURLs)
			if err == nil {
				referenceImagesStr := string(referenceImagesJSON)
				videoGen.ReferenceImageURLs = &referenceImagesStr
				mode := "multiple"
				videoGen.ReferenceMode = &mode
			}
		}
	}

	if err := s.db.Create(videoGen).Error; err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	cache.BumpNamespace(cache.NamespaceDramaList)
	cache.BumpNamespace(cache.NamespaceDramaDetail)
	cache.BumpNamespace(cache.NamespaceVideoList)
	cache.BumpNamespace(cache.NamespaceVideoDetail)
	cache.BumpNamespace(cache.NamespaceStoryboards)
	cache.BumpNamespace(cache.NamespaceVideoDetail)
	cache.BumpNamespace(cache.NamespaceStoryboards)

	s.publishVideoGeneration(videoGen.ID)
	go s.ProcessVideoGeneration(videoGen.ID)

	return videoGen, nil
}

func (s *VideoGenerationService) ProcessVideoGeneration(videoGenID uint) {
	var videoGen models.VideoGeneration
	if err := s.db.First(&videoGen, videoGenID).Error; err != nil {
		s.log.Errorw("Failed to load video generation", "error", err, "id", videoGenID)
		return
	}

	s.db.Model(&videoGen).Update("status", models.VideoStatusProcessing)
	s.publishVideoGeneration(videoGenID)

	cache.BumpNamespace(cache.NamespaceDramaList)
	cache.BumpNamespace(cache.NamespaceDramaDetail)
	cache.BumpNamespace(cache.NamespaceVideoList)

	client, err := s.getVideoClient(videoGen.Provider, videoGen.Model)
	if err != nil {
		s.log.Errorw("Failed to get video client", "error", err, "provider", videoGen.Provider, "model", videoGen.Model)
		s.updateVideoGenError(videoGenID, err.Error())
		return
	}

	s.log.Infow("Starting video generation", "id", videoGenID, "prompt", videoGen.Prompt, "provider", videoGen.Provider)

	var opts []video.VideoOption
	if videoGen.Model != "" {
		opts = append(opts, video.WithModel(videoGen.Model))
	}
	if videoGen.Duration != nil {
		opts = append(opts, video.WithDuration(*videoGen.Duration))
	}
	if videoGen.FPS != nil {
		opts = append(opts, video.WithFPS(*videoGen.FPS))
	}
	if videoGen.AspectRatio != nil {
		opts = append(opts, video.WithAspectRatio(*videoGen.AspectRatio))
	}
	if videoGen.Style != nil {
		opts = append(opts, video.WithStyle(*videoGen.Style))
	}
	if videoGen.MotionLevel != nil {
		opts = append(opts, video.WithMotionLevel(*videoGen.MotionLevel))
	}
	if videoGen.CameraMotion != nil {
		opts = append(opts, video.WithCameraMotion(*videoGen.CameraMotion))
	}
	if videoGen.Seed != nil {
		opts = append(opts, video.WithSeed(*videoGen.Seed))
	}

	// 根据参考图模式添加相应的选项
	if videoGen.ReferenceMode != nil {
		switch *videoGen.ReferenceMode {
		case "first_last":
			// 首尾帧模式
			if videoGen.FirstFrameURL != nil {
				opts = append(opts, video.WithFirstFrame(*videoGen.FirstFrameURL))
			}
			if videoGen.LastFrameURL != nil {
				opts = append(opts, video.WithLastFrame(*videoGen.LastFrameURL))
			}
		case "multiple":
			// 多图模式
			if videoGen.ReferenceImageURLs != nil {
				var imageURLs []string
				if err := json.Unmarshal([]byte(*videoGen.ReferenceImageURLs), &imageURLs); err == nil {
					opts = append(opts, video.WithReferenceImages(imageURLs))
				}
			}
		}
	}

	// 构造imageURL参数（单图模式使用，其他模式传空字符串）
	imageURL := ""
	if videoGen.ImageURL != nil {
		imageURL = *videoGen.ImageURL
	}

	result, err := client.GenerateVideo(imageURL, videoGen.Prompt, opts...)
	if err != nil {
		s.log.Errorw("Video generation API call failed", "error", err, "id", videoGenID)
		s.updateVideoGenError(videoGenID, err.Error())
		return
	}

	if result.TaskID != "" {
		s.db.Model(&videoGen).Updates(map[string]interface{}{
			"task_id": result.TaskID,
			"status":  models.VideoStatusProcessing,
		})
		s.publishVideoGeneration(videoGenID)
		go s.pollTaskStatus(videoGenID, result.TaskID, videoGen.Provider, videoGen.Model)
		return
	}

	if result.VideoURL != "" {
		s.completeVideoGeneration(videoGenID, result.VideoURL, &result.Duration, &result.Width, &result.Height, nil)
		return
	}

	s.updateVideoGenError(videoGenID, "no task ID or video URL returned")
}

func (s *VideoGenerationService) pollTaskStatus(videoGenID uint, taskID string, provider string, model string) {
	client, err := s.getVideoClient(provider, model)
	if err != nil {
		s.log.Errorw("Failed to get video client for polling", "error", err)
		s.updateVideoGenError(videoGenID, "failed to get video client")
		return
	}

	maxAttempts := 300
	interval := 10 * time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		time.Sleep(interval)

		var videoGen models.VideoGeneration
		if err := s.db.First(&videoGen, videoGenID).Error; err != nil {
			s.log.Errorw("Failed to load video generation", "error", err, "id", videoGenID)
			return
		}

		if videoGen.Status != models.VideoStatusProcessing {
			s.log.Infow("Video generation status changed, stopping poll", "id", videoGenID, "status", videoGen.Status)
			return
		}

		result, err := client.GetTaskStatus(taskID)
		if err != nil {
			s.log.Errorw("Failed to get task status", "error", err, "task_id", taskID)
			continue
		}

		if provider == "minimax" && result.Completed && result.FileID != "" {
			episodeID, storyboardID := s.resolveVideoStorageContext(&videoGen)
			videoCategory := buildVideoCategory("videos", videoGen.DramaID, episodeID, storyboardID)
			videoURL, err := s.storeMinimaxVideo(client, result.FileID, videoCategory)
			if err != nil {
				s.updateVideoGenError(videoGenID, err.Error())
				return
			}
			duration := result.Duration
			if duration == 0 && videoGen.Duration != nil {
				duration = *videoGen.Duration
			}
			var durationPtr *int
			if duration > 0 {
				durationPtr = &duration
			}
			s.completeVideoGeneration(videoGenID, videoURL, durationPtr, &result.Width, &result.Height, nil)
			return
		}

		if result.Error != "" {
			s.updateVideoGenError(videoGenID, result.Error)
			return
		}

		if result.Completed {
			if result.VideoURL != "" {
				s.completeVideoGeneration(videoGenID, result.VideoURL, &result.Duration, &result.Width, &result.Height, nil)
				return
			}
			s.updateVideoGenError(videoGenID, "task completed but no video URL")
			return
		}

		s.log.Infow("Video generation in progress", "id", videoGenID, "attempt", attempt+1)
	}

	s.updateVideoGenError(videoGenID, "polling timeout")
}

func buildVideoCategory(base string, dramaID uint, episodeID uint, storyboardID uint) string {
	return fmt.Sprintf("%s/dramas/%d/episodes/%d/storyboards/%d", base, dramaID, episodeID, storyboardID)
}

func (s *VideoGenerationService) resolveVideoStorageContext(videoGen *models.VideoGeneration) (uint, uint) {
	if videoGen == nil {
		return 0, 0
	}
	if videoGen.StoryboardID != nil {
		storyboardID := *videoGen.StoryboardID
		var storyboard models.Storyboard
		if err := s.db.Select("episode_id").Where("id = ?", storyboardID).First(&storyboard).Error; err == nil {
			return storyboard.EpisodeID, storyboardID
		}
		return 0, storyboardID
	}
	return 0, 0
}

func (s *VideoGenerationService) storeURLToLocal(url string, category string) (string, error) {
	if s.localStorage == nil {
		return "", fmt.Errorf("local storage not available")
	}
	if url == "" {
		return "", fmt.Errorf("empty url")
	}
	if s.localStorage.IsLocalURL(url) {
		return url, nil
	}
	if utils.IsDataURI(url) {
		data, mimeType, err := utils.ParseDataURI(url)
		if err != nil {
			return "", err
		}
		return s.localStorage.UploadBytes(data, mimeType, category)
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return s.localStorage.DownloadFromURL(url, category)
	}
	return "", fmt.Errorf("unsupported url format")
}

func (s *VideoGenerationService) completeVideoGeneration(videoGenID uint, videoURL string, duration *int, width *int, height *int, firstFrameURL *string) {
	var localVideoPath string

	var videoGen models.VideoGeneration
	if err := s.db.First(&videoGen, videoGenID).Error; err != nil {
		s.log.Errorw("Failed to load video generation", "error", err, "id", videoGenID)
		return
	}

	episodeID, storyboardID := s.resolveVideoStorageContext(&videoGen)
	videoCategory := buildVideoCategory("videos", videoGen.DramaID, episodeID, storyboardID)
	frameCategory := buildVideoCategory("video_frames", videoGen.DramaID, episodeID, storyboardID)

	finalVideoURL := videoURL
	if s.localStorage != nil && videoURL != "" {
		localURL, err := s.storeURLToLocal(videoURL, videoCategory)
		if err != nil {
			s.log.Warnw("Failed to store video to local storage",
				"error", err,
				"id", videoGenID,
				"original_url", videoURL)
		} else if localURL != "" {
			finalVideoURL = localURL
			localVideoPath = localURL
			s.log.Infow("Video stored to local storage",
				"id", videoGenID,
				"original_url", videoURL,
				"local_url", localURL)
		}
	}

	// 如果视频已下载到本地，探测真实时长
	if localVideoPath != "" && s.ffmpeg != nil {
		if probedDuration, err := s.ffmpeg.GetVideoDuration(localVideoPath); err == nil {
			// 转换为整数秒（向上取整）
			durationInt := int(probedDuration + 0.5)
			duration = &durationInt
			s.log.Infow("Probed video duration",
				"id", videoGenID,
				"duration_seconds", durationInt,
				"duration_float", probedDuration)
		} else {
			s.log.Warnw("Failed to probe video duration, using provided duration",
				"error", err,
				"id", videoGenID,
				"local_path", localVideoPath)
		}
	}

	var finalFirstFrameURL *string
	if firstFrameURL != nil && *firstFrameURL != "" && s.localStorage != nil {
		localFrameURL, err := s.storeURLToLocal(*firstFrameURL, frameCategory)
		if err != nil {
			s.log.Warnw("Failed to store first frame to local storage",
				"error", err,
				"id", videoGenID,
				"original_url", *firstFrameURL)
		} else {
			finalFirstFrameURL = &localFrameURL
			s.log.Infow("First frame stored to local storage",
				"id", videoGenID,
				"original_url", *firstFrameURL,
				"local_url", localFrameURL)
		}
	} else if firstFrameURL != nil {
		finalFirstFrameURL = firstFrameURL
	}

	// 数据库中保持使用原始URL
	updates := map[string]interface{}{
		"status":    models.VideoStatusCompleted,
		"video_url": finalVideoURL,
	}
	if duration != nil {
		updates["duration"] = *duration
	}
	if width != nil {
		updates["width"] = *width
	}
	if height != nil {
		updates["height"] = *height
	}
	if finalFirstFrameURL != nil {
		updates["first_frame_url"] = *finalFirstFrameURL
	}

	if err := s.db.Model(&models.VideoGeneration{}).Where("id = ?", videoGenID).Updates(updates).Error; err != nil {
		s.log.Errorw("Failed to update video generation", "error", err, "id", videoGenID)
		return
	}
	s.publishVideoGeneration(videoGenID)

	if videoGen.StoryboardID != nil {
		var storyboard models.Storyboard
		if err := s.db.Where("id = ?", *videoGen.StoryboardID).First(&storyboard).Error; err != nil {
			s.log.Warnw("Failed to load storyboard for video update", "storyboard_id", *videoGen.StoryboardID, "error", err)
		} else {
			if videoGen.OriginalStoryboardDuration == nil {
				if err := s.db.Model(&models.VideoGeneration{}).
					Where("id = ? AND original_storyboard_duration IS NULL", videoGenID).
					Update("original_storyboard_duration", storyboard.Duration).Error; err != nil {
					s.log.Warnw("Failed to store original storyboard duration", "video_id", videoGenID, "error", err)
				}
			}

			// 更新 Storyboard 的 video_url 和 duration
			storyboardUpdates := map[string]interface{}{
				"video_url": finalVideoURL,
			}
			if duration != nil {
				storyboardUpdates["duration"] = *duration
			}
			if err := s.db.Model(&models.Storyboard{}).Where("id = ?", *videoGen.StoryboardID).Updates(storyboardUpdates).Error; err != nil {
				s.log.Warnw("Failed to update storyboard", "storyboard_id", *videoGen.StoryboardID, "error", err)
			} else {
				s.log.Infow("Updated storyboard with video info", "storyboard_id", *videoGen.StoryboardID, "duration", duration)
			}
		}
	}

	cache.BumpNamespace(cache.NamespaceDramaList)
	cache.BumpNamespace(cache.NamespaceDramaDetail)
	cache.BumpNamespace(cache.NamespaceVideoList)
	cache.BumpNamespace(cache.NamespaceVideoDetail)
	cache.BumpNamespace(cache.NamespaceStoryboards)

	s.log.Infow("Video generation completed", "id", videoGenID, "url", finalVideoURL, "duration", duration)
}

func (s *VideoGenerationService) updateVideoGenError(videoGenID uint, errorMsg string) {
	if err := s.db.Model(&models.VideoGeneration{}).Where("id = ?", videoGenID).Updates(map[string]interface{}{
		"status":    models.VideoStatusFailed,
		"error_msg": errorMsg,
	}).Error; err != nil {
		s.log.Errorw("Failed to update video generation error", "error", err, "id", videoGenID)
	}
	s.publishVideoGeneration(videoGenID)

	cache.BumpNamespace(cache.NamespaceDramaList)
	cache.BumpNamespace(cache.NamespaceDramaDetail)
	cache.BumpNamespace(cache.NamespaceVideoList)
	cache.BumpNamespace(cache.NamespaceVideoDetail)
	cache.BumpNamespace(cache.NamespaceStoryboards)
}

func (s *VideoGenerationService) publishVideoGeneration(videoGenID uint) {
	if s.events == nil {
		return
	}
	var videoGen models.VideoGeneration
	if err := s.db.Where("id = ?", videoGenID).First(&videoGen).Error; err != nil {
		s.log.Warnw("Failed to publish video generation update", "error", err, "id", videoGenID)
		return
	}
	s.events.Publish(&videoGen)
}

func (s *VideoGenerationService) storeMinimaxVideo(client video.VideoClient, fileID string, category string) (string, error) {
	if s.localStorage == nil {
		return "", fmt.Errorf("local storage not available")
	}
	minimaxClient, ok := client.(*video.MinimaxClient)
	if !ok {
		return "", fmt.Errorf("invalid minimax client")
	}

	reader, contentType, err := minimaxClient.RetrieveFileContent(fileID)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	token, err := utils.NewRandomID()
	if err != nil {
		return "", fmt.Errorf("failed to generate file name: %w", err)
	}
	filename := fmt.Sprintf("%s%s", token, minimaxVideoExtension(contentType))
	return s.localStorage.Upload(reader, filename, category)
}

func minimaxVideoExtension(contentType string) string {
	switch {
	case strings.Contains(contentType, "video/quicktime"):
		return ".mov"
	case strings.Contains(contentType, "video/webm"):
		return ".webm"
	case strings.Contains(contentType, "video/mp4"):
		return ".mp4"
	default:
		return ".mp4"
	}
}

func (s *VideoGenerationService) getVideoClient(provider string, modelName string) (video.VideoClient, error) {
	// 根据模型名称获取AI配置
	var config *models.AIServiceConfig
	var err error

	if modelName != "" {
		config, err = s.aiService.GetConfigForModel("video", modelName)
		if err != nil {
			s.log.Warnw("Failed to get config for model, using default", "model", modelName, "error", err)
			config, err = s.aiService.GetDefaultConfig("video")
			if err != nil {
				return nil, fmt.Errorf("no video AI config found: %w", err)
			}
		}
	} else {
		config, err = s.aiService.GetDefaultConfig("video")
		if err != nil {
			return nil, fmt.Errorf("no video AI config found: %w", err)
		}
	}

	// 使用配置中的信息创建客户端
	baseURL := config.BaseURL
	apiKey := config.APIKey
	model := modelName
	if model == "" && len(config.Model) > 0 {
		model = config.Model[0]
	}

	// 根据配置中的 provider 创建对应的客户端
	var endpoint string
	var queryEndpoint string

	switch config.Provider {
	case "chatfire":
		endpoint = "/video/generations"
		queryEndpoint = "/video/task/{taskId}"
		return video.NewChatfireClient(baseURL, apiKey, model, endpoint, queryEndpoint), nil
	case "doubao", "volcengine", "volces":
		endpoint = "/contents/generations/tasks"
		queryEndpoint = "/contents/generations/tasks/{taskId}"
		return video.NewVolcesArkClient(baseURL, apiKey, model, endpoint, queryEndpoint), nil
	case "jimeng":
		accessKey := strings.TrimSpace(apiKey)
		var settings struct {
			AccessKey    string `json:"access_key"`
			SecretKey    string `json:"secret_key"`
			SessionToken string `json:"session_token"`
		}
		if config.Settings != "" {
			if err := json.Unmarshal([]byte(config.Settings), &settings); err != nil {
				return nil, fmt.Errorf("invalid jimeng settings: %w", err)
			}
		}
		if accessKey == "" {
			accessKey = strings.TrimSpace(settings.AccessKey)
		}
		secretKey := strings.TrimSpace(settings.SecretKey)
		if accessKey == "" || secretKey == "" {
			return nil, fmt.Errorf("jimeng access_key or secret_key missing")
		}
		return video.NewJimengCVClient(baseURL, accessKey, secretKey, model, settings.SessionToken), nil
	case "kling":
		return video.NewKlingClient(baseURL, apiKey, model), nil
	case "openai":
		// OpenAI Sora 使用 /v1/videos 端点
		return video.NewOpenAISoraClient(baseURL, apiKey, model), nil
	case "runway":
		return video.NewRunwayClient(baseURL, apiKey, model), nil
	case "pika":
		return video.NewPikaClient(baseURL, apiKey, model), nil
	case "minimax":
		return video.NewMinimaxClient(baseURL, apiKey, model), nil
	default:
		return nil, fmt.Errorf("unsupported video provider: %s", provider)
	}
}

func (s *VideoGenerationService) RecoverPendingTasks() {
	var pendingVideos []models.VideoGeneration
	if err := s.db.Where("status = ? AND task_id != ''", models.VideoStatusProcessing).Find(&pendingVideos).Error; err != nil {
		s.log.Errorw("Failed to load pending video tasks", "error", err)
		return
	}

	s.log.Infow("Recovering pending video generation tasks", "count", len(pendingVideos))

	for _, videoGen := range pendingVideos {
		go s.pollTaskStatus(videoGen.ID, *videoGen.TaskID, videoGen.Provider, videoGen.Model)
	}
}

func (s *VideoGenerationService) GetVideoGeneration(id uint) (*models.VideoGeneration, error) {
	var videoGen models.VideoGeneration
	if err := s.db.First(&videoGen, id).Error; err != nil {
		return nil, err
	}
	return &videoGen, nil
}

func (s *VideoGenerationService) ListVideoGenerations(dramaID *uint, storyboardID *uint, status string, limit int, offset int) ([]*models.VideoGeneration, int64, error) {
	var videos []*models.VideoGeneration
	var total int64

	query := s.db.Model(&models.VideoGeneration{})

	if dramaID != nil {
		query = query.Where("drama_id = ?", *dramaID)
	}
	if storyboardID != nil {
		query = query.Where("storyboard_id = ?", *storyboardID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

func (s *VideoGenerationService) GenerateVideoFromImage(imageGenID uint) (*models.VideoGeneration, error) {
	var imageGen models.ImageGeneration
	if err := s.db.First(&imageGen, imageGenID).Error; err != nil {
		return nil, fmt.Errorf("image generation not found")
	}

	if imageGen.Status != models.ImageStatusCompleted || imageGen.ImageURL == nil {
		return nil, fmt.Errorf("image is not ready")
	}

	// 获取关联的Storyboard以获取时长
	var duration *int
	if imageGen.StoryboardID != nil {
		var storyboard models.Storyboard
		if err := s.db.Where("id = ?", *imageGen.StoryboardID).First(&storyboard).Error; err == nil {
			duration = &storyboard.Duration
			s.log.Infow("Using storyboard duration for video generation",
				"storyboard_id", *imageGen.StoryboardID,
				"duration", storyboard.Duration)
		}
	}

	req := &GenerateVideoRequest{
		DramaID:      fmt.Sprintf("%d", imageGen.DramaID),
		StoryboardID: imageGen.StoryboardID,
		ImageGenID:   &imageGenID,
		ImageURL:     *imageGen.ImageURL,
		Prompt:       imageGen.Prompt,
		Provider:     "doubao",
		Duration:     duration,
	}

	return s.GenerateVideo(req)
}

func (s *VideoGenerationService) BatchGenerateVideosForEpisode(episodeID string) ([]*models.VideoGeneration, error) {
	var episode models.Episode
	if err := s.db.Preload("Storyboards").Where("id = ?", episodeID).First(&episode).Error; err != nil {
		return nil, fmt.Errorf("episode not found")
	}

	var results []*models.VideoGeneration
	for _, storyboard := range episode.Storyboards {
		if storyboard.ImagePrompt == nil {
			continue
		}

		var imageGen models.ImageGeneration
		if err := s.db.Where("storyboard_id = ? AND status = ?", storyboard.ID, models.ImageStatusCompleted).
			Order("created_at DESC").First(&imageGen).Error; err != nil {
			s.log.Warnw("No completed image for storyboard", "storyboard_id", storyboard.ID)
			continue
		}

		videoGen, err := s.GenerateVideoFromImage(imageGen.ID)
		if err != nil {
			s.log.Errorw("Failed to generate video", "storyboard_id", storyboard.ID, "error", err)
			continue
		}

		results = append(results, videoGen)
	}

	return results, nil
}

func (s *VideoGenerationService) DeleteVideoGeneration(id uint) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var videoGen models.VideoGeneration
		if err := tx.Where("id = ?", id).First(&videoGen).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("video generation not found")
			}
			return err
		}

		if err := tx.Delete(&models.VideoGeneration{}, id).Error; err != nil {
			return err
		}

		if videoGen.StoryboardID != nil && videoGen.VideoURL != nil && *videoGen.VideoURL != "" {
			storyboardUpdates := map[string]interface{}{
				"video_url": gorm.Expr("NULL"),
			}
			if videoGen.OriginalStoryboardDuration != nil {
				storyboardUpdates["duration"] = *videoGen.OriginalStoryboardDuration
			}
			if err := tx.Model(&models.Storyboard{}).
				Where("id = ? AND video_url = ?", *videoGen.StoryboardID, *videoGen.VideoURL).
				Updates(storyboardUpdates).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("video_gen_id = ?", id).Delete(&models.Asset{}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	cache.BumpNamespace(cache.NamespaceDramaList)
	cache.BumpNamespace(cache.NamespaceDramaDetail)
	cache.BumpNamespace(cache.NamespaceVideoList)
	cache.BumpNamespace(cache.NamespaceVideoDetail)
	cache.BumpNamespace(cache.NamespaceStoryboards)
	return nil
}
