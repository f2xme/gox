# Queue - 消息队列抽象层

统一的消息队列接口，支持内存队列和分布式队列（RocketMQ、Redis、RabbitMQ 等）。

## 特性

- 统一的 API 接口
- 支持多种消息队列实现
- 类型安全的消息处理
- 支持消息标签和属性
- 支持延迟消息
- 优雅的错误处理

## 安装

```bash
go get github.com/f2xme/gox/queue
```

## 快速开始

### 内存队列

```go
import "github.com/f2xme/gox/queue/adapter/mem"

queue := mem.New()
defer queue.(queue.Closer).Close()

// 发布消息
err := queue.Publish(ctx, "orders", []byte(`{"order_id":"123"}`))

// 订阅消息
sub, err := queue.Subscribe(ctx, "orders", func(ctx context.Context, msg *queue.Message) error {
    log.Printf("Received: %s", msg.Body)
    return nil
})
defer sub.Unsubscribe()
```

### RocketMQ 队列

```go
import "github.com/f2xme/gox/queue/adapter/rocketmq"

queue, err := rocketmq.New(
    rocketmq.WithNameServers([]string{"localhost:9876"}),
    rocketmq.WithGroupName("my-producer-group"),
)
if err != nil {
    log.Fatal(err)
}
defer queue.(queue.Closer).Close()

// 发布带标签的消息
err = queue.PublishWithOptions(ctx, "orders", data, queue.PublishOptions{
    Tags: "urgent",
    Keys: []string{"order-123"},
})

// 订阅特定标签
sub, err := queue.SubscribeWithOptions(ctx, "orders", handler, queue.SubscribeOptions{
    ConsumerGroup: "order-processor",
    Tags:          "urgent||normal",
})
```

## 可用适配器

| 适配器 | 包路径 | 适用场景 |
|--------|--------|----------|
| 内存队列 | `github.com/f2xme/gox/queue/adapter/mem` | 单机应用、测试 |
| RocketMQ | `github.com/f2xme/gox/queue/adapter/rocketmq` | 大规模分布式系统 |

## 核心接口

### Publisher

```go
type Publisher interface {
    Publish(ctx context.Context, topic string, body []byte) error
    PublishWithOptions(ctx context.Context, topic string, body []byte, opts PublishOptions) error
}
```

### Subscriber

```go
type Subscriber interface {
    Subscribe(ctx context.Context, topic string, handler Handler) (Subscription, error)
    SubscribeWithOptions(ctx context.Context, topic string, handler Handler, opts SubscribeOptions) (Subscription, error)
}
```

### Handler

```go
type Handler func(ctx context.Context, msg *Message) error
```

返回 `nil` 表示消息处理成功（ACK），返回 `error` 表示处理失败（NACK）。

## 高级特性

### 消息标签过滤（RocketMQ）

```go
// 发布带标签的消息
queue.PublishWithOptions(ctx, "orders", data, queue.PublishOptions{
    Tags: "urgent",
})

// 订阅特定标签
queue.SubscribeWithOptions(ctx, "orders", handler, queue.SubscribeOptions{
    ConsumerGroup: "processor",
    Tags:          "urgent||normal", // 订阅 urgent 或 normal
})
```

### 延迟消息（RocketMQ）

```go
queue.PublishWithOptions(ctx, "tasks", data, queue.PublishOptions{
    DelayLevel: 3, // 延迟 10 秒
})
```

### 消息属性

```go
queue.PublishWithOptions(ctx, "orders", data, queue.PublishOptions{
    Properties: map[string]string{
        "source": "web",
        "version": "v1",
    },
})
```

## 最佳实践

### 1. 实现幂等性

```go
func handleOrder(ctx context.Context, msg *queue.Message) error {
    var order Order
    json.Unmarshal(msg.Body, &order)

    // 检查是否已处理
    if isProcessed(order.ID) {
        return nil
    }

    // 处理订单
    if err := processOrder(order); err != nil {
        return err
    }

    // 标记为已处理
    markAsProcessed(order.ID)
    return nil
}
```

### 2. 错误处理和重试

```go
func handleMessage(ctx context.Context, msg *queue.Message) error {
    if err := processMessage(msg); err != nil {
        if isRetryable(err) {
            return err // 消息会重新入队
        }
        // 不可重试的错误：记录日志并返回 nil
        log.Printf("permanent error: %v", err)
        return nil
    }
    return nil
}
```

### 3. 使用 context 控制超时

```go
func handleMessage(ctx context.Context, msg *queue.Message) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    return processWithContext(ctx, msg)
}
```

### 4. 优雅关闭

```go
queue := adapter.New()
defer func() {
    if closer, ok := queue.(queue.Closer); ok {
        closer.Close()
    }
}()
```

## 文档

详细文档请参考：
- [队列包文档](https://pkg.go.dev/github.com/f2xme/gox/queue)
- [内存适配器文档](https://pkg.go.dev/github.com/f2xme/gox/queue/adapter/mem)
- [RocketMQ 适配器文档](https://pkg.go.dev/github.com/f2xme/gox/queue/adapter/rocketmq)

## 许可证

MIT License
