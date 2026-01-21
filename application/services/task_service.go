package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/events"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskService struct {
	db  *gorm.DB
	log *logger.Logger
	hub *events.TaskHub
}

func NewTaskService(db *gorm.DB, log *logger.Logger, hub *events.TaskHub) *TaskService {
	return &TaskService{
		db:  db,
		log: log,
		hub: hub,
	}
}

// CreateTask 创建新任务
func (s *TaskService) CreateTask(taskType, resourceID string) (*models.AsyncTask, error) {
	task := &models.AsyncTask{
		ID:         uuid.New().String(),
		Type:       taskType,
		Status:     "pending",
		Progress:   0,
		ResourceID: resourceID,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	s.publishTask(task)
	return task, nil
}

// UpdateTaskStatus 更新任务状态
func (s *TaskService) UpdateTaskStatus(taskID, status string, progress int, message string) error {
	updates := map[string]interface{}{
		"status":     status,
		"progress":   progress,
		"message":    message,
		"updated_at": time.Now(),
	}

	if status == "completed" || status == "failed" {
		now := time.Now()
		updates["completed_at"] = &now
	}

	if err := s.db.Model(&models.AsyncTask{}).
		Where("id = ?", taskID).
		Updates(updates).Error; err != nil {
		return err
	}

	s.publishTaskByID(taskID)
	return nil
}

// UpdateTaskError 更新任务错误
func (s *TaskService) UpdateTaskError(taskID string, err error) error {
	now := time.Now()
	if err := s.db.Model(&models.AsyncTask{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":       "failed",
			"error":        err.Error(),
			"progress":     0,
			"completed_at": &now,
			"updated_at":   time.Now(),
		}).Error; err != nil {
		return err
	}

	s.publishTaskByID(taskID)
	return nil
}

// UpdateTaskResult 更新任务结果
func (s *TaskService) UpdateTaskResult(taskID string, result interface{}) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	now := time.Now()
	if err := s.db.Model(&models.AsyncTask{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":       "completed",
			"progress":     100,
			"result":       string(resultJSON),
			"completed_at": &now,
			"updated_at":   time.Now(),
		}).Error; err != nil {
		return err
	}

	s.publishTaskByID(taskID)
	return nil
}

// GetTask 获取任务信息
func (s *TaskService) GetTask(taskID string) (*models.AsyncTask, error) {
	var task models.AsyncTask
	if err := s.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTasksByResource 获取资源相关的所有任务
func (s *TaskService) GetTasksByResource(resourceID string) ([]*models.AsyncTask, error) {
	var tasks []*models.AsyncTask
	if err := s.db.Where("resource_id = ?", resourceID).
		Order("created_at DESC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) publishTask(task *models.AsyncTask) {
	if s.hub == nil || task == nil {
		return
	}
	s.hub.Publish(task)
}

func (s *TaskService) publishTaskByID(taskID string) {
	if s.hub == nil {
		return
	}
	task, err := s.GetTask(taskID)
	if err != nil {
		s.log.Warnw("Failed to publish task update", "error", err, "task_id", taskID)
		return
	}
	s.hub.Publish(task)
}
