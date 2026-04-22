package httpx

import (
	"encoding/json"
	"testing"
)

func TestNewDataResponse(t *testing.T) {
	resp := NewDataResponse(map[string]string{"name": "alice"})
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

func TestNewDataResponse_CustomMsg(t *testing.T) {
	resp := NewDataResponse(map[string]string{"id": "1"}, "创建成功")
	if resp.Message != "创建成功" {
		t.Errorf("expected Message '创建成功', got %q", resp.Message)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"success":true,"message":"创建成功","data":{"id":"1"}}`
	if string(data) != expected {
		t.Errorf("JSON mismatch\nexpected: %s\ngot:      %s", expected, string(data))
	}
}

func TestNewDataResponse_NilData(t *testing.T) {
	resp := NewDataResponse(nil)
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"success":true,"message":"ok"}`
	if string(data) != expected {
		t.Errorf("JSON mismatch\nexpected: %s\ngot:      %s", expected, string(data))
	}
}

func TestNewDoneResponse(t *testing.T) {
	resp := NewDoneResponse("删除成功")
	if !resp.Success {
		t.Error("expected Success to be true")
	}
	if resp.Message != "删除成功" {
		t.Errorf("expected Message '删除成功', got %q", resp.Message)
	}
	if resp.Data != nil {
		t.Errorf("expected Data to be nil, got %v", resp.Data)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"success":true,"message":"删除成功"}`
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

func TestNewFailResponse_WithData(t *testing.T) {
	fieldErrors := map[string]string{"name": "必填"}
	resp := NewFailResponse("表单校验失败", fieldErrors)
	if resp.Success {
		t.Error("expected Success to be false")
	}
	if resp.Data == nil {
		t.Error("expected Data to be set")
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"success":false,"message":"表单校验失败","data":{"name":"必填"}}`
	if string(data) != expected {
		t.Errorf("JSON mismatch\nexpected: %s\ngot:      %s", expected, string(data))
	}
}
