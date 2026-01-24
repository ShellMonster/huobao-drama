package services

import (
	"strings"

	"github.com/drama-generator/backend/pkg/utils"
)

// parseFramePromptJSON 解析AI返回的JSON格式提示词
func (s *FramePromptService) parseFramePromptJSON(aiResponse string) *SingleFramePrompt {
	var result SingleFramePrompt
	if err := utils.SafeParseAIJSON(aiResponse, &result); err != nil {
		s.log.Warnw("Failed to parse JSON", "error", err)
		return nil
	}

	// 验证必需字段
	if strings.TrimSpace(result.Prompt) == "" {
		s.log.Warnw("Parsed JSON missing prompt field")
		return nil
	}

	return &result
}
