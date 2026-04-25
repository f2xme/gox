// Package captcha 提供验证码生成和验证功能。
//
// # 功能特性
//
//   - 支持多种验证码类型：数字、字母、算术、音频
//   - 灵活的存储后端：内存存储或自定义实现
//   - 可配置的验证码参数：长度、尺寸、噪点等
//   - 自动过期和清理机制
//   - 防止重复使用（验证后自动删除）
//
// # 快速开始
//
// 基本使用：
//
//	package main
//
//	import (
//		"fmt"
//		"time"
//
//		"github.com/f2xme/gox/captcha"
//	)
//
//	func main() {
//		// 创建内存存储（保留 1000 个验证码，5 分钟过期）
//		store := captcha.NewMemoryStore(1000, 5*time.Minute)
//
//		// 创建验证码实例
//		c := captcha.New(store)
//
//		// 生成验证码
//		id, b64s, err := c.Generate()
//		if err != nil {
//			log.Fatalf("生成验证码失败: %v", err)
//		}
//		fmt.Printf("验证码 ID: %s", id)
//		fmt.Printf("验证码图片: data:image/png;base64,%s", b64s)
//
//		// 验证用户输入
//		if c.Verify(id, "1234") {
//			fmt.Println("验证成功")
//		} else {
//			fmt.Println("验证失败")
//		}
//	}
//
// # 验证码类型
//
// 数字验证码（默认）：
//
//	c := captcha.New(store, captcha.WithType(captcha.TypeDigit))
//
// 字母数字混合验证码：
//
//	c := captcha.New(store, captcha.WithType(captcha.TypeString))
//
// 算术表达式验证码：
//
//	c := captcha.New(store, captcha.WithType(captcha.TypeMath))
//
// 音频验证码：
//
//	c := captcha.New(store, captcha.WithType(captcha.TypeAudio))
//
// # 自定义配置
//
// 可以通过选项函数自定义验证码参数：
//
//	// 使用 Option 函数
//	c := captcha.New(store,
//		captcha.WithType(captcha.TypeDigit),
//		captcha.WithLength(6),           // 6 位数字
//		captcha.WithWidth(300),          // 宽度 300px
//		captcha.WithHeight(100),         // 高度 100px
//		captcha.WithNoiseCount(5),       // 5 个噪点
//	)
//
//	// 或直接使用 Options 结构体
//	c := captcha.New(store,
//		func(o *captcha.Options) {
//			o.CaptchaType = captcha.TypeDigit
//			o.Length = 6
//			o.Width = 300
//			o.Height = 100
//			o.NoiseCount = 5
//		},
//	)
//
// # 自定义存储
//
// 可以实现 base64Captcha.Store 接口来使用自定义存储（如 Redis）：
//
//	type RedisStore struct {
//		client *redis.Client
//	}
//
//	func (s *RedisStore) Set(id string, value string) error {
//		return s.client.Set(context.Background(), id, value, 5*time.Minute).Err()
//	}
//
//	func (s *RedisStore) Get(id string, clear bool) string {
//		val, _ := s.client.Get(context.Background(), id).Result()
//		if clear {
//			s.client.Del(context.Background(), id)
//		}
//		return val
//	}
//
//	func (s *RedisStore) Verify(id, answer string, clear bool) bool {
//		val := s.Get(id, clear)
//		return val == answer
//	}
//
//	// 使用 Redis 存储
//	store := &RedisStore{client: redisClient}
//	c := captcha.New(store)
//
// # Web 集成示例
//
// 在 HTTP 服务中使用：
//
//	// 生成验证码接口
//	http.HandleFunc("/captcha/generate", func(w http.ResponseWriter, r *http.Request) {
//		id, b64s, err := c.Generate()
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//			return
//		}
//		json.NewEncoder(w).Encode(map[string]string{
//			"id":    id,
//			"image": b64s,
//		})
//	})
//
//	// 验证验证码接口
//	http.HandleFunc("/captcha/verify", func(w http.ResponseWriter, r *http.Request) {
//		var req struct {
//			ID     string `json:"id"`
//			Answer string `json:"answer"`
//		}
//		json.NewDecoder(r.Body).Decode(&req)
//
//		if c.Verify(req.ID, req.Answer) {
//			json.NewEncoder(w).Encode(map[string]bool{"success": true})
//		} else {
//			json.NewEncoder(w).Encode(map[string]bool{"success": false})
//		}
//	})
//
// # 注意事项
//
//   - 验证码验证成功后会自动删除，防止重复使用
//   - 验证码过期后无法通过验证
//   - 内存存储的 GC 机制会在达到限制时清理最旧的验证码
//   - 音频验证码返回的是 base64 编码的 WAV 音频数据
//   - 建议在生产环境使用 Redis 等持久化存储
package captcha
