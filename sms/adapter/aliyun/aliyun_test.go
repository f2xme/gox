package aliyun_test

import (
	"testing"

	"github.com/f2xme/gox/sms"
	"github.com/f2xme/gox/sms/adapter/aliyun"
)

func TestConnection(t *testing.T) {
	client, err := aliyun.New(
		aliyun.WithAccessKeyID("test-id"),
		aliyun.WithAccessKeySecret("test-secret"),
		aliyun.WithEndpoint("dysmsapi.aliyuncs.com"),
		aliyun.WithSignName("test-sign"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if client == nil {
		t.Fatal("New() returned nil client")
	}

	var _ sms.SMS = client
}
