package s3_test

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/f2xme/gox/oss"
	"github.com/f2xme/gox/oss/adapter/s3"
)

func ExampleNew_r2() {
	storage, err := s3.New(
		s3.WithEndpoint("https://account-id.r2.cloudflarestorage.com"),
		s3.WithRegion("auto"),
		s3.WithCredentials("access-key-id", "access-key-secret"),
		s3.WithBucket("my-bucket"),
		s3.WithUsePathStyle(true),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	link, err := storage.SignURL(context.Background(), "uploads/avatar.png",
		oss.WithMethod(oss.MethodPut),
		oss.WithExpires(15*time.Minute),
		oss.WithSignContentType("image/png"),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 使用此链接上传时，必须发送 Content-Type: image/png。
	u, err := url.Parse(link)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(u.Host, u.Path)
	// Output: account-id.r2.cloudflarestorage.com /my-bucket/uploads/avatar.png
}
