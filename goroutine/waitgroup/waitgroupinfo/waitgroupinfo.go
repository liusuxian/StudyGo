package main

import (
    "fmt"
    "math/rand"
    "sync"
    "time"
    "unsafe"
)

// 扩展一个WaitGroup结构
type WaitGroup struct {
    sync.WaitGroup
}

// 查询 WaitGroup 的当前的计数值
func (wg *WaitGroup) GetCounter() uint32 {
    pointer := unsafe.Pointer(&wg.WaitGroup)
    if (uintptr(pointer)+unsafe.Sizeof(struct{}{}))%8 == 0 {
        // 如果地址是64bit对齐的，数组前两个元素做state，后一个元素做信号量
        return *(*uint32)(unsafe.Pointer(uintptr(pointer) + 4))
    } else {
        // 如果地址是32bit对齐的，数组后两个元素用来做state，它可以用来做64bit的原子操作，第一个元素32bit用来做信号量
        return *(*uint32)(unsafe.Pointer(uintptr(pointer) + 8))
    }
}

// 查询 WaitGroup 的当前的 waiter 数
func (wg *WaitGroup) GetWaiter() uint32 {
    pointer := unsafe.Pointer(&wg.WaitGroup)
    if (uintptr(pointer)+unsafe.Sizeof(struct{}{}))%8 == 0 {
        // 如果地址是64bit对齐的，数组前两个元素做state，后一个元素做信号量
        return *(*uint32)(pointer)
    } else {
        // 如果地址是32bit对齐的，数组后两个元素用来做state，它可以用来做64bit的原子操作，第一个元素32bit用来做信号量
        return *(*uint32)(unsafe.Pointer(uintptr(pointer) + 4))
    }
}

func main() {
    // 使用WaitGroup等待10个goroutine完成
    var wg WaitGroup
    wg.Add(10)
    for i := 0; i < 10; i++ {
        go func() {
            defer wg.Done()
            time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
            fmt.Printf("111 GetCounter: %d, GetWaiter: %d\n", wg.GetCounter(), wg.GetWaiter())
        }()
    }
    // 等待10个goroutine完成
    wg.Wait()
    wg.Add(19999)
    fmt.Printf("222 GetCounter: %d, GetWaiter: %d\n", wg.GetCounter(), wg.GetWaiter())
}
