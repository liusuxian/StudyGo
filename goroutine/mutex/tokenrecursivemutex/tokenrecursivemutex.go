package main

import (
    "fmt"
    "sync"
    "sync/atomic"
)

// TokenRecursiveMutex Token方式的递归锁
type TokenRecursiveMutex struct {
    mu        sync.Mutex
    token     int64 // 当前持有锁的这个goroutine的token
    recursion int32 // 这个goroutine 重入的次数
}

// Lock 请求锁，需要传入token
func (m *TokenRecursiveMutex) Lock(token int64) {
    // 如果传入的token和持有锁的token一致，说明是递归调用
    if atomic.LoadInt64(&m.token) == token {
        m.recursion++
        return
    }
    // 传入的token不一致，说明不是递归调用
    m.mu.Lock()
    // 抢到锁之后记录这个token
    atomic.StoreInt64(&m.token, token)
    m.recursion = 1
}

// Unlock 释放锁
func (m *TokenRecursiveMutex) Unlock(token int64) {
    // 释放其它token持有的锁
    if atomic.LoadInt64(&m.token) != token {
        panic(fmt.Sprintf("wrong the owner(%d): %d!", m.token, token))
    }
    // 当前持有这个锁的token释放锁
    m.recursion--
    // 还没有回退到最初的递归调用
    if m.recursion != 0 {
        return
    }
    atomic.StoreInt64(&m.token, 0) // 没有递归调用了，释放锁
    m.mu.Unlock()
}

func foo(l *TokenRecursiveMutex) {
    fmt.Println("in foo")
    l.Lock(1)
    bar(l) // 重入锁
    l.Unlock(1)
}

func bar(l *TokenRecursiveMutex) {
    l.Lock(1)
    fmt.Println("in bar")
    l.Unlock(1)
}

func main() {
    l := &TokenRecursiveMutex{}
    foo(l)
}
