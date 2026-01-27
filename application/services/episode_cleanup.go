package services

import (
	"fmt"

	models "github.com/drama-generator/backend/domain/models"
	"gorm.io/gorm"
)

func cleanupEpisodeData(tx *gorm.DB, episodeIDs []uint) error {
	if len(episodeIDs) == 0 {
		return nil
	}

	var sceneIDs []uint
	if err := tx.Model(&models.Scene{}).
		Where("episode_id IN ?", episodeIDs).
		Pluck("id", &sceneIDs).Error; err != nil {
		return fmt.Errorf("failed to load scene ids: %w", err)
	}

	var storyboardIDs []uint
	if err := tx.Model(&models.Storyboard{}).
		Where("episode_id IN ?", episodeIDs).
		Pluck("id", &storyboardIDs).Error; err != nil {
		return fmt.Errorf("failed to load storyboard ids: %w", err)
	}

	if len(storyboardIDs) > 0 {
		if err := tx.Table("storyboard_characters").
			Where("storyboard_id IN ?", storyboardIDs).
			Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete storyboard characters: %w", err)
		}

		if err := tx.Table("frame_prompts").
			Where("storyboard_id IN ?", storyboardIDs).
			Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete frame prompts: %w", err)
		}

		if err := tx.Table("frame_prompt_tasks").
			Where("storyboard_id IN ?", storyboardIDs).
			Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete frame prompt tasks: %w", err)
		}

		if err := tx.Where("storyboard_id IN ?", storyboardIDs).
			Delete(&models.ImageGeneration{}).Error; err != nil {
			return fmt.Errorf("failed to delete storyboard image generations: %w", err)
		}

		if err := tx.Where("storyboard_id IN ?", storyboardIDs).
			Delete(&models.VideoGeneration{}).Error; err != nil {
			return fmt.Errorf("failed to delete video generations: %w", err)
		}

		if err := tx.Table("assets").
			Where("storyboard_id IN ?", storyboardIDs).
			Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete storyboard assets: %w", err)
		}
	}

	if err := tx.Table("episode_characters").
		Where("episode_id IN ?", episodeIDs).
		Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to delete episode characters: %w", err)
	}

	if len(sceneIDs) > 0 {
		if err := tx.Where("scene_id IN ?", sceneIDs).
			Delete(&models.ImageGeneration{}).Error; err != nil {
			return fmt.Errorf("failed to delete scene image generations: %w", err)
		}
	}

	if err := tx.Table("video_merges").
		Where("episode_id IN ?", episodeIDs).
		Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to delete video merges: %w", err)
	}

	if err := tx.Table("assets").
		Where("episode_id IN ?", episodeIDs).
		Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to delete episode assets: %w", err)
	}

	var timelineIDs []uint
	if err := tx.Table("timelines").
		Where("episode_id IN ?", episodeIDs).
		Pluck("id", &timelineIDs).Error; err != nil {
		return fmt.Errorf("failed to load timeline ids: %w", err)
	}

	if len(timelineIDs) > 0 {
		trackSubQuery := tx.Table("timeline_tracks").Select("id").Where("timeline_id IN ?", timelineIDs)
		clipSubQuery := tx.Table("timeline_clips").Select("id").Where("track_id IN (?)", trackSubQuery)

		if err := tx.Table("clip_effects").
			Where("clip_id IN (?)", clipSubQuery).
			Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete clip effects: %w", err)
		}

		if err := tx.Table("clip_transitions").
			Where("id IN (?)", tx.Table("timeline_clips").
				Select("transition_in_id").
				Where("track_id IN (?)", trackSubQuery)).
			Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete clip transitions (in): %w", err)
		}

		if err := tx.Table("clip_transitions").
			Where("id IN (?)", tx.Table("timeline_clips").
				Select("transition_out_id").
				Where("track_id IN (?)", trackSubQuery)).
			Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete clip transitions (out): %w", err)
		}

		if err := tx.Table("timeline_clips").
			Where("track_id IN (?)", trackSubQuery).
			Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete timeline clips: %w", err)
		}

		if err := tx.Table("timeline_tracks").
			Where("timeline_id IN ?", timelineIDs).
			Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete timeline tracks: %w", err)
		}

		if err := tx.Table("timelines").
			Where("id IN ?", timelineIDs).
			Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete timelines: %w", err)
		}
	}

	if err := tx.Where("episode_id IN ?", episodeIDs).Delete(&models.Storyboard{}).Error; err != nil {
		return fmt.Errorf("failed to delete storyboards: %w", err)
	}

	if err := tx.Where("episode_id IN ?", episodeIDs).Delete(&models.Scene{}).Error; err != nil {
		return fmt.Errorf("failed to delete scenes: %w", err)
	}

	return nil
}

func collectEpisodeCharacterIDs(tx *gorm.DB, episodeIDs []uint) ([]uint, error) {
	if len(episodeIDs) == 0 {
		return nil, nil
	}

	unique := make(map[uint]struct{})

	var episodeCharacterIDs []uint
	if err := tx.Table("episode_characters").
		Where("episode_id IN ?", episodeIDs).
		Pluck("character_id", &episodeCharacterIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to load episode characters: %w", err)
	}
	for _, id := range episodeCharacterIDs {
		unique[id] = struct{}{}
	}

	var storyboardIDs []uint
	if err := tx.Model(&models.Storyboard{}).
		Where("episode_id IN ?", episodeIDs).
		Pluck("id", &storyboardIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to load storyboard ids: %w", err)
	}

	if len(storyboardIDs) > 0 {
		var storyboardCharacterIDs []uint
		if err := tx.Table("storyboard_characters").
			Where("storyboard_id IN ?", storyboardIDs).
			Pluck("character_id", &storyboardCharacterIDs).Error; err != nil {
			return nil, fmt.Errorf("failed to load storyboard characters: %w", err)
		}
		for _, id := range storyboardCharacterIDs {
			unique[id] = struct{}{}
		}
	}

	result := make([]uint, 0, len(unique))
	for id := range unique {
		result = append(result, id)
	}
	return result, nil
}

func cleanupOrphanCharacters(tx *gorm.DB, dramaID uint, candidateIDs []uint) error {
	if len(candidateIDs) == 0 {
		return nil
	}

	subQueryEpisode := tx.Table("episode_characters").Select("character_id")
	subQueryStoryboard := tx.Table("storyboard_characters").Select("character_id")

	var imageGenCharacterIDs []uint
	if err := tx.Model(&models.ImageGeneration{}).
		Where("character_id IN ?", candidateIDs).
		Distinct().
		Pluck("character_id", &imageGenCharacterIDs).Error; err != nil {
		return fmt.Errorf("failed to load character image generations: %w", err)
	}
	imageGenSet := make(map[uint]struct{}, len(imageGenCharacterIDs))
	for _, id := range imageGenCharacterIDs {
		imageGenSet[id] = struct{}{}
	}

	var candidates []models.Character
	if err := tx.Where("drama_id = ?", dramaID).
		Where("id IN ?", candidateIDs).
		Where("id NOT IN (?)", subQueryEpisode).
		Where("id NOT IN (?)", subQueryStoryboard).
		Find(&candidates).Error; err != nil {
		return fmt.Errorf("failed to load orphan characters: %w", err)
	}

	var deleteIDs []uint
	for _, character := range candidates {
		if hasCharacterImage(character) || hasCharacterReferenceImages(character) || hasCharacterDetails(character) {
			continue
		}
		if _, exists := imageGenSet[character.ID]; exists {
			continue
		}
		deleteIDs = append(deleteIDs, character.ID)
	}

	if len(deleteIDs) == 0 {
		return nil
	}

	if err := tx.Where("id IN ?", deleteIDs).Delete(&models.Character{}).Error; err != nil {
		return fmt.Errorf("failed to delete orphan characters: %w", err)
	}

	return nil
}
