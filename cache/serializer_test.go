package cache

import (
	"testing"
)

// testStruct 是测试用的结构体
type testStruct struct {
	Name  string
	Age   int
	Email string
}

// TestJSONSerializer 测试 JSON 序列化器
// 验证 JSON 序列化和反序列化的正确性
func TestJSONSerializer(t *testing.T) {
	serializer := JSONSerializer
	original := testStruct{
		Name:  "Alice",
		Age:   30,
		Email: "alice@example.com",
	}

	// Marshal
	data, err := serializer.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded testStruct
	err = serializer.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, original.Name)
	}
	if decoded.Age != original.Age {
		t.Errorf("Age mismatch: got %d, want %d", decoded.Age, original.Age)
	}
	if decoded.Email != original.Email {
		t.Errorf("Email mismatch: got %s, want %s", decoded.Email, original.Email)
	}
}

// TestGobSerializer 测试 Gob 序列化器
// 验证 Gob 序列化和反序列化的正确性
func TestGobSerializer(t *testing.T) {
	serializer := GobSerializer
	original := testStruct{
		Name:  "Bob",
		Age:   25,
		Email: "bob@example.com",
	}

	// Marshal
	data, err := serializer.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded testStruct
	err = serializer.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, original.Name)
	}
	if decoded.Age != original.Age {
		t.Errorf("Age mismatch: got %d, want %d", decoded.Age, original.Age)
	}
	if decoded.Email != original.Email {
		t.Errorf("Email mismatch: got %s, want %s", decoded.Email, original.Email)
	}
}

// TestSerializerRoundTrip 测试序列化器的往返转换
// 验证不同序列化器处理复杂数据类型的能力
func TestSerializerRoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		serializer Serializer
	}{
		{"JSON", JSONSerializer},
		{"Gob", GobSerializer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := map[string]interface{}{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			}

			// Marshal
			data, err := tt.serializer.Marshal(original)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// Unmarshal
			var decoded map[string]interface{}
			err = tt.serializer.Unmarshal(data, &decoded)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// Verify string value
			if decoded["key1"] != "value1" {
				t.Errorf("key1 mismatch: got %v, want value1", decoded["key1"])
			}

			// Verify numeric value (JSON may decode as float64)
			switch v := decoded["key2"].(type) {
			case int:
				if v != 42 {
					t.Errorf("key2 mismatch: got %d, want 42", v)
				}
			case float64:
				if v != 42.0 {
					t.Errorf("key2 mismatch: got %f, want 42.0", v)
				}
			default:
				t.Errorf("key2 unexpected type: %T", v)
			}

			// Verify boolean value
			if decoded["key3"] != true {
				t.Errorf("key3 mismatch: got %v, want true", decoded["key3"])
			}
		})
	}
}
