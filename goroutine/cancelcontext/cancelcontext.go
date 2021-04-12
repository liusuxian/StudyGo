package main

import (
    "context"
    "fmt"
    "time"
)

func isCancelled(ctx context.Context) bool {
    select {
    case <-ctx.Done():
        return true
    default:
        return false
    }
}

func task(ctx context.Context) {
    i := 1
    for {
        select {
        case <-ctx.Done():
            fmt.Println("Gracefully exit")
            fmt.Println(ctx.Err())
            return
        default:
            fmt.Println(i)
            time.Sleep(time.Second * 1)
            i++
        }
    }
}

func main() {
    rootCtx := context.Background()
    cancelCtx1, cancelFunc1 := context.WithCancel(rootCtx)
    for i := 0; i < 5; i++ {
        go func(ctx context.Context, x int) {
            for {
                if isCancelled(ctx) {
                    break
                }
                time.Sleep(time.Millisecond * 5)
            }
            fmt.Println(x, "Cancelled")
        }(cancelCtx1, i)
    }
    cancelFunc1()
    time.Sleep(time.Second)

    cancelCtx, cancelFunc := context.WithCancel(rootCtx)
    go task(cancelCtx)
    time.Sleep(time.Second * 3)
    cancelFunc()
    time.Sleep(time.Second * 1)
}
