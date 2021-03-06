package main

import (
    "fmt"
    "sync"
)

// Counter 线程安全的计数器类型
type Counter struct {
    CounterType int
    Name        string

    sync.Mutex
    count uint64
}

// Incr 加1的方法，内部使用互斥锁保护
func (c *Counter) Incr() {
    c.Lock()
    c.count++
    c.Unlock()
}

// Count 得到计数器的值，也需要锁保护
func (c *Counter) Count() uint64 {
    c.Lock()
    defer c.Unlock()
    return c.count
}

func main() {
    // 封装好的计数器
    var counter Counter
    // 使用WaitGroup等待10个goroutine完成
    var wg sync.WaitGroup
    wg.Add(10)
    for i := 0; i < 10; i++ {
        go func() {
            defer wg.Done()
            // 执行10万次累加
            for j := 0; j < 100000; j++ {
                counter.Incr() // 受到锁保护的方法
            }
        }()
    }
    // 等待10个goroutine完成
    wg.Wait()
    fmt.Println(counter.count)
}
