package services

import (
	"encoding/json"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
)

func MigrateLegacyAdImagePromptTypes(db *gorm.DB, log *logger.Logger) error {
	if db == nil {
		return nil
	}

	var images []models.ImageGeneration
	if err := db.Where("image_type = ? AND (ad_prompt_type IS NULL OR ad_prompt_type = '')", "ad").
		Find(&images).Error; err != nil {
		return err
	}

	if len(images) == 0 {
		return nil
	}

	for _, image := range images {
		promptType := models.AdPromptTypeText
		if len(image.ReferenceImages) > 0 {
			var refs []string
			if err := json.Unmarshal(image.ReferenceImages, &refs); err == nil && len(refs) > 0 {
				promptType = models.AdPromptTypeImage
			}
		}

		if err := db.Model(&models.ImageGeneration{}).
			Where("id = ?", image.ID).
			Update("ad_prompt_type", promptType).Error; err != nil {
			log.Warnw("Failed to migrate ad prompt type", "image_id", image.ID, "error", err)
		}
	}

	return nil
}
