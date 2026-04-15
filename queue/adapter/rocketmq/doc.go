/*
Package rocketmqadapter 提供 RocketMQ 消息队列适配器。

# 概述

rocketmqadapter 是基于 Apache RocketMQ Go 客户端的队列实现，支持分布式消息队列的发布和订阅。

# 特性

  - 分布式消息队列
  - 支持消息标签（Tags）过滤
  - 支持消息键（Keys）索引
  - 支持延迟消息
  - 支持集群和广播消费模式
  - 高可用和高吞吐量

# 基本使用

## 创建队列

	import "github.com/f2xme/gox/queue/adapter/rocketmq"

	q, err := rocketmq.New(
		rocketmq.WithNameServers([]string{"localhost:9876"}),
		rocketmq.WithGroupName("my-producer-group"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer q.(queue.Closer).Close()

## 发布消息

	// 简单发布
	err := q.Publish(ctx, "orders", []byte(`{"order_id":"123"}`))

	// 带选项发布
	err := q.PublishWithOptions(ctx, "orders", []byte(`{"order_id":"123"}`), queue.PublishOptions{
		Tags: "urgent",
		Keys: []string{"order-123"},
		Properties: map[string]string{
			"source": "web",
		},
	})

## 订阅消息

	// 简单订阅
	sub, err := q.Subscribe(ctx, "orders", func(ctx context.Context, msg *queue.Message) error {
		log.Printf("Received: %s", msg.Body)
		return nil
	})

	// 带选项订阅
	sub, err := q.SubscribeWithOptions(ctx, "orders", handler, queue.SubscribeOptions{
		ConsumerGroup: "order-processor",
		Tags:          "urgent||normal",
		MaxConcurrency: 10,
	})

	defer sub.Unsubscribe()

# 配置选项

## WithNameServers

设置 RocketMQ Name Server 地址：

	rocketmq.New(rocketmq.WithNameServers([]string{"localhost:9876"}))

## WithCredentials

设置访问凭证：

	rocketmq.New(rocketmq.WithCredentials("accessKey", "secretKey"))

## WithNamespace

设置命名空间：

	rocketmq.New(rocketmq.WithNamespace("dev"))

## WithGroupName

设置生产者组名：

	rocketmq.New(rocketmq.WithGroupName("my-producer-group"))

## WithRetries

设置重试次数：

	rocketmq.New(rocketmq.WithRetries(3))

## WithSendTimeout

设置发送超时：

	rocketmq.New(rocketmq.WithSendTimeout(5 * time.Second))

## WithConsumerModel

设置消费模式（clustering 或 broadcasting）：

	rocketmq.New(rocketmq.WithConsumerModel("broadcasting"))

# 消息标签过滤

RocketMQ 支持使用标签过滤消息：

	// 发布带标签的消息
	q.PublishWithOptions(ctx, "orders", data, queue.PublishOptions{
		Tags: "urgent",
	})

	// 订阅特定标签
	q.SubscribeWithOptions(ctx, "orders", handler, queue.SubscribeOptions{
		ConsumerGroup: "processor",
		Tags:          "urgent||normal", // 订阅 urgent 或 normal 标签
	})

	// 订阅所有标签
	q.SubscribeWithOptions(ctx, "orders", handler, queue.SubscribeOptions{
		ConsumerGroup: "processor",
		Tags:          "*", // 订阅所有消息
	})

# 延迟消息

RocketMQ 支持延迟消息（18 个延迟级别）：

	q.PublishWithOptions(ctx, "tasks", data, queue.PublishOptions{
		DelayLevel: 3, // 延迟级别 3 = 10 秒
	})

延迟级别对应的时间：
  - 1: 1s
  - 2: 5s
  - 3: 10s
  - 4: 30s
  - 5: 1m
  - 6: 2m
  - 7: 3m
  - 8: 4m
  - 9: 5m
  - 10: 6m
  - 11: 7m
  - 12: 8m
  - 13: 9m
  - 14: 10m
  - 15: 20m
  - 16: 30m
  - 17: 1h
  - 18: 2h

# 消费模式

## 集群消费（Clustering）

默认模式，同一消费组内的消费者负载均衡消费消息：

	rocketmq.New(rocketmq.WithConsumerModel("clustering"))

## 广播消费（Broadcasting）

每个消费者都会收到所有消息：

	rocketmq.New(rocketmq.WithConsumerModel("broadcasting"))

# 最佳实践

## 1. 使用消费者组

	q.SubscribeWithOptions(ctx, "orders", handler, queue.SubscribeOptions{
		ConsumerGroup: "order-processor", // 必须指定消费者组
	})

## 2. 设置合理的并发度

	q.SubscribeWithOptions(ctx, "orders", handler, queue.SubscribeOptions{
		ConsumerGroup:  "processor",
		MaxConcurrency: 10, // 限制并发处理数
	})

## 3. 使用消息键进行索引

	q.PublishWithOptions(ctx, "orders", data, queue.PublishOptions{
		Keys: []string{"order-123"}, // 便于查询和追踪
	})

## 4. 优雅关闭

	q, err := rocketmq.New(...)
	defer func() {
		if closer, ok := q.(queue.Closer); ok {
			closer.Close()
		}
	}()

# 注意事项

  - 消费者组名（ConsumerGroup）是必需的
  - Name Server 地址必须正确配置
  - 生产环境建议配置多个 Name Server 地址
  - 延迟消息只支持固定的 18 个延迟级别
  - 广播模式下消息不会重试
*/
package rocketmqadapter
