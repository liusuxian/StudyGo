package main

import (
    "fmt"
    "sync"
    "time"
)

// Counter 一个线程安全的计数器
type Counter struct {
    sync.RWMutex
    count uint64
}

// Incr 使用写锁保护
func (c *Counter) Incr() {
    c.Lock()
    c.count++
    c.Unlock()
}

// Count 使用读锁保护
func (c *Counter) Count() uint64 {
    c.RLock()
    defer c.RUnlock()
    return c.count
}

func main() {
    var counter Counter
    // 10个reader
    for i := 0; i < 10; i++ {
        go func() {
            for {
                // 计数器读操作
                fmt.Println(counter.Count())
                time.Sleep(time.Millisecond)
            }
        }()
    }

    // 一个writer
    for {
        counter.Incr() // 计数器写操作
        time.Sleep(time.Second)
    }
}
