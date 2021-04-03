package main

import (
    "fmt"
    "sync"
)

type SafeSliceQueue struct {
    mu   sync.Mutex
    data []interface{}
}

func NewSafeSliceQueue(n int) (q *SafeSliceQueue) {
    return &SafeSliceQueue{data: make([]interface{}, 0, n)}
}

// Enqueue 把值放在队尾
func (q *SafeSliceQueue) Enqueue(v interface{}) {
    q.mu.Lock()
    q.data = append(q.data, v)
    q.mu.Unlock()
}

// Dequeue 移去队头并返回
func (q *SafeSliceQueue) Dequeue() interface{} {
    q.mu.Lock()
    if len(q.data) == 0 {
        q.mu.Unlock()
        return nil
    }
    v := q.data[0]
    q.data = q.data[1:]
    q.mu.Unlock()
    return v
}

func main() {
    queue := NewSafeSliceQueue(10)
    // 使用WaitGroup等待10个goroutine完成
    var wg sync.WaitGroup
    wg.Add(10)
    for i := 0; i < 5; i++ {
        go func(x int) {
            defer wg.Done()
            queue.Enqueue(x)
        }(i)
    }
    for i := 0; i < 5; i++ {
        go func() {
            defer wg.Done()
            fmt.Println(queue.Dequeue())
        }()
    }

    // 等待10个goroutine完成
    wg.Wait()
    fmt.Println(queue.data)
}
