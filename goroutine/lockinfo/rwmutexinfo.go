package main

import (
    "fmt"
    "sync"
    "sync/atomic"
    "time"
    "unsafe"
)

const rwmutexMaxReaders = 1 << 30

// 扩展一个RWMutex结构
type RWMutex struct {
    sync.RWMutex
}

// 当前reader的数量
func (m *RWMutex) ReaderCount() int {
    v := m.readerCount()
    if v < 0 {
        return v + rwmutexMaxReaders
    }

    return v
}

// 是否有writer
func (m *RWMutex) IsWriter() bool {
    if m.readerCount() < 0 {
        return true
    }

    return false
}

// writer请求锁时需要等待read完成的reader的数量
func (m *RWMutex) ReaderWait() int {
    // readerWait 这个成员变量前有1个mutex+2个uint32+1个int32
    v := atomic.LoadInt32((*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(&m.RWMutex)) + unsafe.Sizeof(sync.Mutex{}) + 2*unsafe.Sizeof(uint32(0)) + unsafe.Sizeof(int32(0)))))
    return int(v)
}

// 当前 readerCount 的值
func (m *RWMutex) readerCount() int {
    // readerCount 这个成员变量前有1个mutex+2个uint32
    v := atomic.LoadInt32((*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(&m.RWMutex)) + unsafe.Sizeof(sync.Mutex{}) + 2*unsafe.Sizeof(uint32(0)))))
    return int(v)
}

func main() {
    var mu RWMutex
    // 启动1000个goroutine
    for i := 0; i < 1000; i++ {
        go func() {
            mu.RLock()
            time.Sleep(time.Duration(2) * time.Second)
            mu.RUnlock()
        }()
    }

    go func() {
        mu.Lock()
        time.Sleep(time.Second)
        mu.Unlock()
    }()

    time.Sleep(time.Duration(2) * time.Second)
    // 输出锁的信息
    fmt.Printf("ReaderCount: %d, IsWriter: %t, ReaderWait: %d\n", mu.ReaderCount(), mu.IsWriter(), mu.ReaderWait())
    time.Sleep(time.Duration(5) * time.Second)
}
