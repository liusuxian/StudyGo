package main

import (
    "fmt"
    "sync"
)

func foo(l sync.Locker) {
    fmt.Println("in foo")
    l.Lock()
    bar(l) // 重入锁
    l.Unlock()
}

func bar(l sync.Locker) {
    l.Lock()
    fmt.Println("in bar")
    l.Unlock()
}

// 重入锁问题
func main() {
    l := &sync.Mutex{}
    foo(l)
}
