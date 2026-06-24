package memory

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/f2xme/gox/oss"
)

// Storage 是基于内存的 OSS 存储实现。
type Storage struct {
	mu      sync.RWMutex
	objects map[string]*object
	buckets map[string]*oss.Bucket
	options Options
}

var _ oss.Storage = (*Storage)(nil)
var _ oss.BucketStorage = (*Storage)(nil)

type object struct {
	key          string
	data         []byte
	contentType  string
	metadata     map[string]string
	etag         string
	lastModified time.Time
}

// Put 上传对象。
func (s *Storage) Put(ctx context.Context, key string, reader io.Reader, opts ...oss.PutOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(key) == "" {
		return oss.NewError(oss.ErrCodeInvalidArgument, "key is required")
	}
	if reader == nil {
		return oss.NewError(oss.ErrCodeInvalidArgument, "reader is required", key)
	}

	s.mu.RLock()
	if !s.defaultBucketExistsLocked() {
		s.mu.RUnlock()
		return oss.NewError(oss.ErrCodeNotFound, "bucket not found")
	}
	s.mu.RUnlock()

	data, err := io.ReadAll(reader)
	if err != nil {
		return oss.WrapError(oss.ErrCodeInternal, "failed to read object", err, key)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	options := oss.ApplyPutOptions(opts...)
	contentType := options.ContentType
	if contentType == "" {
		contentType = oss.DetectContentType(key)
	}

	sum := sha256.Sum256(data)
	item := &object{
		key:          key,
		data:         copyBytes(data),
		contentType:  contentType,
		metadata:     cloneMap(options.Metadata),
		etag:         hex.EncodeToString(sum[:]),
		lastModified: time.Now(),
	}

	s.mu.Lock()
	if !s.defaultBucketExistsLocked() {
		s.mu.Unlock()
		return oss.NewError(oss.ErrCodeNotFound, "bucket not found")
	}
	s.objects[key] = item
	s.mu.Unlock()
	return nil
}

// Get 下载对象。
func (s *Storage) Get(ctx context.Context, key string, opts ...oss.GetOption) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	options := oss.ApplyGetOptions(opts...)

	s.mu.RLock()
	if !s.defaultBucketExistsLocked() {
		s.mu.RUnlock()
		return nil, oss.NewError(oss.ErrCodeNotFound, "bucket not found")
	}
	item, ok := s.objects[key]
	if !ok {
		s.mu.RUnlock()
		return nil, oss.NewError(oss.ErrCodeNotFound, "object not found", key)
	}
	data := copyBytes(item.data)
	s.mu.RUnlock()

	if options.RangeStart >= 0 {
		ranged, err := sliceRange(data, options.RangeStart, options.RangeEnd)
		if err != nil {
			return nil, oss.NewError(oss.ErrCodeInvalidArgument, err.Error(), key)
		}
		data = ranged
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

// Delete 删除对象。
func (s *Storage) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	if !s.defaultBucketExistsLocked() {
		s.mu.Unlock()
		return oss.NewError(oss.ErrCodeNotFound, "bucket not found")
	}
	delete(s.objects, key)
	s.mu.Unlock()
	return nil
}

// Stat 获取对象元信息。
func (s *Storage) Stat(ctx context.Context, key string) (*oss.ObjectInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	if !s.defaultBucketExistsLocked() {
		s.mu.RUnlock()
		return nil, oss.NewError(oss.ErrCodeNotFound, "bucket not found")
	}
	item, ok := s.objects[key]
	if !ok {
		s.mu.RUnlock()
		return nil, oss.NewError(oss.ErrCodeNotFound, "object not found", key)
	}
	info := item.toObjectInfo()
	s.mu.RUnlock()
	return info, nil
}

// Exists 检查对象是否存在。
func (s *Storage) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	s.mu.RLock()
	if !s.defaultBucketExistsLocked() {
		s.mu.RUnlock()
		return false, oss.NewError(oss.ErrCodeNotFound, "bucket not found")
	}
	_, ok := s.objects[key]
	s.mu.RUnlock()
	return ok, nil
}

// List 列出对象。
func (s *Storage) List(ctx context.Context, opts ...oss.ListOption) (*oss.ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	options := oss.ApplyListOptions(opts...)
	if options.Limit < 0 {
		return nil, oss.NewError(oss.ErrCodeInvalidArgument, "limit must not be negative")
	}

	s.mu.RLock()
	if !s.defaultBucketExistsLocked() {
		s.mu.RUnlock()
		return nil, oss.NewError(oss.ErrCodeNotFound, "bucket not found")
	}
	items := make([]*object, 0, len(s.objects))
	for _, item := range s.objects {
		items = append(items, item.clone())
	}
	s.mu.RUnlock()

	sort.Slice(items, func(i, j int) bool {
		return items[i].key < items[j].key
	})

	start, err := parseToken(options.Token, len(items))
	if err != nil {
		return nil, err
	}

	limit := options.Limit
	if limit == 0 {
		limit = len(items)
	}

	result := &oss.ListResult{
		Objects:  make([]*oss.Object, 0),
		Prefixes: make([]string, 0),
	}
	seenPrefixes := make(map[string]struct{})
	count := 0
	for i := start; i < len(items); i++ {
		item := items[i]
		if !strings.HasPrefix(item.key, options.Prefix) {
			continue
		}

		if options.Delimiter != "" {
			rest := strings.TrimPrefix(item.key, options.Prefix)
			if idx := strings.Index(rest, options.Delimiter); idx >= 0 {
				prefix := options.Prefix + rest[:idx+len(options.Delimiter)]
				if _, ok := seenPrefixes[prefix]; ok {
					continue
				}
				seenPrefixes[prefix] = struct{}{}
				if count == limit {
					result.NextToken = strconv.Itoa(i)
					result.Truncated = true
					return result, nil
				}
				result.Prefixes = append(result.Prefixes, prefix)
				count++
				continue
			}
		}

		if count == limit {
			result.NextToken = strconv.Itoa(i)
			result.Truncated = true
			return result, nil
		}
		result.Objects = append(result.Objects, item.toObject())
		count++
	}

	return result, nil
}

// CreateBucket 创建存储桶。
func (s *Storage) CreateBucket(ctx context.Context, bucket string, opts ...oss.BucketOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(bucket) == "" {
		return oss.NewError(oss.ErrCodeInvalidArgument, "bucket is required")
	}

	options := oss.ApplyBucketOptions(opts...)

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.buckets[bucket]; ok {
		return oss.NewError(oss.ErrCodeBucketExists, "bucket already exists")
	}

	s.buckets[bucket] = &oss.Bucket{
		Name:         bucket,
		CreationDate: time.Now(),
		Region:       options.Region,
	}
	return nil
}

// DeleteBucket 删除存储桶。
func (s *Storage) DeleteBucket(ctx context.Context, bucket string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.buckets[bucket]; !ok {
		return oss.NewError(oss.ErrCodeNotFound, "bucket not found")
	}
	if bucket == s.options.BucketName && len(s.objects) > 0 {
		return oss.NewError(oss.ErrCodeBucketNotEmpty, "bucket is not empty")
	}

	delete(s.buckets, bucket)
	return nil
}

// ListBuckets 列出存储桶。
func (s *Storage) ListBuckets(ctx context.Context) ([]*oss.Bucket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	buckets := make([]*oss.Bucket, 0, len(s.buckets))
	for _, bucket := range s.buckets {
		buckets = append(buckets, cloneBucket(bucket))
	}
	s.mu.RUnlock()

	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Name < buckets[j].Name
	})
	return buckets, nil
}

// SignURL 生成测试用预签名 URL。
func (s *Storage) SignURL(ctx context.Context, key string, opts ...oss.SignOption) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	options := oss.ApplySignOptions(opts...)
	switch options.Method {
	case oss.MethodGet, oss.MethodPut, oss.MethodDelete:
	default:
		return "", oss.NewError(oss.ErrCodeInvalidArgument, "unsupported method: "+options.Method, key)
	}
	if options.Expires <= 0 {
		return "", oss.NewError(oss.ErrCodeInvalidArgument, "expires must be positive", key)
	}

	s.mu.RLock()
	if !s.defaultBucketExistsLocked() {
		s.mu.RUnlock()
		return "", oss.NewError(oss.ErrCodeNotFound, "bucket not found")
	}
	s.mu.RUnlock()

	values := url.Values{}
	values.Set("method", options.Method)
	values.Set("expires", strconv.FormatInt(int64(options.Expires/time.Second), 10))
	if options.ContentType != "" {
		values.Set("content-type", options.ContentType)
	}

	return s.options.SignURLBase + "/" + url.PathEscape(key) + "?" + values.Encode(), nil
}

func (o *object) clone() *object {
	return &object{
		key:          o.key,
		data:         copyBytes(o.data),
		contentType:  o.contentType,
		metadata:     cloneMap(o.metadata),
		etag:         o.etag,
		lastModified: o.lastModified,
	}
}

func (o *object) toObject() *oss.Object {
	return &oss.Object{
		Key:          o.key,
		Size:         int64(len(o.data)),
		LastModified: o.lastModified,
		ETag:         o.etag,
		ContentType:  o.contentType,
	}
}

func (o *object) toObjectInfo() *oss.ObjectInfo {
	return &oss.ObjectInfo{
		Key:          o.key,
		Size:         int64(len(o.data)),
		LastModified: o.lastModified,
		ETag:         o.etag,
		ContentType:  o.contentType,
		Metadata:     cloneMap(o.metadata),
	}
}

func copyBytes(value []byte) []byte {
	copied := make([]byte, len(value))
	copy(copied, value)
	return copied
}

func cloneMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for k, v := range values {
		cloned[k] = v
	}
	return cloned
}

func cloneBucket(bucket *oss.Bucket) *oss.Bucket {
	if bucket == nil {
		return nil
	}
	return &oss.Bucket{
		Name:         bucket.Name,
		CreationDate: bucket.CreationDate,
		Region:       bucket.Region,
	}
}

func (s *Storage) defaultBucketExistsLocked() bool {
	_, ok := s.buckets[s.options.BucketName]
	return ok
}

func sliceRange(data []byte, start, end int64) ([]byte, error) {
	if start < 0 || end < start {
		return nil, errInvalidRange
	}
	if start >= int64(len(data)) {
		return nil, errInvalidRange
	}
	if end >= int64(len(data)) {
		end = int64(len(data)) - 1
	}
	return copyBytes(data[start : end+1]), nil
}

var errInvalidRange = &rangeError{}

type rangeError struct{}

func (e *rangeError) Error() string {
	return "invalid range"
}

func parseToken(token string, max int) (int, error) {
	if token == "" {
		return 0, nil
	}
	start, err := strconv.Atoi(token)
	if err != nil || start < 0 || start > max {
		return 0, oss.NewError(oss.ErrCodeInvalidArgument, "invalid token")
	}
	return start, nil
}
