package memory

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/f2xme/gox/sms"
)

func TestClientImplementsSMS(t *testing.T) {
	client, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var _ sms.SMS = client
}

func TestClientSendRecordsMessage(t *testing.T) {
	client := MustNew()
	param := map[string]string{"code": "1234"}

	if err := client.Send(context.Background(), sms.Message{
		Phone:         "13800138000",
		TemplateCode:  "login_code",
		TemplateParam: param,
	}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	param["code"] = "changed"

	if client.Count() != 1 {
		t.Fatalf("Count() = %d, want 1", client.Count())
	}

	last, ok := client.LastMessage()
	if !ok {
		t.Fatal("LastMessage() ok = false, want true")
	}
	if last.Message.Phone != "13800138000" || last.Message.TemplateCode != "login_code" {
		t.Fatalf("LastMessage() = %+v", last.Message)
	}

	gotParam, ok := last.Message.TemplateParam.(map[string]string)
	if !ok {
		t.Fatalf("TemplateParam type = %T, want map[string]string", last.Message.TemplateParam)
	}
	if gotParam["code"] != "1234" {
		t.Fatalf("TemplateParam[code] = %q, want %q", gotParam["code"], "1234")
	}
	if last.SentAt.IsZero() {
		t.Fatal("SentAt is zero")
	}
}

func TestClientMessagesReturnsCopy(t *testing.T) {
	client := MustNew()
	if err := client.Send(context.Background(), sms.Message{
		Phone:         "13800138000",
		TemplateCode:  "login_code",
		TemplateParam: []string{"1234"},
	}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	records := client.Messages()
	records[0].Message.Phone = "changed"
	records[0].Message.TemplateParam.([]string)[0] = "changed"

	records = client.Messages()
	if records[0].Message.Phone != "13800138000" {
		t.Fatalf("Phone = %q, want original value", records[0].Message.Phone)
	}
	if got := records[0].Message.TemplateParam.([]string)[0]; got != "1234" {
		t.Fatalf("TemplateParam[0] = %q, want %q", got, "1234")
	}
}

func TestClientReset(t *testing.T) {
	client := MustNew()
	if err := client.Send(context.Background(), sms.Message{
		Phone:        "13800138000",
		TemplateCode: "login_code",
	}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	client.Reset()

	if client.Count() != 0 {
		t.Fatalf("Count() after Reset = %d, want 0", client.Count())
	}
	if _, ok := client.LastMessage(); ok {
		t.Fatal("LastMessage() ok = true, want false")
	}
}

func TestClientSendValidation(t *testing.T) {
	client := MustNew()
	tests := []struct {
		name    string
		message sms.Message
	}{
		{
			name: "missing phone",
			message: sms.Message{
				TemplateCode: "login_code",
			},
		},
		{
			name: "missing template code",
			message: sms.Message{
				Phone: "13800138000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := client.Send(context.Background(), tt.message); err == nil {
				t.Fatal("Send() error = nil, want error")
			}
		})
	}
}

func TestClientSendContext(t *testing.T) {
	client := MustNew()

	if err := client.Send(nil, sms.Message{}); err == nil {
		t.Fatal("Send(nil) error = nil, want error")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := client.Send(ctx, sms.Message{
		Phone:        "13800138000",
		TemplateCode: "login_code",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Send() error = %v, want context.Canceled", err)
	}
}

func TestClientSendError(t *testing.T) {
	sendErr := errors.New("send failed")
	client := MustNew(WithSendError(sendErr))

	err := client.Send(context.Background(), sms.Message{
		Phone:        "13800138000",
		TemplateCode: "login_code",
	})
	if !errors.Is(err, sendErr) {
		t.Fatalf("Send() error = %v, want %v", err, sendErr)
	}
	if client.Count() != 0 {
		t.Fatalf("Count() = %d, want 0", client.Count())
	}

	client.SetSendError(nil)
	if err := client.Send(context.Background(), sms.Message{
		Phone:        "13800138000",
		TemplateCode: "login_code",
	}); err != nil {
		t.Fatalf("Send() after clearing error = %v", err)
	}
	if client.Count() != 1 {
		t.Fatalf("Count() after clearing error = %d, want 1", client.Count())
	}
}

func TestCloneTemplateParam(t *testing.T) {
	param := map[string]any{
		"codes": []any{"1234", map[string]string{"nested": "value"}},
		"raw":   []byte("hello"),
	}

	cloned := cloneTemplateParam(param).(map[string]any)
	param["codes"].([]any)[1].(map[string]string)["nested"] = "changed"
	param["raw"].([]byte)[0] = 'z'

	if got := cloned["codes"].([]any)[1].(map[string]string)["nested"]; got != "value" {
		t.Fatalf("nested value = %q, want %q", got, "value")
	}
	if got := string(cloned["raw"].([]byte)); got != "hello" {
		t.Fatalf("raw = %q, want %q", got, "hello")
	}

	if got := cloneTemplateParam("value"); !reflect.DeepEqual(got, "value") {
		t.Fatalf("cloneTemplateParam() = %#v, want value", got)
	}
}
