package ai

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type GeminiSDKClient struct {
	client *genai.Client
	model  string
}

func NewGeminiSDKClient(baseURL, apiKey, model string) (*GeminiSDKClient, error) {
	if model == "" {
		model = "gemini-3-pro"
	}

	config := &genai.ClientConfig{
		APIKey:  strings.TrimSpace(apiKey),
		Backend: genai.BackendGeminiAPI,
	}
	if strings.TrimSpace(baseURL) != "" {
		config.HTTPOptions = genai.HTTPOptions{
			BaseURL: strings.TrimSpace(baseURL),
		}
	}

	client, err := genai.NewClient(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &GeminiSDKClient{
		client: client,
		model:  model,
	}, nil
}

func (c *GeminiSDKClient) GenerateText(prompt string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, error) {
	req := &ChatCompletionRequest{Model: c.model}
	for _, option := range options {
		option(req)
	}

	model := req.Model
	if model == "" {
		model = c.model
	}

	var config *genai.GenerateContentConfig
	if systemPrompt != "" || req.Temperature != 0 || req.TopP != 0 || req.MaxTokens > 0 {
		config = &genai.GenerateContentConfig{}
		if systemPrompt != "" {
			config.SystemInstruction = &genai.Content{
				Parts: []*genai.Part{{Text: systemPrompt}},
			}
		}
		if req.Temperature != 0 {
			temperature := float32(req.Temperature)
			config.Temperature = &temperature
		}
		if req.TopP != 0 {
			topP := float32(req.TopP)
			config.TopP = &topP
		}
		if req.MaxTokens > 0 {
			config.MaxOutputTokens = int32(req.MaxTokens)
		}
	}

	result, err := c.client.Models.GenerateContent(context.Background(), model, genai.Text(prompt), config)
	if err != nil {
		return "", err
	}

	text := result.Text()
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("empty response content")
	}

	return text, nil
}

func (c *GeminiSDKClient) TestConnection() error {
	_, err := c.GenerateText("Hello", "")
	return err
}
