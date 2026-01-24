package ai

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/drama-generator/backend/pkg/config"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
)

type OpenAISDKClient struct {
	client openai.Client
	model  string
}

const openaiSDKDefaultTimeout = 10 * time.Minute

func openaiSDKTimeout() time.Duration {
	return config.DurationFromSeconds(config.GetTuning().HTTPTimeout.AISeconds, openaiSDKDefaultTimeout)
}

func NewOpenAISDKClient(baseURL, apiKey, model string) *OpenAISDKClient {
	opts := []option.RequestOption{}
	if strings.TrimSpace(apiKey) != "" {
		opts = append(opts, option.WithAPIKey(strings.TrimSpace(apiKey)))
	}
	timeout := openaiSDKTimeout()
	httpClient := &http.Client{Timeout: timeout}
	opts = append(opts, option.WithHTTPClient(httpClient))
	if normalizedBaseURL := normalizeOpenAIBaseURL(baseURL); normalizedBaseURL != "" {
		opts = append(opts, option.WithBaseURL(normalizedBaseURL))
	}
	client := openai.NewClient(opts...)

	if model == "" {
		defaultModel := strings.TrimSpace(config.GetTuning().Defaults.Models.OpenAIText)
		if defaultModel != "" {
			model = defaultModel
		} else {
			model = "gpt-4o"
		}
	}

	return &OpenAISDKClient{
		client: client,
		model:  model,
	}
}

func normalizeOpenAIBaseURL(apiBase string) string {
	base := strings.TrimSpace(apiBase)
	if base == "" {
		return ""
	}

	base = strings.TrimRight(base, "/")
	if strings.Contains(base, "/chat/completions") {
		base = strings.Split(base, "/chat/completions")[0]
		base = strings.TrimRight(base, "/")
	}
	if strings.Contains(base, "/v1/") {
		base = strings.Split(base, "/v1/")[0] + "/v1"
		return base
	}
	if strings.HasSuffix(base, "/v1") {
		return base
	}
	return base + "/v1"
}

func (c *OpenAISDKClient) GenerateText(prompt string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, error) {
	req := &ChatCompletionRequest{Model: c.model}
	for _, option := range options {
		option(req)
	}
	timeout := openaiSDKTimeout()

	model := req.Model
	if model == "" {
		model = c.model
	}

	messages := []openai.ChatCompletionMessageParamUnion{}
	if strings.TrimSpace(systemPrompt) != "" {
		messages = append(messages, openai.SystemMessage(systemPrompt))
	}
	messages = append(messages, openai.UserMessage(prompt))

	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(model),
		Messages: messages,
	}
	if req.Temperature != 0 {
		params.Temperature = param.NewOpt(req.Temperature)
	}
	if req.TopP != 0 {
		params.TopP = param.NewOpt(req.TopP)
	}
	if req.MaxTokens > 0 {
		params.MaxTokens = param.NewOpt(int64(req.MaxTokens))
	}
	if req.RequireJSON {
		responseFormat := shared.NewResponseFormatJSONObjectParam()
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &responseFormat,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := c.client.Chat.Completions.New(
		ctx,
		params,
		option.WithRequestTimeout(timeout),
	)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	content := resp.Choices[0].Message.Content
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty response content")
	}

	return content, nil
}

func (c *OpenAISDKClient) GenerateVisionText(prompt string, imageURLs []string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, error) {
	req := &ChatCompletionRequest{Model: c.model}
	for _, option := range options {
		option(req)
	}
	timeout := openaiSDKTimeout()

	model := req.Model
	if model == "" {
		model = c.model
	}

	parts := []openai.ChatCompletionContentPartUnionParam{}
	if strings.TrimSpace(prompt) != "" {
		parts = append(parts, openai.TextContentPart(prompt))
	}
	for _, url := range imageURLs {
		trimmed := strings.TrimSpace(url)
		if trimmed == "" {
			continue
		}
		parts = append(parts, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{URL: trimmed}))
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("empty vision prompt")
	}

	messages := []openai.ChatCompletionMessageParamUnion{}
	if strings.TrimSpace(systemPrompt) != "" {
		messages = append(messages, openai.SystemMessage(systemPrompt))
	}
	messages = append(messages, openai.UserMessage(parts))

	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(model),
		Messages: messages,
	}
	if req.Temperature != 0 {
		params.Temperature = param.NewOpt(req.Temperature)
	}
	if req.TopP != 0 {
		params.TopP = param.NewOpt(req.TopP)
	}
	if req.MaxTokens > 0 {
		params.MaxTokens = param.NewOpt(int64(req.MaxTokens))
	}
	if req.RequireJSON {
		responseFormat := shared.NewResponseFormatJSONObjectParam()
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &responseFormat,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := c.client.Chat.Completions.New(
		ctx,
		params,
		option.WithRequestTimeout(timeout),
	)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	content := resp.Choices[0].Message.Content
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty response content")
	}

	return content, nil
}

func (c *OpenAISDKClient) TestConnection() error {
	_, err := c.GenerateText("Hello", "")
	return err
}
