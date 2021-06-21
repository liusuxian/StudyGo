package main

import (
    "fmt"
    "time"
)

func main() {
    ch := make(chan int)
    for i := 1; i < 5; i++ {
        go func(id int) {
            time.Sleep(time.Duration(id*10) * time.Millisecond)
            for {
                <-ch
                fmt.Printf("I am No %d Goroutine\n", id)
                time.Sleep(time.Second)
                ch <- 1
            }
        }(i)
    }
    ch <- 1
    select {}
}
