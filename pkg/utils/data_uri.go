package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
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

// ExtractDataURIPayload returns mime type and raw base64 payload without decoding.
func ExtractDataURIPayload(value string) (string, string, error) {
	if !strings.HasPrefix(value, "data:") {
		return "", "", fmt.Errorf("not a data uri")
	}

	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid data uri")
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

	return mimeType, payload, nil
}

// EncodeFileToDataURI reads a local file and encodes it as a data URI.
func EncodeFileToDataURI(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read header: %w", err)
	}
	mimeType := http.DetectContentType(header[:n])

	reader := io.MultiReader(bytes.NewReader(header[:n]), file)
	var encoded bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &encoded)
	if _, err := io.Copy(encoder, reader); err != nil {
		encoder.Close()
		return "", fmt.Errorf("encode base64: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return "", fmt.Errorf("finalize base64: %w", err)
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded.String()), nil
}
