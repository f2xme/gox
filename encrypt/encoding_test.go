// encrypt/encoding_test.go
package encrypt

import (
	"bytes"
	"testing"
)

func TestEncodeBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte{}, ""},
		{"hello", []byte("hello"), "aGVsbG8="},
		{"binary", []byte{0x00, 0xff, 0x01}, "AP8B"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeBase64(tt.input)
			if got != tt.expected {
				t.Errorf("EncodeBase64() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
		wantErr  bool
	}{
		{"empty", "", []byte{}, false},
		{"hello", "aGVsbG8=", []byte("hello"), false},
		{"binary", "AP8B", []byte{0x00, 0xff, 0x01}, false},
		{"invalid", "!!!invalid!!!", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeBase64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !bytes.Equal(got, tt.expected) {
				t.Errorf("DecodeBase64() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEncodeHex(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte{}, ""},
		{"hello", []byte("hello"), "68656c6c6f"},
		{"binary", []byte{0x00, 0xff, 0x01}, "00ff01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeHex(tt.input)
			if got != tt.expected {
				t.Errorf("EncodeHex() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDecodeHex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
		wantErr  bool
	}{
		{"empty", "", []byte{}, false},
		{"hello", "68656c6c6f", []byte("hello"), false},
		{"binary", "00ff01", []byte{0x00, 0xff, 0x01}, false},
		{"uppercase", "00FF01", []byte{0x00, 0xff, 0x01}, false},
		{"invalid", "zzzz", nil, true},
		{"odd_length", "abc", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeHex(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !bytes.Equal(got, tt.expected) {
				t.Errorf("DecodeHex() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBase64Roundtrip(t *testing.T) {
	data := []byte("roundtrip test data 你好世界")
	encoded := EncodeBase64(data)
	decoded, err := DecodeBase64(encoded)
	if err != nil {
		t.Fatalf("DecodeBase64() error = %v", err)
	}
	if !bytes.Equal(data, decoded) {
		t.Errorf("roundtrip failed: got %v, want %v", decoded, data)
	}
}

func TestHexRoundtrip(t *testing.T) {
	data := []byte("roundtrip test data 你好世界")
	encoded := EncodeHex(data)
	decoded, err := DecodeHex(encoded)
	if err != nil {
		t.Fatalf("DecodeHex() error = %v", err)
	}
	if !bytes.Equal(data, decoded) {
		t.Errorf("roundtrip failed: got %v, want %v", decoded, data)
	}
}
