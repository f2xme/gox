package logx

import (
	"fmt"
	"strings"
	"sync"
)

var callerPrefixes = []string{"app/", "library/", "internal/", "cmd/"}

var callerCache sync.Map

func trimCallerPath(path string) string {
	if cached, ok := callerCache.Load(path); ok {
		return cached.(string)
	}

	trimmed := path
	for _, prefix := range callerPrefixes {
		if idx := strings.Index(path, prefix); idx != -1 {
			trimmed = path[idx:]
			break
		}
	}

	callerCache.Store(path, trimmed)
	return trimmed
}

func formatCaller(file string, line int) string {
	return fmt.Sprintf("%s:%d", trimCallerPath(file), line)
}
