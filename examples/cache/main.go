package main

import (
	"context"
	"fmt"
	"time"

	"github.com/f2xme/gox/cache"
	"github.com/f2xme/gox/cache/adapter/memory"
)

func main() {
	fmt.Println("=== Cache 使用示例 ===")

	// 1. 创建内存缓存实例
	c, err := memory.New()
	if err != nil {
		fmt.Printf("创建缓存失败: %v\n", err)
		return
	}
	// 如果缓存实现了 Closer 接口，在程序结束时关闭
	if closer, ok := c.(cache.Closer); ok {
		defer closer.Close()
	}

	ctx := context.Background()

	// 2. 基本的 Set/Get 操作
	fmt.Println("1. 基本的 Set/Get 操作:")
	key := "user:1001"
	value := []byte(`{"id":1001,"name":"张三","age":25}`)

	// 设置缓存，TTL 为 5 秒
	if err := c.Set(ctx, key, value, 5*time.Second); err != nil {
		fmt.Printf("设置缓存失败: %v\n", err)
		return
	}
	fmt.Printf("已设置缓存: %s\n", key)

	// 获取缓存
	data, err := c.Get(ctx, key)
	if err != nil {
		fmt.Printf("获取缓存失败: %v\n", err)
		return
	}
	fmt.Printf("获取缓存成功: %s\n", string(data))

	// 3. 检查键是否存在
	fmt.Println("\n2. 检查键是否存在:")
	exists, err := c.Exists(ctx, key)
	if err != nil {
		fmt.Printf("检查失败: %v\n", err)
		return
	}
	fmt.Printf("键 %s 是否存在: %v\n", key, exists)

	// 4. 删除缓存
	fmt.Println("\n3. 删除缓存:")
	if err := c.Delete(ctx, key); err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return
	}
	fmt.Printf("已删除键: %s\n", key)

	// 验证删除
	exists, _ = c.Exists(ctx, key)
	fmt.Printf("删除后键是否存在: %v\n", exists)

	// 5. 演示缓存过期
	fmt.Println("\n4. 缓存过期演示:")
	shortKey := "temp:data"
	shortValue := []byte("临时数据")

	// 设置 2 秒过期
	c.Set(ctx, shortKey, shortValue, 2*time.Second)
	fmt.Printf("设置缓存，2秒后过期: %s\n", shortKey)

	// 立即获取
	if data, err := c.Get(ctx, shortKey); err == nil {
		fmt.Printf("立即获取成功: %s\n", string(data))
	}

	// 等待 3 秒后再获取
	fmt.Println("等待 3 秒...")
	time.Sleep(3 * time.Second)

	if _, err := c.Get(ctx, shortKey); err == cache.ErrNotFound {
		fmt.Println("3秒后获取失败: 缓存已过期")
	}

	// 6. 无过期时间的缓存
	fmt.Println("\n5. 无过期时间的缓存:")
	permanentKey := "config:app"
	permanentValue := []byte(`{"version":"1.0.0","env":"production"}`)

	// TTL 为 0 表示永不过期
	c.Set(ctx, permanentKey, permanentValue, 0)
	fmt.Printf("设置永久缓存: %s\n", permanentKey)

	if data, err := c.Get(ctx, permanentKey); err == nil {
		fmt.Printf("获取永久缓存: %s\n", string(data))
	}

	// 7. 批量操作示例
	fmt.Println("\n6. 批量缓存操作:")
	for i := 1; i <= 3; i++ {
		key := fmt.Sprintf("product:%d", i)
		value := []byte(fmt.Sprintf(`{"id":%d,"name":"商品%d"}`, i, i))
		c.Set(ctx, key, value, 10*time.Second)
	}
	fmt.Println("已设置 3 个商品缓存")

	// 验证批量设置
	for i := 1; i <= 3; i++ {
		key := fmt.Sprintf("product:%d", i)
		if exists, _ := c.Exists(ctx, key); exists {
			fmt.Printf("  %s 存在\n", key)
		}
	}

	fmt.Println("\n缓存示例完成")
}
