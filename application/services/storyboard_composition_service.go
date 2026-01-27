package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	models "github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
)

type StoryboardCompositionService struct {
	db       *gorm.DB
	log      *logger.Logger
	imageGen *ImageGenerationService
}

func NewStoryboardCompositionService(db *gorm.DB, log *logger.Logger, imageGen *ImageGenerationService) *StoryboardCompositionService {
	return &StoryboardCompositionService{
		db:       db,
		log:      log,
		imageGen: imageGen,
	}
}

type SceneCharacterInfo struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	ImageURL *string `json:"image_url,omitempty"`
}

type SceneBackgroundInfo struct {
	ID       uint    `json:"id"`
	Location string  `json:"location"`
	Time     string  `json:"time"`
	ImageURL *string `json:"image_url,omitempty"`
	Status   string  `json:"status"`
}

type SceneCompositionInfo struct {
	ID                    uint                 `json:"id"`
	StoryboardNumber      int                  `json:"storyboard_number"`
	Title                 *string              `json:"title"`
	Description           *string              `json:"description"`
	ShotType              *string              `json:"shot_type"`
	Angle                 *string              `json:"angle"`
	Movement              *string              `json:"movement"`
	Location              *string              `json:"location"`
	Time                  *string              `json:"time"`
	Duration              int                  `json:"duration"`
	Dialogue              *string              `json:"dialogue"`
	Action                *string              `json:"action"`
	Result                *string              `json:"result"`
	Atmosphere            *string              `json:"atmosphere"`
	BgmPrompt             *string              `json:"bgm_prompt,omitempty"`
	SoundEffect           *string              `json:"sound_effect,omitempty"`
	ImagePrompt           *string              `json:"image_prompt,omitempty"`
	VideoPrompt           *string              `json:"video_prompt,omitempty"`
	Characters            []SceneCharacterInfo `json:"characters"`
	Background            *SceneBackgroundInfo `json:"background"`
	SceneID               *uint                `json:"scene_id"`
	ComposedImage         *string              `json:"composed_image,omitempty"`
	VideoURL              *string              `json:"video_url,omitempty"`
	ImageGenerationID     *uint                `json:"image_generation_id,omitempty"`
	ImageGenerationStatus *string              `json:"image_generation_status,omitempty"`
	VideoGenerationID     *uint                `json:"video_generation_id,omitempty"`
	VideoGenerationStatus *string              `json:"video_generation_status,omitempty"`
	ReusePrevLastFrame    bool                 `json:"reuse_prev_last_frame"`
	PrevLastPrompt        *string              `json:"prev_last_prompt,omitempty"`
	PrevLastImages        []PrevLastImageInfo  `json:"prev_last_images,omitempty"`
	PrevLastStoryboardID  *uint                `json:"prev_last_storyboard_id,omitempty"`
	PrevLastStoryboardNum *int                 `json:"prev_last_storyboard_number,omitempty"`
}

type PrevLastImageInfo struct {
	ID          uint    `json:"id"`
	StoryboardID uint   `json:"storyboard_id"`
	FrameType   string  `json:"frame_type"`
	ImageURL    *string `json:"image_url,omitempty"`
	Status      string  `json:"status"`
}

type CreateSceneRequest struct {
	DramaID   uint    `json:"drama_id"`
	EpisodeID *uint   `json:"episode_id"`
	Location  string  `json:"location"`
	Time      string  `json:"time"`
	Prompt    *string `json:"prompt"`
}

type UpdateSceneDetailsRequest struct {
	Location *string `json:"location"`
	Time     *string `json:"time"`
	Prompt   *string `json:"prompt"`
}

func (s *StoryboardCompositionService) GetScenesForEpisode(episodeID string) ([]SceneCompositionInfo, error) {
	// 验证权限
	var episode models.Episode
	err := s.db.Preload("Drama").Where("id = ?", episodeID).First(&episode).Error
	if err != nil {
		s.log.Errorw("Episode not found", "episode_id", episodeID, "error", err)
		return nil, fmt.Errorf("episode not found")
	}

	s.log.Infow("GetScenesForEpisode auth check",
		"episode_id", episodeID,
		"drama_id", episode.DramaID)

	// 获取分镜列表
	var storyboards []models.Storyboard
	if err := s.db.Where("episode_id = ?", episodeID).
		Preload("Characters").
		Order("storyboard_number ASC").
		Find(&storyboards).Error; err != nil {
		return nil, fmt.Errorf("failed to load storyboards: %w", err)
	}

	// 获取所有角色（用于匹配角色信息）
	var characters []models.Character
	if err := s.db.Where("drama_id = ?", episode.DramaID).Find(&characters).Error; err != nil {
		s.log.Warnw("Failed to load characters", "error", err)
	}

	bestVisualByName := make(map[string]models.Character)
	visualCountByName := make(map[string]int)
	for _, character := range characters {
		if character.Name == "" {
			continue
		}
		if hasCharacterImage(character) || hasCharacterReferenceImages(character) {
			visualCountByName[character.Name]++
			if existing, ok := bestVisualByName[character.Name]; ok {
				bestVisualByName[character.Name] = preferCharacter(existing, character)
			} else {
				bestVisualByName[character.Name] = character
			}
		}
	}

	// 获取所有场景ID
	var sceneIDs []uint
	for _, storyboard := range storyboards {
		if storyboard.SceneID != nil {
			sceneIDs = append(sceneIDs, *storyboard.SceneID)
		}
	}

	// 批量获取场景信息
	var scenes []models.Scene
	sceneMap := make(map[uint]*models.Scene)
	if len(sceneIDs) > 0 {
		if err := s.db.Where("id IN ?", sceneIDs).Find(&scenes).Error; err == nil {
			for i := range scenes {
				sceneMap[scenes[i].ID] = &scenes[i]
			}
		}
	}

	// 获取分镜的合成图片（从 image_generations 表）
	storyboardIDs := make([]uint, len(storyboards))
	for i, storyboard := range storyboards {
		storyboardIDs[i] = storyboard.ID
	}

	imageGenMap := make(map[uint]string)                      // storyboard_id -> image_url
	imageGenTaskMap := make(map[uint]*models.ImageGeneration) // storyboard_id -> processing task
	if len(storyboardIDs) > 0 {
		var imageGens []models.ImageGeneration
		// 查询已完成的图片生成记录，每个镜头只取最新的一条
		if err := s.db.Where("storyboard_id IN ? AND status = ?", storyboardIDs, models.ImageStatusCompleted).
			Order("created_at DESC").
			Find(&imageGens).Error; err == nil {
			// 为每个镜头保留最新的一条记录
			for _, ig := range imageGens {
				if ig.StoryboardID != nil {
					if _, exists := imageGenMap[*ig.StoryboardID]; !exists {
						if ig.ImageURL != nil {
							imageGenMap[*ig.StoryboardID] = *ig.ImageURL
						}
					}
				}
			}
		}

		// 查询进行中的图片生成任务
		var processingImageGens []models.ImageGeneration
		if err := s.db.Where("storyboard_id IN ? AND status = ?", storyboardIDs, models.ImageStatusProcessing).
			Order("created_at DESC").
			Find(&processingImageGens).Error; err == nil {
			for _, ig := range processingImageGens {
				if ig.StoryboardID != nil {
					if _, exists := imageGenTaskMap[*ig.StoryboardID]; !exists {
						igCopy := ig
						imageGenTaskMap[*ig.StoryboardID] = &igCopy
					}
				}
			}
		}
	}

	// 批量查询进行中的视频生成任务
	videoGenTaskMap := make(map[uint]*models.VideoGeneration) // storyboard_id -> processing task
	if len(storyboardIDs) > 0 {
		var processingVideoGens []models.VideoGeneration
		if err := s.db.Where("scene_id IN ? AND status = ?", storyboardIDs, models.VideoStatusProcessing).
			Order("created_at DESC").
			Find(&processingVideoGens).Error; err == nil {
			for _, vg := range processingVideoGens {
				if vg.StoryboardID != nil {
					if _, exists := videoGenTaskMap[*vg.StoryboardID]; !exists {
						vgCopy := vg
						videoGenTaskMap[*vg.StoryboardID] = &vgCopy
					}
				}
			}
		}
	}

	// 构建返回结果
	prevByStoryboardID := make(map[uint]*models.Storyboard)
	prevStoryboardIDs := make([]uint, 0)
	for i := range storyboards {
		if i == 0 {
			continue
		}
		prevByStoryboardID[storyboards[i].ID] = &storyboards[i-1]
		if storyboards[i].ReusePrevLastFrame {
			prevStoryboardIDs = append(prevStoryboardIDs, storyboards[i-1].ID)
		}
	}

	prevPromptMap := make(map[uint]*models.FramePrompt)
	if len(prevStoryboardIDs) > 0 {
		var prevPrompts []models.FramePrompt
		if err := s.db.Where("storyboard_id IN ? AND frame_type = ?", prevStoryboardIDs, models.FrameTypeLast).
			Order("updated_at DESC").
			Find(&prevPrompts).Error; err == nil {
			for i := range prevPrompts {
				prompt := &prevPrompts[i]
				if _, exists := prevPromptMap[prompt.StoryboardID]; !exists {
					prevPromptMap[prompt.StoryboardID] = prompt
				}
			}
		}
	}

	prevImageMap := make(map[uint][]PrevLastImageInfo)
	if len(prevStoryboardIDs) > 0 {
		var prevImages []models.ImageGeneration
		if err := s.db.Where(
			"storyboard_id IN ? AND frame_type = ? AND status = ?",
			prevStoryboardIDs,
			models.FrameTypeLast,
			models.ImageStatusCompleted,
		).Order("created_at DESC").Find(&prevImages).Error; err == nil {
			for _, img := range prevImages {
				if img.StoryboardID == nil {
					continue
				}
				frameType := ""
				if img.FrameType != nil {
					frameType = *img.FrameType
				}
				prevImageMap[*img.StoryboardID] = append(prevImageMap[*img.StoryboardID], PrevLastImageInfo{
					ID:          img.ID,
					StoryboardID: *img.StoryboardID,
					FrameType:   frameType,
					ImageURL:    img.ImageURL,
					Status:      string(img.Status),
				})
			}
		}
	}

	var result []SceneCompositionInfo
	for _, storyboard := range storyboards {
		storyboardInfo := SceneCompositionInfo{
			ID:               storyboard.ID,
			StoryboardNumber: storyboard.StoryboardNumber,
			Title:            storyboard.Title,
			Description:      storyboard.Description,
			ShotType:         storyboard.ShotType,
			Angle:            storyboard.Angle,
			Movement:         storyboard.Movement,
			Location:         storyboard.Location,
			Time:             storyboard.Time,
			Duration:         storyboard.Duration,
			Action:           storyboard.Action,
			Dialogue:         storyboard.Dialogue,
			Result:           storyboard.Result,
			Atmosphere:       storyboard.Atmosphere,
			BgmPrompt:        storyboard.BgmPrompt,
			SoundEffect:      storyboard.SoundEffect,
			ImagePrompt:      storyboard.ImagePrompt,
			VideoPrompt:      storyboard.VideoPrompt,
			SceneID:          storyboard.SceneID,
			ReusePrevLastFrame: storyboard.ReusePrevLastFrame,
		}

		// 直接使用关联的角色信息，必要时修复同名角色关联
		if len(storyboard.Characters) > 0 {
			needsRepair := false
			desiredIDs := make([]uint, 0, len(storyboard.Characters))
			desiredIDSet := make(map[uint]struct{})
			for _, char := range storyboard.Characters {
				currentChar := char
				if (!hasCharacterImage(char)) && char.Name != "" && visualCountByName[char.Name] == 1 {
					if best, ok := bestVisualByName[char.Name]; ok && best.ID != char.ID {
						currentChar = best
						needsRepair = true
					}
				}

				if _, exists := desiredIDSet[currentChar.ID]; !exists {
					desiredIDs = append(desiredIDs, currentChar.ID)
					desiredIDSet[currentChar.ID] = struct{}{}
				}
				storyboardChar := SceneCharacterInfo{
					ID:       currentChar.ID,
					Name:     currentChar.Name,
					ImageURL: currentChar.ImageURL,
				}
				storyboardInfo.Characters = append(storyboardInfo.Characters, storyboardChar)
			}

			if needsRepair {
				var newCharacters []models.Character
				if err := s.db.Where("id IN ?", desiredIDs).Find(&newCharacters).Error; err == nil {
					if err := s.db.Model(&storyboard).Association("Characters").Replace(&newCharacters); err != nil {
						s.log.Warnw("Failed to repair storyboard characters", "storyboard_id", storyboard.ID, "error", err)
					}
				} else {
					s.log.Warnw("Failed to load characters for repair", "storyboard_id", storyboard.ID, "error", err)
				}
			}
		}

		// 添加场景信息
		if storyboard.SceneID != nil {
			if scene, ok := sceneMap[*storyboard.SceneID]; ok {
				storyboardInfo.Background = &SceneBackgroundInfo{
					ID:       scene.ID,
					Location: scene.Location,
					Time:     scene.Time,
					ImageURL: scene.ImageURL,
					Status:   scene.Status,
				}
			}
		}

		// 添加合成图片
		if imageURL, ok := imageGenMap[storyboard.ID]; ok {
			storyboardInfo.ComposedImage = &imageURL
		}

		// 添加视频URL
		if storyboard.VideoURL != nil {
			storyboardInfo.VideoURL = storyboard.VideoURL
		}

		// 添加进行中的图片生成任务信息
		if imageTask, ok := imageGenTaskMap[storyboard.ID]; ok {
			storyboardInfo.ImageGenerationID = &imageTask.ID
			statusStr := string(imageTask.Status)
			storyboardInfo.ImageGenerationStatus = &statusStr
		}

		// 添加进行中的视频生成任务信息
		if videoTask, ok := videoGenTaskMap[storyboard.ID]; ok {
			storyboardInfo.VideoGenerationID = &videoTask.ID
			statusStr := string(videoTask.Status)
			storyboardInfo.VideoGenerationStatus = &statusStr
		}

		if storyboard.ReusePrevLastFrame {
			if prevStoryboard, ok := prevByStoryboardID[storyboard.ID]; ok {
				prevID := prevStoryboard.ID
				prevNum := prevStoryboard.StoryboardNumber
				storyboardInfo.PrevLastStoryboardID = &prevID
				storyboardInfo.PrevLastStoryboardNum = &prevNum

				if prompt := prevPromptMap[prevID]; prompt != nil {
					promptText := prompt.Prompt
					storyboardInfo.PrevLastPrompt = &promptText
				}
				if images := prevImageMap[prevID]; len(images) > 0 {
					storyboardInfo.PrevLastImages = images
				}
			}
		}

		result = append(result, storyboardInfo)
	}

	return result, nil
}

func (s *StoryboardCompositionService) CreateScene(req *CreateSceneRequest) (*models.Scene, error) {
	if req.DramaID == 0 {
		return nil, fmt.Errorf("invalid drama id")
	}

	location := strings.TrimSpace(req.Location)
	timeValue := strings.TrimSpace(req.Time)
	if location == "" || timeValue == "" {
		return nil, fmt.Errorf("location and time are required")
	}

	// 验证Drama存在
	var drama models.Drama
	if err := s.db.Where("id = ? ", req.DramaID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("drama not found")
		}
		return nil, err
	}

	// 可选校验Episode归属
	if req.EpisodeID != nil {
		var episode models.Episode
		if err := s.db.Where("id = ? AND drama_id = ?", *req.EpisodeID, req.DramaID).First(&episode).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("episode not found")
			}
			return nil, err
		}
	}

	prompt := ""
	if req.Prompt != nil {
		prompt = strings.TrimSpace(*req.Prompt)
	}
	if prompt == "" {
		prompt = fmt.Sprintf("%s, %s", location, timeValue)
	}

	scene := models.Scene{
		DramaID:   req.DramaID,
		EpisodeID: req.EpisodeID,
		Location:  location,
		Time:      timeValue,
		Prompt:    prompt,
		Status:    "pending",
	}

	if err := s.db.Create(&scene).Error; err != nil {
		return nil, err
	}

	s.log.Infow("Scene created", "scene_id", scene.ID, "drama_id", req.DramaID)
	return &scene, nil
}

func (s *StoryboardCompositionService) UpdateSceneDetails(sceneID string, req *UpdateSceneDetailsRequest) error {
	var scene models.Scene
	if err := s.db.Where("id = ?", sceneID).First(&scene).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("scene not found")
		}
		return err
	}

	updates := make(map[string]interface{})
	if req.Location != nil {
		location := strings.TrimSpace(*req.Location)
		if location == "" {
			return fmt.Errorf("location is required")
		}
		updates["location"] = location
	}
	if req.Time != nil {
		timeValue := strings.TrimSpace(*req.Time)
		if timeValue == "" {
			return fmt.Errorf("time is required")
		}
		updates["time"] = timeValue
	}
	if req.Prompt != nil {
		updates["prompt"] = strings.TrimSpace(*req.Prompt)
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	if err := s.db.Model(&scene).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update scene: %w", err)
	}

	s.log.Infow("Scene details updated", "scene_id", sceneID, "updates", updates)
	return nil
}

type UpdateSceneRequest struct {
	SceneID     *uint   `json:"scene_id"`
	Characters  []uint  `json:"characters"` // 改为存储角色ID数组
	Location    *string `json:"location"`
	Time        *string `json:"time"`
	Action      *string `json:"action"`
	Dialogue    *string `json:"dialogue"`
	Description *string `json:"description"`
	Duration    *int    `json:"duration"`
	ImagePrompt *string `json:"image_prompt"`
	VideoPrompt *string `json:"video_prompt"`
}

func (s *StoryboardCompositionService) UpdateScene(sceneID string, req *UpdateSceneRequest) error {
	// 获取分镜并验证权限
	var storyboard models.Storyboard
	err := s.db.Preload("Episode.Drama").Where("id = ?", sceneID).First(&storyboard).Error
	if err != nil {
		return fmt.Errorf("scene not found")
	}

	// 构建更新数据
	updates := make(map[string]interface{})

	// 更新背景ID
	if req.SceneID != nil {
		updates["scene_id"] = req.SceneID
	}

	// 更新角色列表（直接存储ID数组）
	if req.Characters != nil {
		charactersJSON, err := json.Marshal(req.Characters)
		if err != nil {
			return fmt.Errorf("failed to serialize characters: %w", err)
		}
		updates["characters"] = charactersJSON
	}

	// 更新场景信息字段
	if req.Location != nil {
		updates["location"] = req.Location
	}
	if req.Time != nil {
		updates["time"] = req.Time
	}
	if req.Action != nil {
		updates["action"] = req.Action
	}
	if req.Dialogue != nil {
		updates["dialogue"] = req.Dialogue
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.ImagePrompt != nil {
		updates["image_prompt"] = req.ImagePrompt
	}
	if req.VideoPrompt != nil {
		updates["video_prompt"] = req.VideoPrompt
	}

	// 执行更新
	if len(updates) > 0 {
		if err := s.db.Model(&models.Storyboard{}).Where("id = ?", sceneID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update scene: %w", err)
		}
	}

	s.log.Infow("Scene updated", "scene_id", sceneID, "updates", updates)
	return nil
}

type GenerateSceneImageRequest struct {
	SceneID uint   `json:"scene_id"`
	Prompt  string `json:"prompt"`
	Model   string `json:"model"`
}

func (s *StoryboardCompositionService) GenerateSceneImage(req *GenerateSceneImageRequest) (*models.ImageGeneration, error) {
	// 获取场景并验证权限
	var scene models.Scene
	err := s.db.Where("id = ?", req.SceneID).First(&scene).Error
	if err != nil {
		return nil, fmt.Errorf("scene not found")
	}

	// 验证权限：通过DramaID查询Drama
	var drama models.Drama
	if err := s.db.Where("id = ? ", scene.DramaID).First(&drama).Error; err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	// 构建场景图片生成提示词
	prompt := req.Prompt
	if prompt == "" {
		// 使用场景的Prompt字段
		prompt = scene.Prompt
		if prompt == "" {
			// 如果Prompt为空，使用Location和Time构建
			prompt = fmt.Sprintf("%s场景，%s", scene.Location, scene.Time)
		}
		s.log.Infow("Using scene prompt", "scene_id", req.SceneID, "prompt", prompt)
	}

	// 使用imageGen服务直接生成
	if s.imageGen != nil {
		genReq := &GenerateImageRequest{
			SceneID:   &req.SceneID,
			DramaID:   fmt.Sprintf("%d", scene.DramaID),
			ImageType: string(models.ImageTypeScene),
			Prompt:    prompt,
			Model:     req.Model,   // 使用用户指定的模型
			Size:      "2560x1440", // 3,686,400像素，满足doubao模型最低要求（16:9比例）
			Quality:   "standard",
		}
		imageGen, err := s.imageGen.GenerateImage(genReq)
		if err != nil {
			return nil, fmt.Errorf("failed to generate image: %w", err)
		}

		// 更新场景的image_url
		if imageGen.ImageURL != nil {
			scene.ImageURL = imageGen.ImageURL
			scene.Status = "generated"
			if err := s.db.Save(&scene).Error; err != nil {
				s.log.Errorw("Failed to update scene image url", "error", err)
			}
		}

		s.log.Infow("Scene image generation created", "scene_id", req.SceneID, "image_gen_id", imageGen.ID)
		return imageGen, nil
	}

	return nil, fmt.Errorf("image generation service not available")
}

func (s *StoryboardCompositionService) DeleteScene(sceneID string) error {
	var scene models.Scene
	if err := s.db.Where("id = ?", sceneID).First(&scene).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("scene not found")
		}
		return fmt.Errorf("failed to find scene: %w", err)
	}

	// 删除场景
	if err := s.db.Delete(&scene).Error; err != nil {
		return fmt.Errorf("failed to delete scene: %w", err)
	}

	s.log.Infow("Scene deleted successfully", "scene_id", sceneID)
	return nil
}

func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
