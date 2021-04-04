package main

import "sync"

// 一个读写锁保护的线程安全的map
type RWMap struct {
    sync.RWMutex // 读写锁保护下面的map字段
    m            map[int]int
}

// 新建一个RWMap
func NewRWMap(n int) *RWMap {
    return &RWMap{
        m: make(map[int]int, n),
    }
}

// 从map中读取一个值
func (m *RWMap) Get(k int) (int, bool) {
    m.RLock()
    defer m.RUnlock()
    v, existed := m.m[k] // 在锁的保护下从map中读取
    return v, existed
}

// 设置一个键值对
func (m *RWMap) Set(k int, v int) {
    m.Lock() // 锁保护
    defer m.Unlock()
    m.m[k] = v
}

// 删除一个键
func (m *RWMap) Delete(k int) {
    m.Lock() // 锁保护
    defer m.Unlock()
    delete(m.m, k)
}

// map的长度
func (m *RWMap) Len() int {
    m.RLock() // 锁保护
    defer m.RUnlock()
    return len(m.m)
}

// 遍历map
func (m *RWMap) Each(f func(k, v int) bool) {
    m.RLock() //遍历期间一直持有读锁
    defer m.RUnlock()

    for k, v := range m.m {
        if !f(k, v) {
            return
        }
    }
}
