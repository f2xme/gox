package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"
)

// Session 表示一个会话。
type Session struct {
	// ID 是会话唯一标识。
	ID string `json:"id"`
	// Values 存储会话业务数据。
	Values map[string]any `json:"values"`
	// CreatedAt 是会话创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 是会话最后更新时间。
	UpdatedAt time.Time `json:"updated_at"`
	// ExpiresAt 是会话过期时间。
	ExpiresAt time.Time `json:"expires_at"`
}

// Store 定义会话存储接口。
type Store interface {
	// Get 根据 ID 获取会话。
	// 如果会话不存在返回 ErrNotFound。
	Get(ctx context.Context, id string) (*Session, error)

	// Set 保存会话并设置存储层 TTL。
	// 存储实现必须保留 sess.ExpiresAt，ttl 只用于底层过期控制。
	Set(ctx context.Context, sess *Session, ttl time.Duration) error

	// Delete 删除会话。
	// 如果会话不存在不返回错误。
	Delete(ctx context.Context, id string) error
}

// Closer 定义适配器资源释放能力。
type Closer interface {
	// Close 释放适配器资源。
	Close() error
}

// Manager 定义会话管理操作。
type Manager interface {
	// Create 创建并保存新会话。
	Create(ctx context.Context) (*Session, error)

	// Get 根据 ID 获取未过期会话。
	Get(ctx context.Context, id string) (*Session, error)

	// Save 保存会话数据并保留当前过期时间。
	Save(ctx context.Context, sess *Session) error

	// Refresh 延长会话有效期。
	Refresh(ctx context.Context, id string) (*Session, error)

	// Delete 删除会话。
	Delete(ctx context.Context, id string) error

	// Destroy 删除会话。
	Destroy(ctx context.Context, id string) error
}

type manager struct {
	store Store
	opts  Options
}

// New 创建会话管理器。
func New(store Store, opts ...Option) (Manager, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	if store == nil {
		return nil, ErrNilStore
	}
	if err := options.validate(); err != nil {
		return nil, err
	}

	return &manager{store: store, opts: options}, nil
}

// MustNew 创建会话管理器，失败时退出程序。
func MustNew(store Store, opts ...Option) Manager {
	m, err := New(store, opts...)
	if err != nil {
		log.Fatalf("session: create manager failed: %v", err)
	}
	return m
}

// Create 创建并保存新会话。
func (m *manager) Create(ctx context.Context) (*Session, error) {
	id, err := generateID(m.opts.IDLength)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	sess := &Session{
		ID:        id,
		Values:    make(map[string]any),
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(m.opts.TTL),
	}

	if err := m.store.Set(ctx, sess, m.opts.TTL); err != nil {
		return nil, err
	}
	return sess, nil
}

// Get 根据 ID 获取未过期会话。
func (m *manager) Get(ctx context.Context, id string) (*Session, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	sess, err := m.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if isExpired(sess, time.Now()) {
		_ = m.store.Delete(ctx, id)
		return nil, ErrExpired
	}
	if sess.Values == nil {
		sess.Values = make(map[string]any)
	}
	return sess, nil
}

// Save 保存会话数据并保留当前过期时间。
func (m *manager) Save(ctx context.Context, sess *Session) error {
	if sess == nil || sess.ID == "" {
		return ErrInvalidID
	}

	now := time.Now()
	if isExpired(sess, now) {
		_ = m.store.Delete(ctx, sess.ID)
		return ErrExpired
	}
	if sess.Values == nil {
		sess.Values = make(map[string]any)
	}
	sess.UpdatedAt = now

	ttl := time.Until(sess.ExpiresAt)
	if ttl <= 0 {
		return ErrExpired
	}
	return m.store.Set(ctx, sess, ttl)
}

// Refresh 延长会话有效期。
func (m *manager) Refresh(ctx context.Context, id string) (*Session, error) {
	sess, err := m.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	sess.UpdatedAt = now
	sess.ExpiresAt = now.Add(m.opts.TTL)
	if err := m.store.Set(ctx, sess, m.opts.TTL); err != nil {
		return nil, err
	}
	return sess, nil
}

// Delete 删除会话。
func (m *manager) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidID
	}
	return m.store.Delete(ctx, id)
}

// Destroy 删除会话。
func (m *manager) Destroy(ctx context.Context, id string) error {
	return m.Delete(ctx, id)
}

func generateID(length int) (string, error) {
	if length <= 0 {
		return "", ErrInvalidID
	}

	bytes := make([]byte, (length+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func isExpired(sess *Session, now time.Time) bool {
	if sess == nil {
		return true
	}
	return !sess.ExpiresAt.IsZero() && !sess.ExpiresAt.After(now)
}
