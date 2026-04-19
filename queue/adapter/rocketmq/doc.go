/*
Package rocketmq 基于 Apache RocketMQ 5.x Go SDK 实现 `queue.Queue`，
提供面向 Proxy 的消息发布与订阅适配器。

# 功能特性

  - 使用 `WithEndpoint` 指定 RocketMQ 5.x Proxy 地址
  - 支持消息标签（Tags）过滤
  - 支持消息键（Keys）和自定义属性
  - 支持 RocketMQ 固定延迟等级消息（1-18 级）
  - 支持通过 `queue.AdvancedPublisher` 获取发送后的消息 ID
  - 支持 `NewWithConfig` / `MustNewWithConfig` 从配置中心创建实例

# 快速开始

	import (
		"context"

		"github.com/f2xme/gox/queue"
		"github.com/f2xme/gox/queue/adapter/rocketmq"
	)

	func example(ctx context.Context) error {
		q, err := rocketmq.New(
			rocketmq.WithEndpoint("localhost:8081"),
			rocketmq.WithNamespace("dev"),
		)
		if err != nil {
			return err
		}
		defer q.(queue.Closer).Close()

		if err := q.Publish(ctx, "orders", []byte(`{"order_id":"123"}`)); err != nil {
			return err
		}

		sub, err := q.SubscribeWithOptions(ctx, "orders", func(ctx context.Context, msg *queue.Message) error {
			return nil
		}, queue.SubscribeOptions{
			ConsumerGroup: "order-consumer",
			Tags:          "*",
		})
		if err != nil {
			return err
		}
		defer sub.Unsubscribe()

		return nil
	}

# 创建实例

使用 `New` 通过 Option 创建队列：

	q, err := rocketmq.New(
		rocketmq.WithEndpoint("localhost:8081"),
		rocketmq.WithCredentials("accessKey", "secretKey"),
		rocketmq.WithNamespace("production"),
		rocketmq.WithSendTimeout(5*time.Second),
	)

使用 `MustNew` 在初始化失败时直接终止程序：

	q := rocketmq.MustNew(
		rocketmq.WithEndpoint("localhost:8081"),
	)

使用 `NewWithConfig` 从 `config.Config` 读取配置：

	q, err := rocketmq.NewWithConfig(cfg)

支持的配置键：

  - `queue.rocketmq.endpoint`
  - `queue.rocketmq.accessKey`
  - `queue.rocketmq.secretKey`
  - `queue.rocketmq.namespace`
  - `queue.rocketmq.retries`
  - `queue.rocketmq.sendTimeout`
  - `queue.rocketmq.consumerModel`

# 发布消息

简单发布：

	err := q.Publish(ctx, "orders", []byte(`{"order_id":"123"}`))

带选项发布：

	err := q.PublishWithOptions(ctx, "orders", []byte(`{"order_id":"123"}`), queue.PublishOptions{
		Tags: "urgent",
		Keys: []string{"order-123"},
		Properties: map[string]string{
			"source": "web",
		},
		DelayLevel: 3,
	})

获取发送结果（消息 ID）：

	if ap, ok := q.(queue.AdvancedPublisher); ok {
		result, err := ap.PublishAndGetResult(ctx, "orders", body, queue.PublishOptions{})
		if err != nil {
			return err
		}
		_ = result.MessageID
	}

# 订阅消息

`Subscribe` 会使用默认消费者组 `DEFAULT_CONSUMER_GROUP`，更推荐在生产环境显式调用
`SubscribeWithOptions` 指定消费者组和过滤条件：

	sub, err := q.SubscribeWithOptions(ctx, "orders", handler, queue.SubscribeOptions{
		ConsumerGroup: "order-processor",
		Tags:          "urgent||normal",
	})

标签规则：

  - `"*"` 表示订阅全部消息
  - `"tagA||tagB"` 表示订阅多个标签
  - 空字符串会被自动视为 `"*"`

# 可用选项

## WithEndpoint

设置 RocketMQ 5.x Proxy 地址：

	rocketmq.New(rocketmq.WithEndpoint("localhost:8081"))

## WithCredentials

设置访问凭证：

	rocketmq.New(rocketmq.WithCredentials("accessKey", "secretKey"))

## WithNamespace

设置命名空间：

	rocketmq.New(rocketmq.WithNamespace("dev"))

## WithRetries

设置发送失败重试次数。该字段会被记录到适配器配置中，但 RocketMQ 5.x SDK
当前由底层自行处理重试策略：

	rocketmq.New(rocketmq.WithRetries(3))

## WithSendTimeout

设置发送超时：

	rocketmq.New(rocketmq.WithSendTimeout(5 * time.Second))

## WithConsumerModel

设置消费模式标识：

	rocketmq.New(rocketmq.WithConsumerModel(rocketmq.ConsumerModelClustering))

其中 `ConsumerModelBroadcasting` 目前仅作为兼容字段保留；RocketMQ 5.x Go SDK
及当前适配器实现并未真正切换到广播消费行为。

# 延迟消息

RocketMQ 支持固定延迟等级，适配器会将 `queue.PublishOptions.DelayLevel`
转换为绝对投递时间。支持的等级如下：

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

# 关闭资源

该适配器实现了 `queue.Closer`。调用 `Close` 会停止所有已注册订阅，
并优雅关闭底层 Producer：

	if closer, ok := q.(queue.Closer); ok {
		_ = closer.Close()
	}

# 注意事项

  - 当前接入点是 RocketMQ 5.x Proxy，不再使用旧版 NameServer 配置
  - `SubscribeWithOptions` 必须提供 `ConsumerGroup`
  - `Subscribe` 仅适合快速示例，生产环境应显式指定消费者组
  - `MaxConcurrency` 和 `AutoCommit` 目前未映射到底层 SDK 行为
  - 延迟消息仅支持固定的 1-18 级
*/
package rocketmq
