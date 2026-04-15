/*
Package queue 提供统一的消息队列抽象层。

queue 包定义了消息队列的标准接口，支持内存队列和分布式队列（RocketMQ、Redis、RabbitMQ、Kafka 等）。
通过统一的 API，你可以轻松地在不同的消息队列实现之间切换，而无需修改业务代码。

# 功能特性

  - 统一的消息队列抽象接口
  - 支持多种消息队列实现（内存、RocketMQ、Redis、RabbitMQ）
  - 类型安全的消息处理
  - 支持发布订阅模式
  - 支持消息标签过滤（RocketMQ）
  - 支持延迟消息（RocketMQ）
  - 支持消息属性和元数据
  - 线程安全

# 快速开始

基本使用：

	package main

	import (
		"context"
		"encoding/json"
		"log"

		"github.com/f2xme/gox/queue"
		"github.com/f2xme/gox/queue/adapter/mem"
	)

	func main() {
		ctx := context.Background()

		// 创建内存队列
		q := mem.New()
		defer q.(queue.Closer).Close()

		// 订阅主题
		sub, err := q.Subscribe(ctx, "orders", func(ctx context.Context, msg *queue.Message) error {
			var order Order
			json.Unmarshal(msg.Body, &order)

			// 处理订单
			if err := processOrder(order); err != nil {
				return err // 返回错误，消息会重新入队
			}

			return nil // 返回 nil，消息被确认
		})
		if err != nil {
			log.Fatal(err)
		}
		defer sub.Unsubscribe()

		// 发布消息
		data, _ := json.Marshal(Order{ID: "123", Amount: 9900})
		err = q.Publish(ctx, "orders", data)
		if err != nil {
			log.Fatal(err)
		}
	}

# 核心接口

Queue 接口结合了发布和订阅功能：

	type Queue interface {
		Publisher
		Subscriber
	}

Publisher 接口定义发布操作：

	type Publisher interface {
		Publish(ctx context.Context, topic string, body []byte) error
		PublishWithOptions(ctx context.Context, topic string, body []byte, opts PublishOptions) error
	}

Subscriber 接口定义订阅操作：

	type Subscriber interface {
		Subscribe(ctx context.Context, topic string, handler Handler) (Subscription, error)
		SubscribeWithOptions(ctx context.Context, topic string, handler Handler, opts SubscribeOptions) (Subscription, error)
	}

Handler 是消息处理函数：

	type Handler func(ctx context.Context, msg *Message) error

返回 nil 表示消息处理成功（ACK），返回 error 表示处理失败（NACK）。

# 可用适配器

内存队列：

	import "github.com/f2xme/gox/queue/adapter/mem"

	queue := mem.New()

特性：进程内通信、零外部依赖、高性能、适合单机应用

RocketMQ 队列：

	import "github.com/f2xme/gox/queue/adapter/rocketmq"

	queue, err := rocketmq.New(
		rocketmq.WithNameServers([]string{"localhost:9876"}),
		rocketmq.WithGroupName("my-producer-group"),
	)

特性：分布式消息队列、支持消息标签过滤、支持延迟消息、支持集群和广播消费、高可用和高吞吐量

# 使用示例

发布消息：

	// 发布简单消息
	err := queue.Publish(ctx, "orders", []byte(`{"order_id":"123"}`))

	// 发布结构化消息
	data, _ := json.Marshal(order)
	err := queue.Publish(ctx, "orders", data)

	// 发布带选项的消息（RocketMQ）
	err := queue.PublishWithOptions(ctx, "orders", data, queue.PublishOptions{
		Tags: "urgent",
		Keys: []string{"order-123"},
		Properties: map[string]string{
			"source": "web",
		},
		DelayLevel: 3, // 延迟 10 秒
	})

订阅消息：

	// 订阅主题
	sub, err := queue.Subscribe(ctx, "orders", func(ctx context.Context, msg *queue.Message) error {
		var order Order
		json.Unmarshal(msg.Body, &order)

		// 处理订单
		if err := processOrder(order); err != nil {
			return err // 返回错误，消息会重新入队
		}

		return nil // 返回 nil，消息被确认
	})

	// 取消订阅
	defer sub.Unsubscribe()

	// 带选项订阅（RocketMQ）
	sub, err := queue.SubscribeWithOptions(ctx, "orders", handler, queue.SubscribeOptions{
		ConsumerGroup:  "order-processor",
		Tags:           "urgent||normal", // 过滤标签
		MaxConcurrency: 10,               // 限制并发
	})

多个订阅者：

	// 订阅者 1：发送邮件
	queue.Subscribe(ctx, "orders", func(ctx context.Context, msg *queue.Message) error {
		return sendEmail(msg.Body)
	})

	// 订阅者 2：更新库存
	queue.Subscribe(ctx, "orders", func(ctx context.Context, msg *queue.Message) error {
		return updateInventory(msg.Body)
	})

优雅关闭：

	queue := adapter.New()
	defer func() {
		if closer, ok := queue.(queue.Closer); ok {
			closer.Close()
		}
	}()

# 消息处理模式

点对点（Point-to-Point）：

一个消息只被一个消费者处理：

	// 生产者
	queue.Publish(ctx, "tasks", taskData)

	// 消费者 1
	queue.Subscribe(ctx, "tasks", handler1)

	// 消费者 2
	queue.Subscribe(ctx, "tasks", handler2)

	// 每个消息只会被 handler1 或 handler2 处理一次

发布订阅（Pub/Sub）：

一个消息被所有订阅者处理：

	// 发布者
	queue.Publish(ctx, "events", eventData)

	// 订阅者 1
	queue.Subscribe(ctx, "events", handler1)

	// 订阅者 2
	queue.Subscribe(ctx, "events", handler2)

	// 每个消息会被 handler1 和 handler2 都处理

# 最佳实践

使用结构化消息：

	// 推荐：使用 JSON
	type OrderMessage struct {
		OrderID string `json:"order_id"`
		Amount  int64  `json:"amount"`
	}

	data, _ := json.Marshal(OrderMessage{OrderID: "123", Amount: 9900})
	queue.Publish(ctx, "orders", data)

实现幂等性：

	func handleOrder(ctx context.Context, msg *queue.Message) error {
		var order OrderMessage
		json.Unmarshal(msg.Body, &order)

		// 检查是否已处理
		if isProcessed(order.OrderID) {
			return nil // 已处理，直接返回
		}

		// 处理订单
		processOrder(order)

		// 标记为已处理
		markAsProcessed(order.OrderID)

		return nil
	}

错误处理和重试：

	func handleMessage(ctx context.Context, msg *queue.Message) error {
		// 可重试的错误：返回 error
		if err := processMessage(msg); err != nil {
			if isRetryable(err) {
				return err // 消息会重新入队
			}
			// 不可重试的错误：记录日志并返回 nil
			log.Printf("permanent error: %v", err)
			return nil // 消息被确认，不再重试
		}
		return nil
	}

使用 context 控制超时：

	func handleMessage(ctx context.Context, msg *queue.Message) error {
		// 设置处理超时
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		return processWithContext(ctx, msg)
	}

# 线程安全

所有队列实现都是线程安全的，可以在多个 goroutine 中并发使用。
*/
package queue
