# Cache Redesign

## Goal

Redesign the `cache` package around stable capability interfaces while preserving the existing feature set: basic byte cache operations, batch operations, typed wrappers, expiration control, conditional writes, counters, scanning, memory adapter, Redis adapter, and optional locking.

Breaking changes are allowed. The redesign should favor clear naming, consistent semantics, and maintainable adapter implementations over preserving the current API shape.

## Current Problems

The current package has accumulated features without a clean capability model:

- `MultiCacheV2` exposes versioned interface naming instead of capability naming.
- Method names mix `Multi` and single-key terminology.
- TTL handling differs across operations and adapters.
- Locking, cache storage, expiration, scanning, and counters are documented as one broad surface even though they are separate capabilities.
- The memory adapter mixes storage, expiration, eviction, scanning, counters, and locking in large files.
- The memory scanner only implements a partial glob matcher.
- `Typed.GetOrSet` hides cache write failures and uses a loader without context.
- Redis lock renewal uses the caller context for renewal and unlock work, which can make behavior unclear after cancellation.

## Public API Design

The root `cache` package will define small capability interfaces. Implementations may satisfy any subset.

```go
type Store interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type BatchStore interface {
	Store
	GetMany(ctx context.Context, keys []string) (map[string][]byte, error)
	SetMany(ctx context.Context, items map[string][]byte, ttl time.Duration) error
	DeleteMany(ctx context.Context, keys []string) error
	ExistsMany(ctx context.Context, keys []string) (map[string]bool, error)
}

type Expirer interface {
	TTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Persist(ctx context.Context, key string) error
}

type ConditionalStore interface {
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
	SetXX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
	Swap(ctx context.Context, key string, value []byte, ttl time.Duration) ([]byte, error)
}
```

Additional optional capabilities remain separate:

- `Counter`: integer and float increments.
- `Scanner`: cursor-based key scanning.
- `Locker`: lock acquisition and release.
- `LockMetadata`: optional lock inspection.
- `Closer`: resource cleanup.

`Cache`, `MultiCache`, `MultiCacheV2`, `Advanced`, and `LockerV2` should be removed or replaced by compatibility aliases only if implementation impact demands a short transition. New docs and tests should use the new capability names.

## Naming

Use `Many` for batch operations:

- `GetMulti` becomes `GetMany`.
- `SetMulti` becomes `SetMany`.
- `DeleteMulti` becomes `DeleteMany`.
- `ExistsMulti` becomes `ExistsMany`.

Use `Swap` instead of `GetSet` to describe an atomic replacement that returns the old value.

Use `GetOrLoad` instead of `GetOrSet` in typed cache wrappers because the function loads a value on cache miss and then stores it.

## TTL Semantics

Define shared TTL constants:

```go
const (
	NoExpiration time.Duration = 0
	KeepTTL      time.Duration = -1
)
```

Rules:

- `Set(..., NoExpiration)` stores a value without expiration.
- `TTL` returns `ErrNotFound` when the key does not exist.
- `TTL` returns `NoExpiration, nil` when the key exists and has no expiration.
- `Expire(..., NoExpiration)` removes expiration.
- `Persist` removes expiration and returns `ErrNotFound` when the key does not exist.
- `SetXX` and `Swap` support `KeepTTL`.
- `SetNX` treats `KeepTTL` as invalid because there is no existing TTL to preserve.

## Errors

Keep a small error surface in the root package:

- `ErrNotFound`: key or lock metadata does not exist.
- `ErrLocked`: a lock cannot be acquired immediately.
- `ErrInvalidTTL`: an operation received an unsupported TTL value.

Adapter errors should be wrapped with `%w` when adding context so callers can use `errors.Is`.

## Memory Adapter

The memory adapter should be reorganized by capability:

- `store.go`: `Store`, `BatchStore`, copy-on-read/write behavior.
- `expiration.go`: TTL, expire, persist, cleanup loop.
- `eviction.go`: LRU/LFU policies.
- `counter.go`: integer and float increments.
- `scan.go`: scanner using `path.Match` for standard glob matching.
- `lock.go`: process-local locking.
- `options.go`: configuration.

The adapter should continue copying returned and stored byte slices to avoid external mutation of internal state.

Expired keys should be removed lazily on access and eagerly by cleanup. Batch and scan operations should ignore expired keys and clean them when practical.

## Redis Adapter

The Redis adapter should implement the same capabilities with Redis-native operations where possible:

- Basic operations use `GET`, `SET`, `DEL`, and `EXISTS`.
- Batch operations use pipeline or native multi-key commands where semantics match.
- TTL operations map Redis `TTL` results to shared package semantics.
- `Swap` should remain atomic with Lua when preserving or changing TTL is required.
- Locks use unique tokens and Lua-based unlock.
- Lock renewal should use an internal renewal context and stop channel rather than relying only on the caller context after acquisition.

Redis-specific client configuration remains in `adapter/redis`.

## Typed Wrapper

`Typed[T]` remains in the root package and wraps a `Store`.

Methods:

- `Get`
- `Set`
- `Delete`
- `Exists`
- `GetOrLoad`
- `GetMany`
- `SetMany`
- `DeleteMany`

`GetOrLoad` signature:

```go
func (t *Typed[T]) GetOrLoad(
	ctx context.Context,
	key string,
	ttl time.Duration,
	load func(context.Context) (T, error),
) (T, error)
```

The default behavior should return cache write errors after a successful load. An option may allow returning the loaded value while ignoring write errors for applications that prefer availability over cache correctness.

Serializers remain pluggable. JSON remains the default serializer.

## Tests

Add capability-focused contract tests where possible:

- Store contract.
- BatchStore contract.
- Expirer contract.
- ConditionalStore contract.
- Counter contract.
- Scanner contract.
- Locker contract.

Memory adapter tests should cover expiration cleanup, lazy expiration, copy safety, eviction behavior, glob scanning, counters, and locks.

Redis tests should keep integration behavior explicit and skip when Redis is unavailable.

Typed wrapper tests should cover serialization failures, `ErrNotFound` behavior, `GetOrLoad` singleflight behavior, write-error handling policy, and batch fallback.

## Migration Notes

Users should update imports and method calls:

- `cache.Cache` to `cache.Store`.
- `cache.MultiCacheV2` to `cache.BatchStore`.
- `GetMulti` to `GetMany`.
- `SetMulti` to `SetMany`.
- `DeleteMulti` to `DeleteMany`.
- `ExistsMulti` to `ExistsMany`.
- `GetSet` to `Swap`.
- `GetOrSet` to `GetOrLoad`.
- `ErrLockFailed` to `ErrLocked`.

The redesign should update docs, examples, and tests in the same change so the package presents one coherent API.
