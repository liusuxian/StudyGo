package main

import (
    "fmt"
    "sync"
    "sync/atomic"
)

// 一个功能更加强大的Once
type Once struct {
    sync.Mutex
    done uint32
}

// 传入的函数f有返回值error，如果初始化失败，需要返回失败的error
// Do方法会把这个error返回给调用者
func (o *Once) Do(f func() error) error {
    // fast path
    if atomic.LoadUint32(&o.done) == 1 {
        return nil
    }
    return o.slowDo(f)
}

// 如果还没有初始化
func (o *Once) slowDo(f func() error) error {
    o.Lock()
    defer o.Unlock()
    var err error
    // 双检查，还没有初始化
    if o.done == 0 {
        err = f()
        // 初始化成功才将标记置为已初始化
        if err == nil {
            atomic.StoreUint32(&o.done, 1)
        }
    }
    return err
}

// Done 返回此Once是否执行成功过
// 如果执行成功过则返回true
// 如果没有执行成功过或者正在执行，返回false
func (o *Once) Done() bool {
    return atomic.LoadUint32(&o.done) == 1
}

func main() {
    var once Once
    fmt.Println(once.Done()) //false

    // 第一个初始化函数
    f1 := func() error {
        fmt.Println("in f1")
        return nil
    }
    _ = once.Do(f1)          // 打印出 in f1
    fmt.Println(once.Done()) //true

    // 第二个初始化函数
    f2 := func() error {
        fmt.Println("in f2")
        return nil
    }
    _ = once.Do(f2) // 无输出
}
