package testkit

import (
	"testing"

	"github.com/f2xme/gox/httpx"
)

// EngineFactory 创建一个待验证的 httpx.Engine。
type EngineFactory struct {
	// Name 是子测试名称。
	Name string
	// New 创建一个新的 httpx.Engine。每个子测试都会调用一次。
	New func(t testing.TB) httpx.Engine
}

// RunEngineSuite 使用同一套测试用例验证多个 httpx.Engine 实现。
func RunEngineSuite(t *testing.T, factories []EngineFactory, suite func(t *testing.T, client *Client)) {
	t.Helper()
	for _, factory := range factories {
		factory := factory
		t.Run(factory.Name, func(t *testing.T) {
			if factory.New == nil {
				t.Fatal("testkit: engine factory New is nil")
			}
			engine := factory.New(t)
			client := New(t, engine)
			defer client.Close()
			suite(t, client)
		})
	}
}
