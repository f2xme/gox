package internal

import (
	"net/http"

	"github.com/f2xme/gox/httpx"
)

// JSONError writes a JSON error response with the given status code and message.
func JSONError(ctx httpx.Context, code int, message string) {
	_ = ctx.JSON(code, map[string]any{
		"error":   http.StatusText(code),
		"message": message,
	})
}

// JSONErrorSimple writes a simple JSON error response.
func JSONErrorSimple(ctx httpx.Context, code int, message string) {
	_ = ctx.JSON(code, map[string]any{
		"error": message,
	})
}
