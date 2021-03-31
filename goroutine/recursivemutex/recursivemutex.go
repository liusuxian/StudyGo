package main

import (
    "fmt"
    "github.com/petermattis/goid"
    "runtime"
    "strconv"
    "strings"
    "sync"
    "sync/atomic"
)

// 获取 goroutine id
func GoID() int64 {
    var buf [64]byte
    n := runtime.Stack(buf[:], false)
    // 得到id字符串
    idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
    id, err := strconv.Atoi(idField)
    if err != nil {
        panic(fmt.Sprintf("cannot get goroutine id: %v", err))
    }
    return int64(id)
}

// RecursiveMutex 包装一个Mutex，实现可重入
type RecursiveMutex struct {
    mu        sync.Mutex
    owner     int64 // 当前持有锁的goroutine id
    recursion int32 // 这个goroutine 重入的次数
}

// Lock
func (m *RecursiveMutex) Lock() {
    gid := goid.Get()
    // 如果当前持有锁的goroutine就是这次调用的goroutine,说明是重入
    if atomic.LoadInt64(&m.owner) == gid {
        m.recursion++
        return
    }
    m.mu.Lock()
    // 获得锁的goroutine第一次调用，记录下它的goroutine id,调用次数加1
    atomic.StoreInt64(&m.owner, gid)
    m.recursion = 1
}

// Unlock
func (m *RecursiveMutex) Unlock() {
    gid := goid.Get()
    // 非持有锁的goroutine尝试释放锁，错误的使用
    if atomic.LoadInt64(&m.owner) != gid {
        panic(fmt.Sprintf("wrong the owner(%d): %d!", m.owner, gid))
    }
    // 调用次数减1
    m.recursion--
    // 如果这个goroutine还没有完全释放，则直接返回
    if m.recursion != 0 {
        return
    }
    // 此goroutine最后一次调用，需要释放锁
    atomic.StoreInt64(&m.owner, -1)
    m.mu.Unlock()
}

func foo(l *RecursiveMutex) {
    fmt.Println("in foo")
    l.Lock()
    bar(l) // 重入锁
    l.Unlock()
}

func bar(l *RecursiveMutex) {
    l.Lock()
    fmt.Println("in bar")
    l.Unlock()
}

func main() {
    l := &RecursiveMutex{}
    foo(l)
}
