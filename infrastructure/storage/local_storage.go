package storage

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/drama-generator/backend/pkg/utils"
)

type LocalStorage struct {
	basePath string
	baseURL  string
}

func NewLocalStorage(basePath, baseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}, nil
}

func (s *LocalStorage) Upload(file io.Reader, filename string, category string) (string, error) {
	dir := filepath.Join(s.basePath, category)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create category directory: %w", err)
	}

	filePath := filepath.Join(dir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", s.baseURL, category, filename)
	return url, nil
}

func (s *LocalStorage) UploadBytes(data []byte, contentType string, category string) (string, error) {
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	ext := getFileExtension("", contentType)
	token, err := utils.NewRandomID()
	if err != nil {
		return "", fmt.Errorf("failed to generate file name: %w", err)
	}
	filename := fmt.Sprintf("%s%s", token, ext)
	return s.saveReader(bytes.NewReader(data), category, filename)
}

func (s *LocalStorage) Delete(url string) error {
	return nil
}

func (s *LocalStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, path)
}

func (s *LocalStorage) BaseURL() string {
	return s.baseURL
}

func (s *LocalStorage) IsLocalURL(url string) bool {
	if s.baseURL == "" {
		return false
	}
	return strings.HasPrefix(url, s.baseURL+"/")
}

// ResolvePath converts a local URL back to the filesystem path.
func (s *LocalStorage) ResolvePath(rawURL string) (string, error) {
	if !s.IsLocalURL(rawURL) {
		return "", fmt.Errorf("not a local url")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	baseURL, err := url.Parse(s.baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base url: %w", err)
	}

	rel := strings.TrimPrefix(parsedURL.Path, baseURL.Path)
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" {
		return "", fmt.Errorf("invalid local url path")
	}

	return filepath.Join(s.basePath, filepath.FromSlash(rel)), nil
}

// ToDataURL reads a local URL and converts it to a data URI.
func (s *LocalStorage) ToDataURL(rawURL string) (string, error) {
	path, err := s.ResolvePath(rawURL)
	if err != nil {
		return "", err
	}
	return utils.EncodeFileToDataURI(path)
}

// DownloadFromURL 从远程URL下载文件到本地存储
func (s *LocalStorage) DownloadFromURL(url, category string) (string, error) {
	// 发送HTTP请求下载文件
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}

	// 从URL或Content-Type推断文件扩展名
	ext := getFileExtension(url, resp.Header.Get("Content-Type"))

	// 创建目录
	dir := filepath.Join(s.basePath, category)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create category directory: %w", err)
	}

	// 生成唯一文件名
	token, err := utils.NewRandomID()
	if err != nil {
		return "", fmt.Errorf("failed to generate file name: %w", err)
	}
	filename := fmt.Sprintf("%s%s", token, ext)
	filePath := filepath.Join(dir, filename)

	// 保存文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// 返回本地URL
	localURL := fmt.Sprintf("%s/%s/%s", s.baseURL, category, filename)
	return localURL, nil
}

func (s *LocalStorage) saveReader(reader io.Reader, category string, filename string) (string, error) {
	dir := filepath.Join(s.basePath, category)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create category directory: %w", err)
	}
	filePath := filepath.Join(dir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}
	return fmt.Sprintf("%s/%s/%s", s.baseURL, category, filename), nil
}

// getFileExtension 从URL或Content-Type推断文件扩展名
func getFileExtension(url, contentType string) string {
	// 首先尝试从URL获取扩展名
	if idx := strings.LastIndex(url, "."); idx != -1 {
		ext := url[idx:]
		// 只取扩展名部分，忽略查询参数
		if qIdx := strings.Index(ext, "?"); qIdx != -1 {
			ext = ext[:qIdx]
		}
		if len(ext) <= 5 { // 合理的扩展名长度
			return ext
		}
	}

	// 根据Content-Type推断扩展名
	switch {
	case strings.Contains(contentType, "image/jpeg"):
		return ".jpg"
	case strings.Contains(contentType, "image/png"):
		return ".png"
	case strings.Contains(contentType, "image/gif"):
		return ".gif"
	case strings.Contains(contentType, "image/webp"):
		return ".webp"
	case strings.Contains(contentType, "video/mp4"):
		return ".mp4"
	case strings.Contains(contentType, "video/webm"):
		return ".webm"
	case strings.Contains(contentType, "video/quicktime"):
		return ".mov"
	default:
		return ".bin"
	}
}
