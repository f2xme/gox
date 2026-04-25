package uni

import (
	"reflect"
	"testing"

	"github.com/f2xme/gox/sms"
)

func TestNewAllowsEmptyAccessKeySecret(t *testing.T) {
	client, err := New(
		WithAccessKeyID("test-id"),
		WithSignName("test-sign"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if client == nil {
		t.Fatal("New() returned nil client")
	}

	var _ sms.SMS = client
}

func TestNewRequiresAccessKeyID(t *testing.T) {
	_, err := New(WithSignName("test-sign"))
	if err == nil {
		t.Fatal("New() error = nil, want error")
	}
}

func TestNewRequiresSignName(t *testing.T) {
	_, err := New(WithAccessKeyID("test-id"))
	if err == nil {
		t.Fatal("New() error = nil, want error")
	}
}

func TestNormalizeTemplateData(t *testing.T) {
	tests := []struct {
		name    string
		param   any
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "nil",
			param: nil,
			want:  nil,
		},
		{
			name:  "map string string",
			param: map[string]string{"code": "6666"},
			want:  map[string]string{"code": "6666"},
		},
		{
			name:  "map string any",
			param: map[string]any{"code": 6666},
			want:  map[string]string{"code": "6666"},
		},
		{
			name:    "invalid type",
			param:   []string{"6666"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeTemplateData(tt.param)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeTemplateData() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("normalizeTemplateData() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
