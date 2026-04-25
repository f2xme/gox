package cache_test

import (
	"context"
	"testing"

	"github.com/f2xme/gox/cache"
	"github.com/f2xme/gox/cache/adapter/memory"
)

func TestCounter_Incr(t *testing.T) {
	c, err := memory.New()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.(cache.Closer).Close()

	counter, ok := c.(cache.Counter)
	if !ok {
		t.Fatal("memory cache should implement Counter interface")
	}

	ctx := context.Background()

	// 测试从 0 开始递增
	val, err := counter.Incr(ctx, "counter:1", 1)
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}

	// 再次递增
	val, err = counter.Incr(ctx, "counter:1", 5)
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}
	if val != 6 {
		t.Errorf("Expected 6, got %d", val)
	}

	// 测试递减
	val, err = counter.Incr(ctx, "counter:1", -3)
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}
	if val != 3 {
		t.Errorf("Expected 3, got %d", val)
	}
}

func TestCounter_IncrFloat(t *testing.T) {
	c, err := memory.New()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.(cache.Closer).Close()

	counter, ok := c.(cache.Counter)
	if !ok {
		t.Fatal("memory cache should implement Counter interface")
	}

	ctx := context.Background()

	// 测试从 0.0 开始递增
	val, err := counter.IncrFloat(ctx, "balance", 10.5)
	if err != nil {
		t.Fatalf("IncrFloat failed: %v", err)
	}
	if val != 10.5 {
		t.Errorf("Expected 10.5, got %f", val)
	}

	// 再次递增
	val, err = counter.IncrFloat(ctx, "balance", 5.25)
	if err != nil {
		t.Fatalf("IncrFloat failed: %v", err)
	}
	if val != 15.75 {
		t.Errorf("Expected 15.75, got %f", val)
	}

	// 测试递减
	val, err = counter.IncrFloat(ctx, "balance", -3.5)
	if err != nil {
		t.Fatalf("IncrFloat failed: %v", err)
	}
	if val != 12.25 {
		t.Errorf("Expected 12.25, got %f", val)
	}
}

func TestCounter_Concurrent(t *testing.T) {
	c, err := memory.New()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.(cache.Closer).Close()

	counter, ok := c.(cache.Counter)
	if !ok {
		t.Fatal("memory cache should implement Counter interface")
	}

	ctx := context.Background()
	key := "concurrent:counter"

	// 并发递增
	const goroutines = 100
	const increments = 10

	done := make(chan bool, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < increments; j++ {
				_, err := counter.Incr(ctx, key, 1)
				if err != nil {
					t.Errorf("Incr failed: %v", err)
				}
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// 验证最终值
	val, err := counter.Incr(ctx, key, 0)
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}

	expected := int64(goroutines * increments)
	if val != expected {
		t.Errorf("Expected %d, got %d", expected, val)
	}
}
