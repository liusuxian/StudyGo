package main

import (
    "fmt"
    "sync"
    "sync/atomic"
    "time"
    "unsafe"
)

// 复制Mutex定义的常量
const (
    mutexLocked      = 1 << iota // 加锁标识位置
    mutexWoken                   // 唤醒标识位置
    mutexStarving                // 锁饥饿标识位置
    mutexWaiterShift = iota      // 标识waiter的起始bit位置
)

// Mutex 扩展一个Mutex结构
type Mutex struct {
    sync.Mutex
}

// Count 当前等待者的数量
func (m *Mutex) Count() int {
    // 获取state字段的值
    v := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.Mutex)))
    v = v >> mutexWaiterShift // 得到等待者的数值
    v = v + (v & mutexLocked) // 再加上锁持有者的数量，0或者1
    return int(v)
}

// IsLocked 锁是否被持有
func (m *Mutex) IsLocked() bool {
    state := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.Mutex)))
    return state&mutexLocked == mutexLocked
}

// IsWoken 是否有等待者被唤醒
func (m *Mutex) IsWoken() bool {
    state := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.Mutex)))
    return state&mutexWoken == mutexWoken
}

// IsStarving 锁是否处于饥饿状态
func (m *Mutex) IsStarving() bool {
    state := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.Mutex)))
    return state&mutexStarving == mutexStarving
}

func count() {
    var mu Mutex
    // 启动1000个goroutine
    for i := 0; i < 1000; i++ {
        go func() {
            mu.Lock()
            time.Sleep(time.Second)
            mu.Unlock()
        }()
    }

    time.Sleep(time.Second)
    // 输出锁的信息
    fmt.Printf("waitings: %d, isLocked: %t, woken: %t,  starving: %t\n", mu.Count(), mu.IsLocked(), mu.IsWoken(), mu.IsStarving())
}

func main() {
    count()
}
