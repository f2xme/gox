package idgen

import (
	"crypto/rand"
	"io"
	"sync"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

var ulidEntropyPool = sync.Pool{
	New: func() interface{} {
		return ulid.Monotonic(rand.Reader, 0)
	},
}

// UUID 生成随机 UUID (v4)
func UUID() uuid.UUID {
	return uuid.New()
}

// UUIDString 生成随机 UUID (v4) 并返回字符串形式
func UUIDString() string {
	return uuid.New().String()
}

// ULID 生成 ULID（通用唯一字典序可排序标识符）
// ULID 是时间有序的，适合用作数据库主键
func ULID() ulid.ULID {
	entropy := ulidEntropyPool.Get().(io.Reader)
	defer ulidEntropyPool.Put(entropy)
	return ulid.MustNew(ulid.Now(), entropy)
}

// ULIDString 生成 ULID 并返回字符串形式
func ULIDString() string {
	return ULID().String()
}
