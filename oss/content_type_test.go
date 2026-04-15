package oss

import "testing"

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "jpeg image",
			key:      "photo.jpg",
			expected: "image/jpeg",
		},
		{
			name:     "png image",
			key:      "image.png",
			expected: "image/png",
		},
		{
			name:     "pdf document",
			key:      "document.pdf",
			expected: "application/pdf",
		},
		{
			name:     "text file",
			key:      "readme.txt",
			expected: "text/plain; charset=utf-8",
		},
		{
			name:     "json file",
			key:      "data.json",
			expected: "application/json",
		},
		{
			name:     "no extension",
			key:      "file",
			expected: "application/octet-stream",
		},
		{
			name:     "unknown extension",
			key:      "file.unknown",
			expected: "application/octet-stream",
		},
		{
			name:     "uppercase extension",
			key:      "PHOTO.JPG",
			expected: "image/jpeg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectContentType(tt.key)
			if result != tt.expected {
				t.Errorf("DetectContentType(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}
