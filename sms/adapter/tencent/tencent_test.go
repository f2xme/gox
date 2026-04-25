package tencent_test

import (
	"testing"

	"github.com/f2xme/gox/sms"
	"github.com/f2xme/gox/sms/adapter/tencent"
)

func TestConnection(t *testing.T) {
	client, err := tencent.New(
		tencent.WithSecretID("test-id"),
		tencent.WithSecretKey("test-key"),
		tencent.WithRegion("ap-guangzhou"),
		tencent.WithAppID("test-appid"),
		tencent.WithSignName("test-sign"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if client == nil {
		t.Fatal("New() returned nil client")
	}

	var _ sms.SMS = client
}
