package models

import (
	"time"

	"gorm.io/gorm"
)

type FramePromptTaskStatus string

const (
	FramePromptTaskPending    FramePromptTaskStatus = "pending"
	FramePromptTaskProcessing FramePromptTaskStatus = "processing"
	FramePromptTaskCompleted  FramePromptTaskStatus = "completed"
	FramePromptTaskFailed     FramePromptTaskStatus = "failed"
)

type FramePromptTask struct {
	ID           uint                  `gorm:"primaryKey;autoIncrement" json:"id"`
	StoryboardID uint                  `gorm:"not null;index" json:"storyboard_id"`
	FrameType    string                `gorm:"size:20;not null;index" json:"frame_type"`
	Status       FramePromptTaskStatus `gorm:"size:20;not null;index" json:"status"`
	ErrorMsg     *string               `gorm:"type:text" json:"error_msg,omitempty"`
	CreatedAt    time.Time             `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time             `gorm:"not null;autoUpdateTime" json:"updated_at"`
	CompletedAt  *time.Time            `json:"completed_at,omitempty"`
	DeletedAt    gorm.DeletedAt         `gorm:"index" json:"-"`
}

func (t *FramePromptTask) TableName() string {
	return "frame_prompt_tasks"
}
