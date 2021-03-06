// Code generated by gotemplate. DO NOT EDIT.

package xlistener

import (
	"sync"
)

//template type Concurrent(KType,VType,KeyHash)

// A thread safe map.
// To avoid lock bottlenecks this map is dived to several (DefaultShardCount) map shards.

var DefaultShardCountIntFdInfo = uint64(32)

type IntFdInfo struct {
	shardedList  []*shardedIntFdInfo
	shardedCount uint64
}

type shardedIntFdInfo struct {
	items map[int]*fdInfo
	sync.RWMutex
}

// Tuple Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type TupleIntFdInfo struct {
	Key int
	Val *fdInfo
}

// NewWithSharedCount 返回协程安全版本
func NewWithSharedCountIntFdInfo(sharedCount uint64) *IntFdInfo {
	p := &IntFdInfo{
		shardedCount: sharedCount,
		shardedList:  make([]*shardedIntFdInfo, sharedCount),
	}
	for i := uint64(0); i < sharedCount; i++ {
		p.shardedList[i] = &shardedIntFdInfo{items: make(map[int]*fdInfo)}
	}
	return p
}

// New 返回协程安全版本
func NewIntFdInfo() *IntFdInfo {
	return NewWithSharedCountIntFdInfo(DefaultShardCountIntFdInfo)
}

// GetShard Returns shard under given key.
func (m *IntFdInfo) GetShard(key int) *shardedIntFdInfo {
	return m.shardedList[func(k int) uint64 {
		return uint64(k)
	}(key)%m.shardedCount]
}

// IsEmpty checks if map is empty.
func (m *IntFdInfo) IsEmpty() bool {
	return m.Count() == 0
}

func (m *IntFdInfo) Set(key int, value *fdInfo) {
	shard := m.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

// Keys get all keys
func (m *IntFdInfo) Keys() []int {
	var ret []int
	for _, shard := range m.shardedList {
		shard.RLock()
		for key := range shard.items {
			ret = append(ret, key)
		}
		shard.RUnlock()
	}
	return ret
}

// MGet multiple get by keys
func (m *IntFdInfo) MGet(keys ...int) map[int]*fdInfo {
	data := make(map[int]*fdInfo)
	for _, key := range keys {
		if val, ok := m.Get(key); ok {
			data[key] = val
		}
	}
	return data
}

// GetAll get all values
func (m *IntFdInfo) GetAll() map[int]*fdInfo {
	data := make(map[int]*fdInfo)

	for _, shard := range m.shardedList {
		shard.RLock()
		for key, val := range shard.items {
			data[key] = val
		}
		shard.RUnlock()
	}
	return data
}

// Clear all values
func (m *IntFdInfo) Clear() {
	for _, shard := range m.shardedList {
		shard.Lock()
		shard.items = make(map[int]*fdInfo)
		shard.Unlock()
	}
}

// MSet multiple set
func (m *IntFdInfo) MSet(data map[int]*fdInfo) {
	for key, value := range data {
		m.Set(key, value)
	}
}

// SetNX like redis SETNX
// return true if the key was set
// return false if the key was not set
func (m *IntFdInfo) SetNX(key int, value *fdInfo) bool {
	shard := m.GetShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return true
}

func (m *IntFdInfo) LockFuncWithKey(key int, f func(m map[int]*fdInfo)) {
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	f(shard.items)
}

func (m *IntFdInfo) RLockFuncWithKey(key int, f func(m map[int]*fdInfo)) {
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()
	f(shard.items)
}

func (m *IntFdInfo) LockFunc(f func(data map[int]*fdInfo)) {
	for _, shard := range m.shardedList {
		shard.Lock()
		f(shard.items)
		shard.Unlock()
	}
}

func (m *IntFdInfo) RLockFunc(f func(m map[int]*fdInfo)) {
	for _, shard := range m.shardedList {
		shard.RLock()
		f(shard.items)
		shard.RUnlock()
	}
}

func (m *IntFdInfo) doSetWithLockCheck(key int, val *fdInfo) (result *fdInfo, isSet bool) {
	shard := m.GetShard(key)
	shard.Lock()

	if got, ok := shard.items[key]; ok {
		shard.Unlock()
		return got, false
	}

	shard.items[key] = val
	isSet = true
	shard.Unlock()
	return
}

func (m *IntFdInfo) doSetWithLockCheckWithFunc(key int, f func(key int) *fdInfo) (result *fdInfo, isSet bool) {
	shard := m.GetShard(key)
	shard.Lock()

	if got, ok := shard.items[key]; ok {
		shard.Unlock()
		return got, false
	}

	shard.items[key] = f(key)
	isSet = true
	shard.Unlock()
	return
}

// GetOrSetFunc 获取或者设定数值，f在lock外执行
func (m *IntFdInfo) GetOrSetFunc(key int, f func(key int) *fdInfo) (result *fdInfo, isSet bool) {
	if v, ok := m.Get(key); ok {
		return v, false
	}
	return m.doSetWithLockCheck(key, f(key))
}

// GetOrSetFuncLock 获取或者设定数值，f在lock内执行
func (m *IntFdInfo) GetOrSetFuncLock(key int, f func(key int) *fdInfo) (result *fdInfo, isSet bool) {
	if v, ok := m.Get(key); ok {
		return v, false
	}
	return m.doSetWithLockCheckWithFunc(key, f)
}

// GetOrSet 获取或设定元素
func (m *IntFdInfo) GetOrSet(key int, val *fdInfo) (*fdInfo, bool) {
	if v, ok := m.Get(key); ok {
		return v, false
	}
	return m.doSetWithLockCheck(key, val)
}

func (m *IntFdInfo) Get(key int) (*fdInfo, bool) {
	shard := m.GetShard(key)
	shard.RLock()
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

func (m *IntFdInfo) Len() int  { return m.Count() }
func (m *IntFdInfo) Size() int { return m.Count() }
func (m *IntFdInfo) Count() int {
	count := 0
	for i := uint64(0); i < m.shardedCount; i++ {
		shard := m.shardedList[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

func (m *IntFdInfo) Has(key int) bool {
	shard := m.GetShard(key)
	shard.RLock()
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

func (m *IntFdInfo) Remove(key int) {
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

func (m *IntFdInfo) GetAndRemove(key int) (*fdInfo, bool) {
	shard := m.GetShard(key)
	shard.Lock()
	val, ok := shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return val, ok
}

// Iter Returns an iterator which could be used in a for range loop.
func (m *IntFdInfo) Iter() <-chan TupleIntFdInfo {
	ch := make(chan TupleIntFdInfo)
	go func() {
		for _, shard := range m.shardedList {
			shard.RLock()
			for key, val := range shard.items {
				ch <- TupleIntFdInfo{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

// IterBuffered Returns a buffered iterator which could be used in a for range loop.
func (m *IntFdInfo) IterBuffered() <-chan TupleIntFdInfo {
	ch := make(chan TupleIntFdInfo, m.Count())
	go func() {
		// Foreach shard.
		for _, shard := range m.shardedList {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				ch <- TupleIntFdInfo{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}
