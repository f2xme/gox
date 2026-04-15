package httpx

import (
	"encoding/json"
	"testing"
)

func TestNewSuccessResponse(t *testing.T) {
	resp := NewSuccessResponse(map[string]string{"name": "alice"})
	if !resp.Success {
		t.Error("expected Success to be true")
	}
	if resp.Message != "ok" {
		t.Errorf("expected Message 'ok', got %q", resp.Message)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"success":true,"message":"ok","data":{"name":"alice"}}`
	if string(data) != expected {
		t.Errorf("JSON mismatch\nexpected: %s\ngot:      %s", expected, string(data))
	}
}

func TestNewFailResponse(t *testing.T) {
	resp := NewFailResponse("something went wrong")
	if resp.Success {
		t.Error("expected Success to be false")
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"success":false,"message":"something went wrong"}`
	if string(data) != expected {
		t.Errorf("JSON mismatch\nexpected: %s\ngot:      %s", expected, string(data))
	}
}

func TestNewSuccessResponse_NilData(t *testing.T) {
	resp := NewSuccessResponse(nil)
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"success":true,"message":"ok"}`
	if string(data) != expected {
		t.Errorf("JSON mismatch\nexpected: %s\ngot:      %s", expected, string(data))
	}
}
