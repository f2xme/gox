package idgen

import (
	"strconv"
	"sync/atomic"
)

// AutoIncrement 线程安全的自增 ID 生成器
type AutoIncrement struct {
	counter int64
}

// NewAutoIncrement 创建从给定值开始的新 AutoIncrement 生成器
func NewAutoIncrement(start int64) *AutoIncrement {
	return &AutoIncrement{counter: start - 1}
}

// Next 返回下一个 ID
func (a *AutoIncrement) Next() int64 {
	return atomic.AddInt64(&a.counter, 1)
}

// NextN 返回接下来的 n 个 ID
func (a *AutoIncrement) NextN(n int) []int64 {
	if n <= 0 {
		return nil
	}

	start := atomic.AddInt64(&a.counter, int64(n))
	start = start - int64(n) + 1

	ids := make([]int64, n)
	for i := 0; i < n; i++ {
		ids[i] = start + int64(i)
	}
	return ids
}

// Generate 实现 Generator 接口
func (a *AutoIncrement) Generate() string {
	return strconv.FormatInt(a.Next(), 10)
}
