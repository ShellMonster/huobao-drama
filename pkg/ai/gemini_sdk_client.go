package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/utils"
	"google.golang.org/genai"
)

type GeminiSDKClient struct {
	client *genai.Client
	model  string
}

func NewGeminiSDKClient(baseURL, apiKey, model string) (*GeminiSDKClient, error) {
	if model == "" {
		model = strings.TrimSpace(config.GetTuning().Defaults.Models.GeminiText)
		if model == "" {
			model = "gemini-3-pro"
		}
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
	if systemPrompt != "" || req.Temperature != 0 || req.TopP != 0 || req.MaxTokens > 0 || req.RequireJSON {
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
		if req.RequireJSON {
			config.ResponseMIMEType = "application/json"
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

func (c *GeminiSDKClient) GenerateVisionText(prompt string, imageURLs []string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, error) {
	req := &ChatCompletionRequest{Model: c.model}
	for _, option := range options {
		option(req)
	}

	model := req.Model
	if model == "" {
		model = c.model
	}

	var config *genai.GenerateContentConfig
	if systemPrompt != "" || req.Temperature != 0 || req.TopP != 0 || req.MaxTokens > 0 || req.RequireJSON {
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
		if req.RequireJSON {
			config.ResponseMIMEType = "application/json"
		}
	}

	parts := make([]*genai.Part, 0, len(imageURLs)+1)
	hasImage := false
	for _, image := range imageURLs {
		data, mimeType, err := resolveVisionImage(image)
		if err != nil {
			continue
		}
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				Data:     data,
				MIMEType: mimeType,
			},
		})
		hasImage = true
	}
	if strings.TrimSpace(prompt) != "" {
		parts = append(parts, &genai.Part{Text: prompt})
	}
	if !hasImage {
		return c.GenerateText(prompt, systemPrompt, options...)
	}

	contents := []*genai.Content{{
		Role:  genai.RoleUser,
		Parts: parts,
	}}

	result, err := c.client.Models.GenerateContent(context.Background(), model, contents, config)
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

func resolveVisionImage(image string) ([]byte, string, error) {
	image = strings.TrimSpace(image)
	if image == "" {
		return nil, "", fmt.Errorf("empty image reference")
	}

	if strings.HasPrefix(image, "http://") || strings.HasPrefix(image, "https://") {
		resp, err := http.Get(image)
		if err != nil {
			return nil, "", fmt.Errorf("download image: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("download image failed with status: %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", fmt.Errorf("read image data: %w", err)
		}

		mimeType := resp.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = http.DetectContentType(data)
		}
		return data, mimeType, nil
	}

	if utils.IsDataURI(image) {
		return utils.ParseDataURI(image)
	}

	data, err := base64.StdEncoding.DecodeString(image)
	if err != nil {
		return nil, "", fmt.Errorf("decode base64 image: %w", err)
	}
	return data, "image/jpeg", nil
}
