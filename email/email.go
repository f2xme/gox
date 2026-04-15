package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

// ContentType 邮件内容类型
type ContentType string

const (
	ContentTypePlain ContentType = "text/plain"
	ContentTypeHTML  ContentType = "text/html"
)

// SendText 发送纯文本邮件
func (c *Client) SendText(to, subject, content string) error {
	return c.Send(to, subject, ContentTypePlain, content)
}

// SendHTML 发送 HTML 邮件
func (c *Client) SendHTML(to, subject, content string) error {
	return c.Send(to, subject, ContentTypeHTML, content)
}

// Send 发送邮件
func (c *Client) Send(to, subject string, contentType ContentType, content string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", msg.FormatAddress(c.dialer.Username, c.name))
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody(string(contentType), content)

	if err := c.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
