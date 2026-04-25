package aliyun_test

import (
	"testing"

	"github.com/f2xme/gox/sms/adapter/aliyun"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name            string
		accessKeyID     string
		accessKeySecret string
		endpoint        string
		signName        string
		wantPanic       bool
	}{
		{
			name:            "valid_config",
			accessKeyID:     "test-id",
			accessKeySecret: "test-secret",
			endpoint:        "test-endpoint",
			signName:        "test-sign",
			wantPanic:       false,
		},
		{
			name:            "empty_access_key_id",
			accessKeyID:     "",
			accessKeySecret: "test-secret",
			endpoint:        "test-endpoint",
			signName:        "test-sign",
			wantPanic:       true,
		},
		{
			name:            "empty_access_key_secret",
			accessKeyID:     "test-id",
			accessKeySecret: "",
			endpoint:        "test-endpoint",
			signName:        "test-sign",
			wantPanic:       true,
		},
		{
			name:            "empty_sign_name",
			accessKeyID:     "test-id",
			accessKeySecret: "test-secret",
			endpoint:        "test-endpoint",
			signName:        "",
			wantPanic:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := aliyun.New(
				aliyun.WithAccessKeyID(tt.accessKeyID),
				aliyun.WithAccessKeySecret(tt.accessKeySecret),
				aliyun.WithEndpoint(tt.endpoint),
				aliyun.WithSignName(tt.signName),
			)
			if (err != nil) != tt.wantPanic {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantPanic)
			}
		})
	}
}
