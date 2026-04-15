// encrypt/hash_test.go
package encrypt

import "testing"

func TestMD5(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte(""), "d41d8cd98f00b204e9800998ecf8427e"},
		{"123456", []byte("123456"), "e10adc3949ba59abbe56e057f20f883e"},
		{"hello", []byte("hello"), "5d41402abc4b2a76b9719d911017c592"},
		{"chinese", []byte("你好"), "7eca689f0d3389d9dea66ae112e5cfd7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MD5(tt.input)
			if got != tt.expected {
				t.Errorf("MD5() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSHA1(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte(""), "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"hello", []byte("hello"), "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"123456", []byte("123456"), "7c4a8d09ca3762af61e59520943dc26494f8941b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SHA1(tt.input)
			if got != tt.expected {
				t.Errorf("SHA1() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSHA256(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte(""), "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"hello", []byte("hello"), "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"123456", []byte("123456"), "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SHA256(tt.input)
			if got != tt.expected {
				t.Errorf("SHA256() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSHA512(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte(""), "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
		{"hello", []byte("hello"), "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SHA512(tt.input)
			if got != tt.expected {
				t.Errorf("SHA512() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestBlake3(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte(""), "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262"},
		{"123456", []byte("123456"), "7adb787627ad5ee341fa0ba46a956e78fd85c39e195119bb260d5181b4f1e4ba"},
		{"hello", []byte("hello"), "ea8f163db38682925e4491c5e58d4bb3506ef8c14eb78a86e908c5624a67200f"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Blake3(tt.input)
			if got != tt.expected {
				t.Errorf("Blake3() = %q, want %q", got, tt.expected)
			}
		})
	}
}
