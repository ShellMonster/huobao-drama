package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/drama-generator/backend/pkg/config"
)

// MiniMax Hailuo 支持的模型
const (
	// ModelHailuo23 全新视频生成模型，肢体动作、面部表情、物理表现与指令遵循再度突破
	// 支持：文生视频、图生视频
	// 时长：768P(6s/10s), 1080P(6s)
	ModelHailuo23 = "MiniMax-Hailuo-2.3"

	// ModelHailuo23Fast 全新图生视频模型，物理表现与指令遵循具佳，更快更优惠
	// 支持：图生视频
	// 时长：768P(6s/10s), 1080P(6s)
	ModelHailuo23Fast = "MiniMax-Hailuo-2.3-Fast"

	// ModelHailuo02 新一代视频生成模型，1080p 原生，SOTA 指令遵循，极致物理表现
	// 支持：文生视频、图生视频、首尾帧模式
	// 时长：768P(6s/10s), 1080P(6s)
	ModelHailuo02 = "MiniMax-Hailuo-02"
)

// MiniMax Hailuo 支持的分辨率
const (
	Resolution768P  = "768P"
	Resolution1080P = "1080P"
)

// MiniMax Hailuo 支持的时长（秒）
const (
	Duration6s  = 6
	Duration10s = 10
)

const minimaxDefaultBaseURL = "https://api.minimaxi.com"

// MinimaxClient Minimax视频生成客户端
type MinimaxClient struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

type MinimaxSubjectReference struct {
	Type  string   `json:"type"`
	Image []string `json:"image"`
}

type MinimaxRequest struct {
	Prompt           string                    `json:"prompt"`
	FirstFrameImage  string                    `json:"first_frame_image,omitempty"`
	LastFrameImage   string                    `json:"last_frame_image,omitempty"`
	SubjectReference []MinimaxSubjectReference `json:"subject_reference,omitempty"`
	Model            string                    `json:"model"`
	Duration         int                       `json:"duration,omitempty"`
	Resolution       string                    `json:"resolution,omitempty"`
}

// MinimaxCreateResponse 创建任务的响应
type MinimaxCreateResponse struct {
	TaskID   string `json:"task_id"`
	BaseResp struct {
		StatusCode int    `json:"status_code"`
		StatusMsg  string `json:"status_msg"`
	} `json:"base_resp"`
}

// MinimaxQueryResponse 查询任务状态的响应
type MinimaxQueryResponse struct {
	TaskID      string `json:"task_id"`
	Status      string `json:"status"` // Preparing, Queueing, Processing, Success, Fail
	FileID      string `json:"file_id"`
	VideoWidth  int    `json:"video_width"`
	VideoHeight int    `json:"video_height"`
	BaseResp    struct {
		StatusCode int    `json:"status_code"`
		StatusMsg  string `json:"status_msg"`
	} `json:"base_resp"`
}

func NewMinimaxClient(baseURL, apiKey, model string) *MinimaxClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = minimaxDefaultBaseURL
	}
	timeout := config.DurationFromSeconds(config.GetTuning().HTTPTimeout.VideoSeconds, 300*time.Second)
	return &MinimaxClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GenerateVideo 生成视频（支持首尾帧和主体参考）
// 步骤1：创建任务，返回 task_id
func (c *MinimaxClient) GenerateVideo(imageURL, prompt string, opts ...VideoOption) (*VideoResult, error) {
	options := &VideoOptions{
		Duration:   defaultVideoDuration(6),
		Resolution: defaultVideoResolution("1080P"),
	}

	for _, opt := range opts {
		opt(options)
	}

	model := c.Model
	if options.Model != "" {
		model = options.Model
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("model is required")
	}

	duration, resolution := normalizeMinimaxSettings(model, options.Duration, options.Resolution)
	reqBody := MinimaxRequest{
		Prompt:   prompt,
		Model:    model,
		Duration: duration,
	}

	// 设置分辨率
	reqBody.Resolution = resolution

	// 支持首帧图片
	if options.FirstFrameURL != "" {
		reqBody.FirstFrameImage = options.FirstFrameURL
	} else if imageURL != "" {
		reqBody.FirstFrameImage = imageURL
	}

	if strings.TrimSpace(reqBody.FirstFrameImage) == "" {
		return nil, fmt.Errorf("first_frame_image is required")
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// 步骤1：创建任务，POST 请求
	endpoint, err := joinMinimaxURL(c.BaseURL, "/v1/video_generation")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result MinimaxCreateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("minimax error: %s", result.BaseResp.StatusMsg)
	}

	// 第一步只返回 task_id，状态为 Processing
	videoResult := &VideoResult{
		TaskID:    result.TaskID,
		Status:    "Processing",
		Completed: false,
	}

	return videoResult, nil
}

// GetTaskStatus 查询任务状态
// 步骤2：查询任务状态，如果成功则进入步骤3获取文件下载地址
func (c *MinimaxClient) GetTaskStatus(taskID string) (*VideoResult, error) {
	// 步骤2：查询任务状态
	endpoint, err := joinMinimaxURL(c.BaseURL, "/v1/query/video_generation")
	if err != nil {
		return nil, err
	}
	endpoint = endpoint + "?task_id=" + url.QueryEscape(taskID)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var queryResult MinimaxQueryResponse
	if err := json.Unmarshal(body, &queryResult); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if queryResult.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("minimax error: %s", queryResult.BaseResp.StatusMsg)
	}

	videoResult := &VideoResult{
		TaskID:    queryResult.TaskID,
		Status:    queryResult.Status,
		Width:     queryResult.VideoWidth,
		Height:    queryResult.VideoHeight,
		Completed: false,
	}

	status := strings.ToLower(queryResult.Status)
	if status == "success" && queryResult.FileID != "" {
		videoResult.FileID = queryResult.FileID
		videoResult.Completed = true
	} else if status == "fail" || status == "failed" {
		videoResult.Error = "video generation failed"
		videoResult.Completed = true
	}

	return videoResult, nil
}

// RetrieveFileContent 下载文件内容（返回数据流与 Content-Type）
func (c *MinimaxClient) RetrieveFileContent(fileID string) (io.ReadCloser, string, error) {
	endpoint, err := joinMinimaxURL(c.BaseURL, "/v1/files/retrieve_content")
	if err != nil {
		return nil, "", err
	}
	endpoint = endpoint + "?file_id=" + url.QueryEscape(fileID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	contentType := resp.Header.Get("Content-Type")
	return resp.Body, contentType, nil
}

func joinMinimaxURL(baseURL, path string) (string, error) {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = minimaxDefaultBaseURL
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base url: %w", err)
	}
	base.Path = strings.TrimSuffix(base.Path, "/")
	base.Path = strings.TrimSuffix(base.Path, "/v1")

	ref, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	return base.ResolveReference(ref).String(), nil
}

func normalizeMinimaxSettings(model string, duration int, resolution string) (int, string) {
	if duration >= 8 {
		duration = Duration10s
	} else {
		duration = Duration6s
	}

	model = strings.TrimSpace(model)
	resolution = strings.TrimSpace(resolution)

	switch model {
	case ModelHailuo23, ModelHailuo23Fast:
		if duration == Duration10s {
			return duration, Resolution768P
		}
		if resolution == "" {
			resolution = Resolution768P
		}
		return duration, resolution
	case ModelHailuo02:
		if duration == Duration10s {
			if resolution == "" {
				resolution = Resolution768P
			}
			if resolution == Resolution1080P {
				resolution = Resolution768P
			}
			return duration, resolution
		}
		if resolution == "" {
			resolution = Resolution768P
		}
		return duration, resolution
	default:
		// 其他模型仅支持 720P + 6s
		return Duration6s, "720P"
	}
}
