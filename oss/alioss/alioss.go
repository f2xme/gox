package alioss

import (
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/f2xme/gox/oss"
)

// Storage 阿里云 OSS 存储实现
type Storage struct {
	client       *alioss.Client
	bucket       string
	bucketHandle *alioss.Bucket
}

// convertError 转换阿里云错误为统一错误
func (s *Storage) convertError(err error, key string) error {
	if err == nil {
		return nil
	}

	code := oss.ErrCodeInternal
	message := err.Error()

	if serviceErr, ok := err.(alioss.ServiceError); ok {
		message = serviceErr.Message
		switch serviceErr.Code {
		case "NoSuchKey":
			code = oss.ErrCodeNotFound
		case "AccessDenied":
			code = oss.ErrCodeAccessDenied
		case "InvalidArgument":
			code = oss.ErrCodeInvalidArgument
		case "BucketNotEmpty":
			code = oss.ErrCodeBucketNotEmpty
		case "BucketAlreadyExists":
			code = oss.ErrCodeBucketExists
		}
	}

	return &oss.Error{Code: code, Message: message, Key: key, Err: err}
}
func (s *Storage) Put(ctx context.Context, key string, reader io.Reader, opts ...oss.PutOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	options := &oss.PutOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.ContentType == "" {
		options.ContentType = oss.DetectContentType(key)
	}

	sdkOpts := make([]alioss.Option, 0, 1+len(options.Metadata))
	sdkOpts = append(sdkOpts, alioss.ContentType(options.ContentType))
	for k, v := range options.Metadata {
		sdkOpts = append(sdkOpts, alioss.Meta(k, v))
	}

	if err := s.bucketHandle.PutObject(key, reader, sdkOpts...); err != nil {
		return s.convertError(err, key)
	}

	return nil
}

// Get 下载对象
func (s *Storage) Get(ctx context.Context, key string, opts ...oss.GetOption) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	options := &oss.GetOptions{Start: -1}
	for _, opt := range opts {
		opt(options)
	}

	var sdkOpts []alioss.Option
	if options.Start >= 0 {
		sdkOpts = append(sdkOpts, alioss.Range(options.Start, options.End))
	}

	body, err := s.bucketHandle.GetObject(key, sdkOpts...)
	if err != nil {
		return nil, s.convertError(err, key)
	}

	return body, nil
}

// Delete 删除对象
func (s *Storage) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := s.bucketHandle.DeleteObject(key); err != nil {
		return s.convertError(err, key)
	}

	return nil
}

// Head 获取对象元信息
func (s *Storage) Head(ctx context.Context, key string) (*oss.ObjectInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	meta, err := s.bucketHandle.GetObjectMeta(key)
	if err != nil {
		return nil, s.convertError(err, key)
	}

	info := &oss.ObjectInfo{
		Key:         key,
		ContentType: meta.Get("Content-Type"),
		ETag:        meta.Get("ETag"),
	}

	if sizeStr := meta.Get("Content-Length"); sizeStr != "" {
		if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			info.Size = size
		}
	}

	if modStr := meta.Get("Last-Modified"); modStr != "" {
		if t, err := time.Parse(time.RFC1123, modStr); err == nil {
			info.LastModified = t
		}
	}

	for k := range meta {
		if metaKey, ok := strings.CutPrefix(k, "X-Oss-Meta-"); ok {
			if info.Metadata == nil {
				info.Metadata = make(map[string]string)
			}
			info.Metadata[metaKey] = meta.Get(k)
		}
	}

	return info, nil
}

// Exists 检查对象是否存在
func (s *Storage) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	exists, err := s.bucketHandle.IsObjectExist(key)
	if err != nil {
		return false, s.convertError(err, key)
	}

	return exists, nil
}

// List 列出对象
func (s *Storage) List(ctx context.Context, opts ...oss.ListOption) ([]*oss.Object, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	options := oss.DefaultListOptions()
	for _, opt := range opts {
		opt(options)
	}

	sdkOpts := make([]alioss.Option, 0, 4)
	if options.Prefix != "" {
		sdkOpts = append(sdkOpts, alioss.Prefix(options.Prefix))
	}
	if options.Delimiter != "" {
		sdkOpts = append(sdkOpts, alioss.Delimiter(options.Delimiter))
	}
	if options.MaxKeys > 0 {
		sdkOpts = append(sdkOpts, alioss.MaxKeys(options.MaxKeys))
	}
	if options.Marker != "" {
		sdkOpts = append(sdkOpts, alioss.Marker(options.Marker))
	}

	result, err := s.bucketHandle.ListObjects(sdkOpts...)
	if err != nil {
		return nil, s.convertError(err, "")
	}

	objects := make([]*oss.Object, len(result.Objects))
	for i, obj := range result.Objects {
		objects[i] = &oss.Object{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
			ETag:         obj.ETag,
			ContentType:  obj.Type,
		}
	}

	return objects, nil
}

// CreateBucket 创建存储桶
func (s *Storage) CreateBucket(ctx context.Context, bucket string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.client.CreateBucket(bucket); err != nil {
		return s.convertError(err, "")
	}
	return nil
}

// DeleteBucket 删除存储桶
func (s *Storage) DeleteBucket(ctx context.Context, bucket string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.client.DeleteBucket(bucket); err != nil {
		return s.convertError(err, "")
	}
	return nil
}

// ListBuckets 列出所有存储桶
func (s *Storage) ListBuckets(ctx context.Context) ([]*oss.Bucket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result, err := s.client.ListBuckets()
	if err != nil {
		return nil, s.convertError(err, "")
	}

	buckets := make([]*oss.Bucket, len(result.Buckets))
	for i, b := range result.Buckets {
		buckets[i] = &oss.Bucket{
			Name:         b.Name,
			CreationDate: b.CreationDate,
			Region:       b.Location,
		}
	}

	return buckets, nil
}

// PresignedURL 生成预签名 URL
func (s *Storage) PresignedURL(ctx context.Context, key string, opts ...oss.PresignedOption) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	options := oss.DefaultPresignedOptions()
	for _, opt := range opts {
		opt(options)
	}

	var url string
	var err error
	switch options.Method {
	case oss.MethodGet:
		url, err = s.bucketHandle.SignURL(key, alioss.HTTPGet, int64(options.Expires.Seconds()))
	case oss.MethodPut:
		url, err = s.bucketHandle.SignURL(key, alioss.HTTPPut, int64(options.Expires.Seconds()))
	case oss.MethodDelete:
		url, err = s.bucketHandle.SignURL(key, alioss.HTTPDelete, int64(options.Expires.Seconds()))
	default:
		return "", oss.NewError(oss.ErrCodeInvalidArgument, "unsupported method: "+options.Method)
	}

	if err != nil {
		return "", s.convertError(err, key)
	}

	return url, nil
}
