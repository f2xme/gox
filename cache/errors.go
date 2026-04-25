package cache

import "errors"

// ErrNotFound 当键在缓存中不存在时返回。
var ErrNotFound = errors.New("cache: key not found")

// ErrLocked 当无法立即获取锁时返回。
var ErrLocked = errors.New("cache: lock already held")

// ErrInvalidTTL 当操作收到不支持的 TTL 值时返回。
var ErrInvalidTTL = errors.New("cache: invalid ttl")
