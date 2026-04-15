package alioss

import (
	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/f2xme/gox/oss"
)

// New 创建一个新的阿里云 OSS 存储实例
//
// 参数：
//   - endpoint: OSS 端点地址（如 oss-cn-hangzhou.aliyuncs.com）
//   - accessKeyID: Access Key ID
//   - accessKeySecret: Access Key Secret
//   - bucket: 存储桶名称
//   - opts: 可选配置函数
//
// 返回值：
//   - *Storage: 存储实例
//   - error: 创建失败时返回错误
func New(endpoint, accessKeyID, accessKeySecret, bucket string, opts ...Option) (*Storage, error) {
	o := defaultOptions()
	o.Endpoint = endpoint
	o.AccessKeyID = accessKeyID
	o.AccessKeySecret = accessKeySecret
	o.Bucket = bucket

	for _, opt := range opts {
		opt(&o)
	}

	if err := o.Validate(); err != nil {
		return nil, err
	}

	client, err := alioss.New(o.Endpoint, o.AccessKeyID, o.AccessKeySecret, o.buildClientOptions()...)
	if err != nil {
		return nil, oss.WrapError(oss.ErrCodeInternal, "failed to create client", err)
	}

	bucketHandle, err := client.Bucket(o.Bucket)
	if err != nil {
		return nil, oss.WrapError(oss.ErrCodeInternal, "failed to get bucket", err)
	}

	return &Storage{
		client:       client,
		bucket:       o.Bucket,
		bucketHandle: bucketHandle,
	}, nil
}
