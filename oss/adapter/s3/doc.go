// Package s3 使用 AWS SDK v2 实现 oss.Storage 和 oss.BucketStorage，
// 可连接 AWS S3、Cloudflare R2 及兼容 S3 API 的服务。
//
// # Cloudflare R2
//
//	storage, err := s3.New(
//		s3.WithEndpoint("https://<ACCOUNT_ID>.r2.cloudflarestorage.com"),
//		s3.WithRegion("auto"),
//		s3.WithCredentials(accessKeyID, accessKeySecret),
//		s3.WithBucket("my-bucket"),
//		s3.WithUsePathStyle(true),
//	)
//	if err != nil {
//		return err
//	}
//	err = storage.Put(ctx, "hello.txt", strings.NewReader("hello"),
//		oss.WithMetadata(map[string]string{"author": "gox"}),
//	)
//
// Endpoint 使用 R2 S3 API 地址，而非公开访问域名或 r2.dev 地址。
// R2 不支持 Bucket ACL，创建桶时不要传入 oss.WithBucketACL。
// 参考：https://developers.cloudflare.com/r2/api/s3/api/
//
// # AWS S3 与 MinIO
//
// AWS S3 可省略 Endpoint，使用 WithRegion 指定桶的地域，默认 us-east-1。
// 必须显式提供凭证；临时凭证额外使用 WithSecurityToken。
// MinIO 通常需要 WithEndpoint("http://localhost:9000") 和 WithUsePathStyle(true)。
// WithHTTPClient 可传入带 Timeout 或自定义 Transport 的 *http.Client。
//
// # 上传、分页与预签名
//
// Put 接受普通 io.Reader，SDK 自动分片上传大对象并在失败后尝试清理分片。
// 上传过程中读取 reader 的阻塞行为由调用方负责；应保证 reader 能及时返回。
// List 返回一页（最多 1000 项），NextToken 传给 oss.WithToken 获取下一页。
// List 不返回 ContentType，需要时调用 Stat。
// ListBuckets 自动合并所有分页。
//
// SignURL 支持 GET、PUT、DELETE，有效期为 1 秒至 7 天；临时凭证到期后链接也会失效。
// PUT 使用 oss.WithSignContentType 后，客户端必须发送相同的 Content-Type 请求头。
// R2 预签名 URL 必须使用 S3 API 域名，不能替换成自定义公开域名。
//
// # 测试
//
// 默认测试使用本地 HTTP 服务，不访问云端。显式运行真实服务测试：
//
//	GOX_S3_ENDPOINT=https://<ACCOUNT_ID>.r2.cloudflarestorage.com \
//	GOX_S3_REGION=auto GOX_S3_BUCKET=my-bucket \
//	GOX_S3_ACCESS_KEY_ID=... GOX_S3_ACCESS_KEY_SECRET=... \
//	go test -run TestIntegration -v ./...
//
// 测试在既有桶内创建唯一对象并清理，不创建或删除桶。
package s3
