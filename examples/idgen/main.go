package main

import (
	"fmt"
	"sync"

	"github.com/f2xme/gox/idgen"
)

func main() {
	fmt.Println("=== IDGen 使用示例 ===")

	// 1. UUID 生成
	fmt.Println("1. UUID 生成:")
	uuid1 := idgen.UUIDString()
	uuid2 := idgen.UUIDString()
	fmt.Printf("UUID 1: %s\n", uuid1)
	fmt.Printf("UUID 2: %s\n", uuid2)
	fmt.Printf("UUID 长度: %d 字符\n", len(uuid1))

	// 2. ULID 生成（时间排序的 UUID）
	fmt.Println("\n2. ULID 生成:")
	ulid1 := idgen.ULIDString()
	ulid2 := idgen.ULIDString()
	fmt.Printf("ULID 1: %s\n", ulid1)
	fmt.Printf("ULID 2: %s\n", ulid2)
	fmt.Println("ULID 特点: 时间排序，适合作为数据库主键")

	// 3. Snowflake ID 生成
	fmt.Println("\n3. Snowflake ID 生成:")
	snowflake1, err := idgen.Snowflake()
	if err != nil {
		fmt.Printf("生成 Snowflake ID 失败: %v\n", err)
		return
	}
	snowflake2, _ := idgen.Snowflake()
	fmt.Printf("Snowflake 1: %d\n", snowflake1)
	fmt.Printf("Snowflake 2: %d\n", snowflake2)

	// 4. 解析 Snowflake ID
	fmt.Println("\n4. 解析 Snowflake ID:")
	info := idgen.ParseSnowflake(snowflake1)
	fmt.Printf("时间戳: %s\n", info.Timestamp.Format("2006-01-02 15:04:05.000"))
	fmt.Printf("节点 ID: %d\n", info.NodeID)
	fmt.Printf("序列号: %d\n", info.Sequence)

	// 5. 使用指定节点 ID 生成 Snowflake
	fmt.Println("\n5. 使用指定节点 ID:")
	nodeID := int64(100)
	snowflakeWithNode, err := idgen.SnowflakeWithNode(nodeID)
	if err != nil {
		fmt.Printf("生成失败: %v\n", err)
		return
	}
	fmt.Printf("节点 %d 的 Snowflake ID: %d\n", nodeID, snowflakeWithNode)

	infoWithNode := idgen.ParseSnowflake(snowflakeWithNode)
	fmt.Printf("验证节点 ID: %d\n", infoWithNode.NodeID)

	// 6. 并发生成 ID（测试并发安全性）
	fmt.Println("\n6. 并发生成 ID 测试:")
	const goroutines = 10
	const idsPerGoroutine = 100

	var wg sync.WaitGroup
	idMap := sync.Map{}
	duplicates := 0

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := idgen.Snowflake()
				if err != nil {
					continue
				}
				if _, exists := idMap.LoadOrStore(id, true); exists {
					duplicates++
				}
			}
		}()
	}
	wg.Wait()

	totalIDs := goroutines * idsPerGoroutine
	fmt.Printf("生成了 %d 个 ID\n", totalIDs)
	fmt.Printf("重复 ID 数量: %d\n", duplicates)
	if duplicates == 0 {
		fmt.Println("✓ 并发安全测试通过，无重复 ID")
	}

	// 7. 不同 ID 类型的比较
	fmt.Println("\n7. 不同 ID 类型的特点:")
	fmt.Println("UUID:")
	fmt.Println("  - 全局唯一")
	fmt.Println("  - 无序")
	fmt.Println("  - 36 字符（带连字符）")
	fmt.Println("  - 适合: 分布式系统、无序场景")

	fmt.Println("\nULID:")
	fmt.Println("  - 全局唯一")
	fmt.Println("  - 时间排序")
	fmt.Println("  - 26 字符")
	fmt.Println("  - 适合: 数据库主键、需要排序的场景")

	fmt.Println("\nSnowflake:")
	fmt.Println("  - 全局唯一（需要配置节点 ID）")
	fmt.Println("  - 时间排序")
	fmt.Println("  - 64 位整数")
	fmt.Println("  - 适合: 高性能场景、分布式系统")

	// 8. 实际应用示例
	fmt.Println("\n8. 实际应用示例:")

	// 订单 ID（使用 ULID，时间排序）
	orderID := idgen.ULIDString()
	fmt.Printf("订单 ID: %s\n", orderID)

	// 用户 ID（使用 Snowflake，高性能）
	userID, _ := idgen.Snowflake()
	fmt.Printf("用户 ID: %d\n", userID)

	// 会话 ID（使用 UUID，全局唯一）
	sessionID := idgen.UUIDString()
	fmt.Printf("会话 ID: %s\n", sessionID)

	fmt.Println("\nID 生成示例完成")
}
