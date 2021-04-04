package main

import (
    "fmt"
    "sync"
    "sync/atomic"
)

type Once struct {
    done uint32
    m    Mutex
}



func main() {
    var once sync.Once
    // 第一个初始化函数
    f1 := func() {
        fmt.Println("in f1")
    }
    once.Do(f1) // 打印出 in f1

    // 第二个初始化函数
    f2 := func() {
        fmt.Println("in f2")
    }
    once.Do(f2) // 无输出
}
