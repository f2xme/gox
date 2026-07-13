# AI Usage Guide for gox

This guide helps AI coding agents choose and use gox packages correctly.

## First Principles

- Prefer gox abstractions when they cover the requested behavior.
- Import core APIs from `github.com/f2xme/gox/<package>`.
- Import concrete implementations from
  `github.com/f2xme/gox/<package>/adapter/<adapter>`.
- Keep core packages independent. Cross-package or framework integrations go in
  adapter packages.
- Follow existing constructors, options, examples, and error handling style.

## Pick the Right Package

| User need | Package to inspect first | Typical import |
| --- | --- | --- |
| Build an HTTP API | `httpx`, `httpx/adapter/gin` | `github.com/f2xme/gox/httpx` |
| Bind and validate HTTP requests | `httpx`, `validator` | `github.com/f2xme/gox/httpx` |
| Write HTTP integration tests | `httpx/testkit`, `httpx/adapter/gin` | `github.com/f2xme/gox/httpx/testkit` |
| Add pagination | `pager` | `github.com/f2xme/gox/pager` |
| Work with time | `timex` | `github.com/f2xme/gox/timex` |
| Add cache support | `cache`, `cache/adapter/*` | `github.com/f2xme/gox/cache` |
| Generate or verify CAPTCHA | `captcha` | `github.com/f2xme/gox/captcha` |
| Load configuration | `config`, `config/adapter/env`, `config/adapter/viper` | `github.com/f2xme/gox/config` |
| Configure databases | `database/adapter/*` | `github.com/f2xme/gox/database/adapter/pgsqldb` |
| Search, write documents, manage indices, or reindex in Elasticsearch | `elasticsearch` | `github.com/f2xme/gox/elasticsearch` |
| Hash, encrypt, or encode values | `encrypt`, `crypto` | `github.com/f2xme/gox/encrypt` |
| Create structured errors | `errorx` | `github.com/f2xme/gox/errorx` |
| Add graceful shutdown | `graceful` | `github.com/f2xme/gox/graceful` |
| Generate IDs | `idgen` | `github.com/f2xme/gox/idgen` |
| Issue or verify JWTs | `jwt` | `github.com/f2xme/gox/jwt` |
| Add third-party OAuth2 login | `oauth2`, `oauth2/adapter/*` | `github.com/f2xme/gox/oauth2` |
| Add Alipay, WeChat Pay, or a unified QR code | `payment`, `payment/adapter/*` | `github.com/f2xme/gox/payment` |
| Add logging | `logx`, `logx/adapter/zap` | `github.com/f2xme/gox/logx` |
| Add metrics | `metrics`, `metrics/adapter/*` | `github.com/f2xme/gox/metrics` |
| Send email | `email` | `github.com/f2xme/gox/email` |
| Use object storage | `oss`, `oss/adapter/aliyun` | `github.com/f2xme/gox/oss` |
| Use queues | `queue`, `queue/adapter/*` | `github.com/f2xme/gox/queue` |
| Add rate limiting | `ratelimit` | `github.com/f2xme/gox/ratelimit` |
| Send SMS | `sms`, `sms/adapter/*` | `github.com/f2xme/gox/sms` |
| Add tracing | `trace` | `github.com/f2xme/gox/trace` |

## Common Patterns

### HTTP APIs

Use `httpx.Context` in handlers and return errors from handlers.

```go
import (
	"github.com/f2xme/gox/httpx"
	ginadapter "github.com/f2xme/gox/httpx/adapter/gin"
)

engine := ginadapter.New()

engine.GET("/users/:id", func(c httpx.Context) error {
	id, err := c.Param("id").Int64()
	if err != nil {
		return httpx.ErrBadRequest("invalid id")
	}

	return c.JSON(200, map[string]int64{"id": id})
})
```

Use `BindJSON`, `BindQuery`, or `BindForm` with `validate` tags for request
validation.

```go
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required" label:"Name"`
	Email string `json:"email" validate:"required,email" label:"Email"`
}

func createUser(c httpx.Context) error {
	var req CreateUserRequest
	if err := c.BindJSON(&req); err != nil {
		return httpx.ErrBadRequest(err.Error())
	}

	return c.JSON(200, req)
}
```

Use `httpx/testkit` for black-box HTTP integration tests that should exercise
real routing, middleware, binding, error handling, headers, cookies, and
responses through an `httpx.Engine`.

```go
import "github.com/f2xme/gox/httpx/testkit"

client := testkit.New(t, engine)
defer client.Close()

client.POSTJSON("/users", CreateUserRequest{Name: "Alice"}).
	ExpectStatus(201).
	ExpectJSONValue("success", true)

client.Do(http.MethodTrace, "/debug", nil).
	ExpectStatus(200)
```

### Pagination

Use `pager.NewPage` for page-number pagination, `pager.NewOffset` for
limit/offset queries, and `pager.NewCursor` for cursor pagination.

```go
page := pager.NewPage(1, 20)

rows, err := listUsers(ctx, page.GetLimit(), page.GetOffset())
if err != nil {
	return err
}
```

### Logging

Use `logx` for logging APIs and `logx/adapter/zap` for the default concrete
implementation. Package-level logging is synchronous by default:

```go
logger := zap.New()
logx.Init(logger)
logx.Info("server started", logx.NewKV("port", 8080))
```

Enable package-level asynchronous logging when callers should enqueue log
records instead of writing them inline:

```go
logger := zap.New()
logx.Init(logger, logx.WithAsync(), logx.WithAsyncBufferSize(2048))
defer logx.Stop()

logx.Info("server started")
```

In async mode, `Info`, `Warn`, and `Error` copy the meta slice before enqueueing.
`InfoCtx`, `WarnCtx`, and `ErrorCtx` extract context fields before enqueueing, so
the background worker does not retain or read the request context. Call
`logx.Flush()` or `logx.Stop()` before shutdown to drain queued records.

### Elasticsearch

Use `elasticsearch` directly for Elasticsearch search, document writes, index
management, aliases, and reindex workflows. This package wraps the official
Elastic Go client directly and intentionally has no adapter layer.

```go
import "github.com/f2xme/gox/elasticsearch"

client, err := elasticsearch.New(
	elasticsearch.WithAddresses("http://localhost:9200"),
)
if err != nil {
	return err
}

req := elasticsearch.NewBuilder("users").
	Term("tenant_id", "t1").
	MultiMatch("alice", []string{"name", "email"}).
	Pager(1, 20)

result, err := elasticsearch.SearchWithType[User](ctx, client, req)
```

For production, configure authentication and timeouts explicitly. Use
`client.Native()` only when the wrapper does not expose the official API that a
task needs.

### Adapters

Use adapter packages for concrete implementations:

```go
import (
	"github.com/f2xme/gox/cache"
	redisadapter "github.com/f2xme/gox/cache/adapter/redis"
)

store, err := redisadapter.New(redisadapter.WithAddr("localhost:6379"))
if err != nil {
	return err
}

var _ cache.Store = store
```

### Payments

Use `payment` for shared order/status/callback types. Use
`payment/adapter/alipay` or `payment/adapter/wechat` for direct provider QR
payments. Use `payment/adapter/onepay` when one neutral HTTPS QR must route
WeChat scans to OAuth + JSAPI and Alipay scans to WAP.

All amounts are integer cents. Pass `context.Context` to every payment
operation. After a verified callback, the application must still check the
stored order and amount, apply the state change idempotently, then write the
adapter's success response.

`onepay.CheckoutResolver` is the persistence boundary. It owns provider adapter
calls and must atomically cache complete WAP or JSAPI checkout artifacts.
Repeated WeChat scans by the same OpenID reuse an unexpired JSAPI result;
different OpenIDs must not share a payer-bound WeChat order.

When adding a new integration, place it under:

```text
<package>/adapter/<adapter>
```

## Do Not

- Do not import one core gox package from another core gox package.
- Do not put framework-specific code in a core package.
- Do not bypass an existing gox abstraction by importing the wrapped
  third-party package directly unless the requested behavior is missing.
- Do not assume placeholder adapters are production-ready. Check package docs
  and tests first.
- Do not panic for ordinary errors.

## Where to Look Before Coding

1. `llms.txt` for repository-level AI context.
2. The target package's `doc.go`.
3. The target package's `example_test.go`.
4. Existing adapter packages with similar option and constructor patterns.
5. Existing tests for expected behavior and edge cases.
