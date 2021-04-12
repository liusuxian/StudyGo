package main

import (
    "context"
    "fmt"
    "time"
)

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
    cancelCtx, cancelFunc := context.WithDeadline(rootCtx, time.Now().Add(time.Second*5))
    defer cancelFunc()
    go task(cancelCtx)
    time.Sleep(time.Second * 6)
}
