package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/f2xme/gox/queue"
	"github.com/f2xme/gox/queue/adapter/mem"
)

type OrderMessage struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

func main() {
	fmt.Println("=== queue 包使用示例 ===")

	// 创建内存队列
	q := mem.New()
	ctx := context.Background()

	// 示例 1: 发布消息
	fmt.Println("\n示例 1: 发布消息")
	order := OrderMessage{
		OrderID: "ORDER-001",
		Amount:  9900,
	}
	data, err := json.Marshal(order)
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return
	}

	err = q.Publish(ctx, "orders", data)
	if err != nil {
		fmt.Printf("发布失败: %v\n", err)
	} else {
		fmt.Printf("✓ 已发布订单消息: %s\n", order.OrderID)
	}

	// 示例 2: 订阅消息
	fmt.Println("\n示例 2: 订阅消息")
	sub, err := q.Subscribe(ctx, "orders", func(ctx context.Context, msg *queue.Message) error {
		var order OrderMessage
		json.Unmarshal(msg.Body, &order)
		fmt.Printf("  → 收到订单: %s, 金额: %.2f 元\n", order.OrderID, float64(order.Amount)/100)
		return nil
	})
	if err != nil {
		fmt.Printf("订阅失败: %v\n", err)
		return
	}
	defer sub.Unsubscribe()

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 示例 3: 发布多条消息
	fmt.Println("\n示例 3: 发布多条消息")
	orders := []OrderMessage{
		{OrderID: "ORDER-002", Amount: 15000},
		{OrderID: "ORDER-003", Amount: 8800},
		{OrderID: "ORDER-004", Amount: 12500},
	}

	for _, order := range orders {
		data, err := json.Marshal(order)
		if err != nil {
			fmt.Printf("序列化失败: %v\n", err)
			continue
		}
		if err := q.Publish(ctx, "orders", data); err != nil {
			fmt.Printf("发布失败: %v\n", err)
			continue
		}
		fmt.Printf("✓ 已发布: %s\n", order.OrderID)
	}

	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n=== 示例结束 ===")
}
