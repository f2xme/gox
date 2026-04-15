/*
Package queue 提供统一的消息队列抽象层。

# 概述

queue 包定义了消息队列的标准接口，支持内存队列和分布式队列（Redis、RabbitMQ、Kafka 等）。
通过统一的 API，你可以轻松地在不同的消息队列实现之间切换，而无需修改业务代码。

# 核心接口

## Queue - 消息队列接口

结合了发布和订阅功能：

	type Queue interface {
		Publisher
		Subscriber
	}

## Publisher - 发布者接口

	type Publisher interface {
		Publish(ctx context.Context, topic string, body []byte) error
	}

## Subscriber - 订阅者接口

	type Subscriber interface {
		Subscribe(ctx context.Context, topic string, handler Handler) (Subscription, error)
	}

## Handler - 消息处理函数

	type Handler func(ctx context.Context, msg *Message) error

返回 nil 表示消息处理成功（ACK），返回 error 表示处理失败（NACK）。

# 使用示例

## 发布消息

	// 发布简单消息
	err := queue.Publish(ctx, "orders", []byte(`{"order_id":"123"}`))

	// 发布结构化消息
	data, _ := json.Marshal(order)
	err := queue.Publish(ctx, "orders", data)

## 订阅消息

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

## 多个订阅者

	// 订阅者 1：发送邮件
	queue.Subscribe(ctx, "orders", func(ctx context.Context, msg *queue.Message) error {
		return sendEmail(msg.Body)
	})

	// 订阅者 2：更新库存
	queue.Subscribe(ctx, "orders", func(ctx context.Context, msg *queue.Message) error {
		return updateInventory(msg.Body)
	})

## 优雅关闭

	queue := adapter.New()
	defer func() {
		if closer, ok := queue.(queue.Closer); ok {
			closer.Close()
		}
	}()

# 可用适配器

## 内存队列

	import "github.com/f2xme/gox/queue/adapter/mem"

	queue := mem.New()

特性：
  - 进程内通信
  - 零外部依赖
  - 高性能
  - 适合单机应用

## Redis 队列

	import (
		"github.com/f2xme/gox/queue/adapter/redisadapter"
		"github.com/redis/go-redis/v9"
	)

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	queue := redisadapter.New(
		redisadapter.WithClient(rdb),
	)

特性：
  - 分布式队列
  - 持久化
  - 支持多消费者
  - 适合分布式应用

## RabbitMQ 队列

	import "github.com/f2xme/gox/queue/adapter/rabbitmqadapter"

	queue := rabbitmqadapter.New(
		rabbitmqadapter.WithURL("amqp://guest:guest@localhost:5672/"),
	)

特性：
  - 企业级消息队列
  - 支持多种消息模式
  - 高可用
  - 适合复杂的消息路由

# 消息处理模式

## 点对点（Point-to-Point）

一个消息只被一个消费者处理：

	// 生产者
	queue.Publish(ctx, "tasks", taskData)

	// 消费者 1
	queue.Subscribe(ctx, "tasks", handler1)

	// 消费者 2
	queue.Subscribe(ctx, "tasks", handler2)

	// 每个消息只会被 handler1 或 handler2 处理一次

## 发布订阅（Pub/Sub）

一个消息被所有订阅者处理：

	// 发布者
	queue.Publish(ctx, "events", eventData)

	// 订阅者 1
	queue.Subscribe(ctx, "events", handler1)

	// 订阅者 2
	queue.Subscribe(ctx, "events", handler2)

	// 每个消息会被 handler1 和 handler2 都处理

# 最佳实践

## 1. 使用结构化消息

	// 推荐：使用 JSON
	type OrderMessage struct {
		OrderID string `json:"order_id"`
		Amount  int64  `json:"amount"`
	}

	data, _ := json.Marshal(OrderMessage{OrderID: "123", Amount: 9900})
	queue.Publish(ctx, "orders", data)

## 2. 实现幂等性

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

## 3. 错误处理和重试

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

## 4. 使用 context 控制超时

	func handleMessage(ctx context.Context, msg *queue.Message) error {
		// 设置处理超时
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		return processWithContext(ctx, msg)
	}

## 5. 优雅关闭

	// 等待所有消息处理完成
	sub, _ := queue.Subscribe(ctx, "tasks", handler)

	// 收到关闭信号
	<-shutdownChan

	// 停止接收新消息
	sub.Unsubscribe()

	// 等待正在处理的消息完成
	time.Sleep(5 * time.Second)

## 6. 监控和日志

	func handleMessage(ctx context.Context, msg *queue.Message) error {
		start := time.Now()
		defer func() {
			log.Printf("processed message %s in %v", msg.ID, time.Since(start))
		}()

		return processMessage(msg)
	}

# 消息可靠性

## 消息确认（ACK）

	// 处理成功：返回 nil
	return nil

	// 处理失败：返回 error
	return errors.New("processing failed")

## 死信队列（DLQ）

对于多次重试失败的消息，可以发送到死信队列：

	if retryCount > maxRetries {
		queue.Publish(ctx, "dlq", msg.Body)
		return nil // 确认原消息
	}

# 性能考虑

  - 内存队列：纳秒级延迟，适合高频通信
  - Redis 队列：毫秒级延迟，适合分布式场景
  - RabbitMQ：毫秒级延迟，适合复杂路由
  - 批量发布可以提高吞吐量

# 线程安全

所有队列实现都应该是线程安全的，可以在多个 goroutine 中并发使用。
*/
package queue
