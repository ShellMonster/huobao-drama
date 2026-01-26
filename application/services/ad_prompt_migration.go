package services

import (
	"encoding/json"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
)

func MigrateLegacyAdPromptItems(db *gorm.DB, log *logger.Logger) error {
	if db == nil {
		return nil
	}

	var prompts []models.AdImagePrompt
	if err := db.Table("ad_image_prompts AS p").
		Select("p.*").
		Joins("LEFT JOIN ad_image_prompt_items AS i ON i.prompt_id = p.id").
		Where("i.id IS NULL").
		Where("p.prompts IS NOT NULL AND p.prompts != '' AND p.prompts != '[]'").
		Order("p.id ASC").
		Find(&prompts).Error; err != nil {
		return err
	}

	if len(prompts) == 0 {
		return nil
	}

	service := NewAdImagePromptService(db, log, nil)
	for _, prompt := range prompts {
		if len(prompt.Prompts) == 0 {
			continue
		}
		var list []string
		if err := json.Unmarshal(prompt.Prompts, &list); err != nil {
			log.Warnw("Failed to parse legacy ad prompts", "prompt_id", prompt.ID, "error", err)
			continue
		}
		if len(list) == 0 {
			continue
		}
		if _, err := service.createPromptItems(db, &prompt, list); err != nil {
			log.Warnw("Failed to migrate legacy ad prompts", "prompt_id", prompt.ID, "error", err)
			continue
		}
		log.Infow("Migrated legacy ad prompts", "prompt_id", prompt.ID, "count", len(list))
	}

	return nil
}
