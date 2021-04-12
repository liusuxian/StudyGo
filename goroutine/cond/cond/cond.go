package main

import (
    "fmt"
    "math/rand"
    "sync"
    "time"
)

func main() {
    c := sync.NewCond(&sync.Mutex{})
    var ready int

    for i := 0; i < 10; i++ {
        go func(i int) {
            time.Sleep(time.Duration(rand.Int63n(10)) * time.Second)

            // 加锁更改等待条件
            c.L.Lock()
            ready++
            c.L.Unlock()

            fmt.Printf("运动员#%d 已准备就绪\n", i)
            // 广播唤醒所有的等待者
            // 因为裁判员只有一个，所以这里也可以直接替换成Signal方法调用
            c.Broadcast()
        }(i)
    }

    c.L.Lock()
    for ready != 10 {
        c.Wait()
        fmt.Println("裁判员被唤醒一次")
    }
    c.L.Unlock()

    // 所有的运动员是否就绪
    fmt.Println("所有运动员都准备就绪。比赛开始，3，2，1, ......")
}
