package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/f2xme/gox/oss"
)

// Storage 是并发安全的 S3 对象存储适配器。
type Storage struct {
	client    *awss3.Client
	presigner *awss3.PresignClient
	uploader  *transfermanager.Client
	bucket    string
	region    string
}

var _ oss.Storage = (*Storage)(nil)
var _ oss.BucketStorage = (*Storage)(nil)

// Put 上传对象；普通 io.Reader 和大对象由 SDK 自动缓冲、分片上传。
func (s *Storage) Put(ctx context.Context, key string, reader io.Reader, opts ...oss.PutOption) error {
	if err := validateKey(ctx, key); err != nil {
		return err
	}
	if reader == nil {
		return oss.NewError(oss.ErrCodeInvalidArgument, "reader is required", key)
	}
	o := oss.ApplyPutOptions(opts...)
	if o.ContentType == "" {
		o.ContentType = oss.DetectContentType(key)
	}
	_, err := s.uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: &s.bucket, Key: &key, Body: reader, ContentType: &o.ContentType, Metadata: o.Metadata,
	})
	return convertError(err, key)
}

// Get 下载对象；调用方负责关闭返回的流。RangeEnd 为负数时读取到对象末尾。
func (s *Storage) Get(ctx context.Context, key string, opts ...oss.GetOption) (io.ReadCloser, error) {
	if err := validateKey(ctx, key); err != nil {
		return nil, err
	}
	o := oss.ApplyGetOptions(opts...)
	in := &awss3.GetObjectInput{Bucket: &s.bucket, Key: &key}
	if o.RangeStart >= 0 {
		r := fmt.Sprintf("bytes=%d-", o.RangeStart)
		if o.RangeEnd >= 0 {
			if o.RangeEnd < o.RangeStart {
				return nil, oss.NewError(oss.ErrCodeInvalidArgument, "range end must not precede start", key)
			}
			r += fmt.Sprint(o.RangeEnd)
		}
		in.Range = &r
	}
	out, err := s.client.GetObject(ctx, in)
	if err != nil {
		return nil, convertError(err, key)
	}
	return out.Body, nil
}

// Delete 删除对象；删除不存在的对象遵循 S3 的幂等语义。
func (s *Storage) Delete(ctx context.Context, key string) error {
	if err := validateKey(ctx, key); err != nil {
		return err
	}
	_, err := s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{Bucket: &s.bucket, Key: &key})
	return convertError(err, key)
}

// Stat 获取对象元信息。
func (s *Storage) Stat(ctx context.Context, key string) (*oss.ObjectInfo, error) {
	if err := validateKey(ctx, key); err != nil {
		return nil, err
	}
	out, err := s.client.HeadObject(ctx, &awss3.HeadObjectInput{Bucket: &s.bucket, Key: &key})
	if err != nil {
		return nil, convertError(err, key)
	}
	return &oss.ObjectInfo{
		Key: key, Size: aws.ToInt64(out.ContentLength), LastModified: aws.ToTime(out.LastModified),
		ETag: aws.ToString(out.ETag), ContentType: aws.ToString(out.ContentType), Metadata: out.Metadata,
	}, nil
}

// Exists 检查对象是否存在；权限错误仍返回 error。
func (s *Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.Stat(ctx, key)
	if oss.IsNotFound(err) {
		return false, nil
	}
	return err == nil, err
}

// List 返回一页对象；NextToken 可传给 oss.WithToken 获取下一页。
// S3 列表不返回 ContentType，该字段为空；需要时使用 Stat。
func (s *Storage) List(ctx context.Context, opts ...oss.ListOption) (*oss.ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	o := oss.ApplyListOptions(opts...)
	if o.Limit < 0 || o.Limit > 1000 {
		return nil, oss.NewError(oss.ErrCodeInvalidArgument, "limit must be between 0 and 1000")
	}
	in := &awss3.ListObjectsV2Input{Bucket: &s.bucket, Prefix: &o.Prefix, Delimiter: &o.Delimiter, EncodingType: types.EncodingTypeUrl}
	if o.Limit > 0 {
		in.MaxKeys = aws.Int32(int32(o.Limit))
	}
	if o.Token != "" {
		in.ContinuationToken = &o.Token
	}
	out, err := s.client.ListObjectsV2(ctx, in)
	if err != nil {
		return nil, convertError(err, "")
	}
	decodeName := func(value *string) (string, error) {
		name := aws.ToString(value)
		if out.EncodingType == types.EncodingTypeUrl {
			return url.PathUnescape(name)
		}
		return name, nil
	}
	result := &oss.ListResult{NextToken: aws.ToString(out.NextContinuationToken), Truncated: aws.ToBool(out.IsTruncated)}
	for _, obj := range out.Contents {
		key, err := decodeName(obj.Key)
		if err != nil {
			return nil, convertError(err, "")
		}
		result.Objects = append(result.Objects, &oss.Object{Key: key, Size: aws.ToInt64(obj.Size), LastModified: aws.ToTime(obj.LastModified), ETag: aws.ToString(obj.ETag)})
	}
	for _, prefix := range out.CommonPrefixes {
		name, err := decodeName(prefix.Prefix)
		if err != nil {
			return nil, convertError(err, "")
		}
		result.Prefixes = append(result.Prefixes, name)
	}
	return result, nil
}

// SignURL 生成 GET、PUT 或 DELETE 预签名 URL，有效期为 1 秒至 7 天。
// PUT 指定 ContentType 后，使用 URL 时必须发送相同的 Content-Type 请求头。
func (s *Storage) SignURL(ctx context.Context, key string, opts ...oss.SignOption) (string, error) {
	if err := validateKey(ctx, key); err != nil {
		return "", err
	}
	o := oss.ApplySignOptions(opts...)
	if o.Expires < time.Second || o.Expires > 7*24*time.Hour {
		return "", oss.NewError(oss.ErrCodeInvalidArgument, "expires must be between 1 second and 7 days", key)
	}
	presignOpts := func(p *awss3.PresignOptions) { p.Expires = o.Expires }
	var out *v4.PresignedHTTPRequest
	var err error
	switch o.Method {
	case oss.MethodGet:
		out, err = s.presigner.PresignGetObject(ctx, &awss3.GetObjectInput{Bucket: &s.bucket, Key: &key}, presignOpts)
	case oss.MethodPut:
		in := &awss3.PutObjectInput{Bucket: &s.bucket, Key: &key}
		if o.ContentType != "" {
			in.ContentType = &o.ContentType
		}
		out, err = s.presigner.PresignPutObject(ctx, in, presignOpts, func(p *awss3.PresignOptions) {
			if o.ContentType == "" {
				return
			}
			p.ClientOptions = append(p.ClientOptions, func(c *awss3.Options) {
				c.APIOptions = append(c.APIOptions, func(stack *middleware.Stack) error {
					// SDK 默认删除未知长度请求的 Content-Type；保留显式指定的签名约束。
					_, err := stack.Build.Remove("RemoveContentTypeHeader")
					return err
				})
			})
		})
	case oss.MethodDelete:
		out, err = s.presigner.PresignDeleteObject(ctx, &awss3.DeleteObjectInput{Bucket: &s.bucket, Key: &key}, presignOpts)
	default:
		return "", oss.NewError(oss.ErrCodeInvalidArgument, "unsupported method: "+o.Method, key)
	}
	if err != nil {
		return "", convertError(err, key)
	}
	return out.URL, nil
}

// CreateBucket 创建存储桶。Region 默认使用客户端地域；ACL 是否支持取决于服务商。
func (s *Storage) CreateBucket(ctx context.Context, bucket string, opts ...oss.BucketOption) error {
	if err := validateBucket(ctx, bucket); err != nil {
		return err
	}
	o := oss.ApplyBucketOptions(opts...)
	region := o.Region
	if region == "" {
		region = s.region
	}
	in := &awss3.CreateBucketInput{Bucket: &bucket, ACL: types.BucketCannedACL(o.ACL)}
	if region != "us-east-1" && region != "auto" {
		in.CreateBucketConfiguration = &types.CreateBucketConfiguration{LocationConstraint: types.BucketLocationConstraint(region)}
	}
	_, err := s.client.CreateBucket(ctx, in, func(o *awss3.Options) { o.Region = region })
	return convertError(err, "")
}

// DeleteBucket 删除空存储桶。
func (s *Storage) DeleteBucket(ctx context.Context, bucket string) error {
	if err := validateBucket(ctx, bucket); err != nil {
		return err
	}
	_, err := s.client.DeleteBucket(ctx, &awss3.DeleteBucketInput{Bucket: &bucket})
	return convertError(err, "")
}

// ListBuckets 列出所有可见存储桶，自动遍历服务端分页。
func (s *Storage) ListBuckets(ctx context.Context) ([]*oss.Bucket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	in := &awss3.ListBucketsInput{MaxBuckets: aws.Int32(1000)}
	var buckets []*oss.Bucket
	for {
		out, err := s.client.ListBuckets(ctx, in)
		if err != nil {
			return nil, convertError(err, "")
		}
		for _, b := range out.Buckets {
			buckets = append(buckets, &oss.Bucket{Name: aws.ToString(b.Name), CreationDate: aws.ToTime(b.CreationDate), Region: aws.ToString(b.BucketRegion)})
		}
		next := aws.ToString(out.ContinuationToken)
		if response, ok := awsmiddleware.GetRawResponse(out.ResultMetadata).(*smithyhttp.Response); ok {
			// R2 的 XML ContinuationToken 回显当前令牌；下一页令牌通过扩展头提供。
			// https://developers.cloudflare.com/r2/api/s3/extensions/#listbuckets
			switch response.Header.Get("Cf-Is-Truncated") {
			case "true":
				next = response.Header.Get("Cf-Next-Continuation-Token")
				if next == "" {
					return nil, oss.NewError(oss.ErrCodeInternal, "truncated bucket list has no next token")
				}
			case "false":
				next = ""
			}
		}
		if next == "" {
			return buckets, nil
		}
		if next == aws.ToString(in.ContinuationToken) {
			return nil, oss.NewError(oss.ErrCodeInternal, "bucket list repeated continuation token")
		}
		in.ContinuationToken = aws.String(next)
	}
}

func validateKey(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if key == "" {
		return oss.NewError(oss.ErrCodeInvalidArgument, "key is required")
	}
	return nil
}

func validateBucket(ctx context.Context, bucket string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(bucket) == "" {
		return oss.NewError(oss.ErrCodeInvalidArgument, "bucket is required")
	}
	return nil
}

func convertError(err error, key string) error {
	if err == nil {
		return nil
	}
	code := oss.ErrCodeInternal
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NoSuchBucket", "NotFound":
			code = oss.ErrCodeNotFound
		case "AccessDenied", "InvalidAccessKeyId", "SignatureDoesNotMatch", "ExpiredToken", "InvalidToken":
			code = oss.ErrCodeAccessDenied
		case "InvalidArgument", "InvalidRequest", "InvalidRange", "InvalidBucketName", "MalformedXML":
			code = oss.ErrCodeInvalidArgument
		case "BucketNotEmpty":
			code = oss.ErrCodeBucketNotEmpty
		case "BucketAlreadyExists", "BucketAlreadyOwnedByYou":
			code = oss.ErrCodeBucketExists
		}
	}
	// HEAD 错误可能只有 HTTP 状态码，没有 XML 错误体。
	var responseErr *smithyhttp.ResponseError
	if code == oss.ErrCodeInternal && errors.As(err, &responseErr) {
		switch responseErr.HTTPStatusCode() {
		case 404:
			code = oss.ErrCodeNotFound
		case 401, 403:
			code = oss.ErrCodeAccessDenied
		case 400, 416:
			code = oss.ErrCodeInvalidArgument
		}
	}
	return oss.WrapError(code, err.Error(), err, key)
}
