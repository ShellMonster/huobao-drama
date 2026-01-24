package storage

import (
	"io"

	"github.com/drama-generator/backend/pkg/config"
)

// Storage 抽象存储接口，支持本地或对象存储切换。
type Storage interface {
	Upload(file io.Reader, filename string, category string) (string, error)
	UploadBytes(data []byte, contentType string, category string) (string, error)
	Delete(url string) error
	GetURL(path string) string
	BaseURL() string
	IsLocalURL(url string) bool
	ResolvePath(rawURL string) (string, error)
	ToDataURL(rawURL string) (string, error)
	DownloadFromURL(url string, category string) (string, error)
}

func NewStorage(cfg config.StorageConfig) (Storage, error) {
	switch cfg.Type {
	case "", "local":
		return NewLocalStorage(cfg.LocalPath, cfg.BaseURL)
	default:
		// 兜底使用本地存储，避免因未实现对象存储而阻塞启动
		return NewLocalStorage(cfg.LocalPath, cfg.BaseURL)
	}
}

var _ Storage = (*LocalStorage)(nil)
