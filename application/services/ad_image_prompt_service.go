package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
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

type adPromptTemplate struct {
	Key       string
	Label     string
	LabelEN   string
	Ratios    []string
	Layout    string
	LayoutEN  string
	Copy      string
	CopyEN    string
	Visual    string
	VisualEN  string
	UseCase   string
	UseCaseEN string
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
	context, resolvedBrandID, resolvedSpecID, spec, err := s.buildAdPromptContext(req.DramaID, req.BrandID, req.SpecID)
	if err != nil {
		return nil, err
	}
	isEnglish := isEnglishPrompt(req.Prompt)
	systemPrompt := adPromptSystemText(isEnglish)
	templates := selectAdPromptTemplates(spec)
	userPrompt := adPromptUserText(req.Prompt, context, req.Count, isEnglish, templates)

	text, err := s.aiService.GenerateTextJSON(userPrompt, systemPrompt, ai.WithMaxTokens(2500))
	if err != nil {
		return nil, err
	}
	prompts, err := parsePromptList(text)
	if err != nil {
		return nil, err
	}
	if len(prompts) > 0 {
		_ = s.savePromptRecord(models.AdPromptTypeText, req.DramaID, resolvedBrandID, resolvedSpecID, &req.Prompt, nil, prompts)
	}
	return prompts, nil
}

func (s *AdImagePromptService) GeneratePromptsFromImage(req *GenerateAdImagePromptRequest) ([]string, error) {
	if req.Count <= 0 {
		req.Count = 10
	}
	context, resolvedBrandID, resolvedSpecID, spec, err := s.buildAdPromptContext(req.DramaID, req.BrandID, req.SpecID)
	if err != nil {
		return nil, err
	}
	systemPrompt := adPromptSystemText(false)
	templates := selectAdPromptTemplates(spec)
	userPrompt := adPromptUserImageText(context, req.Count, templates)
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
	prompts, err := parsePromptList(text)
	if err != nil {
		return nil, err
	}
	if len(prompts) > 0 {
		_ = s.savePromptRecord(models.AdPromptTypeImage, req.DramaID, resolvedBrandID, resolvedSpecID, nil, &req.ImageURL, prompts)
	}
	return prompts, nil
}

type GetAdPromptRequest struct {
	DramaID string `form:"drama_id" binding:"required"`
	BrandID *uint  `form:"brand_id"`
	SpecID  *uint  `form:"spec_id"`
}

type LatestAdPromptResult struct {
	TextPrompts  []string `json:"text_prompts"`
	ImagePrompts []string `json:"image_prompts"`
}

func (s *AdImagePromptService) GetLatestPrompts(req *GetAdPromptRequest) (*LatestAdPromptResult, error) {
	_, resolvedBrandID, resolvedSpecID, _, err := s.buildAdPromptContext(req.DramaID, req.BrandID, req.SpecID)
	if err != nil {
		return nil, err
	}

	textPrompts, err := s.loadLatestPromptList(req.DramaID, resolvedBrandID, resolvedSpecID, models.AdPromptTypeText)
	if err != nil {
		return nil, err
	}
	imagePrompts, err := s.loadLatestPromptList(req.DramaID, resolvedBrandID, resolvedSpecID, models.AdPromptTypeImage)
	if err != nil {
		return nil, err
	}

	return &LatestAdPromptResult{
		TextPrompts:  textPrompts,
		ImagePrompts: imagePrompts,
	}, nil
}

func (s *AdImagePromptService) buildAdPromptContext(dramaID string, brandID *uint, specID *uint) (string, *uint, *uint, *models.BrandSpec, error) {
	var drama models.Drama
	if err := s.db.Where("id = ?", dramaID).First(&drama).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil, nil, nil, fmt.Errorf("drama not found")
		}
		return "", nil, nil, nil, err
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
				return "", nil, nil, nil, fmt.Errorf("brand not found")
			}
			return "", nil, nil, nil, err
		}
		brand = &brandModel
	}

	var spec *models.BrandSpec
	if resolvedSpecID != nil {
		if resolvedBrandID == nil {
			return "", nil, nil, nil, fmt.Errorf("brand_id is required when spec_id is provided")
		}
		var specModel models.BrandSpec
		if err := s.db.Where("id = ? AND brand_id = ?", *resolvedSpecID, *resolvedBrandID).First(&specModel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil, nil, nil, fmt.Errorf("brand spec not found")
			}
			return "", nil, nil, nil, err
		}
		spec = &specModel
	}

	return buildAdContextText(brand, spec), resolvedBrandID, resolvedSpecID, spec, nil
}

func (s *AdImagePromptService) savePromptRecord(promptType string, dramaID string, brandID *uint, specID *uint, sourceText *string, sourceImageURL *string, prompts []string) error {
	data, err := json.Marshal(prompts)
	if err != nil {
		return err
	}
	dramaIDParsed, err := strconv.ParseUint(dramaID, 10, 32)
	if err != nil {
		return err
	}
	record := models.AdImagePrompt{
		DramaID:        uint(dramaIDParsed),
		BrandID:        brandID,
		SpecID:         specID,
		PromptType:     promptType,
		SourceText:     sourceText,
		SourceImageURL: sourceImageURL,
		Prompts:        data,
	}
	return s.db.Create(&record).Error
}

func (s *AdImagePromptService) loadLatestPromptList(dramaID string, brandID *uint, specID *uint, promptType string) ([]string, error) {
	query := s.db.Model(&models.AdImagePrompt{}).
		Where("drama_id = ? AND prompt_type = ?", dramaID, promptType)
	if brandID == nil {
		query = query.Where("brand_id IS NULL")
	} else {
		query = query.Where("brand_id = ?", *brandID)
	}
	if specID == nil {
		query = query.Where("spec_id IS NULL")
	} else {
		query = query.Where("spec_id = ?", *specID)
	}

	var record models.AdImagePrompt
	if err := query.Order("created_at DESC").First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []string{}, nil
		}
		return nil, err
	}

	var prompts []string
	if err := json.Unmarshal(record.Prompts, &prompts); err != nil {
		return nil, err
	}
	return prompts, nil
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

func defaultAdPromptTemplates() []adPromptTemplate {
	return []adPromptTemplate{
		{
			Key:       "splash-vertical",
			Label:     "App开屏竖版",
			LabelEN:   "App Splash Vertical",
			Ratios:    []string{"9:16", "3:5", "2:3"},
			Layout:    "上方主视觉，中下文案区，底部CTA/品牌安全区",
			LayoutEN:  "Top hero visual, mid-lower copy area, bottom CTA/brand safe zone",
			Copy:      "文案块置于底部或下半部留白区，确保安全区",
			CopyEN:    "Place copy in the bottom or lower whitespace safe area",
			Visual:    "强主视觉、层次清晰、商业棚拍级打光",
			VisualEN:  "Strong hero focus, layered composition, studio-grade lighting",
			UseCase:   "适合开屏/竖版海报",
			UseCaseEN: "Best for splash screens / vertical posters",
		},
		{
			Key:       "feed-horizontal",
			Label:     "信息流横版",
			LabelEN:   "In-feed Horizontal",
			Ratios:    []string{"16:9", "3:2", "4:3"},
			Layout:    "左侧主视觉、右侧独立文案区",
			LayoutEN:  "Left hero visual, right-side copy block",
			Copy:      "右侧安全区放标题/副标题/卖点",
			CopyEN:    "Place headline/subhead/benefits in the right safe area",
			Visual:    "主物突出，辅物点缀，商业广告质感",
			VisualEN:  "Hero-first focus with supporting elements, commercial polish",
			UseCase:   "适合信息流横版/电商横幅",
			UseCaseEN: "Best for horizontal feeds / ecommerce banners",
		},
		{
			Key:       "feed-square",
			Label:     "信息流方形",
			LabelEN:   "In-feed Square",
			Ratios:    []string{"1:1", "4:5"},
			Layout:    "上图下文案或中心视觉+底部文案条",
			LayoutEN:  "Top visual + bottom copy, or centered visual with bottom copy strip",
			Copy:      "底部安全区放短文案/CTA",
			CopyEN:    "Place short copy/CTA in the bottom safe area",
			Visual:    "简洁高对比，适合移动端信息流",
			VisualEN:  "Clean, high contrast, mobile-friendly",
			UseCase:   "适合方图/竖方信息流",
			UseCaseEN: "Best for square or portrait feeds",
		},
		{
			Key:       "banner-horizontal",
			Label:     "横幅Banner",
			LabelEN:   "Horizontal Banner",
			Ratios:    []string{"3:1", "2:1"},
			Layout:    "中心产品主体+左右装饰，文案块置于一侧",
			LayoutEN:  "Centered product focus with side decorations, copy block on one side",
			Copy:      "文案块与主视觉分区明显，保留留白",
			CopyEN:    "Clear separation between copy and visuals, preserve whitespace",
			Visual:    "横向延展、构图平衡、视觉冲击力强",
			VisualEN:  "Wide composition, balanced layout, strong visual impact",
			UseCase:   "适合横幅/活动banner",
			UseCaseEN: "Best for banners / campaign headers",
		},
	}
}

func normalizeRatio(input string) string {
	re := regexp.MustCompile(`(\d+)\s*[:xX/]\s*(\d+)`)
	match := re.FindStringSubmatch(input)
	if len(match) == 3 {
		return fmt.Sprintf("%s:%s", match[1], match[2])
	}
	return strings.TrimSpace(input)
}

func parseAspectRatios(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var ratios []string
	if err := json.Unmarshal(raw, &ratios); err == nil {
		return normalizeRatioList(ratios)
	}
	var ratioStr string
	if err := json.Unmarshal(raw, &ratioStr); err == nil {
		return normalizeRatioList(extractRatiosFromText(ratioStr))
	}
	return normalizeRatioList(extractRatiosFromText(string(raw)))
}

func normalizeRatioList(items []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(items))
	for _, item := range items {
		ratio := normalizeRatio(item)
		if ratio == "" {
			continue
		}
		if _, ok := seen[ratio]; ok {
			continue
		}
		seen[ratio] = struct{}{}
		result = append(result, ratio)
	}
	return result
}

func extractRatiosFromText(text string) []string {
	re := regexp.MustCompile(`\d+\s*[:xX/]\s*\d+`)
	matches := re.FindAllString(text, -1)
	result := make([]string, 0, len(matches))
	for _, item := range matches {
		result = append(result, item)
	}
	return result
}

func matchesRatio(templateRatios []string, ratios []string) bool {
	if len(ratios) == 0 {
		return false
	}
	ratioSet := map[string]struct{}{}
	for _, ratio := range ratios {
		ratioSet[normalizeRatio(ratio)] = struct{}{}
	}
	for _, ratio := range templateRatios {
		if _, ok := ratioSet[normalizeRatio(ratio)]; ok {
			return true
		}
	}
	return false
}

func selectAdPromptTemplates(spec *models.BrandSpec) []adPromptTemplate {
	templates := defaultAdPromptTemplates()
	if spec == nil {
		return templates
	}

	ratios := parseAspectRatios(spec.AspectRatios)
	if len(ratios) > 0 {
		matched := make([]adPromptTemplate, 0, len(templates))
		for _, template := range templates {
			if matchesRatio(template.Ratios, ratios) {
				matched = append(matched, template)
			}
		}
		if len(matched) > 0 {
			return matched
		}
	}

	name := strings.ToLower(spec.Name)
	desc := ""
	if spec.Description != nil {
		desc = strings.ToLower(*spec.Description)
	}
	keywordText := name + " " + desc
	filtered := make([]adPromptTemplate, 0, len(templates))
	for _, template := range templates {
		if strings.Contains(keywordText, "开屏") && template.Key == "splash-vertical" {
			filtered = append(filtered, template)
			continue
		}
		if strings.Contains(keywordText, "splash") && template.Key == "splash-vertical" {
			filtered = append(filtered, template)
			continue
		}
		if strings.Contains(keywordText, "信息流") && template.Key == "feed-horizontal" {
			filtered = append(filtered, template)
			continue
		}
		if (strings.Contains(keywordText, "feed") || strings.Contains(keywordText, "in-feed")) && template.Key == "feed-horizontal" {
			filtered = append(filtered, template)
			continue
		}
		if strings.Contains(keywordText, "横幅") || strings.Contains(keywordText, "banner") {
			if template.Key == "banner-horizontal" {
				filtered = append(filtered, template)
				continue
			}
		}
		if strings.Contains(keywordText, "方图") || strings.Contains(keywordText, "方形") {
			if template.Key == "feed-square" {
				filtered = append(filtered, template)
				continue
			}
		}
		if strings.Contains(keywordText, "square") && template.Key == "feed-square" {
			filtered = append(filtered, template)
			continue
		}
	}
	if len(filtered) > 0 {
		return filtered
	}
	return templates
}

func formatAdPromptTemplates(templates []adPromptTemplate, isEnglish bool) string {
	if len(templates) == 0 {
		if isEnglish {
			return "No template constraints."
		}
		return "无模板约束。"
	}
	lines := make([]string, 0, len(templates))
	for i, template := range templates {
		index := i + 1
		if isEnglish {
			label := template.LabelEN
			if label == "" {
				label = template.Label
			}
			layout := template.LayoutEN
			if layout == "" {
				layout = template.Layout
			}
			copyHint := template.CopyEN
			if copyHint == "" {
				copyHint = template.Copy
			}
			visual := template.VisualEN
			if visual == "" {
				visual = template.Visual
			}
			useCase := template.UseCaseEN
			if useCase == "" {
				useCase = template.UseCase
			}
			lines = append(lines, fmt.Sprintf(
				"%d) %s (Ratios: %s). Layout: %s. Copy area: %s. Visual: %s. Use case: %s.",
				index,
				label,
				strings.Join(template.Ratios, " / "),
				layout,
				copyHint,
				visual,
				useCase,
			))
			continue
		}
		lines = append(lines, fmt.Sprintf(
			"%d) %s（适用比例：%s）。版式：%s。文案区：%s。视觉：%s。用途：%s。",
			index,
			template.Label,
			strings.Join(template.Ratios, " / "),
			template.Layout,
			template.Copy,
			template.Visual,
			template.UseCase,
		))
	}
	return strings.Join(lines, "\n")
}

func adPromptSystemText(isEnglish bool) string {
	if isEnglish {
		return "You are a senior advertising creative and image prompt expert, focused on app splash screens, in-feed ads, and banner layouts. Generate concise, production-ready prompts."
	}
	return "你是资深广告创意与图像提示词专家，擅长App开屏、信息流与横幅广告版式，负责生成可直接用于AI图像生成的广告提示词。"
}

func adPromptUserText(demand, context string, count int, isEnglish bool, templates []adPromptTemplate) string {
	if isEnglish {
		return fmt.Sprintf(`Ad requirement:
%s

Brand & spec context:
%s

Available layout templates (prioritize matching allowed ratios and try to cover different templates):
%s

Please generate %d high-quality ad image prompts. Each prompt must include:
- Aspect ratio + layout structure (e.g., left visual/right copy, top visual/bottom copy)
- Scene, subject, action, composition, lighting, mood
- Clear copy-safe area and ad placement zone
- One short English ad copy line suitable for the layout

Style: commercial studio lighting, realistic photography or high-fidelity 3D, clean typography. No brand logos, no price numbers, no cartoon style, no distortion.
Return only a JSON object, no other text. Do NOT include markdown code blocks. Start with { and end with }.

Example:
{
  "prompts": [
    "16:9 horizontal feed ad, left-side hero product arrangement on a clean counter, right-side copy block safe area, bright commercial lighting, upbeat mood, English copy: \"Limited time, shop now\"",
    "9:16 app splash, top hero scene with products, bottom copy-safe area, warm studio lighting, premium red & gold palette, English copy: \"Festive deals today\""
  ]
}`, strings.TrimSpace(demand), context, formatAdPromptTemplates(templates, true), count)
	}
	return fmt.Sprintf(`广告需求描述：
%s

品牌与规范信息：
%s

可用广告版式模板（优先匹配允许比例，尽量覆盖不同模板）：
%s

请生成%d条高质量广告图像提示词。每条提示词必须包含：
- 画面比例与版式结构（例如左视觉右文案、上图下文）
- 场景、主体、动作、构图、光线、情绪
- 明确的文案安全区与广告露出位置
- 一句符合需求的中文广告文案

风格要求：商业棚拍级打光，高写实摄影或高写实3D风格，中文排版正确，不出现品牌标识、价格数字、卡通或畸变。
仅返回JSON对象（包含prompts数组），不要输出其它内容，不要包含markdown代码块。直接以 { 开头，以 } 结尾。

示例：
{
  "prompts": [
    "16:9信息流横版，左侧主视觉为写实年货组合陈列，右侧独立文案区与安全区，商业棚拍级打光，酒红与金色质感，文案“年货别等，今日到手”",
    "9:16开屏竖版，上方主视觉为高质感礼盒与产品组合，底部文案区留白，暖光柔影，促销氛围，文案“限时好价，立即囤货”"
  ]
}`, strings.TrimSpace(demand), context, formatAdPromptTemplates(templates, false), count)
}

func adPromptUserImageText(context string, count int, templates []adPromptTemplate) string {
	return fmt.Sprintf(`请基于参考图理解场景、人物、动作与构图，生成%d条广告图像提示词。
结合品牌与规范信息：
%s

可用广告版式模板（优先匹配允许比例，尽量覆盖不同模板）：
%s

每条提示词需包含场景、主体、动作、构图、光线、情绪，并在不破坏主体的前提下强化广告版式（明确文案区、安全区、露出位）。
请加入一句简短中文广告文案，适配对应版式位置。
风格要求：商业棚拍级打光，高写实摄影或高写实3D风格，中文排版正确，不出现品牌标识、价格数字、卡通或畸变。
仅返回JSON对象（包含prompts数组），不要输出其它内容，不要包含markdown代码块。直接以 { 开头，以 } 结尾。

示例：
{
  "prompts": [
    "9:16开屏竖版，保留海边沙滩主体人物，底部留出文案安全区，暖色侧逆光，高质感，文案“夏日清凉，即刻畅享”",
    "16:9信息流横版，保留夜间霓虹街景主视觉，右侧独立文案区，冷暖对比光，文案“城市夜色，点亮你的故事”"
  ]
}`, count, context, formatAdPromptTemplates(templates, false))
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
