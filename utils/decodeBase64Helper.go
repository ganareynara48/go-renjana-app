package utils

import (
	"encoding/base64"
	"errors"
	"strings"
)

func DecodeBase64Image(data string) ([]byte, string, error) {
	if !strings.HasPrefix(data, "data:image/") {
		return nil, "", errors.New("invalid image format")
	}

	parts := strings.SplitN(data, ",", 2)
	if len(parts) != 2 {
		return nil, "", errors.New("invalid base64 data")
	}

	meta := parts[0]
	raw := parts[1]

	var ext string
	switch {
	case strings.Contains(meta, "image/jpeg"):
		ext = ".jpg"
	case strings.Contains(meta, "image/png"):
		ext = ".png"
	case strings.Contains(meta, "image/webp"):
		ext = ".webp"
	default:
		return nil, "", errors.New("unsupported image type")
	}

	dataDecoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, "", errors.New("failed to decode base64 image")
	}

	return dataDecoded, ext, nil
}
