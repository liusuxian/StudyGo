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

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    for i := 0; i < 5; i++ {
        go func(ctx context.Context, x int) {
            for {
                if isCancelled(ctx) {
                    break
                }
                time.Sleep(time.Millisecond * 5)
            }
            fmt.Println(x, "Cancelled")
        }(ctx, i)
    }
    cancel()
    time.Sleep(time.Second)
}
