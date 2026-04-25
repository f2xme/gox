package cache_test

import (
	"context"
	"fmt"

	"github.com/f2xme/gox/cache"
	"github.com/f2xme/gox/cache/adapter/memory"
)

func Example_counter() {
	// 创建缓存实例
	c, _ := memory.New()
	defer c.(cache.Closer).Close()

	counter := c.(cache.Counter)
	ctx := context.Background()

	// 递增页面浏览量
	views, _ := counter.Incr(ctx, "page:views", 1)
	fmt.Println("Page views:", views)

	// 再次递增
	views, _ = counter.Incr(ctx, "page:views", 1)
	fmt.Println("Page views:", views)

	// 递减配额
	quota, _ := counter.Incr(ctx, "api:quota", 100)
	fmt.Println("Initial quota:", quota)

	quota, _ = counter.Incr(ctx, "api:quota", -1)
	fmt.Println("After one request:", quota)

	// 浮点数计数器
	balance, _ := counter.IncrFloat(ctx, "account:balance", 100.50)
	fmt.Println("Initial balance:", balance)

	balance, _ = counter.IncrFloat(ctx, "account:balance", -10.25)
	fmt.Println("After withdrawal:", balance)

	// Output:
	// Page views: 1
	// Page views: 2
	// Initial quota: 100
	// After one request: 99
	// Initial balance: 100.5
	// After withdrawal: 90.25
}
