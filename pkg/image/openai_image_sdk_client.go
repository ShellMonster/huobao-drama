package image

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/drama-generator/backend/pkg/config"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
)

type OpenAIImageSDKClient struct {
	client openai.Client
	model  string
}

func NewOpenAIImageSDKClient(baseURL, apiKey, model string) *OpenAIImageSDKClient {
	opts := []option.RequestOption{}
	if strings.TrimSpace(apiKey) != "" {
		opts = append(opts, option.WithAPIKey(strings.TrimSpace(apiKey)))
	}
	if strings.TrimSpace(baseURL) != "" {
		opts = append(opts, option.WithBaseURL(strings.TrimSpace(baseURL)))
	}
	client := openai.NewClient(opts...)

	if model == "" {
		model = strings.TrimSpace(config.GetTuning().Defaults.Models.OpenAIImage)
		if model == "" {
			model = openai.ImageModelDallE3
		}
	}

	return &OpenAIImageSDKClient{
		client: client,
		model:  model,
	}
}

func (c *OpenAIImageSDKClient) GenerateImage(prompt string, opts ...ImageOption) (*ImageResult, error) {
	options := &ImageOptions{
		Size:    defaultImageSize("1024x1024"),
		Quality: defaultImageQuality("standard"),
	}

	for _, opt := range opts {
		opt(options)
	}

	model := c.model
	if options.Model != "" {
		model = options.Model
	}

	params := openai.ImageGenerateParams{
		Prompt: prompt,
		N:      param.NewOpt(int64(1)),
	}
	if model != "" {
		params.Model = openai.ImageModel(model)
	}
	if options.Size != "" {
		params.Size = openai.ImageGenerateParamsSize(options.Size)
	}
	if options.Quality != "" {
		params.Quality = openai.ImageGenerateParamsQuality(options.Quality)
	}
	if options.Style != "" {
		params.Style = openai.ImageGenerateParamsStyle(options.Style)
	}
	params.ResponseFormat = openai.ImageGenerateParamsResponseFormatURL

	timeout := config.DurationFromSeconds(config.GetTuning().HTTPTimeout.ImageSeconds, 10*time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resp, err := c.client.Images.Generate(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no image generated")
	}

	imageURL := resp.Data[0].URL
	if imageURL == "" && resp.Data[0].B64JSON != "" {
		imageURL = "data:image/png;base64," + resp.Data[0].B64JSON
	}
	if imageURL == "" {
		return nil, fmt.Errorf("no image url in response")
	}

	width, height := parseImageSize(options.Size)

	return &ImageResult{
		Status:    "completed",
		ImageURL:  imageURL,
		Completed: true,
		Width:     width,
		Height:    height,
	}, nil
}

func (c *OpenAIImageSDKClient) GetTaskStatus(taskID string) (*ImageResult, error) {
	return nil, fmt.Errorf("not supported for OpenAI image generation (synchronous)")
}
