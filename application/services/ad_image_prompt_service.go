package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/ai"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/utils"
	"gorm.io/gorm"
)

type AdImagePromptService struct {
	db        *gorm.DB
	aiService *AIService
	log       *logger.Logger
	storage   storage.Storage
}

func NewAdImagePromptService(db *gorm.DB, log *logger.Logger, localStorage storage.Storage) *AdImagePromptService {
	return &AdImagePromptService{
		db:        db,
		aiService: NewAIService(db, log),
		log:       log,
		storage:   localStorage,
	}
}

type GenerateAdTextPromptRequest struct {
	DramaID string `json:"drama_id" binding:"required"`
	BrandID *uint  `json:"brand_id"`
	SpecID  *uint  `json:"spec_id"`
	Prompt  string `json:"prompt" binding:"required,min=5,max=4000"`
	Count   int    `json:"count"`
}

type GenerateAdImagePromptRequest struct {
	DramaID  string `json:"drama_id" binding:"required"`
	BrandID  *uint  `json:"brand_id"`
	SpecID   *uint  `json:"spec_id"`
	ImageURL string `json:"image_url" binding:"required"`
	Count    int    `json:"count"`
}

func (s *AdImagePromptService) GenerateTextPrompts(req *GenerateAdTextPromptRequest) ([]string, error) {
	if req.Count <= 0 {
		req.Count = 10
	}
	context, err := s.buildAdPromptContext(req.DramaID, req.BrandID, req.SpecID)
	if err != nil {
		return nil, err
	}
	isEnglish := isEnglishPrompt(req.Prompt)
	systemPrompt := adPromptSystemText(isEnglish)
	userPrompt := adPromptUserText(req.Prompt, context, req.Count, isEnglish)

	text, err := s.aiService.GenerateTextJSON(userPrompt, systemPrompt, ai.WithMaxTokens(2500))
	if err != nil {
		return nil, err
	}
	return parsePromptList(text)
}

func (s *AdImagePromptService) GeneratePromptsFromImage(req *GenerateAdImagePromptRequest) ([]string, error) {
	if req.Count <= 0 {
		req.Count = 10
	}
	context, err := s.buildAdPromptContext(req.DramaID, req.BrandID, req.SpecID)
	if err != nil {
		return nil, err
	}
	systemPrompt := adPromptSystemText(false)
	userPrompt := adPromptUserImageText(context, req.Count)
	imageURL := req.ImageURL
	if s.storage != nil && s.storage.IsLocalURL(imageURL) {
		if dataURL, err := s.storage.ToDataURL(imageURL); err == nil {
			imageURL = dataURL
		}
	}

	text, err := s.aiService.GenerateVisionTextJSON(userPrompt, []string{imageURL}, systemPrompt, ai.WithMaxTokens(2500))
	if err != nil {
		return nil, err
	}
	return parsePromptList(text)
}

func (s *AdImagePromptService) buildAdPromptContext(dramaID string, brandID *uint, specID *uint) (string, error) {
	var drama models.Drama
	if err := s.db.Where("id = ?", dramaID).First(&drama).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("drama not found")
		}
		return "", err
	}

	resolvedBrandID := brandID
	if resolvedBrandID == nil {
		resolvedBrandID = drama.BrandID
	}
	resolvedSpecID := specID
	if resolvedSpecID == nil {
		resolvedSpecID = drama.SpecID
	}

	var brand *models.Brand
	if resolvedBrandID != nil {
		var brandModel models.Brand
		if err := s.db.Where("id = ?", *resolvedBrandID).First(&brandModel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", fmt.Errorf("brand not found")
			}
			return "", err
		}
		brand = &brandModel
	}

	var spec *models.BrandSpec
	if resolvedSpecID != nil {
		if resolvedBrandID == nil {
			return "", fmt.Errorf("brand_id is required when spec_id is provided")
		}
		var specModel models.BrandSpec
		if err := s.db.Where("id = ? AND brand_id = ?", *resolvedSpecID, *resolvedBrandID).First(&specModel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", fmt.Errorf("brand spec not found")
			}
			return "", err
		}
		spec = &specModel
	}

	return buildAdContextText(brand, spec), nil
}

func buildAdContextText(brand *models.Brand, spec *models.BrandSpec) string {
	parts := []string{}
	if brand != nil {
		name := brand.Name
		if brand.DisplayName != nil && *brand.DisplayName != "" {
			name = *brand.DisplayName
		}
		parts = append(parts, fmt.Sprintf("品牌名称：%s", name))
		if brand.Description != nil && *brand.Description != "" {
			parts = append(parts, fmt.Sprintf("品牌描述：%s", *brand.Description))
		}
		if brand.LogoURL != nil && *brand.LogoURL != "" {
			parts = append(parts, fmt.Sprintf("品牌Logo：%s", *brand.LogoURL))
		}
	}
	if spec != nil {
		parts = append(parts, fmt.Sprintf("素材规范名称：%s", spec.Name))
		if spec.Description != nil && *spec.Description != "" {
			parts = append(parts, fmt.Sprintf("规范描述：%s", *spec.Description))
		}
		if len(spec.AllowedSizes) > 0 {
			parts = append(parts, fmt.Sprintf("允许尺寸：%s", strings.TrimSpace(string(spec.AllowedSizes))))
		}
		if len(spec.AspectRatios) > 0 {
			parts = append(parts, fmt.Sprintf("允许比例：%s", strings.TrimSpace(string(spec.AspectRatios))))
		}
		if len(spec.SafeArea) > 0 {
			parts = append(parts, fmt.Sprintf("安全区：%s", strings.TrimSpace(string(spec.SafeArea))))
		}
		if len(spec.LogoRules) > 0 {
			parts = append(parts, fmt.Sprintf("Logo露出规则：%s", strings.TrimSpace(string(spec.LogoRules))))
		}
		if len(spec.TextRules) > 0 {
			parts = append(parts, fmt.Sprintf("文字规则：%s", strings.TrimSpace(string(spec.TextRules))))
		}
	}

	if len(parts) == 0 {
		return "无品牌与规范约束"
	}
	return strings.Join(parts, "\n")
}

func adPromptSystemText(isEnglish bool) string {
	if isEnglish {
		return "You are a senior advertising creative and image prompt expert. Generate concise, production-ready prompts."
	}
	return "你是资深广告创意与图像提示词专家，负责生成可直接用于AI图像生成的广告提示词。"
}

func adPromptUserText(demand, context string, count int, isEnglish bool) string {
	if isEnglish {
		return fmt.Sprintf(`Ad requirement:
%s

Brand & spec context:
%s

Please generate %d image prompts. Each prompt should include scene, subject, action, composition, lighting, mood, and ad placement hints.
Return only a JSON object, no other text. Do NOT include markdown code blocks. Start with { and end with }.

Example:
{
  "prompts": [
    "A clean modern kitchen at sunrise, a young mother preparing breakfast, warm light, soft shadows, cozy mood, empty space on the right for ad copy",
    "A bright city street in the afternoon, confident office worker holding a coffee, medium shot, crisp lighting, upbeat mood, billboard space above"
  ]
}`, strings.TrimSpace(demand), context, count)
	}
	return fmt.Sprintf(`广告需求描述：
%s

品牌与规范信息：
%s

请生成%d条图像提示词，每条包含场景、主体、动作、构图、光线、情绪与可用的广告露出位置。
仅返回JSON对象（包含prompts数组），不要输出其它内容，不要包含markdown代码块。直接以 { 开头，以 } 结尾。

示例：
{
  "prompts": [
    "清晨的现代厨房，年轻母亲准备早餐，中景构图，暖光柔影，温馨情绪，右侧留出广告文案位置",
    "午后城市街道，干练上班族手持咖啡，半身构图，明亮光线，积极氛围，上方预留广告牌空间"
  ]
}`, strings.TrimSpace(demand), context, count)
}

func adPromptUserImageText(context string, count int) string {
	return fmt.Sprintf(`请基于参考图理解场景、人物、动作与构图，生成%d条广告图像提示词。
结合品牌与规范信息：
%s

每条提示词需包含场景、主体、动作、构图、光线、情绪与可用的广告露出位置。
仅返回JSON对象（包含prompts数组），不要输出其它内容，不要包含markdown代码块。直接以 { 开头，以 } 结尾。

示例：
{
  "prompts": [
    "海边沙滩黄昏，人物手持饮料微笑，侧逆光，金色氛围，画面右下留出品牌露出",
    "夜间霓虹街景，人物站立看向镜头，中景构图，冷暖对比光，画面上方留出标题区"
  ]
}`, count, context)
}

func parsePromptList(text string) ([]string, error) {
	prompts, err := utils.ParseAIJSONList[string](text, []string{"prompts", "items"}, false)
	if err == nil {
		return normalizePrompts(prompts), nil
	}

	if lines := parsePromptLines(text); len(lines) > 0 {
		return normalizePrompts(lines), nil
	}

	return nil, fmt.Errorf("failed to parse prompt list")
}

var promptLinePattern = regexp.MustCompile(`^\s*(?:[-*•]|\d+[.)、]|\(\d+\)|（\d+）)\s*(.+)$`)

func parsePromptLines(text string) []string {
	cleaned := strings.TrimSpace(text)
	if cleaned == "" {
		return nil
	}
	cleaned = regexp.MustCompile("(?m)^```json\\s*").ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile("(?m)^```\\s*").ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile("(?m)```\\s*$").ReplaceAllString(cleaned, "")
	lines := strings.Split(cleaned, "\n")
	items := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		match := promptLinePattern.FindStringSubmatch(trimmed)
		if len(match) < 2 {
			continue
		}
		item := strings.TrimSpace(match[1])
		item = strings.Trim(item, "\"'`，,.;；。")
		if item == "" {
			continue
		}
		items = append(items, item)
	}
	if len(items) < 2 {
		return nil
	}
	return items
}

func normalizePrompts(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}
