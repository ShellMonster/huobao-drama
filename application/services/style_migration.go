package services

import (
	"fmt"
	"strings"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/cache"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
)

var legacyStyleKeys = []string{
	"realistic",
	"anime_cinematic",
	"anime_isekai",
	"fantasy_cartoon",
	"ink_wash",
}

func MigrateLegacyStyleKeys(db *gorm.DB, log *logger.Logger) error {
	if len(legacyStyleKeys) == 0 {
		return nil
	}

	var updated int64
	if err := db.Model(&models.Drama{}).
		Where("style IN ?", legacyStyleKeys).
		Count(&updated).Error; err != nil {
		return err
	}
	if updated > 0 {
		if err := db.Model(&models.Drama{}).
			Where("style IN ?", legacyStyleKeys).
			Update("style", realisticStyleKey).Error; err != nil {
			return err
		}
	}

	var removed int64
	if err := db.Model(&models.Style{}).
		Where("key IN ?", legacyStyleKeys).
		Count(&removed).Error; err != nil {
		return err
	}
	if removed > 0 {
		if err := db.Where("key IN ?", legacyStyleKeys).
			Delete(&models.Style{}).Error; err != nil {
			return err
		}
		if err := ensureDefaultStyle(db); err != nil {
			return err
		}
	}

	if updated > 0 {
		cache.BumpNamespace(cache.NamespaceDramaDetail)
		cache.BumpNamespace(cache.NamespaceDramaList)
	}
	if updated > 0 || removed > 0 {
		cache.BumpNamespace(cache.NamespaceStyles)
		log.Infow("Legacy style key migration completed", "dramas_updated", updated, "styles_removed", removed)
	}

	return nil
}

type promptRow struct {
	ID     uint   `gorm:"column:id"`
	Prompt string `gorm:"column:prompt"`
}

type storyboardPromptRow struct {
	ID          uint    `gorm:"column:id"`
	ImagePrompt *string `gorm:"column:image_prompt"`
	VideoPrompt *string `gorm:"column:video_prompt"`
}

func MigrateLegacyStylePlaceholders(db *gorm.DB, log *logger.Logger) error {
	updated := 0

	count, err := migratePromptTable(db, "image_generations", "prompt")
	if err != nil {
		return err
	}
	updated += count

	count, err = migratePromptTable(db, "video_generations", "prompt")
	if err != nil {
		return err
	}
	updated += count

	count, err = migratePromptTable(db, "frame_prompts", "prompt")
	if err != nil {
		return err
	}
	updated += count

	count, err = migratePromptTable(db, "scenes", "prompt")
	if err != nil {
		return err
	}
	updated += count

	count, err = migrateStoryboardPrompts(db)
	if err != nil {
		return err
	}
	updated += count

	if updated > 0 {
		cache.BumpNamespace(cache.NamespaceDramaDetail)
		cache.BumpNamespace(cache.NamespaceStoryboards)
		cache.BumpNamespace(cache.NamespaceImageList)
		cache.BumpNamespace(cache.NamespaceImageDetail)
		cache.BumpNamespace(cache.NamespaceVideoList)
		cache.BumpNamespace(cache.NamespaceVideoDetail)
		log.Infow("Legacy style prompt migration completed", "updated", updated)
	}

	return nil
}

func migratePromptTable(db *gorm.DB, table string, column string) (int, error) {
	where, args := legacyPromptWhere(column)
	if where == "" {
		return 0, nil
	}

	updated := 0
	rows := []promptRow{}
	err := db.Table(table).
		Select(fmt.Sprintf("id, %s as prompt", column)).
		Where(where, args...).
		FindInBatches(&rows, 200, func(tx *gorm.DB, batch int) error {
			for _, row := range rows {
				newPrompt := replaceLegacyStyleTokensWithPlaceholder(row.Prompt)
				if newPrompt == row.Prompt {
					continue
				}
				if err := db.Table(table).Where("id = ?", row.ID).Update(column, newPrompt).Error; err != nil {
					return err
				}
				updated++
			}
			return nil
		}).Error

	return updated, err
}

func migrateStoryboardPrompts(db *gorm.DB) (int, error) {
	where, args := legacyPromptWhere("image_prompt", "video_prompt")
	if where == "" {
		return 0, nil
	}

	updated := 0
	rows := []storyboardPromptRow{}
	err := db.Table("storyboards").
		Select("id, image_prompt, video_prompt").
		Where(where, args...).
		FindInBatches(&rows, 200, func(tx *gorm.DB, batch int) error {
			for _, row := range rows {
				changes := map[string]interface{}{}
				if row.ImagePrompt != nil {
					newPrompt := replaceLegacyStyleTokensWithPlaceholder(*row.ImagePrompt)
					if newPrompt != *row.ImagePrompt {
						changes["image_prompt"] = newPrompt
					}
				}
				if row.VideoPrompt != nil {
					newPrompt := replaceLegacyStyleTokensWithPlaceholder(*row.VideoPrompt)
					if newPrompt != *row.VideoPrompt {
						changes["video_prompt"] = newPrompt
					}
				}
				if len(changes) == 0 {
					continue
				}
				if err := db.Table("storyboards").Where("id = ?", row.ID).Updates(changes).Error; err != nil {
					return err
				}
				updated += len(changes)
			}
			return nil
		}).Error

	return updated, err
}

func legacyPromptWhere(columns ...string) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	for _, column := range columns {
		columnClause, columnArgs := legacyPromptLike(column)
		if columnClause == "" {
			continue
		}
		clauses = append(clauses, fmt.Sprintf("(%s)", columnClause))
		args = append(args, columnArgs...)
	}

	if len(clauses) == 0 {
		return "", nil
	}

	return strings.Join(clauses, " OR "), args
}

func legacyPromptLike(column string) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	lowerColumn := fmt.Sprintf("lower(%s)", column)
	for _, token := range legacyEnglishStyleTokens {
		clauses = append(clauses, fmt.Sprintf("%s LIKE ?", lowerColumn))
		args = append(args, "%"+token+"%")
	}
	for _, token := range legacyChineseStyleTokens {
		clauses = append(clauses, fmt.Sprintf("%s LIKE ?", column))
		args = append(args, "%"+token+"%")
	}

	if len(clauses) == 0 {
		return "", nil
	}
	return strings.Join(clauses, " OR "), args
}
