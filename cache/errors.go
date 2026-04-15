package cache

import "errors"

// ErrNotFound 当键在缓存中不存在时返回。
var ErrNotFound = errors.New("cache: key not found")

// ErrLockFailed 当无法获取锁时返回。
var ErrLockFailed = errors.New("cache: failed to acquire lock")
