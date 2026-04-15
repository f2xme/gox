package main

import (
	"errors"
	"fmt"

	"github.com/f2xme/gox/logx"
	"github.com/f2xme/gox/logx/adapter/zap"
)

func main() {
	// 初始化日志（使用 Zap 适配器）
	logger := zap.New()
	logx.Init(logger)

	fmt.Println("=== logx 使用示例 ===")

	// 1. 基本日志记录
	fmt.Println("1. 基本日志记录:")
	logx.Info("应用程序启动")
	logx.Warn("这是一个警告消息")

	// 2. 结构化日志（带字段）
	fmt.Println("\n2. 结构化日志:")
	logx.Info("用户登录",
		logx.NewKV("user_id", 12345),
		logx.NewKV("username", "zhangsan"),
		logx.NewKV("ip", "192.168.1.100"),
	)

	// 3. 错误日志
	fmt.Println("\n3. 错误日志:")
	err := errors.New("数据库连接失败")
	logx.Error(err,
		logx.NewKV("database", "mysql"),
		logx.NewKV("host", "localhost:3306"),
	)

	// 4. 多个字段的日志
	fmt.Println("\n4. 多字段日志:")
	logx.Info("订单创建成功",
		logx.NewKV("order_id", "ORD-2024-001"),
		logx.NewKV("user_id", 12345),
		logx.NewKV("amount", 99.99),
		logx.NewKV("currency", "CNY"),
	)

	// 5. 业务流程日志
	fmt.Println("\n5. 业务流程日志:")
	processOrder()

	// 6. 刷新日志缓冲区（确保所有日志写入）
	if err := logx.Flush(); err != nil {
		fmt.Printf("刷新日志失败: %v\n", err)
	}

	fmt.Println("\n日志示例完成")
}

// processOrder 模拟订单处理流程
func processOrder() {
	orderID := "ORD-2024-001"

	logx.Info("开始处理订单", logx.NewKV("order_id", orderID))

	// 模拟验证
	logx.Info("验证订单信息", logx.NewKV("order_id", orderID))

	// 模拟支付
	logx.Info("处理支付", logx.NewKV("order_id", orderID), logx.NewKV("amount", 99.99))

	// 模拟库存扣减
	logx.Info("扣减库存", logx.NewKV("order_id", orderID), logx.NewKV("product_id", "PROD-001"))

	logx.Info("订单处理完成", logx.NewKV("order_id", orderID), logx.NewKV("status", "success"))
}
