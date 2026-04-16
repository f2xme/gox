package serializer

import (
	"bytes"
	"testing"
)

type testStruct struct {
	Name string `json:"name" xml:"name" msgpack:"name"`
	Age  int    `json:"age" xml:"age" msgpack:"age"`
}

func TestJSONSerializer(t *testing.T) {
	s := NewJSON()
	testSerializer(t, s, "JSON")
}

func TestXMLSerializer(t *testing.T) {
	s := NewXML()
	testSerializer(t, s, "XML")
}

func TestMsgPackSerializer(t *testing.T) {
	s := NewMsgPack()
	testSerializer(t, s, "MsgPack")
}

func testSerializer(t *testing.T, s Serializer, name string) {
	t.Run(name+"_Marshal", func(t *testing.T) {
		obj := testStruct{Name: "Alice", Age: 30}
		data, err := s.Marshal(obj)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("Marshal returned empty data")
		}

		var result testStruct
		err = s.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if result.Name != obj.Name || result.Age != obj.Age {
			t.Fatalf("Unmarshal mismatch: got %+v, want %+v", result, obj)
		}
	})

	t.Run(name+"_Encode", func(t *testing.T) {
		obj := testStruct{Name: "Bob", Age: 25}
		var buf bytes.Buffer
		err := s.Encode(&buf, obj)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		var result testStruct
		err = s.Decode(&buf, &result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if result.Name != obj.Name || result.Age != obj.Age {
			t.Fatalf("Decode mismatch: got %+v, want %+v", result, obj)
		}
	})

	t.Run(name+"_ContentType", func(t *testing.T) {
		ct := s.ContentType()
		if ct == "" {
			t.Fatal("ContentType returned empty string")
		}
	})
}
