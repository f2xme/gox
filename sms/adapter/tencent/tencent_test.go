package tencent_test

import (
	"testing"

	"github.com/f2xme/gox/sms/adapter/tencent"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		secretID  string
		secretKey string
		region    string
		appID     string
		signName  string
		wantPanic bool
	}{
		{
			name:      "valid_config",
			secretID:  "test-id",
			secretKey: "test-key",
			region:    "test-region",
			appID:     "test-appid",
			signName:  "test-sign",
			wantPanic: false,
		},
		{
			name:      "empty_secret_id",
			secretID:  "",
			secretKey: "test-key",
			region:    "test-region",
			appID:     "test-appid",
			signName:  "test-sign",
			wantPanic: true,
		},
		{
			name:      "empty_secret_key",
			secretID:  "test-id",
			secretKey: "",
			region:    "test-region",
			appID:     "test-appid",
			signName:  "test-sign",
			wantPanic: true,
		},
		{
			name:      "empty_app_id",
			secretID:  "test-id",
			secretKey: "test-key",
			region:    "test-region",
			appID:     "",
			signName:  "test-sign",
			wantPanic: true,
		},
		{
			name:      "empty_sign_name",
			secretID:  "test-id",
			secretKey: "test-key",
			region:    "test-region",
			appID:     "test-appid",
			signName:  "",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tencent.New(
				tencent.WithSecretID(tt.secretID),
				tencent.WithSecretKey(tt.secretKey),
				tencent.WithRegion(tt.region),
				tencent.WithAppID(tt.appID),
				tencent.WithSignName(tt.signName),
			)
			if (err != nil) != tt.wantPanic {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantPanic)
			}
		})
	}
}
