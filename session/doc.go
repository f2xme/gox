/*
Package session 提供通用会话管理能力。

session 包只处理会话 ID、数据、过期时间和存储抽象，不绑定 HTTP Cookie、
Gin 中间件或认证登录态。调用方可以在 Web 层自行把 session ID 写入 Cookie、
Header 或其他传输位置。

# 功能特性

  - 管理会话 ID、过期时间和键值数据
  - 通过 Store 接口适配内存、Redis 等不同存储
  - 支持刷新、保存、删除和销毁会话

# 快速开始

	import (
		"context"

		"github.com/f2xme/gox/session"
		"github.com/f2xme/gox/session/adapter/memory"
	)

	func main() {
		store, _ := memory.New()
		manager, _ := session.New(store)

		ctx := context.Background()
		sess, _ := manager.Create(ctx)

		sess.Values["user_id"] = "1001"
		_ = manager.Save(ctx, sess)

		loaded, _ := manager.Get(ctx, sess.ID)
		_ = loaded.Values["user_id"]
	}

# 存储适配器

memory 适合单进程应用和测试场景；redis 适合多实例服务共享会话。
*/
package session
