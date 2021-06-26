package main

import (
    "fmt"
    "golang.org/x/sync/singleflight"
    "log"
    "sync"
    "sync/atomic"
    "time"
)

var (
    sf           = singleflight.Group{}
    requestCount = int64(0)
    resp         = make(chan int64, 0)
    wg           sync.WaitGroup
)

func main() {
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            do, err, _ := sf.Do("number", Request)
            if err != nil {
                log.Println(err)
            }
            log.Println("resp", do)
            defer wg.Done()
        }()
    }
    time.Sleep(time.Second)
    resp <- atomic.LoadInt64(&requestCount)
    wg.Wait()
}

func Request() (interface{}, error) {
    fmt.Println("Request")
    atomic.AddInt64(&requestCount, 1)
    return <-resp, nil
}
