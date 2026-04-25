package memory

// evictionPolicy 定义缓存淘汰策略的接口。
type evictionPolicy interface {
	// onAccess 在通过 Get 访问条目时调用。
	onAccess(key string)

	// onSet 在设置条目时调用。
	onSet(key string)

	// selectVictim 返回要淘汰的键。
	// 如果无法选择受害者则返回空字符串。
	selectVictim() string

	// remove 在从缓存中删除条目时调用。
	remove(key string)
}

// lruPolicy 实现最近最少使用淘汰策略。
type lruPolicy struct {
	accessOrder map[string]uint64 // key -> 单调递增的访问序列
	nextOrder   uint64            // 注意：理论上在 2^64 次访问时溢出（实际上不可能）
}

// newLRUPolicy 创建一个新的 LRU 淘汰策略。
func newLRUPolicy() *lruPolicy {
	return &lruPolicy{
		accessOrder: make(map[string]uint64),
	}
}

// onAccess 更新键的访问顺序。
func (p *lruPolicy) onAccess(key string) {
	p.nextOrder++
	p.accessOrder[key] = p.nextOrder
}

// onSet 更新键的访问顺序。
func (p *lruPolicy) onSet(key string) {
	p.nextOrder++
	p.accessOrder[key] = p.nextOrder
}

// selectVictim 返回具有最旧访问顺序的键。
func (p *lruPolicy) selectVictim() string {
	var victim string
	var oldestOrder uint64
	first := true

	for key, order := range p.accessOrder {
		if first || order < oldestOrder {
			oldestOrder = order
			victim = key
			first = false
		}
	}

	return victim
}

// remove 从跟踪中删除键。
func (p *lruPolicy) remove(key string) {
	delete(p.accessOrder, key)
}

// lfuPolicy 实现最不经常使用淘汰策略。
type lfuPolicy struct {
	accessCount map[string]int // key -> 访问计数
}

// newLFUPolicy 创建一个新的 LFU 淘汰策略。
func newLFUPolicy() *lfuPolicy {
	return &lfuPolicy{
		accessCount: make(map[string]int),
	}
}

// onAccess 增加键的访问计数。
func (p *lfuPolicy) onAccess(key string) {
	p.accessCount[key]++
}

// onSet 如果键是新的则初始化其访问计数。
func (p *lfuPolicy) onSet(key string) {
	if _, exists := p.accessCount[key]; !exists {
		p.accessCount[key] = 0
	}
}

// selectVictim 返回具有最低访问计数的键。
func (p *lfuPolicy) selectVictim() string {
	var victim string
	var lowestCount int = -1

	for key, count := range p.accessCount {
		if lowestCount == -1 || count < lowestCount {
			lowestCount = count
			victim = key
		}
	}

	return victim
}

// remove 从跟踪中删除键。
func (p *lfuPolicy) remove(key string) {
	delete(p.accessCount, key)
}
