package oss

import (
	"mime"
	"path/filepath"
	"strings"
)

// DetectContentType 根据文件扩展名检测 Content-Type
func DetectContentType(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	if ext == "" {
		return "application/octet-stream"
	}

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		return "application/octet-stream"
	}

	return contentType
}
