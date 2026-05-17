package queue

import "errors"

// ErrClosed 表示操作已关闭的队列。
var ErrClosed = errors.New("queue: closed")

// ErrFull 表示队列已满，无法继续接收消息。
var ErrFull = errors.New("queue: full")
