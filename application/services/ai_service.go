package services

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/ai"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
)

type AIService struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewAIService(db *gorm.DB, log *logger.Logger) *AIService {
	return &AIService{
		db:  db,
		log: log,
	}
}

const (
	defaultRetryAttempts = 3
	defaultRetryDelay    = 2 * time.Second
	defaultRetryMaxDelay = 12 * time.Second
	defaultRetryJitterMs = 500
)

type AIRequest struct {
	ServiceType        string
	Model              string
	Prompt             string
	SystemPrompt       string
	ImageURLs          []string
	RequireJSON        bool
	Options            []func(*ai.ChatCompletionRequest)
	MaxAttempts        int
	RetryDelay         time.Duration
	RetryMaxDelay      time.Duration
	AllowJSONFallback  bool
	AllowModelFallback bool
}

type CreateAIConfigRequest struct {
	ServiceType   string            `json:"service_type" binding:"required,oneof=text image video"`
	Name          string            `json:"name" binding:"required,min=1,max=100"`
	Provider      string            `json:"provider" binding:"required"`
	BaseURL       string            `json:"base_url" binding:"required,url"`
	APIKey        string            `json:"api_key" binding:"required"`
	Model         models.ModelField `json:"model" binding:"required"`
	Endpoint      string            `json:"endpoint"`
	QueryEndpoint string            `json:"query_endpoint"`
	Priority      int               `json:"priority"`
	IsDefault     bool              `json:"is_default"`
	Settings      string            `json:"settings"`
}

type UpdateAIConfigRequest struct {
	Name          string             `json:"name" binding:"omitempty,min=1,max=100"`
	Provider      string             `json:"provider"`
	BaseURL       string             `json:"base_url" binding:"omitempty,url"`
	APIKey        string             `json:"api_key"`
	Model         *models.ModelField `json:"model"`
	Endpoint      string             `json:"endpoint"`
	QueryEndpoint string             `json:"query_endpoint"`
	Priority      *int               `json:"priority"`
	IsDefault     bool               `json:"is_default"`
	IsActive      bool               `json:"is_active"`
	Settings      string             `json:"settings"`
}

type TestConnectionRequest struct {
	BaseURL  string            `json:"base_url" binding:"required,url"`
	APIKey   string            `json:"api_key" binding:"required"`
	Model    models.ModelField `json:"model" binding:"required"`
	Provider string            `json:"provider"`
	Endpoint string            `json:"endpoint"`
}

func (s *AIService) CreateConfig(req *CreateAIConfigRequest) (*models.AIServiceConfig, error) {
	// 根据 provider 和 service_type 自动设置 endpoint
	endpoint := req.Endpoint
	queryEndpoint := req.QueryEndpoint

	if endpoint == "" {
		switch req.Provider {
		case "gemini", "google":
			if req.ServiceType == "text" {
				endpoint = "/v1beta/models/{model}:generateContent"
			} else if req.ServiceType == "image" {
				endpoint = "/v1beta/models/{model}:generateContent"
			}
		case "openai":
			if req.ServiceType == "text" {
				endpoint = "/chat/completions"
			} else if req.ServiceType == "image" {
				endpoint = "/images/generations"
			} else if req.ServiceType == "video" {
				endpoint = "/videos"
				if queryEndpoint == "" {
					queryEndpoint = "/videos/{taskId}"
				}
			}
		case "chatfire":
			if req.ServiceType == "text" {
				endpoint = "/chat/completions"
			} else if req.ServiceType == "image" {
				endpoint = "/images/generations"
			} else if req.ServiceType == "video" {
				endpoint = "/video/generations"
				if queryEndpoint == "" {
					queryEndpoint = "/video/task/{taskId}"
				}
			}
		case "doubao", "volcengine", "volces":
			if req.ServiceType == "video" {
				endpoint = "/contents/generations/tasks"
				if queryEndpoint == "" {
					queryEndpoint = "/generations/tasks/{taskId}"
				}
			}
		default:
			// 默认使用 OpenAI 格式
			if req.ServiceType == "text" {
				endpoint = "/chat/completions"
			} else if req.ServiceType == "image" {
				endpoint = "/images/generations"
			}
		}
	}

	config := &models.AIServiceConfig{
		ServiceType:   req.ServiceType,
		Name:          req.Name,
		Provider:      req.Provider,
		BaseURL:       req.BaseURL,
		APIKey:        req.APIKey,
		Model:         req.Model,
		Endpoint:      endpoint,
		QueryEndpoint: queryEndpoint,
		Priority:      req.Priority,
		IsDefault:     req.IsDefault,
		IsActive:      true,
		Settings:      req.Settings,
	}

	if err := s.db.Create(config).Error; err != nil {
		s.log.Errorw("Failed to create AI config", "error", err)
		return nil, err
	}

	s.log.Infow("AI config created", "config_id", config.ID, "provider", req.Provider, "endpoint", endpoint)
	return config, nil
}

func (s *AIService) GetConfig(configID uint) (*models.AIServiceConfig, error) {
	var config models.AIServiceConfig
	err := s.db.Where("id = ? ", configID).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("config not found")
		}
		return nil, err
	}
	return &config, nil
}

func (s *AIService) ListConfigs(serviceType string) ([]models.AIServiceConfig, error) {
	var configs []models.AIServiceConfig
	query := s.db

	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}

	err := query.Order("priority DESC, created_at DESC").Find(&configs).Error
	if err != nil {
		s.log.Errorw("Failed to list AI configs", "error", err)
		return nil, err
	}

	return configs, nil
}

func (s *AIService) UpdateConfig(configID uint, req *UpdateAIConfigRequest) (*models.AIServiceConfig, error) {
	var config models.AIServiceConfig
	if err := s.db.Where("id = ? ", configID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("config not found")
		}
		return nil, err
	}

	tx := s.db.Begin()

	// 不再需要is_default独占逻辑

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Provider != "" {
		updates["provider"] = req.Provider
	}
	if req.BaseURL != "" {
		updates["base_url"] = req.BaseURL
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	if req.Model != nil && len(*req.Model) > 0 {
		updates["model"] = *req.Model
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}

	// 如果提供了 provider，根据 provider 和 service_type 自动设置 endpoint
	if req.Provider != "" && req.Endpoint == "" {
		provider := req.Provider
		serviceType := config.ServiceType

		switch provider {
		case "gemini", "google":
			if serviceType == "text" || serviceType == "image" {
				updates["endpoint"] = "/v1beta/models/{model}:generateContent"
			}
		case "openai":
			if serviceType == "text" {
				updates["endpoint"] = "/chat/completions"
			} else if serviceType == "image" {
				updates["endpoint"] = "/images/generations"
			} else if serviceType == "video" {
				updates["endpoint"] = "/videos"
				updates["query_endpoint"] = "/videos/{taskId}"
			}
		case "chatfire":
			if serviceType == "text" {
				updates["endpoint"] = "/chat/completions"
			} else if serviceType == "image" {
				updates["endpoint"] = "/images/generations"
			} else if serviceType == "video" {
				updates["endpoint"] = "/video/generations"
				updates["query_endpoint"] = "/video/task/{taskId}"
			}
		}
	} else if req.Endpoint != "" {
		updates["endpoint"] = req.Endpoint
	}

	// 允许清空query_endpoint，所以不检查是否为空
	updates["query_endpoint"] = req.QueryEndpoint
	if req.Settings != "" {
		updates["settings"] = req.Settings
	}
	updates["is_default"] = req.IsDefault
	updates["is_active"] = req.IsActive

	if err := tx.Model(&config).Updates(updates).Error; err != nil {
		tx.Rollback()
		s.log.Errorw("Failed to update AI config", "error", err)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	s.log.Infow("AI config updated", "config_id", configID)
	return &config, nil
}

func (s *AIService) DeleteConfig(configID uint) error {
	result := s.db.Where("id = ? ", configID).Delete(&models.AIServiceConfig{})

	if result.Error != nil {
		s.log.Errorw("Failed to delete AI config", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("config not found")
	}

	s.log.Infow("AI config deleted", "config_id", configID)
	return nil
}

func (s *AIService) TestConnection(req *TestConnectionRequest) error {
	s.log.Infow("TestConnection called", "baseURL", req.BaseURL, "provider", req.Provider, "endpoint", req.Endpoint, "modelCount", len(req.Model))

	// 使用第一个模型进行测试
	model := ""
	if len(req.Model) > 0 {
		model = req.Model[0]
	}
	s.log.Infow("Using model for test", "model", model, "provider", req.Provider)

	// 根据 provider 参数选择客户端
	var client ai.AIClient
	var endpoint string

	switch req.Provider {
	case "gemini", "google":
		// Gemini
		s.log.Infow("Using Gemini client", "baseURL", req.BaseURL)
		endpoint = "/v1beta/models/{model}:generateContent"
		client = ai.NewGeminiClient(req.BaseURL, req.APIKey, model, endpoint)
	case "openai", "chatfire":
		// OpenAI 格式（包括 chatfire 等）
		s.log.Infow("Using OpenAI-compatible client", "baseURL", req.BaseURL, "provider", req.Provider)
		endpoint = req.Endpoint
		if endpoint == "" {
			endpoint = "/chat/completions"
		}
		client = ai.NewOpenAIClient(req.BaseURL, req.APIKey, model, endpoint)
	default:
		// 默认使用 OpenAI 格式
		s.log.Infow("Using default OpenAI-compatible client", "baseURL", req.BaseURL)
		endpoint = req.Endpoint
		if endpoint == "" {
			endpoint = "/chat/completions"
		}
		client = ai.NewOpenAIClient(req.BaseURL, req.APIKey, model, endpoint)
	}

	s.log.Infow("Calling TestConnection on client", "endpoint", endpoint)
	err := client.TestConnection()
	if err != nil {
		s.log.Errorw("TestConnection failed", "error", err)
	} else {
		s.log.Infow("TestConnection succeeded")
	}
	return err
}

func (s *AIService) GetDefaultConfig(serviceType string) (*models.AIServiceConfig, error) {
	var config models.AIServiceConfig
	// 按优先级降序获取第一个激活的配置
	err := s.db.Where("service_type = ? AND is_active = ?", serviceType, true).
		Order("priority DESC, created_at DESC").
		First(&config).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no active config found")
		}
		return nil, err
	}

	return &config, nil
}

// GetConfigForModel 根据服务类型和模型名称获取优先级最高的激活配置
func (s *AIService) GetConfigForModel(serviceType string, modelName string) (*models.AIServiceConfig, error) {
	var configs []models.AIServiceConfig
	err := s.db.Where("service_type = ? AND is_active = ?", serviceType, true).
		Order("priority DESC, created_at DESC").
		Find(&configs).Error

	if err != nil {
		return nil, err
	}

	// 查找包含指定模型的配置
	for _, config := range configs {
		for _, model := range config.Model {
			if model == modelName {
				return &config, nil
			}
		}
	}

	return nil, errors.New("no active config found for model: " + modelName)
}

func (s *AIService) GetAIClient(serviceType string) (ai.AIClient, error) {
	config, err := s.GetDefaultConfig(serviceType)
	if err != nil {
		return nil, err
	}
	return s.buildClientFromConfig(config, "")
}

// GetAIClientForModel 根据服务类型和模型名称获取对应的AI客户端
func (s *AIService) GetAIClientForModel(serviceType string, modelName string) (ai.AIClient, error) {
	config, err := s.GetConfigForModel(serviceType, modelName)
	if err != nil {
		return nil, err
	}
	return s.buildClientFromConfig(config, modelName)
}

func (s *AIService) GenerateText(prompt string, systemPrompt string, options ...func(*ai.ChatCompletionRequest)) (string, error) {
	return s.GenerateTextWithModel(prompt, systemPrompt, "", false, options...)
}

func (s *AIService) GenerateTextJSON(prompt string, systemPrompt string, options ...func(*ai.ChatCompletionRequest)) (string, error) {
	return s.GenerateTextWithModel(prompt, systemPrompt, "", true, options...)
}

func (s *AIService) GenerateVisionText(prompt string, imageURLs []string, systemPrompt string, options ...func(*ai.ChatCompletionRequest)) (string, error) {
	return s.GenerateVisionTextWithModel(prompt, imageURLs, systemPrompt, "", false, options...)
}

func (s *AIService) GenerateVisionTextJSON(prompt string, imageURLs []string, systemPrompt string, options ...func(*ai.ChatCompletionRequest)) (string, error) {
	return s.GenerateVisionTextWithModel(prompt, imageURLs, systemPrompt, "", true, options...)
}

func (s *AIService) GenerateTextWithModel(prompt string, systemPrompt string, model string, requireJSON bool, options ...func(*ai.ChatCompletionRequest)) (string, error) {
	req := &AIRequest{
		ServiceType:        "text",
		Model:              model,
		Prompt:             prompt,
		SystemPrompt:       systemPrompt,
		RequireJSON:        requireJSON,
		Options:            options,
		AllowJSONFallback:  true,
		AllowModelFallback: true,
	}
	return s.Generate(req)
}

func (s *AIService) GenerateVisionTextWithModel(prompt string, imageURLs []string, systemPrompt string, model string, requireJSON bool, options ...func(*ai.ChatCompletionRequest)) (string, error) {
	req := &AIRequest{
		ServiceType:        "text",
		Model:              model,
		Prompt:             prompt,
		SystemPrompt:       systemPrompt,
		ImageURLs:          imageURLs,
		RequireJSON:        requireJSON,
		Options:            options,
		AllowJSONFallback:  true,
		AllowModelFallback: true,
	}
	return s.Generate(req)
}

func (s *AIService) Generate(req *AIRequest) (string, error) {
	if req == nil {
		return "", fmt.Errorf("AI request is nil")
	}
	serviceType := req.ServiceType
	if serviceType == "" {
		serviceType = "text"
	}

	client, cfg, err := s.getClientWithConfig(serviceType, req.Model, req.AllowModelFallback)
	if err != nil {
		return "", fmt.Errorf("failed to get AI client: %w", err)
	}

	call := func(requireJSON bool) (string, error) {
		options := buildAIOptions(req.Options, requireJSON)
		if len(req.ImageURLs) == 0 {
			return client.GenerateText(req.Prompt, req.SystemPrompt, options...)
		}
		visionClient, ok := client.(ai.VisionClient)
		if !ok {
			return "", fmt.Errorf("current provider does not support vision inputs")
		}
		return visionClient.GenerateVisionText(req.Prompt, req.ImageURLs, req.SystemPrompt, options...)
	}

	result, err := s.callWithRetry(req, call, req.RequireJSON)
	if err != nil && req.RequireJSON && req.AllowJSONFallback && isJSONModeUnsupported(err) {
		s.log.Warnw("JSON mode unsupported, retrying without response_format",
			"provider", safeConfigProvider(cfg),
			"model", safeConfigModel(cfg, req.Model),
			"error", err)
		return s.callWithRetry(req, call, false)
	}
	return result, err
}

func (s *AIService) getClientWithConfig(serviceType string, model string, allowFallback bool) (ai.AIClient, *models.AIServiceConfig, error) {
	if model != "" {
		cfg, err := s.GetConfigForModel(serviceType, model)
		if err == nil {
			client, buildErr := s.buildClientFromConfig(cfg, model)
			return client, cfg, buildErr
		}
		if !allowFallback {
			return nil, nil, err
		}
		s.log.Warnw("Failed to resolve model config, fallback to default", "model", model, "error", err)
	}

	cfg, err := s.GetDefaultConfig(serviceType)
	if err != nil {
		return nil, nil, err
	}
	client, buildErr := s.buildClientFromConfig(cfg, "")
	return client, cfg, buildErr
}

func (s *AIService) buildClientFromConfig(config *models.AIServiceConfig, modelOverride string) (ai.AIClient, error) {
	model := modelOverride
	if model == "" && len(config.Model) > 0 {
		model = config.Model[0]
	}

	endpoint := config.Endpoint
	if endpoint == "" {
		switch config.Provider {
		case "gemini", "google":
			endpoint = "/v1beta/models/{model}:generateContent"
		default:
			endpoint = "/chat/completions"
		}
	}

	switch config.Provider {
	case "gemini", "google":
		client, err := ai.NewGeminiSDKClient(config.BaseURL, config.APIKey, model)
		if err != nil {
			return nil, err
		}
		return client, nil
	case "openai":
		return ai.NewOpenAISDKClient(config.BaseURL, config.APIKey, model), nil
	default:
		return ai.NewOpenAIClient(config.BaseURL, config.APIKey, model, endpoint), nil
	}
}

func (s *AIService) callWithRetry(req *AIRequest, call func(requireJSON bool) (string, error), requireJSON bool) (string, error) {
	attempts := req.MaxAttempts
	if attempts <= 0 {
		attempts = defaultRetryAttempts
	}
	delay := req.RetryDelay
	if delay <= 0 {
		delay = defaultRetryDelay
	}
	maxDelay := req.RetryMaxDelay
	if maxDelay <= 0 {
		maxDelay = defaultRetryMaxDelay
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		resp, err := call(requireJSON)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !shouldRetryAIError(err) || attempt == attempts {
			break
		}
		sleep := delay
		if sleep > maxDelay {
			sleep = maxDelay
		}
		jitter := time.Duration(rand.Intn(defaultRetryJitterMs+1)) * time.Millisecond
		time.Sleep(sleep + jitter)
		delay = minDuration(delay*2, maxDelay)
	}
	return "", lastErr
}

func buildAIOptions(base []func(*ai.ChatCompletionRequest), requireJSON bool) []func(*ai.ChatCompletionRequest) {
	options := make([]func(*ai.ChatCompletionRequest), 0, len(base)+1)
	options = append(options, base...)
	options = append(options, ai.WithRequireJSON(requireJSON))
	return options
}

func shouldRetryAIError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout"),
		strings.Contains(msg, "temporarily unavailable"),
		strings.Contains(msg, "connection reset"),
		strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "connection closed"),
		strings.Contains(msg, "eof"),
		strings.Contains(msg, "rate limit"),
		strings.Contains(msg, "429"),
		strings.Contains(msg, "502"),
		strings.Contains(msg, "503"),
		strings.Contains(msg, "504"),
		strings.Contains(msg, "gateway"),
		strings.Contains(msg, "service unavailable"):
		return true
	default:
		return false
	}
}

func isJSONModeUnsupported(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "response_format") ||
		strings.Contains(msg, "json_object") ||
		strings.Contains(msg, "response_mime_type") ||
		strings.Contains(msg, "responsemimetype") ||
		strings.Contains(msg, "response mime")
}

func safeConfigProvider(cfg *models.AIServiceConfig) string {
	if cfg == nil {
		return ""
	}
	return cfg.Provider
}

func safeConfigModel(cfg *models.AIServiceConfig, override string) string {
	if override != "" {
		return override
	}
	if cfg == nil || len(cfg.Model) == 0 {
		return ""
	}
	return cfg.Model[0]
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
