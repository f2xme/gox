# Captcha 包重新设计

## 概述

重新设计 gox/captcha 包，采用适配器模式支持多种存储后端，抽象生成器接口支持自定义验证码类型，提供清晰的 API 和灵活的配置选项。

## 设计目标

1. **独立的包设计**：captcha 作为独立包，不与其他 gox 包耦合
2. **支持多个适配器**：内存适配器和基于 gox/cache 的适配器
3. **可扩展的生成器**：抽象 Generator 接口，支持自定义验证码生成逻辑
4. **清晰的职责分离**：存储、生成、验证各自独立
5. **灵活的配置**：支持快捷方式和完全自定义两种使用方式

## 架构设计

### 目录结构

```
captcha/
├── captcha.go              # 核心接口和实现
├── store.go                # Store 接口定义
├── options.go              # 配置选项
├── errors.go               # 错误定义
├── doc.go                  # 包文档
├── adapter/                # 适配器实现
│   ├── memory/             # 内存适配器
│   │   ├── memory.go
│   │   ├── memory_test.go
│   │   ├── options.go
│   │   └── doc.go
│   └── cache/              # 基于 gox/cache 的适配器
│       ├── cache.go
│       ├── cache_test.go
│       ├── options.go
│       └── doc.go
└── generator/              # 生成器接口和实现
    ├── generator.go        # Generator 接口定义
    └── base64/             # 基于 base64Captcha 的实现
        ├── base64.go
        ├── base64_test.go
        ├── options.go
        └── doc.go
```

### 核心接口

#### Store 接口

Store 接口定义验证码存储的通用操作，所有适配器必须实现此接口。

```go
package captcha

// Store 定义验证码存储接口
type Store interface {
    // Set 存储验证码答案
    // ttl 为 0 表示使用适配器的默认过期时间
    Set(ctx context.Context, id string, answer string, ttl time.Duration) error
    
    // Get 获取验证码答案
    // 如果验证码不存在或已过期，返回 ErrNotFound
    Get(ctx context.Context, id string) (string, error)
    
    // Delete 删除验证码
    // 如果验证码不存在不返回错误
    Delete(ctx context.Context, id string) error
    
    // Exists 检查验证码是否存在且未过期
    Exists(ctx context.Context, id string) (bool, error)
}
```

**设计要点：**
- 所有方法支持 context，便于超时控制和链路追踪
- Get 返回 ErrNotFound 而不是空字符串，明确区分"不存在"和"空答案"
- Delete 幂等，多次删除不报错
- TTL 可以在调用时指定，也可以使用适配器默认值（传 0）

#### Generator 接口

Generator 接口定义验证码生成器，负责生成验证码内容和答案。

```go
package generator

// Generator 定义验证码生成器接口
type Generator interface {
    // Generate 生成验证码内容和答案
    // data: base64 编码的验证码数据（图片或音频）
    // answer: 验证码答案（用于验证用户输入）
    Generate() (data string, answer string, err error)
    
    // Type 返回生成器类型标识
    // 用于日志、监控和调试
    Type() string
}
```

**设计要点：**
- Generate 返回 base64 编码的数据，统一格式
- answer 是明文答案，由 Store 存储
- Type() 用于标识生成器类型，便于日志和监控

#### Captcha 接口

Captcha 接口是对外暴露的主要接口，提供完整的验证码生命周期管理。

```go
package captcha

// Captcha 定义验证码服务接口
type Captcha interface {
    // Generate 生成验证码
    // 返回验证码 ID 和 base64 编码的数据
    Generate(ctx context.Context) (id string, data string, err error)
    
    // Verify 验证验证码答案
    // 验证成功返回 true，失败返回 false
    // 不会自动删除验证码，需要手动调用 Delete
    Verify(ctx context.Context, id string, answer string) (bool, error)
    
    // Delete 删除验证码
    // 通常在验证成功后调用，防止重复使用
    Delete(ctx context.Context, id string) error
    
    // Regenerate 重新生成验证码内容（保持相同 ID）
    // 用于"看不清，换一张"的场景
    Regenerate(ctx context.Context, id string) (data string, err error)
}
```

**设计要点：**
- Generate 自动生成随机 ID
- Verify 不自动删除，由用户控制生命周期
- Delete 显式调用，清晰明确
- Regenerate 支持"换一张"功能，保持 ID 不变

### 核心实现

#### Captcha 实现

```go
type captcha struct {
    store     Store
    generator Generator
    opts      Options
}

type Options struct {
    TTL       time.Duration // 默认过期时间，默认 5 分钟
    IDLength  int           // ID 长度，默认 20
    Generator Generator     // 自定义生成器
}

// New 创建 Captcha 实例（通用构造函数）
func New(store Store, opts ...Option) Captcha

// NewWithMemory 创建使用内存存储的 Captcha（便捷函数）
func NewWithMemory(opts ...Option) Captcha

// NewWithCache 创建使用 cache 存储的 Captcha（便捷函数）
func NewWithCache(c cache.Cache, opts ...Option) Captcha
```

**实现细节：**

1. **Generate 实现：**
   - 生成随机 ID（使用 crypto/rand）
   - 调用 generator.Generate() 获取数据和答案
   - 调用 store.Set() 存储答案
   - 返回 ID 和数据

2. **Verify 实现：**
   - 调用 store.Get() 获取答案
   - 比较答案（忽略大小写和前后空格）
   - 返回比较结果

3. **Delete 实现：**
   - 直接调用 store.Delete()

4. **Regenerate 实现：**
   - 调用 generator.Generate() 生成新内容
   - 调用 store.Set() 更新答案（保持 ID 不变）
   - 返回新数据

#### 配置选项

```go
// Captcha 层选项
func WithTTL(ttl time.Duration) Option
func WithIDLength(length int) Option
func WithGenerator(gen Generator) Option

// 便捷选项（用于 NewWithMemory 和 NewWithCache）
// 这些选项会创建默认的 base64 生成器
func WithType(t generator.CaptchaType) Option
func WithSize(width, height int) Option
func WithLength(length int) Option
func WithNoiseCount(count int) Option
func WithLanguage(lang string) Option
```

### 适配器实现

#### 内存适配器 (adapter/memory)

内存适配器提供轻量级的单机存储方案，适合开发和小规模部署。

```go
package memory

type memoryStore struct {
    mu    sync.RWMutex
    items map[string]*item
    opts  Options
    stopCh chan struct{}
}

type item struct {
    answer     string
    expiration int64 // Unix 纳秒，0 表示无过期
}

type Options struct {
    TTL             time.Duration // 默认过期时间，默认 5 分钟
    CleanupInterval time.Duration // 清理间隔，默认 1 分钟
    MaxSize         int           // 最大条目数，0 表示无限制
}

// New 创建内存适配器
func New(opts ...Option) captcha.Store

// 配置选项
func WithTTL(ttl time.Duration) Option
func WithCleanupInterval(interval time.Duration) Option
func WithMaxSize(size int) Option
```

**实现细节：**

1. **过期清理：**
   - 启动后台 goroutine 定期清理过期条目
   - Get/Exists 时惰性检查过期
   - Close() 停止清理 goroutine

2. **容量限制：**
   - 当达到 MaxSize 时，使用 LRU 策略淘汰最旧的条目
   - 淘汰前先清理过期条目

3. **并发安全：**
   - 使用 sync.RWMutex 保护 map
   - 读操作使用 RLock，写操作使用 Lock

#### Cache 适配器 (adapter/cache)

Cache 适配器基于 gox/cache 包，支持分布式场景。

```go
package cache

type cacheStore struct {
    cache cache.Cache
    opts  Options
}

type Options struct {
    TTL    time.Duration // 默认过期时间，默认 5 分钟
    Prefix string        // key 前缀，默认 "captcha:"
}

// New 创建 cache 适配器
func New(c cache.Cache, opts ...Option) captcha.Store

// 配置选项
func WithTTL(ttl time.Duration) Option
func WithPrefix(prefix string) Option
```

**实现细节：**

1. **Key 管理：**
   - 所有 key 添加前缀（默认 "captcha:"）
   - 避免与其他缓存数据冲突

2. **过期管理：**
   - 依赖 cache 包的 TTL 机制
   - 不需要自己实现清理逻辑

3. **错误处理：**
   - cache.ErrNotFound 转换为 captcha.ErrNotFound
   - 其他错误直接返回

### 生成器实现

#### Base64 生成器 (generator/base64)

Base64 生成器基于 base64Captcha 库，支持多种验证码类型。

```go
package base64

type base64Generator struct {
    driver base64Captcha.Driver
    opts   Options
}

type Options struct {
    Type       CaptchaType // 验证码类型，默认 TypeDigit
    Width      int         // 宽度，默认 240
    Height     int         // 高度，默认 80
    Length     int         // 长度，默认 4
    NoiseCount int         // 噪点数量，默认 1
    Language   string      // 音频语言，默认 "en"
}

type CaptchaType int

const (
    TypeDigit  CaptchaType = iota // 数字验证码
    TypeString                     // 字母数字混合
    TypeMath                       // 算术表达式
    TypeAudio                      // 音频验证码
)

// New 创建 base64 生成器
func New(opts ...Option) generator.Generator

// 配置选项
func WithType(t CaptchaType) Option
func WithSize(width, height int) Option
func WithLength(length int) Option
func WithNoiseCount(count int) Option
func WithLanguage(lang string) Option
```

**实现细节：**

1. **Driver 创建：**
   - 根据 CaptchaType 创建对应的 base64Captcha.Driver
   - TypeDigit → DriverDigit
   - TypeString → DriverString
   - TypeMath → DriverMath
   - TypeAudio → DriverAudio

2. **Generate 实现：**
   - 调用 driver.GenerateIdQuestionAnswer()
   - 返回 base64 数据和答案
   - 忽略 base64Captcha 生成的 ID（我们自己生成）

3. **Type() 实现：**
   - 返回 "base64"

## 错误处理

```go
package captcha

var (
    // ErrNotFound 验证码不存在或已过期
    ErrNotFound = errors.New("captcha: not found")
    
    // ErrGenerateFailed 生成验证码失败
    ErrGenerateFailed = errors.New("captcha: generate failed")
    
    // ErrInvalidID 无效的验证码 ID
    ErrInvalidID = errors.New("captcha: invalid id")
)
```

**错误处理原则：**
- Store.Get 返回 ErrNotFound 表示不存在或过期
- Generator.Generate 返回具体错误，Captcha 层包装为 ErrGenerateFailed
- Verify 返回 (false, nil) 表示答案错误，返回 (false, err) 表示系统错误

## 使用示例

### 基础使用（内存）

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/f2xme/gox/captcha"
)

func main() {
    // 最简单的方式，使用默认配置
    c := captcha.NewWithMemory()
    
    ctx := context.Background()
    
    // 生成验证码
    id, data, err := c.Generate(ctx)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("验证码 ID: %s\n", id)
    fmt.Printf("验证码图片: %s\n", data) // data:image/png;base64,...
    
    // 验证用户输入
    ok, err := c.Verify(ctx, id, "1234")
    if err != nil {
        panic(err)
    }
    
    if ok {
        fmt.Println("验证成功")
        // 验证成功后删除，防止重复使用
        c.Delete(ctx, id)
    } else {
        fmt.Println("验证失败")
    }
}
```

### 自定义配置

```go
// 自定义验证码类型和参数
c := captcha.NewWithMemory(
    captcha.WithTTL(10*time.Minute),
    captcha.WithType(generator.TypeMath),
    captcha.WithSize(300, 100),
    captcha.WithLength(6),
)
```

### 使用 Cache 适配器（分布式）

```go
import (
    "github.com/f2xme/gox/cache/adapter/mem"
    "github.com/f2xme/gox/captcha"
)

// 使用 gox/cache 的内存实现
cache := mem.New()
c := captcha.NewWithCache(cache,
    captcha.WithTTL(10*time.Minute),
    captcha.WithType(generator.TypeString),
)

// 或者使用 Redis
import "github.com/f2xme/gox/cache/adapter/redis"

redisCache := redis.New(redisClient)
c := captcha.NewWithCache(redisCache)
```

### 完全自定义

```go
import (
    "github.com/f2xme/gox/captcha"
    "github.com/f2xme/gox/captcha/adapter/memory"
    "github.com/f2xme/gox/captcha/generator/base64"
)

// 自定义存储
store := memory.New(
    memory.WithTTL(15*time.Minute),
    memory.WithMaxSize(10000),
)

// 自定义生成器
gen := base64.New(
    base64.WithType(base64.TypeDigit),
    base64.WithLength(6),
)

// 组合
c := captcha.New(store,
    captcha.WithGenerator(gen),
    captcha.WithIDLength(32),
)
```

### Web 集成示例

```go
// 生成验证码接口
http.HandleFunc("/api/captcha/generate", func(w http.ResponseWriter, r *http.Request) {
    id, data, err := c.Generate(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]string{
        "id":   id,
        "data": data,
    })
})

// 验证验证码接口
http.HandleFunc("/api/captcha/verify", func(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ID     string `json:"id"`
        Answer string `json:"answer"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    ok, err := c.Verify(r.Context(), req.ID, req.Answer)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    if ok {
        // 验证成功，删除验证码
        c.Delete(r.Context(), req.ID)
        json.NewEncoder(w).Encode(map[string]bool{"success": true})
    } else {
        json.NewEncoder(w).Encode(map[string]bool{"success": false})
    }
})

// 刷新验证码接口
http.HandleFunc("/api/captcha/regenerate", func(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ID string `json:"id"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    data, err := c.Regenerate(r.Context(), req.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]string{
        "data": data,
    })
})
```

## 扩展性

### 自定义存储适配器

实现 Store 接口即可：

```go
package myadapter

import (
    "context"
    "time"
    
    "github.com/f2xme/gox/captcha"
)

type myStore struct {
    // 自定义字段
}

func New() captcha.Store {
    return &myStore{}
}

func (s *myStore) Set(ctx context.Context, id string, answer string, ttl time.Duration) error {
    // 实现存储逻辑
}

func (s *myStore) Get(ctx context.Context, id string) (string, error) {
    // 实现获取逻辑
}

func (s *myStore) Delete(ctx context.Context, id string) error {
    // 实现删除逻辑
}

func (s *myStore) Exists(ctx context.Context, id string) (bool, error) {
    // 实现检查逻辑
}
```

### 自定义生成器

实现 Generator 接口即可：

```go
package mygen

import "github.com/f2xme/gox/captcha/generator"

type myGenerator struct {
    // 自定义字段
}

func New() generator.Generator {
    return &myGenerator{}
}

func (g *myGenerator) Generate() (string, string, error) {
    // 生成验证码数据和答案
    // 返回 base64 编码的数据
    data := "data:image/svg+xml;base64,..."
    answer := "ABC123"
    return data, answer, nil
}

func (g *myGenerator) Type() string {
    return "custom"
}
```

## 迁移指南

### 从旧版本迁移

旧版本代码：

```go
store := captcha.NewMemoryStore(1000, 5*time.Minute)
c := captcha.New(store, captcha.WithType(captcha.TypeDigit))

id, b64s, err := c.Generate()
if c.Verify(id, "1234") {
    // 验证成功（自动删除）
}
```

新版本代码：

```go
c := captcha.NewWithMemory(
    captcha.WithTTL(5*time.Minute),
    captcha.WithType(generator.TypeDigit),
)

ctx := context.Background()
id, b64s, err := c.Generate(ctx)
ok, err := c.Verify(ctx, id, "1234")
if ok {
    c.Delete(ctx, id) // 需要手动删除
}
```

**主要变化：**
1. 所有方法需要传入 context
2. Verify 不再自动删除，需要手动调用 Delete
3. 生成器类型从 captcha.TypeDigit 改为 generator.TypeDigit
4. NewMemoryStore 改为 NewWithMemory

## 测试策略

### 单元测试

1. **Store 接口测试：**
   - 测试 Set/Get/Delete/Exists 基本操作
   - 测试过期机制
   - 测试并发安全

2. **Generator 接口测试：**
   - 测试各种类型的验证码生成
   - 测试生成的数据格式
   - 测试答案的正确性

3. **Captcha 核心测试：**
   - 测试 Generate/Verify/Delete/Regenerate
   - 测试错误处理
   - 测试配置选项

### 集成测试

1. **内存适配器集成测试：**
   - 测试完整的生成-验证-删除流程
   - 测试过期清理
   - 测试容量限制

2. **Cache 适配器集成测试：**
   - 测试与 gox/cache 的集成
   - 测试 Redis 后端（需要 Redis 环境）

### 性能测试

1. **并发性能测试：**
   - 测试高并发下的生成和验证性能
   - 测试内存使用情况

2. **压力测试：**
   - 测试大量验证码的存储和清理
   - 测试长时间运行的稳定性

## 实现计划

### 阶段 1：核心接口和内存适配器

1. 定义 Store、Generator、Captcha 接口
2. 实现 Captcha 核心逻辑
3. 实现内存适配器
4. 实现 base64 生成器
5. 编写单元测试

### 阶段 2：Cache 适配器

1. 实现 cache 适配器
2. 编写集成测试
3. 测试与 gox/cache 的兼容性

### 阶段 3：文档和示例

1. 编写包文档
2. 编写使用示例
3. 编写迁移指南
4. 更新 README

## 设计权衡

### 为什么分离 Verify 和 Delete？

**优点：**
- 用户可以控制验证码的生命周期
- 支持多次验证（如果需要）
- 更灵活的错误处理

**缺点：**
- 用户需要记得手动删除
- 代码稍微冗长

**决策：** 选择分离，因为灵活性更重要，且可以通过文档和示例引导用户正确使用。

### 为什么使用适配器模式而不是直接实现？

**优点：**
- 易于扩展新的存储后端
- 符合开闭原则
- 便于测试和 mock

**缺点：**
- 增加了抽象层
- 初次使用需要理解更多概念

**决策：** 选择适配器模式，因为扩展性是核心需求，且与 gox/cache 的设计风格一致。

### 为什么抽象 Generator 接口？

**优点：**
- 支持自定义验证码类型（SVG、Canvas、AI 生成等）
- 不绑定 base64Captcha 库
- 便于测试

**缺点：**
- 增加了复杂度
- 大多数用户只会使用默认生成器

**决策：** 选择抽象，因为验证码类型的多样性是未来趋势，且接口设计简单。

## 总结

本设计采用适配器模式和接口抽象，提供了清晰、灵活、可扩展的验证码解决方案。核心特性包括：

1. **清晰的职责分离**：存储、生成、验证各自独立
2. **灵活的配置**：支持快捷方式和完全自定义
3. **易于扩展**：新增适配器或生成器只需实现接口
4. **生产就绪**：支持分布式场景，提供完整的错误处理
5. **符合 gox 风格**：与 cache 包设计风格一致

设计遵循 Go 最佳实践，接口定义在使用方，实现在子包，便于测试和维护。
