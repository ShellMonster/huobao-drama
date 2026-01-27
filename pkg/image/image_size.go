package image

import (
	"strconv"
	"strings"
)

func parseImageSize(size string) (int, int) {
	normalized := strings.TrimSpace(strings.ToLower(size))
	normalized = strings.ReplaceAll(normalized, "×", "x")
	normalized = strings.ReplaceAll(normalized, "*", "x")
	normalized = strings.ReplaceAll(normalized, " ", "")

	switch strings.ToUpper(normalized) {
	case "1K":
		return 1024, 1024
	case "2K":
		return 2048, 2048
	case "4K":
		return 4096, 4096
	}

	parts := strings.Split(normalized, "x")
	if len(parts) != 2 {
		return 0, 0
	}

	width, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0
	}
	height, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0
	}

	return width, height
}
