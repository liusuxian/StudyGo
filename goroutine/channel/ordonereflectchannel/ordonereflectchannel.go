package main

import (
    "fmt"
    "reflect"
    "time"
)

func or(channels ...<-chan interface{}) <-chan interface{} {
    //特殊情况，只有0个或者1个
    switch len(channels) {
    case 0:
        return nil
    case 1:
        return channels[0]
    }

    orDone := make(chan interface{})
    go func() {
        defer close(orDone)
        // 利用反射构建SelectCase
        var cases []reflect.SelectCase
        for _, c := range channels {
            cases = append(cases, reflect.SelectCase{
                Dir:  reflect.SelectRecv,
                Chan: reflect.ValueOf(c),
            })
        }

        // 随机选择一个可用的case
        reflect.Select(cases)
    }()

    return orDone
}

func sig(after time.Duration) <-chan interface{} {
    c := make(chan interface{})
    go func() {
        defer close(c)
        time.Sleep(after)
    }()

    return c
}

func main() {
    start := time.Now()

    <-or(
        sig(10*time.Second),
        sig(20*time.Second),
        sig(30*time.Second),
        sig(40*time.Second),
        sig(50*time.Second),
        sig(01*time.Minute),
    )

    fmt.Printf("done after %v", time.Since(start))
}
