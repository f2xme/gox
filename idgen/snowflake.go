package idgen

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	// Epoch 自定义纪元（2020-01-01 00:00:00 UTC）
	epoch int64 = 1577836800000

	nodeBits     = 10
	sequenceBits = 12

	nodeMax     = -1 ^ (-1 << nodeBits)
	sequenceMax = -1 ^ (-1 << sequenceBits)

	timeShift = nodeBits + sequenceBits
	nodeShift = sequenceBits
)

var (
	defaultNode     int64
	defaultNodeOnce sync.Once

	mu       sync.Mutex
	lastTime int64
	sequence int64
)

// SnowflakeInfo 包含解析后的 Snowflake ID 信息
type SnowflakeInfo struct {
	Timestamp time.Time
	NodeID    int64
	Sequence  int64
}

// Snowflake 使用默认节点 ID 生成 Snowflake ID
// 节点 ID 从 NODE_ID 环境变量获取，
// 如果未设置则从 MAC 地址派生
// 如果 ID 生成失败则返回错误
func Snowflake() (int64, error) {
	defaultNodeOnce.Do(func() {
		defaultNode = getDefaultNode()
	})
	return SnowflakeWithNode(defaultNode)
}

// SnowflakeWithNode 使用指定的节点 ID 生成 Snowflake ID
// 节点 ID 必须在 0 到 1023 之间
// 如果节点 ID 超出范围则返回错误
func SnowflakeWithNode(nodeID int64) (int64, error) {
	if nodeID < 0 || nodeID > nodeMax {
		return 0, fmt.Errorf("node ID out of range: %d (must be 0-%d)", nodeID, nodeMax)
	}

	mu.Lock()
	defer mu.Unlock()

	now := time.Now().UnixMilli() - epoch

	if now == lastTime {
		sequence = (sequence + 1) & sequenceMax
		if sequence == 0 {
			for now <= lastTime {
				mu.Unlock()
				time.Sleep(time.Millisecond)
				mu.Lock()
				now = time.Now().UnixMilli() - epoch
			}
		}
	} else {
		sequence = 0
	}

	lastTime = now

	return (now << timeShift) | (nodeID << nodeShift) | sequence, nil
}

// ParseSnowflake 解析 Snowflake ID 并返回其组成部分
func ParseSnowflake(id int64) SnowflakeInfo {
	timestamp := (id >> timeShift) + epoch
	nodeID := (id >> nodeShift) & nodeMax
	seq := id & sequenceMax

	return SnowflakeInfo{
		Timestamp: time.UnixMilli(timestamp),
		NodeID:    nodeID,
		Sequence:  seq,
	}
}

func getDefaultNode() int64 {
	// Try environment variable first
	if nodeStr := os.Getenv("NODE_ID"); nodeStr != "" {
		if node, err := strconv.ParseInt(nodeStr, 10, 64); err == nil {
			if node >= 0 && node <= nodeMax {
				return node
			}
		}
	}

	// Derive from MAC address
	if interfaces, err := net.Interfaces(); err == nil {
		for _, iface := range interfaces {
			if len(iface.HardwareAddr) >= 6 {
				// Use last 10 bits of MAC address
				mac := iface.HardwareAddr
				node := int64(mac[4])<<8 | int64(mac[5])
				return node & nodeMax
			}
		}
	}

	// Fallback to 0
	return 0
}
