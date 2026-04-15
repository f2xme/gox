// Package idgen 提供多种 ID 生成算法。
//
// 支持 Snowflake、UUID、ULID、ShortID 和 AutoIncrement 等多种 ID 生成策略，
// 所有生成器都实现了 Generator 接口，可以互换使用。
//
// # 功能特性
//
//   - Snowflake：基于 Twitter Snowflake 算法的分布式唯一 ID，趋势递增可排序
//   - UUID：标准 UUID v4，全局唯一无需协调
//   - ULID：可排序的唯一标识符，结合 UUID 唯一性和 Snowflake 可排序性
//   - ShortID：短小的 URL 安全 ID，适合短链接场景
//   - AutoIncrement：线程安全的自增 ID，适合单机应用
//   - Generator 接口：统一的 ID 生成抽象，支持依赖注入
//   - 线程安全：所有生成器都支持并发使用
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/f2xme/gox/idgen"
//	)
//
//	func main() {
//		// Snowflake ID（分布式场景）
//		id, _ := idgen.Snowflake()
//		fmt.Println(id) // 1234567890123456789
//
//		// UUID（全局唯一）
//		uuid := idgen.UUIDString()
//		fmt.Println(uuid) // "550e8400-e29b-41d4-a716-446655440000"
//
//		// ULID（可排序）
//		ulid := idgen.ULIDString()
//		fmt.Println(ulid) // "01ARZ3NDEKTSV4RRFFQ69G5FAV"
//
//		// ShortID（短链接）
//		shortID, _ := idgen.ShortID()
//		fmt.Println(shortID) // "V1StGXR8"
//
//		// AutoIncrement（自增）
//		gen := idgen.NewAutoIncrement(1000)
//		fmt.Println(gen.Next()) // 1001
//		fmt.Println(gen.Next()) // 1002
//	}
//
// # 使用 Generator 接口
//
// 通过 Generator 接口实现依赖注入：
//
//	type UserService struct {
//		idGen idgen.Generator
//	}
//
//	func NewUserService(idGen idgen.Generator) *UserService {
//		return &UserService{idGen: idGen}
//	}
//
//	func (s *UserService) CreateUser(name string) (*User, error) {
//		user := &User{
//			ID:   s.idGen.Generate(),
//			Name: name,
//		}
//		return user, db.Save(user)
//	}
//
//	// 使用时注入具体实现
//	service := NewUserService(idgen.GeneratorFunc(idgen.UUIDString))
//
// # 算法选择指南
//
// 不同场景推荐使用的算法：
//
//   - 分布式系统：Snowflake（趋势递增，支持多节点）
//   - 需要排序：Snowflake / ULID（按时间排序）
//   - 全局唯一：UUID / ULID（无需协调）
//   - 短链接：ShortID（短小，URL 安全）
//   - 单机应用：AutoIncrement（简单高效）
//   - 数据库主键：Snowflake / ULID（可排序，索引友好）
//
// # Snowflake 详解
//
// Snowflake 算法生成 64 位整数 ID，结构如下：
//
//   - 41 位时间戳（毫秒级，可用 69 年）
//   - 10 位节点 ID（支持 1024 个节点）
//   - 12 位序列号（每毫秒可生成 4096 个 ID）
//
// 使用指定节点 ID：
//
//	id, err := idgen.SnowflakeWithNode(1)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// 解析 Snowflake ID：
//
//	info := idgen.ParseSnowflake(id)
//	fmt.Printf("时间: %v, 节点: %d, 序列: %d\n",
//		info.Timestamp, info.NodeID, info.Sequence)
//
// # 最佳实践
//
// 1. 在分布式系统中为每个节点分配唯一的节点 ID：
//
//	nodeID := getNodeIDFromConfig() // 从配置读取
//	id, _ := idgen.SnowflakeWithNode(nodeID)
//
// 2. 使用依赖注入而不是直接创建：
//
//	// 推荐
//	func NewService(idGen idgen.Generator) *Service {
//		return &Service{idGen: idGen}
//	}
//
//	// 不推荐
//	func NewService() *Service {
//		return &Service{idGen: idgen.GeneratorFunc(idgen.UUIDString)}
//	}
//
// 3. 复用生成器实例，避免在循环中重复创建：
//
//	// 推荐
//	gen := idgen.NewAutoIncrement(0)
//	for i := 0; i < 1000; i++ {
//		id := gen.Generate()
//	}
//
//	// 不推荐
//	for i := 0; i < 1000; i++ {
//		gen := idgen.NewAutoIncrement(0)
//		id := gen.Generate()
//	}
//
// # 线程安全
//
// 所有 ID 生成器都是线程安全的，可以在多个 goroutine 中并发使用。
package idgen
