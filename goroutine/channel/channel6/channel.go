package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup
    wg.Add(2)
    stop := make(chan struct{})

    go func(wg *sync.WaitGroup) {
        defer wg.Done()
        defer fmt.Println(1111)
        defer fmt.Println(2222)
        <-stop
        fmt.Println("Exit1")
    }(&wg)

    go func(wg *sync.WaitGroup) {
        defer wg.Done()
        <-stop
        fmt.Println("Exit2")
    }(&wg)

    close(stop)
    wg.Wait()
}
