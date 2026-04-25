# Cache 包增强设计文档

**日期**: 2026-04-25  
**状态**: 待审阅  
**作者**: Claude Code

## 概述

本设计文档描述了对 `github.com/f2xme/gox/cache` 包的功能增强,主要包括:
1. 高级缓存操作 (TTL 查询、SetNX/SetXX、GetSet、Expire)
2. 批量存在性检查 (ExistsMulti)
3. 键扫描功能 (Scan)
4. 分布式锁增强 (自动续期、可重入锁、元数据查询)
5. Typed 包装器改进 (防缓存击穿、批量操作)
6. 退避策略优化 (decorrelated jitter)

## 设计原则

- **向后兼容**: 所有现有接口和行为保持不变
- **接口隔离**: 新功能通过独立接口提供,用户按需使用
- **渐进式增强**: 通过类型断言使用新功能,不强制所有实现都支持
- **自动降级**: Typed 包装器在底层不支持批量操作时自动降级为循环调用

## 架构设计

### 接口层次结构

```
Cache (基础接口)
├── Advanced (高级操作)
├── MultiCache (批量操作)
│   └── MultiCacheV2 (扩展批量操作)
├── Scanner (键扫描)
├── Locker (基础锁)
│   └── LockerV2 (增强锁)
├── LockMetadata (锁元数据)
├── Counter (计数器)
└── Closer (资源管理)
```

所有接口都是可选的,实现者可以根据能力选择性实现。

## 详细设计

### 1. Advanced 接口 - 高级缓存操作

```go
// Advanced 提供高级缓存操作
type Advanced interface {
    // TTL 返回键的剩余过期时间
    // 如果键不存在返回 ErrNotFound
    // 如果键没有设置过期时间返回 0, nil
    TTL(ctx context.Context, key string) (time.Duration, error)
    
    // SetNX 仅当键不存在时设置值
    // 返回 true 表示设置成功, false 表示键已存在
    SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
    
    // SetXX 仅当键存在时更新值
    // 返回 true 表示更新成功, false 表示键不存在
    SetXX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
    
    // GetSet 原子性地获取旧值并设置新值
    // 如果键不存在返回 nil, ErrNotFound
    // ttl 为 0 表示不改变原有过期时间 (如果有), -1 表示移除过期时间
    GetSet(ctx context.Context, key string, value []byte, ttl time.Duration) ([]byte, error)
    
    // Expire 更新键的过期时间而不修改值
    // 如果键不存在返回 ErrNotFound
    Expire(ctx context.Context, key string, ttl time.Duration) error
}
```

**使用场景**:
- TTL: 监控缓存健康状态,决定是否需要刷新
- SetNX: 实现分布式锁、防止重复提交
- SetXX: 仅更新已存在的缓存,避免意外创建
- GetSet: 原子性地更新配置并获取旧值
- Expire: 延长热点数据的生命周期

**实现要点**:
- Redis: 直接使用对应的 Redis 命令
- mem: 使用互斥锁保证原子性
- GetSet 的 ttl 参数语义:
  - `ttl > 0`: 设置新的过期时间
  - `ttl == 0`: 保持原有过期时间不变
  - `ttl == -1`: 移除过期时间,键永不过期

### 2. Scanner 接口 - 键扫描

```go
// Scanner 提供键遍历功能
type Scanner interface {
    // Scan 使用游标迭代匹配 pattern 的键
    // pattern 支持 glob 模式: * 匹配任意字符, ? 匹配单个字符, [abc] 匹配字符集
    // cursor 为 0 表示开始新的迭代, 返回的 cursor 为 0 表示迭代结束
    // count 是每次迭代的建议返回数量 (实际可能更多或更少)
    Scan(ctx context.Context, pattern string, cursor uint64, count int64) (keys []string, nextCursor uint64, err error)
}
```

**使用场景**:
- 调试: 查看某个前缀下的所有缓存键
- 批量清理: 删除匹配模式的所有键
- 统计分析: 统计不同类型缓存的数量

**实现要点**:
- Redis: 直接使用 SCAN 命令
- mem: 
  - 收集所有未过期的键并排序 (保证迭代顺序一致)
  - 使用 `filepath.Match` 进行 glob 匹配
  - 通过 cursor 实现分页

**注意事项**:
- Scan 是阻塞操作,大量键时可能影响性能
- 建议在低峰期或只读副本上执行
- 不推荐在生产环境频繁使用

### 3. MultiCacheV2 接口 - 扩展批量操作

```go
// MultiCacheV2 扩展 MultiCache 接口
type MultiCacheV2 interface {
    MultiCache
    
    // ExistsMulti 批量检查键是否存在
    // 返回 map 中每个键对应其存在状态
    ExistsMulti(ctx context.Context, keys []string) (map[string]bool, error)
}
```

**使用场景**:
- 批量预检: 在批量获取前先检查哪些键存在
- 缓存命中率统计: 批量检查缓存是否命中

**实现要点**:
- Redis: 使用 Pipeline 批量执行 EXISTS 命令
- mem: 加读锁后遍历检查

### 4. LockerV2 接口 - 增强的分布式锁

```go
// LockerV2 扩展 Locker 接口,提供自动续期和可重入锁
type LockerV2 interface {
    Locker
    
    // LockWithRenewal 获取带自动续期的锁
    // renewInterval 指定续期间隔,通常设置为 ttl 的 1/3 到 1/2
    // 返回的 unlock 函数会自动停止续期并释放锁
    LockWithRenewal(ctx context.Context, key string, ttl, renewInterval time.Duration) (unlock func() error, err error)
    
    // TryLockWithRenewal 非阻塞版本的 LockWithRenewal
    TryLockWithRenewal(ctx context.Context, key string, ttl, renewInterval time.Duration) (unlock func() error, err error)
    
    // LockReentrant 获取可重入锁
    // ownerID 是调用者的唯一标识 (如 requestID, traceID 等)
    // 同一个 ownerID 可以多次获取同一把锁,每次获取会增加重入计数
    // unlock 时会减少计数,计数归零时才真正释放锁
    LockReentrant(ctx context.Context, key string, ownerID string, ttl time.Duration) (unlock func() error, err error)
    
    // TryLockReentrant 非阻塞版本的 LockReentrant
    TryLockReentrant(ctx context.Context, key string, ownerID string, ttl time.Duration) (unlock func() error, err error)
}

// LockInfo 包含锁的元数据
type LockInfo struct {
    Owner      string        // 锁持有者的标识
    AcquiredAt time.Time     // 锁获取时间
    TTL        time.Duration // 锁的剩余有效期
    Reentrant  bool          // 是否为可重入锁
    Count      int           // 重入计数 (仅可重入锁有效)
}

// LockMetadata 提供锁元数据查询功能
type LockMetadata interface {
    // GetLockInfo 查询锁的当前状态
    // 如果锁不存在或已过期返回 ErrNotFound
    GetLockInfo(ctx context.Context, key string) (LockInfo, error)
}
```

**使用场景**:
- LockWithRenewal: 长时间运行的任务 (数据导入、报表生成等)
- LockReentrant: 递归调用、嵌套事务、分布式请求链路
- LockMetadata: 监控锁的持有情况,排查死锁问题

**实现要点**:

#### 自动续期
- 获取锁成功后启动后台 goroutine
- 使用 ticker 定期执行 EXPIRE 命令刷新 TTL
- unlock 时关闭 goroutine 并释放锁
- 续期间隔建议为 TTL 的 1/3 到 1/2

#### 可重入锁
- Redis 实现:
  - 使用 Hash 存储锁信息: `{owner: "xxx", count: 2}`
  - Lua 脚本保证原子性:
    ```lua
    -- 获取锁
    local current = redis.call("HGET", KEYS[1], "owner")
    if current == false then
        redis.call("HSET", KEYS[1], "owner", ARGV[1], "count", 1)
        redis.call("PEXPIRE", KEYS[1], ARGV[2])
        return 1
    elseif current == ARGV[1] then
        redis.call("HINCRBY", KEYS[1], "count", 1)
        redis.call("PEXPIRE", KEYS[1], ARGV[2])
        return 1
    else
        return 0
    end
    
    -- 释放锁
    local current = redis.call("HGET", KEYS[1], "owner")
    if current ~= ARGV[1] then
        return 0
    end
    local count = redis.call("HINCRBY", KEYS[1], "count", -1)
    if count <= 0 then
        redis.call("DEL", KEYS[1])
    end
    return 1
    ```

- mem 实现:
  - 扩展 lockEntry 结构体,添加 owner 和 count 字段
  - 获取锁时检查 owner,相同则增加计数
  - 释放锁时减少计数,归零时清除锁

#### 锁元数据
- Redis: 使用 HGETALL 获取 Hash 中的所有字段
- mem: 直接读取 lockEntry 结构体

### 5. 退避策略优化 - Decorrelated Jitter

当前 Lock 方法使用固定的指数退避,在高并发下容易产生惊群效应。改用 decorrelated jitter 算法:

```go
func (r *redisCache) Lock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
    const (
        initialBackoff = 10 * time.Millisecond
        maxBackoff     = 100 * time.Millisecond
    )
    
    sleep := initialBackoff
    for {
        unlock, err := r.TryLock(ctx, key, ttl)
        if err == nil {
            return unlock, nil
        }
        if err != cache.ErrLockFailed {
            return nil, err
        }
        
        // Decorrelated jitter: sleep = min(max, random(base, sleep * 3))
        sleep = time.Duration(rand.Int63n(int64(sleep*3-initialBackoff))) + initialBackoff
        if sleep > maxBackoff {
            sleep = maxBackoff
        }
        
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(sleep):
        }
    }
}
```

**优势**:
- 避免惊群效应: 随机化的退避时间分散了重试请求
- 更快的平均获取时间: 相比固定指数退避,decorrelated jitter 在高并发下表现更好
- 业界最佳实践: AWS 推荐的算法

### 6. Typed 包装器增强

#### 6.1 防止缓存击穿 - singleflight

使用 `golang.org/x/sync/singleflight` 确保同一 key 的并发请求只执行一次加载函数:

```go
import "golang.org/x/sync/singleflight"

type Typed[T any] struct {
    cache      Cache
    serializer Serializer
    group      singleflight.Group // 新增
}

func (t *Typed[T]) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (T, error)) (T, error) {
    var zero T
    
    // 先尝试从缓存获取
    value, err := t.Get(ctx, key)
    if err == nil {
        return value, nil
    }
    if err != ErrNotFound {
        return zero, err
    }
    
    // 使用 singleflight 确保同一 key 只有一个 goroutine 执行 fn
    result, err, _ := t.group.Do(key, func() (interface{}, error) {
        // 再次检查缓存 (可能已被其他 goroutine 填充)
        if value, err := t.Get(ctx, key); err == nil {
            return value, nil
        }
        
        // 执行加载函数
        value, err := fn()
        if err != nil {
            return zero, err
        }
        
        // 存入缓存
        _ = t.Set(ctx, key, value, ttl)
        return value, nil
    })
    
    if err != nil {
        return zero, err
    }
    return result.(T), nil
}
```

**效果**:
- 100 个并发请求同时访问同一个不存在的 key,只会执行 1 次 fn()
- 进程级别的去重,适用于 99% 的缓存击穿场景
- 如需跨进程防击穿,用户可以在调用前自己使用 Locker 加锁

#### 6.2 批量操作 - 自动降级

```go
func (t *Typed[T]) GetMulti(ctx context.Context, keys []string) (map[string]T, error) {
    if mc, ok := t.cache.(MultiCache); ok {
        // 使用批量操作
        dataMap, err := mc.GetMulti(ctx, keys)
        if err != nil {
            return nil, err
        }
        
        result := make(map[string]T, len(dataMap))
        for key, data := range dataMap {
            var value T
            if err := t.serializer.Unmarshal(data, &value); err != nil {
                return nil, fmt.Errorf("unmarshal key %s: %w", key, err)
            }
            result[key] = value
        }
        return result, nil
    }
    
    // 降级为循环调用
    result := make(map[string]T)
    for _, key := range keys {
        value, err := t.Get(ctx, key)
        if err == ErrNotFound {
            continue // 跳过不存在的键
        }
        if err != nil {
            return nil, err
        }
        result[key] = value
    }
    return result, nil
}

func (t *Typed[T]) SetMulti(ctx context.Context, items map[string]T, ttl time.Duration) error {
    if mc, ok := t.cache.(MultiCache); ok {
        // 序列化所有值
        dataMap := make(map[string][]byte, len(items))
        for key, value := range items {
            data, err := t.serializer.Marshal(value)
            if err != nil {
                return fmt.Errorf("marshal key %s: %w", key, err)
            }
            dataMap[key] = data
        }
        return mc.SetMulti(ctx, dataMap, ttl)
    }
    
    // 降级为循环调用
    for key, value := range items {
        if err := t.Set(ctx, key, value, ttl); err != nil {
            return err
        }
    }
    return nil
}

func (t *Typed[T]) DeleteMulti(ctx context.Context, keys []string) error {
    if mc, ok := t.cache.(MultiCache); ok {
        return mc.DeleteMulti(ctx, keys)
    }
    
    // 降级为循环调用
    for _, key := range keys {
        if err := t.Delete(ctx, key); err != nil {
            return err
        }
    }
    return nil
}
```

**优势**:
- 用户无需关心底层实现,API 总是可用
- 从 mem 切换到 redis 时,批量操作自动获得性能提升
- 符合抽象层的职责,屏蔽底层差异

### 7. 错误处理

新增错误类型:

```go
var (
    ErrNotFound    = errors.New("cache: key not found")
    ErrLockFailed  = errors.New("cache: failed to acquire lock")
    ErrNoTTL       = errors.New("cache: key has no expiration")  // 新增
    ErrLockExpired = errors.New("cache: lock has expired")       // 新增
)
```

**使用场景**:
- ErrNoTTL: TTL 查询时键永不过期 (已废弃,改为返回 0, nil)
- ErrLockExpired: unlock 时锁已过期或被其他人持有

## 实现计划

### Phase 1: 核心功能 (优先级: 高)

1. **Advanced 接口**
   - 在 `cache.go` 中定义接口
   - redis 适配器实现 (使用 Redis 命令和 Lua 脚本)
   - mem 适配器实现 (使用互斥锁)
   - 单元测试

2. **MultiCacheV2.ExistsMulti**
   - 扩展 MultiCache 接口
   - redis 实现 (Pipeline)
   - mem 实现 (遍历检查)
   - 单元测试

3. **Typed 批量操作和 singleflight**
   - 添加 `golang.org/x/sync/singleflight` 依赖
   - 实现 GetMulti/SetMulti/DeleteMulti
   - 改进 GetOrSet 使用 singleflight
   - 并发测试

### Phase 2: 锁增强 (优先级: 高)

4. **LockerV2.LockWithRenewal**
   - 定义接口
   - redis 实现 (后台 goroutine 续期)
   - mem 实现 (后台 goroutine 续期)
   - 单元测试和集成测试

5. **LockerV2.LockReentrant**
   - 定义接口
   - redis 实现 (Lua 脚本 + Hash)
   - mem 实现 (扩展 lockEntry)
   - 单元测试和并发测试

6. **LockMetadata 接口**
   - 定义接口和 LockInfo 结构体
   - redis 实现 (HGETALL)
   - mem 实现 (读取 lockEntry)
   - 单元测试

7. **退避策略优化**
   - 修改 redis 和 mem 的 Lock 方法
   - 使用 decorrelated jitter 算法
   - 性能测试对比

### Phase 3: 扫描和完善 (优先级: 中)

8. **Scanner 接口**
   - 定义接口
   - redis 实现 (SCAN 命令)
   - mem 实现 (glob 匹配 + 游标分页)
   - 单元测试

9. **测试完善**
   - 集成测试 (使用 testcontainers-go 启动 Redis)
   - 并发测试 (race detector)
   - 性能测试 (benchmark)

10. **文档更新**
    - 更新 doc.go,添加新接口的使用示例
    - 添加"高级功能"章节
    - 添加"最佳实践"章节
    - 更新 README.md

## 测试策略

### 单元测试
- 每个新方法都有对应的测试用例
- 测试正常流程和边界情况 (key 不存在、已过期、并发等)
- 使用 table-driven tests
- 目标覆盖率: 80%+

### 集成测试
- redis 适配器需要真实 Redis 实例
- 使用 `testcontainers-go` 启动临时 Redis 容器
- 测试与真实 Redis 的交互

### 并发测试
- 测试 GetOrSet 的 singleflight 机制
- 测试可重入锁的并发场景
- 测试自动续期的竞态条件
- 使用 `go test -race` 检测竞态

### 性能测试
- Benchmark 批量操作 vs 循环调用
- Benchmark JSON vs Gob 序列化
- Benchmark 锁的获取和释放
- Benchmark decorrelated jitter vs 固定退避

## 向后兼容性

### 接口兼容
- 所有现有接口 (Cache, MultiCache, Locker, Counter, Closer) 保持不变
- 新接口通过类型断言使用,不影响现有代码
- 现有的 mem 和 redis 适配器继续工作

### 行为兼容
- 现有方法的行为不变
- 新方法的错误返回遵循现有约定
- TTL 语义保持一致 (0 表示永不过期)

### 迁移路径

```go
// 旧代码无需修改
cache := mem.New()
cache.Set(ctx, "key", []byte("value"), time.Minute)

// 新功能通过类型断言使用
if adv, ok := cache.(cache.Advanced); ok {
    ttl, _ := adv.TTL(ctx, "key")
    fmt.Println("Remaining TTL:", ttl)
}

// Typed 包装器自动支持批量操作
typed := cache.NewTyped[User](cache)
users, _ := typed.GetMulti(ctx, []string{"user:1", "user:2"})
```

## 文档更新

### doc.go 更新内容

1. **新增"高级功能"章节**
   - Advanced 接口的使用示例
   - Scanner 接口的使用示例
   - LockerV2 接口的使用示例

2. **新增"性能优化"章节**
   - 批量操作的性能提升数据
   - singleflight 的效果说明
   - 何时使用批量操作

3. **新增"最佳实践"章节**
   - 何时使用可重入锁 vs 普通锁
   - 如何选择续期间隔 (建议 TTL 的 1/3 到 1/2)
   - Scan 的正确使用方式 (避免阻塞生产环境)
   - 如何选择 ownerID (requestID, traceID 等)

4. **更新示例代码**
   - 添加完整的使用示例
   - 添加错误处理示例
   - 添加性能对比示例

## 依赖变更

新增依赖:
```go
require (
    golang.org/x/sync v0.6.0 // singleflight
)
```

## 风险和缓解

### 风险 1: singleflight 可能导致请求阻塞
- **描述**: 如果 fn() 执行时间过长,其他 goroutine 会一直等待
- **缓解**: 
  - 文档中明确说明 fn() 应该快速返回
  - 建议用户在 fn() 中使用 context 控制超时
  - 考虑添加 GetOrSetWithTimeout 方法

### 风险 2: 自动续期可能导致 goroutine 泄漏
- **描述**: 如果用户忘记调用 unlock,续期 goroutine 会一直运行
- **缓解**:
  - 文档中强调必须使用 defer unlock()
  - 续期 goroutine 在锁过期后自动退出
  - 添加日志记录异常情况

### 风险 3: 可重入锁的 ownerID 冲突
- **描述**: 不同请求使用相同的 ownerID 会导致锁被错误释放
- **缓解**:
  - 文档中建议使用全局唯一的 ID (UUID, traceID 等)
  - 提供示例代码展示正确用法
  - 考虑添加 ownerID 格式验证

### 风险 4: Scan 操作可能阻塞生产环境
- **描述**: 大量键时 Scan 可能影响性能
- **缓解**:
  - 文档中明确警告不推荐在生产环境频繁使用
  - 建议在低峰期或只读副本上执行
  - mem 实现使用读锁,减少阻塞

## 成功标准

1. **功能完整性**
   - 所有列出的功能都已实现
   - redis 和 mem 适配器都支持新接口
   - Typed 包装器支持批量操作

2. **测试覆盖**
   - 单元测试覆盖率 ≥ 80%
   - 所有并发场景都有测试
   - 集成测试通过

3. **性能指标**
   - 批量操作性能提升 ≥ 50% (相比循环调用)
   - singleflight 减少 ≥ 90% 的重复加载
   - decorrelated jitter 减少 ≥ 30% 的平均获取时间

4. **文档质量**
   - doc.go 包含所有新功能的示例
   - 最佳实践章节清晰易懂
   - 错误处理示例完整

5. **向后兼容**
   - 现有代码无需修改即可运行
   - 所有现有测试通过
   - 无破坏性变更

## 未来扩展

以下功能不在本次设计范围内,但可以在未来考虑:

1. **分布式事务**: 支持多 key 的原子操作
2. **Pub/Sub**: 支持发布订阅模式
3. **Stream**: 支持消息流
4. **Sorted Set**: 支持有序集合
5. **Geo**: 支持地理位置操作
6. **指标监控**: 内置 metrics (命中率、延迟等)
7. **日志接口**: 可注入自定义 logger

## 总结

本设计通过渐进式扩展的方式,在保持向后兼容的前提下,为 cache 包添加了丰富的高级功能。核心设计原则是接口隔离和自动降级,确保用户可以按需使用新功能,同时不影响现有代码。

实现分为三个阶段,优先实现核心功能和锁增强,最后完善扫描和文档。所有新功能都有完整的测试覆盖和文档说明。

通过本次增强,cache 包将成为一个功能完整、性能优异、易于使用的缓存抽象层,满足从简单到复杂的各种使用场景。
