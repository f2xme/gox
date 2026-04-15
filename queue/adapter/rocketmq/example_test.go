package rocketmqadapter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/f2xme/gox/queue"
	rocketmqadapter "github.com/f2xme/gox/queue/adapter/rocketmq"
)

// Order 示例订单结构
type Order struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
	Status  string `json:"status"`
}

func Example() {
	ctx := context.Background()

	// 创建 RocketMQ 队列
	q, err := rocketmqadapter.New(
		rocketmqadapter.WithNameServers([]string{"localhost:9876"}),
		rocketmqadapter.WithGroupName("example-group"),
	)
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}

	// 发布简单消息
	err = q.Publish(ctx, "orders", []byte("hello"))
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}

	fmt.Println("Message published")
	// Output: Message published
}

func ExampleNew_withOptions() {
	// 创建带完整配置的 RocketMQ 队列
	q, err := rocketmqadapter.New(
		rocketmqadapter.WithNameServers([]string{"localhost:9876"}),
		rocketmqadapter.WithCredentials("access-key", "secret-key"),
		rocketmqadapter.WithNamespace("production"),
		rocketmqadapter.WithGroupName("order-service"),
		rocketmqadapter.WithRetries(3),
	)
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}
	_ = q

	fmt.Println("Queue created with options")
	// Output: Queue created with options
}

func ExampleQueue_PublishWithOptions() {
	ctx := context.Background()
	q, _ := rocketmqadapter.New(
		rocketmqadapter.WithNameServers([]string{"localhost:9876"}),
		rocketmqadapter.WithGroupName("example-group"),
	)

	// 发布带标签和延迟的消息
	order := Order{
		OrderID: "12345",
		Amount:  9900,
		Status:  "pending",
	}

	data, _ := json.Marshal(order)
	err := q.PublishWithOptions(ctx, "orders", data, queue.PublishOptions{
		Tags:       "urgent",
		DelayLevel: 3, // 延迟 10 秒
	})
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}

	fmt.Println("Order published with options")
	// Output: Order published with options
}

func ExampleQueue_Subscribe() {
	ctx := context.Background()
	q, _ := rocketmqadapter.New(
		rocketmqadapter.WithNameServers([]string{"localhost:9876"}),
		rocketmqadapter.WithGroupName("example-group"),
	)

	// 订阅消息
	handler := func(ctx context.Context, msg *queue.Message) error {
		var order Order
		if err := json.Unmarshal(msg.Body, &order); err != nil {
			return err
		}
		fmt.Printf("Processing order: %s\n", order.OrderID)
		return nil
	}

	sub, err := q.Subscribe(ctx, "orders", handler)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	fmt.Println("Subscribed to orders topic")
	// Output: Subscribed to orders topic
}
