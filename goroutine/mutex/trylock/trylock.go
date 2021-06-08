package main

import (
    "fmt"
    "math/rand"
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

// TryLock 尝试获取锁
func (m *Mutex) TryLock() bool {
    // 如果能成功抢到锁
    if atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.Mutex)), 0, mutexLocked) {
        return true
    }

    // 如果处于唤醒、加锁或者饥饿状态，这次请求就不参与竞争了，返回false
    old := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.Mutex)))
    if old&(mutexLocked|mutexStarving|mutexWoken) != 0 {
        return false
    }

    // 尝试在竞争的状态下请求锁
    vnew := old | mutexLocked
    return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.Mutex)), old, vnew)
}

func try() {
    var mu Mutex
    // 启动一个goroutine持有一段时间的锁
    go func() {
        mu.Lock()
        time.Sleep(time.Duration(rand.Intn(2)) * time.Second)
        mu.Unlock()
    }()

    time.Sleep(time.Second)

    ok := mu.TryLock() // 尝试获取到锁
    if ok {
        // 获取成功
        fmt.Println("got the lock")
        // do something
        mu.Unlock()
        return
    }

    // 没有获取到
    fmt.Println("can't get the lock")
}

func main() {
    try()
}
