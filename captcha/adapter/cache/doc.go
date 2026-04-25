// Package cache 提供基于 gox/cache 包的存储适配器。
//
// cache 适配器适合分布式场景，支持多种后端（内存、Redis 等）。
//
// 示例：
//
//	import (
//		"github.com/f2xme/gox/cache/adapter/mem"
//		cacheadapter "github.com/f2xme/gox/captcha/adapter/cache"
//	)
//
//	c := mem.New()
//	store := cacheadapter.New(c,
//		cacheadapter.WithTTL(10*time.Minute),
//		cacheadapter.WithPrefix("mycaptcha:"),
//	)
package cache
