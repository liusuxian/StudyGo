package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    rootCtx := context.Background()
    cancelCtx, cancelFunc := context.WithCancel(rootCtx)
    child := context.WithValue(cancelCtx, "name", "lsx")

    go func() {
        for {
            select {
            case <-child.Done():
                fmt.Println("it's over")
                fmt.Println(child.Err())
                return
            default:
                res := child.Value("name")
                fmt.Println("name:", res)
                time.Sleep(1 * time.Second)
            }
        }
    }()

    go func() {
        time.Sleep(3 * time.Second)
        cancelFunc()
    }()

    time.Sleep(5 * time.Second)
}
