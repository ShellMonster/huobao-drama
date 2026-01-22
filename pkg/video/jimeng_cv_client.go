package video

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/drama-generator/backend/pkg/utils"
	"github.com/volcengine/volc-sdk-golang/service/visual"
)

const (
	jimengDefaultRegion  = "cn-north-1"
	jimengDefaultBaseURL = "https://visual.volcengineapi.com"
	jimengDefaultScheme  = "https"
)

// JimengCVClient handles Jimeng Visual API (CVSync2AsyncSubmitTask/GetResult).
type JimengCVClient struct {
	BaseURL      string
	AccessKey    string
	SecretKey    string
	SessionToken string
	Model        string
	Region       string
	Timeout      time.Duration
}

type jimengSubmitRequest struct {
	ReqKey           string   `json:"req_key"`
	Prompt           string   `json:"prompt,omitempty"`
	BinaryDataBase64 []string `json:"binary_data_base64,omitempty"`
	ImageURLs        []string `json:"image_urls,omitempty"`
	Seed             *int64   `json:"seed,omitempty"`
	Frames           *int     `json:"frames,omitempty"`
	AspectRatio      string   `json:"aspect_ratio,omitempty"`
}

type jimengSubmitResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      struct {
		TaskID string `json:"task_id"`
	} `json:"data"`
}

type jimengResultRequest struct {
	ReqKey  string `json:"req_key"`
	TaskID  string `json:"task_id"`
	ReqJSON string `json:"req_json,omitempty"`
}

type jimengResultResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      struct {
		Status    string `json:"status"`
		VideoURL  string `json:"video_url"`
		Tagged    bool   `json:"aigc_meta_tagged"`
	} `json:"data"`
}

func NewJimengCVClient(baseURL, accessKey, secretKey, model, sessionToken string) *JimengCVClient {
	if baseURL == "" {
		baseURL = jimengDefaultBaseURL
	}
	return &JimengCVClient{
		BaseURL:      baseURL,
		AccessKey:    accessKey,
		SecretKey:    secretKey,
		SessionToken: sessionToken,
		Model:        model,
		Region:       jimengDefaultRegion,
		Timeout:      10 * time.Minute,
	}
}

func (c *JimengCVClient) GenerateVideo(imageURL, prompt string, opts ...VideoOption) (*VideoResult, error) {
	options := &VideoOptions{
		Duration:    5,
		AspectRatio: "16:9",
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

	reqKey := model

	images, err := selectJimengImages(model, imageURL, options)
	if err != nil {
		return nil, err
	}

	frames := mapDurationToFrames(options.Duration)

	useBinary := false
	for _, image := range images {
		if utils.IsDataURI(image) {
			useBinary = true
			break
		}
	}

	var binaryImages []string
	if useBinary {
		for _, image := range images {
			if !utils.IsDataURI(image) {
				return nil, fmt.Errorf("jimeng requires all images to be data uri when using base64 input")
			}
			_, payload, err := utils.ExtractDataURIPayload(image)
			if err != nil {
				return nil, fmt.Errorf("parse data uri: %w", err)
			}
			binaryImages = append(binaryImages, payload)
		}
	}

	requestBody := jimengSubmitRequest{
		ReqKey:      reqKey,
		Prompt:      strings.TrimSpace(prompt),
		Frames:      &frames,
		AspectRatio: "",
	}
	if useBinary {
		requestBody.BinaryDataBase64 = binaryImages
	} else {
		requestBody.ImageURLs = images
	}

	if options.Seed != 0 {
		seed := options.Seed
		requestBody.Seed = &seed
	}

	// Only include aspect_ratio for text-to-video.
	if len(images) == 0 && options.AspectRatio != "" {
		requestBody.AspectRatio = options.AspectRatio
	}

	client, err := c.newVisualClient()
	if err != nil {
		return nil, err
	}

	respMap, status, err := client.CVSync2AsyncSubmitTask(requestBody)
	if err != nil {
		return nil, fmt.Errorf("jimeng sdk submit failed (status %d): %w", status, err)
	}

	var resp jimengSubmitResponse
	if err := decodeJimengResponse(respMap, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if resp.Code != 10000 {
		message := resp.Message
		if message == "" {
			message = "unknown error"
		}
		return nil, fmt.Errorf("jimeng submit failed: %s", message)
	}
	if resp.Data.TaskID == "" {
		return nil, fmt.Errorf("jimeng submit missing task_id")
	}

	return &VideoResult{
		TaskID: resp.Data.TaskID,
		Status: "submitted",
	}, nil
}

func (c *JimengCVClient) GetTaskStatus(taskID string) (*VideoResult, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}

	model := c.Model
	if model == "" {
		return nil, fmt.Errorf("model is required")
	}

	requestBody := jimengResultRequest{
		ReqKey: model,
		TaskID: taskID,
	}

	client, err := c.newVisualClient()
	if err != nil {
		return nil, err
	}

	respMap, status, err := client.CVSync2AsyncGetResult(requestBody)
	if err != nil {
		return nil, fmt.Errorf("jimeng sdk query failed (status %d): %w", status, err)
	}

	var resp jimengResultResponse
	if err := decodeJimengResponse(respMap, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	result := &VideoResult{
		TaskID: taskID,
		Status: resp.Data.Status,
	}

	if resp.Code != 10000 {
		if resp.Message != "" {
			result.Error = resp.Message
		} else {
			result.Error = fmt.Sprintf("jimeng error code %d", resp.Code)
		}
		return result, nil
	}

	switch resp.Data.Status {
	case "done":
		if resp.Data.VideoURL != "" {
			result.VideoURL = resp.Data.VideoURL
			result.Completed = true
			return result, nil
		}
		result.Error = "task completed but video_url empty"
		return result, nil
	case "not_found", "expired":
		result.Error = fmt.Sprintf("task %s", resp.Data.Status)
		return result, nil
	default:
		// in_queue / generating
		return result, nil
	}
}

func (c *JimengCVClient) newVisualClient() (*visual.Visual, error) {
	if strings.TrimSpace(c.AccessKey) == "" || strings.TrimSpace(c.SecretKey) == "" {
		return nil, fmt.Errorf("missing access key or secret key")
	}

	client := visual.NewInstance()
	client.Client.SetAccessKey(c.AccessKey)
	client.Client.SetSecretKey(c.SecretKey)
	if c.SessionToken != "" {
		client.Client.SetSessionToken(c.SessionToken)
	}
	if c.Region != "" {
		client.SetRegion(c.Region)
	}

	scheme, host, err := splitJimengBaseURL(c.BaseURL)
	if err != nil {
		return nil, err
	}
	client.SetHost(host)
	client.SetSchema(scheme)

	if c.Timeout > 0 {
		client.Client.SetCustomTimeout(c.Timeout)
	}

	return client, nil
}

func splitJimengBaseURL(baseURL string) (string, string, error) {
	if baseURL == "" {
		baseURL = jimengDefaultBaseURL
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid base url: %w", err)
	}

	if parsed.Scheme == "" && parsed.Host == "" {
		parsed, err = url.Parse(jimengDefaultScheme + "://" + baseURL)
		if err != nil {
			return "", "", fmt.Errorf("invalid base url: %w", err)
		}
	}

	scheme := parsed.Scheme
	host := parsed.Host
	if host == "" {
		host = parsed.Path
	}
	if scheme == "" {
		scheme = jimengDefaultScheme
	}
	if host == "" {
		return "", "", fmt.Errorf("invalid base url: host missing")
	}
	return scheme, host, nil
}

func decodeJimengResponse(input map[string]interface{}, target interface{}) error {
	raw, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

func mapDurationToFrames(duration int) int {
	switch {
	case duration >= 10:
		return 241
	case duration <= 0:
		return 121
	case duration <= 5:
		return 121
	default:
		return 241
	}
}

func selectJimengImages(model, imageURL string, options *VideoOptions) ([]string, error) {
	if len(options.ReferenceImageURLs) > 0 {
		return nil, fmt.Errorf("jimeng does not support multiple reference images")
	}

	modelLower := strings.ToLower(model)
	isFirstTail := strings.Contains(modelLower, "first_tail")
	isFirstOnly := strings.Contains(modelLower, "_first_") && !isFirstTail
	isTi2v := strings.Contains(modelLower, "ti2v") || strings.Contains(modelLower, "t2v")

	var images []string
	if options.FirstFrameURL != "" || options.LastFrameURL != "" {
		if isFirstTail {
			if options.FirstFrameURL == "" || options.LastFrameURL == "" {
				return nil, fmt.Errorf("jimeng first_tail requires both first and last frame images")
			}
			images = []string{options.FirstFrameURL, options.LastFrameURL}
		} else {
			url := options.FirstFrameURL
			if url == "" {
				url = imageURL
			}
			if url != "" {
				images = []string{url}
			}
		}
	} else if imageURL != "" {
		images = []string{imageURL}
	}

	if isFirstTail {
		if len(images) != 2 {
			return nil, fmt.Errorf("jimeng first_tail requires two images")
		}
		return images, nil
	}

	if isFirstOnly {
		if len(images) != 1 {
			return nil, fmt.Errorf("jimeng first frame requires one image")
		}
		return images, nil
	}

	if !isTi2v && len(images) > 0 {
		// Unknown model type; allow single image but warn by validation.
		return images, nil
	}

	if isTi2v {
		if len(images) > 1 {
			return nil, fmt.Errorf("jimeng ti2v supports at most one image")
		}
		// ti2v allows text-only or single image.
		return images, nil
	}

	return images, nil
}
