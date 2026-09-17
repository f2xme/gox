package s3

import (
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/f2xme/gox/oss"
)

// New 创建 S3 存储实例；构造过程不访问网络。
func New(opts ...Option) (*Storage, error) {
	o := Options{Region: "us-east-1"}
	for _, opt := range opts {
		if opt == nil {
			return nil, oss.NewError(oss.ErrCodeInvalidArgument, "option must not be nil")
		}
		opt(&o)
	}
	return NewWithOptions(&o)
}

// NewWithOptions 使用配置创建 S3 存储实例；空 Region 默认使用 us-east-1。
func NewWithOptions(opts *Options) (*Storage, error) {
	if opts == nil {
		return nil, oss.NewError(oss.ErrCodeInvalidArgument, "options is required")
	}
	o := *opts
	if strings.TrimSpace(o.Bucket) == "" || strings.TrimSpace(o.AccessKeyID) == "" || strings.TrimSpace(o.AccessKeySecret) == "" {
		return nil, oss.NewError(oss.ErrCodeInvalidArgument, "bucket and credentials are required")
	}
	if o.Region == "" {
		o.Region = "us-east-1"
	}
	if strings.TrimSpace(o.Region) != o.Region {
		return nil, oss.NewError(oss.ErrCodeInvalidArgument, "invalid region")
	}
	if o.Endpoint != "" {
		u, err := url.Parse(o.Endpoint)
		if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return nil, oss.NewError(oss.ErrCodeInvalidArgument, "endpoint must be an HTTP(S) URL without credentials, query or fragment")
		}
	}
	cfg := awss3.Options{
		Region:       o.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(o.AccessKeyID, o.AccessKeySecret, o.SecurityToken),
		UsePathStyle: o.UsePathStyle,
		// 兼容 R2 等服务，避免 SDK 自动附加可选的 checksum/trailer。
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	}
	if o.Endpoint != "" {
		cfg.BaseEndpoint = aws.String(o.Endpoint)
	}
	if o.HTTPClient != nil {
		cfg.HTTPClient = o.HTTPClient
	}
	client := awss3.New(cfg)
	return &Storage{
		client:    client,
		presigner: awss3.NewPresignClient(client),
		uploader: transfermanager.New(client, func(o *transfermanager.Options) {
			o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
			o.FailTimeout = 30 * time.Second
		}),
		bucket: o.Bucket,
		region: o.Region,
	}, nil
}
