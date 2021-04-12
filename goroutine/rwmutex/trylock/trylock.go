package main

import (
    "fmt"
    "math/rand"
    "sync"
    "sync/atomic"
    "time"
    "unsafe"
)

// 扩展一个RWMutex结构
type RWMutex struct {
    sync.RWMutex
}

// 当前 readerCount 的值
func (m *RWMutex) readerCount() int {
    // readerCount 这个成员变量前有1个mutex+2个uint32
    v := atomic.LoadInt32((*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(&m.RWMutex)) + unsafe.Sizeof(sync.Mutex{}) + 2*unsafe.Sizeof(uint32(0)))))
    return int(v)
}

// 尝试获取写锁
func (m *RWMutex) TryLock() bool {
    if m.readerCount() < 0 {
        // 已经有写锁了
        return false
    }

    // 尝试获取写锁
    m.Lock()
    return true
}

func main() {
    var mu RWMutex
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
