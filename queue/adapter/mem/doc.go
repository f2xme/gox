// Package memadapter provides an in-memory queue implementation.
//
// This package implements the queue.Queue interface with an in-memory pub/sub backend
// using Go channels. It supports multiple topics and concurrent subscribers with
// configurable buffer sizes.
//
// Basic usage:
//
//	q := memadapter.New()
//	defer q.Close()
//
//	ctx := context.Background()
//
//	// Subscribe to a topic
//	sub, err := q.Subscribe(ctx, "events", func(ctx context.Context, msg *queue.Message) error {
//		fmt.Printf("Received: %s\n", msg.Body)
//		return nil
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer sub.Unsubscribe()
//
//	// Publish a message
//	err = q.Publish(ctx, "events", []byte("hello"))
//
// Configuration options:
//
//	// Using Option functions
//	q := memadapter.New(
//		memadapter.WithBufferSize(128), // Set channel buffer size
//	)
//
//	// Or using Options struct directly
//	q := memadapter.New(
//		func(o *memadapter.Options) {
//			o.BufferSize = 128
//		},
//	)
//
// The queue uses Go channels internally for message delivery. Each subscription
// gets its own buffered channel. Messages are delivered to all active subscribers
// of a topic concurrently.
//
// Thread Safety:
//
// All operations are thread-safe and can be called concurrently from multiple
// goroutines. The queue uses read-write locks to protect internal state.
//
// Resource Management:
//
// Always call Close() when done to properly clean up all subscriptions and
// release resources. Subscriptions should also call Unsubscribe() when no
// longer needed.
package memadapter
