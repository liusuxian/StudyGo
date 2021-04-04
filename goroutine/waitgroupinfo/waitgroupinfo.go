package main

import (
    "fmt"
    "sync"
    "unsafe"
)

// 扩展一个WaitGroup结构
type WaitGroup struct {
    sync.WaitGroup
}

// 查询WaitGroup的当前的计数值
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

func main() {
    a := WaitGroup{}
    a.Add(19999)
    fmt.Println(a.GetCounter())
}
