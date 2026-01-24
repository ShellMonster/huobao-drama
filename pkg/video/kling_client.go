package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/drama-generator/backend/pkg/config"
)

const (
	klingDefaultBaseURL = "https://api-beijing.klingai.com"
	klingOmniModelName  = "kling-video-o1"
)

// KlingClient handles Kling video generation APIs.
type KlingClient struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

type klingOmniImage struct {
	ImageURL string `json:"image_url"`
	Type     string `json:"type,omitempty"`
}

type klingOmniRequest struct {
	ModelName      string           `json:"model_name,omitempty"`
	Prompt         string           `json:"prompt"`
	ImageList      []klingOmniImage `json:"image_list,omitempty"`
	Mode           string           `json:"mode,omitempty"`
	AspectRatio    string           `json:"aspect_ratio,omitempty"`
	Duration       string           `json:"duration,omitempty"`
	CallbackURL    string           `json:"callback_url,omitempty"`
	ExternalTaskID string           `json:"external_task_id,omitempty"`
}

type klingImageRequest struct {
	ModelName      string `json:"model_name,omitempty"`
	Mode           string `json:"mode,omitempty"`
	Duration       string `json:"duration,omitempty"`
	Image          string `json:"image,omitempty"`
	ImageTail      string `json:"image_tail,omitempty"`
	Prompt         string `json:"prompt,omitempty"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	CallbackURL    string `json:"callback_url,omitempty"`
	ExternalTaskID string `json:"external_task_id,omitempty"`
}

type klingTaskResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      struct {
		TaskID        string `json:"task_id"`
		TaskStatus    string `json:"task_status"`
		TaskStatusMsg string `json:"task_status_msg"`
		TaskResult    struct {
			Videos []struct {
				ID       string `json:"id"`
				URL      string `json:"url"`
				Duration string `json:"duration"`
			} `json:"videos"`
		} `json:"task_result"`
	} `json:"data"`
}

func NewKlingClient(baseURL, apiKey, model string) *KlingClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = klingDefaultBaseURL
	}
	timeout := config.DurationFromSeconds(config.GetTuning().HTTPTimeout.VideoSeconds, 10*time.Minute)
	return &KlingClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *KlingClient) GenerateVideo(imageURL, prompt string, opts ...VideoOption) (*VideoResult, error) {
	options := &VideoOptions{
		Duration:    defaultVideoDuration(5),
		AspectRatio: defaultVideoAspectRatio("16:9"),
	}
	for _, opt := range opts {
		opt(options)
	}

	model := c.Model
	if options.Model != "" {
		model = options.Model
	}
	if model == "" {
		return nil, fmt.Errorf("model is required")
	}

	if isKlingOmniModel(model) {
		return c.generateOmniVideo(model, imageURL, prompt, options)
	}

	return c.generateImageVideo(model, imageURL, prompt, options)
}

func (c *KlingClient) GetTaskStatus(taskID string) (*VideoResult, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	if c.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	path := "/v1/videos/image2video/" + url.PathEscape(taskID)
	if isKlingOmniModel(c.Model) {
		path = "/v1/videos/omni-video/" + url.PathEscape(taskID)
	}

	respBody, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp klingTaskResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	result := &VideoResult{
		TaskID: taskID,
		Status: resp.Data.TaskStatus,
	}

	if resp.Code != 0 {
		if resp.Message != "" {
			result.Error = resp.Message
		} else {
			result.Error = fmt.Sprintf("kling error code %d", resp.Code)
		}
		return result, nil
	}

	switch resp.Data.TaskStatus {
	case "succeed":
		if len(resp.Data.TaskResult.Videos) > 0 && resp.Data.TaskResult.Videos[0].URL != "" {
			result.VideoURL = resp.Data.TaskResult.Videos[0].URL
			result.Completed = true
			if duration, err := strconv.Atoi(resp.Data.TaskResult.Videos[0].Duration); err == nil {
				result.Duration = duration
			}
			return result, nil
		}
		result.Error = "task succeeded but video url empty"
		return result, nil
	case "failed":
		if resp.Data.TaskStatusMsg != "" {
			result.Error = resp.Data.TaskStatusMsg
		} else {
			result.Error = "task failed"
		}
		return result, nil
	default:
		return result, nil
	}
}

func (c *KlingClient) generateOmniVideo(model, imageURL, prompt string, options *VideoOptions) (*VideoResult, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required for Omni-Video")
	}

	images, hasFirstFrame, err := buildOmniImages(imageURL, options)
	if err != nil {
		return nil, err
	}

	if len(images) > 7 {
		return nil, fmt.Errorf("omni-video supports up to 7 reference images")
	}

	request := klingOmniRequest{
		ModelName: model,
		Prompt:    prompt,
	}
	if len(images) > 0 {
		request.ImageList = images
	}

	if !hasFirstFrame {
		if options.AspectRatio != "" {
			request.AspectRatio = options.AspectRatio
		} else {
			request.AspectRatio = "16:9"
		}
	}

	request.Duration = normalizeOmniDuration(options.Duration, hasFirstFrame, len(images) == 0)

	respBody, err := c.doRequest(http.MethodPost, "/v1/videos/omni-video", request)
	if err != nil {
		return nil, err
	}

	var resp klingTaskResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("kling omni submit failed: %s", resp.Message)
	}
	if resp.Data.TaskID == "" {
		return nil, fmt.Errorf("kling omni submit missing task_id")
	}

	return &VideoResult{
		TaskID: resp.Data.TaskID,
		Status: resp.Data.TaskStatus,
	}, nil
}

func (c *KlingClient) generateImageVideo(model, imageURL, prompt string, options *VideoOptions) (*VideoResult, error) {
	image, imageTail, err := buildImage2VideoImages(imageURL, options)
	if err != nil {
		return nil, err
	}
	if image == "" && imageTail == "" {
		return nil, fmt.Errorf("image or image_tail is required")
	}

	request := klingImageRequest{
		ModelName: model,
		Image:     image,
		ImageTail: imageTail,
	}

	if prompt = strings.TrimSpace(prompt); prompt != "" {
		request.Prompt = prompt
	}

	request.Duration = normalizeImageDuration(options.Duration)

	respBody, err := c.doRequest(http.MethodPost, "/v1/videos/image2video", request)
	if err != nil {
		return nil, err
	}

	var resp klingTaskResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("kling image2video submit failed: %s", resp.Message)
	}
	if resp.Data.TaskID == "" {
		return nil, fmt.Errorf("kling image2video submit missing task_id")
	}

	return &VideoResult{
		TaskID: resp.Data.TaskID,
		Status: resp.Data.TaskStatus,
	}, nil
}

func (c *KlingClient) doRequest(method, path string, payload interface{}) ([]byte, error) {
	fullURL, err := joinKlingURL(c.BaseURL, path)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(c.APIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.APIKey))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("kling api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func joinKlingURL(baseURL, path string) (string, error) {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = klingDefaultBaseURL
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base url: %w", err)
	}
	ref, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	return base.ResolveReference(ref).String(), nil
}

func isKlingOmniModel(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return strings.Contains(model, klingOmniModelName)
}

func buildOmniImages(imageURL string, options *VideoOptions) ([]klingOmniImage, bool, error) {
	if options.FirstFrameURL != "" || options.LastFrameURL != "" {
		if options.FirstFrameURL == "" {
			return nil, false, fmt.Errorf("first frame is required when end frame is provided")
		}
		images := []klingOmniImage{
			{ImageURL: options.FirstFrameURL, Type: "first_frame"},
		}
		if options.LastFrameURL != "" {
			images = append(images, klingOmniImage{ImageURL: options.LastFrameURL, Type: "end_frame"})
		}
		return images, true, nil
	}

	if len(options.ReferenceImageURLs) > 0 {
		images := make([]klingOmniImage, 0, len(options.ReferenceImageURLs))
		for _, url := range options.ReferenceImageURLs {
			url = strings.TrimSpace(url)
			if url == "" {
				continue
			}
			images = append(images, klingOmniImage{ImageURL: url})
		}
		return images, false, nil
	}

	if imageURL != "" {
		return []klingOmniImage{{ImageURL: imageURL}}, false, nil
	}

	return nil, false, nil
}

func buildImage2VideoImages(imageURL string, options *VideoOptions) (string, string, error) {
	if len(options.ReferenceImageURLs) > 1 {
		return "", "", fmt.Errorf("image2video does not support multiple reference images")
	}

	image := strings.TrimSpace(imageURL)
	if options.FirstFrameURL != "" {
		image = options.FirstFrameURL
	}
	if image == "" && len(options.ReferenceImageURLs) == 1 {
		image = strings.TrimSpace(options.ReferenceImageURLs[0])
	}

	return image, strings.TrimSpace(options.LastFrameURL), nil
}

func normalizeImageDuration(duration int) string {
	if duration >= 8 {
		return "10"
	}
	return "5"
}

func normalizeOmniDuration(duration int, hasFirstFrame bool, textOnly bool) string {
	if hasFirstFrame || textOnly {
		if duration >= 8 {
			return "10"
		}
		return "5"
	}

	if duration < 3 {
		duration = 3
	}
	if duration > 10 {
		duration = 10
	}
	return strconv.Itoa(duration)
}
