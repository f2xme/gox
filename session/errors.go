package session

import "errors"

// ErrNotFound 表示会话不存在。
var ErrNotFound = errors.New("session: not found")

// ErrExpired 表示会话已过期。
var ErrExpired = errors.New("session: expired")

// ErrInvalidID 表示会话 ID 无效。
var ErrInvalidID = errors.New("session: invalid id")

// ErrInvalidTTL 表示会话 TTL 无效。
var ErrInvalidTTL = errors.New("session: invalid ttl")

// ErrNilStore 表示存储适配器为空。
var ErrNilStore = errors.New("session: nil store")

// ErrInvalidSession 表示会话认证信息无效。
var ErrInvalidSession = errors.New("session: invalid session")
