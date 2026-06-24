package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/f2xme/gox/sms"
)

// Client 是基于内存的短信服务实现。
type Client struct {
	mu      sync.RWMutex
	options Options
	records []SentMessage
}

var _ sms.SMS = (*Client)(nil)

// SentMessage 定义一条内存发送记录。
type SentMessage struct {
	// Message 发送的短信消息。
	Message sms.Message
	// SentAt 记录发送成功的时间。
	SentAt time.Time
}

// Send 发送短信并记录消息。
func (c *Client) Send(ctx context.Context, message sms.Message) error {
	if ctx == nil {
		return fmt.Errorf("memory sms: context cannot be nil")
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("memory sms: context error: %w", err)
	}
	if err := validateMessage(message); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.options.SendError != nil {
		return c.options.SendError
	}

	c.records = append(c.records, SentMessage{
		Message: cloneMessage(message),
		SentAt:  time.Now(),
	})
	return nil
}

// Messages 返回所有已发送短信记录的副本。
func (c *Client) Messages() []SentMessage {
	c.mu.RLock()
	defer c.mu.RUnlock()

	records := make([]SentMessage, len(c.records))
	for i, record := range c.records {
		records[i] = cloneSentMessage(record)
	}
	return records
}

// LastMessage 返回最后一条已发送短信记录。
func (c *Client) LastMessage() (SentMessage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.records) == 0 {
		return SentMessage{}, false
	}
	return cloneSentMessage(c.records[len(c.records)-1]), true
}

// Count 返回已发送短信记录数量。
func (c *Client) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.records)
}

// Reset 清空所有已发送短信记录。
func (c *Client) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = c.records[:0]
}

// SetSendError 设置发送短信时固定返回的错误。
func (c *Client) SetSendError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.options.SendError = err
}

func validateMessage(message sms.Message) error {
	if strings.TrimSpace(message.Phone) == "" {
		return fmt.Errorf("memory sms: phone is required")
	}
	if strings.TrimSpace(message.TemplateCode) == "" {
		return fmt.Errorf("memory sms: template code is required")
	}
	return nil
}

func cloneSentMessage(record SentMessage) SentMessage {
	return SentMessage{
		Message: cloneMessage(record.Message),
		SentAt:  record.SentAt,
	}
}

func cloneMessage(message sms.Message) sms.Message {
	return sms.Message{
		Phone:         message.Phone,
		TemplateCode:  message.TemplateCode,
		TemplateParam: cloneTemplateParam(message.TemplateParam),
	}
}

func cloneTemplateParam(param any) any {
	switch v := param.(type) {
	case nil:
		return nil
	case map[string]string:
		cloned := make(map[string]string, len(v))
		for key, value := range v {
			cloned[key] = value
		}
		return cloned
	case map[string]any:
		cloned := make(map[string]any, len(v))
		for key, value := range v {
			cloned[key] = cloneTemplateParam(value)
		}
		return cloned
	case []string:
		cloned := make([]string, len(v))
		copy(cloned, v)
		return cloned
	case []any:
		cloned := make([]any, len(v))
		for i, value := range v {
			cloned[i] = cloneTemplateParam(value)
		}
		return cloned
	case []byte:
		cloned := make([]byte, len(v))
		copy(cloned, v)
		return cloned
	default:
		return v
	}
}
