package image

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/drama-generator/backend/pkg/config"
	"google.golang.org/genai"
)

type GeminiImageSDKClient struct {
	client *genai.Client
	model  string
}

func NewGeminiImageSDKClient(baseURL, apiKey, model string) (*GeminiImageSDKClient, error) {
	if model == "" {
		model = strings.TrimSpace(config.GetTuning().Defaults.Models.GeminiImage)
		if model == "" {
			model = "gemini-3-pro-image-preview"
		}
	}

	timeout := config.DurationFromSeconds(config.GetTuning().HTTPTimeout.ImageSeconds, 10*time.Minute)
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DisableKeepAlives:   true,
			ForceAttemptHTTP2:   false,
			MaxIdleConns:        0,
			MaxIdleConnsPerHost: 0,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	clientConfig := &genai.ClientConfig{
		APIKey:     strings.TrimSpace(apiKey),
		Backend:    genai.BackendGeminiAPI,
		HTTPClient: httpClient,
	}
	if strings.TrimSpace(baseURL) != "" {
		apiBase := strings.TrimRight(strings.TrimSpace(baseURL), "/")
		clientConfig.HTTPOptions = genai.HTTPOptions{
			BaseURL: apiBase,
		}
	}

	client, err := genai.NewClient(context.Background(), clientConfig)
	if err != nil {
		return nil, err
	}

	return &GeminiImageSDKClient{
		client: client,
		model:  model,
	}, nil
}

func (c *GeminiImageSDKClient) GenerateImage(prompt string, opts ...ImageOption) (*ImageResult, error) {
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

	promptText := prompt
	if options.NegativePrompt != "" {
		promptText += fmt.Sprintf("\n\nNegative prompt: %s", options.NegativePrompt)
	}

	imageSize := normalizeGeminiImageSize(options.Size, options.Width, options.Height)
	aspectRatio := normalizeGeminiAspectRatio(options.Size, options.Width, options.Height)
	if imageSize == "" && aspectRatio == "" && options.Size != "" {
		promptText += fmt.Sprintf("\n\nImage size: %s", options.Size)
	}

	parts := make([]*genai.Part, 0, len(options.ReferenceImages)+1)
	for _, ref := range options.ReferenceImages {
		imageBytes, mimeType, err := resolveImageBytes(ref)
		if err != nil {
			continue
		}
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				Data:     imageBytes,
				MIMEType: mimeType,
			},
		})
	}
	parts = append(parts, &genai.Part{Text: promptText})

	contents := []*genai.Content{{
		Role:  genai.RoleUser,
		Parts: parts,
	}}

	genConfig := &genai.GenerateContentConfig{
		ResponseModalities: []string{"TEXT", "IMAGE"},
	}
	if imageSize != "" || aspectRatio != "" {
		genConfig.ImageConfig = &genai.ImageConfig{
			ImageSize:   imageSize,
			AspectRatio: aspectRatio,
		}
	}

	timeout := config.DurationFromSeconds(config.GetTuning().HTTPTimeout.ImageSeconds, 10*time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := c.client.Models.GenerateContent(ctx, model, contents, genConfig)
	if err != nil {
		return nil, err
	}

	dataURI, err := extractInlineImage(resp)
	if err != nil {
		return nil, err
	}

	width, height := resolveImageDimensions(options.Size, options.Width, options.Height)

	return &ImageResult{
		Status:    "completed",
		ImageURL:  dataURI,
		Completed: true,
		Width:     width,
		Height:    height,
	}, nil
}

func (c *GeminiImageSDKClient) GetTaskStatus(taskID string) (*ImageResult, error) {
	return nil, fmt.Errorf("not supported for Gemini image generation (synchronous)")
}

func resolveImageBytes(image string) ([]byte, string, error) {
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

	if strings.HasPrefix(image, "data:") {
		return parseDataURI(image)
	}

	data, err := base64.StdEncoding.DecodeString(image)
	if err != nil {
		return nil, "", fmt.Errorf("decode base64 image: %w", err)
	}
	return data, "image/jpeg", nil
}

func parseDataURI(value string) ([]byte, string, error) {
	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid data uri")
	}

	header := parts[0]
	dataPart := parts[1]

	mimeType := "image/jpeg"
	if strings.HasPrefix(header, "data:") {
		meta := strings.TrimPrefix(header, "data:")
		if semi := strings.Index(meta, ";"); semi != -1 {
			if meta[:semi] != "" {
				mimeType = meta[:semi]
			}
		}
	}

	data, err := base64.StdEncoding.DecodeString(dataPart)
	if err != nil {
		return nil, "", fmt.Errorf("decode data uri: %w", err)
	}

	return data, mimeType, nil
}

func extractInlineImage(resp *genai.GenerateContentResponse) (string, error) {
	if resp == nil || len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	for _, candidate := range resp.Candidates {
		if candidate.Content == nil {
			continue
		}
		for _, part := range candidate.Content.Parts {
			if part == nil || part.InlineData == nil || len(part.InlineData.Data) == 0 {
				continue
			}
			mimeType := strings.TrimSpace(part.InlineData.MIMEType)
			if mimeType == "" {
				mimeType = "image/jpeg"
			}
			encoded := base64.StdEncoding.EncodeToString(part.InlineData.Data)
			return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
		}
	}

	return "", fmt.Errorf("no inline image data in response")
}

func normalizeGeminiImageSize(size string, width, height int) string {
	normalized := strings.ToUpper(strings.TrimSpace(size))
	if normalized == "1K" || normalized == "2K" || normalized == "4K" {
		return normalized
	}

	w, h := resolveImageDimensions(size, width, height)
	if w == 0 || h == 0 {
		return ""
	}

	longSide := w
	if h > w {
		longSide = h
	}
	switch {
	case longSide <= 1024:
		return "1K"
	case longSide <= 2048:
		return "2K"
	default:
		return "4K"
	}
}

func normalizeGeminiAspectRatio(size string, width, height int) string {
	w, h := resolveImageDimensions(size, width, height)
	if w == 0 || h == 0 {
		return ""
	}

	target := float64(w) / float64(h)
	allowed := map[string]float64{
		"1:1":  1.0,
		"2:3":  2.0 / 3.0,
		"3:2":  3.0 / 2.0,
		"3:4":  3.0 / 4.0,
		"4:3":  4.0 / 3.0,
		"9:16": 9.0 / 16.0,
		"16:9": 16.0 / 9.0,
		"21:9": 21.0 / 9.0,
	}
	const epsilon = 0.02
	best := ""
	bestDiff := 1.0
	for key, value := range allowed {
		diff := value - target
		if diff < 0 {
			diff = -diff
		}
		if diff < bestDiff {
			bestDiff = diff
			best = key
		}
	}
	if bestDiff <= epsilon {
		return best
	}
	return ""
}

func resolveImageDimensions(size string, width, height int) (int, int) {
	if width > 0 && height > 0 {
		return width, height
	}
	if size == "" {
		return 0, 0
	}
	return parseImageSize(size)
}
