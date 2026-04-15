package queue

import "errors"

// ErrClosed is returned when operating on a closed queue.
var ErrClosed = errors.New("queue: closed")

// ErrFull is returned when the queue is full and cannot accept more messages.
var ErrFull = errors.New("queue: full")
