package ai

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
)

type OpenAISDKClient struct {
	client openai.Client
	model  string
}

func NewOpenAISDKClient(baseURL, apiKey, model string) *OpenAISDKClient {
	opts := []option.RequestOption{}
	if strings.TrimSpace(apiKey) != "" {
		opts = append(opts, option.WithAPIKey(strings.TrimSpace(apiKey)))
	}
	if strings.TrimSpace(baseURL) != "" {
		opts = append(opts, option.WithBaseURL(strings.TrimSpace(baseURL)))
	}
	client := openai.NewClient(opts...)

	if model == "" {
		model = "gpt-4o"
	}

	return &OpenAISDKClient{
		client: client,
		model:  model,
	}
}

func (c *OpenAISDKClient) GenerateText(prompt string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, error) {
	req := &ChatCompletionRequest{Model: c.model}
	for _, option := range options {
		option(req)
	}

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

	resp, err := c.client.Chat.Completions.New(context.Background(), params)
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
