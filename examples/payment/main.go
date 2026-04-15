package main

import "fmt"

func main() {
	fmt.Println("=== payment 包使用示例 ===")
	fmt.Println("注意：payment 需要配置真实的支付服务商")
	fmt.Println()
	fmt.Println("可用适配器：")
	fmt.Println("- wechatadapter - 微信支付")
	fmt.Println("- alipayadapter - 支付宝")
	fmt.Println()
	fmt.Println("典型流程：")
	fmt.Println("1. 创建订单（Order）")
	fmt.Println("2. 调用 Pay() 发起支付")
	fmt.Println("3. 调用 Query() 查询支付状态")
	fmt.Println("4. 调用 Refund() 申请退款")
	fmt.Println("5. 调用 Close() 关闭订单")
}
