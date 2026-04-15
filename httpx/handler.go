package httpx

// Handler is the unified request handler function.
type Handler func(ctx Context) error

// Middleware wraps a Handler to add cross-cutting behavior (onion model).
type Middleware func(next Handler) Handler

// ErrorHandler handles errors returned by Handler functions.
// It is responsible for writing the HTTP response based on the error type.
type ErrorHandler func(ctx Context, err error)
