package httpadapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"go.yaml.in/yaml/v3"
)

func parseConfig(body []byte, format Format, sourceURL string, header http.Header) (map[string]any, error) {
	format = resolveFormat(format, sourceURL, header)

	var values map[string]any
	switch format {
	case JSON:
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.UseNumber()
		if err := decoder.Decode(&values); err != nil {
			return nil, fmt.Errorf("httpadapter: parse json: %w", err)
		}
		var trailing any
		if err := decoder.Decode(&trailing); err != io.EOF {
			if err == nil {
				return nil, fmt.Errorf("httpadapter: parse json: trailing data")
			}
			return nil, fmt.Errorf("httpadapter: parse json trailing data: %w", err)
		}
	case YAML:
		if err := yaml.Unmarshal(body, &values); err != nil {
			return nil, fmt.Errorf("httpadapter: parse yaml: %w", err)
		}
	default:
		return nil, fmt.Errorf("httpadapter: unsupported config format %q", format)
	}

	if values == nil {
		values = make(map[string]any)
	}
	return normalizeMap(values), nil
}

func resolveFormat(format Format, sourceURL string, header http.Header) Format {
	if format != Auto {
		return format
	}

	contentType := strings.ToLower(header.Get("Content-Type"))
	switch {
	case strings.Contains(contentType, "json"):
		return JSON
	case strings.Contains(contentType, "yaml"), strings.Contains(contentType, "x-yaml"):
		return YAML
	}

	if u, err := url.Parse(sourceURL); err == nil {
		switch strings.ToLower(path.Ext(u.Path)) {
		case ".json":
			return JSON
		case ".yaml", ".yml":
			return YAML
		}
	}

	return YAML
}

func normalizeMap(values map[string]any) map[string]any {
	result := make(map[string]any, len(values))
	for k, v := range values {
		result[k] = normalizeValue(v)
	}
	return result
}

func normalizeValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return normalizeMap(v)
	case map[any]any:
		m := make(map[string]any, len(v))
		for key, val := range v {
			m[fmt.Sprint(key)] = normalizeValue(val)
		}
		return m
	case []any:
		items := make([]any, len(v))
		for i, item := range v {
			items[i] = normalizeValue(item)
		}
		return items
	default:
		return value
	}
}
