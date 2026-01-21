package utils

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// IsDataURI reports whether the value is a data URI.
func IsDataURI(value string) bool {
	return strings.HasPrefix(value, "data:")
}

// ParseDataURI parses a data URI and returns decoded bytes and mime type.
// Expected format: data:<mime>;base64,<payload>
func ParseDataURI(value string) ([]byte, string, error) {
	if !strings.HasPrefix(value, "data:") {
		return nil, "", fmt.Errorf("not a data uri")
	}

	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid data uri")
	}

	header := parts[0]
	payload := parts[1]

	mimeType := "application/octet-stream"
	meta := strings.TrimPrefix(header, "data:")
	if semi := strings.Index(meta, ";"); semi != -1 {
		if meta[:semi] != "" {
			mimeType = meta[:semi]
		}
	} else if meta != "" {
		mimeType = meta
	}

	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, "", fmt.Errorf("decode data uri: %w", err)
	}

	return data, mimeType, nil
}
