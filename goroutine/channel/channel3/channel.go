package main

import (
    "fmt"
    "time"
)

func main() {
    ch1 := make(chan int)
    ch2 := make(chan int)
    ch3 := make(chan int)
    ch4 := make(chan int)

    go func() {
        for {
            fmt.Println("I'm goroutine 1")
            time.Sleep(time.Second)
            ch2 <- 1 //I'm done, you turn
            <-ch1
        }
    }()

    go func() {
        for {
            <-ch2
            fmt.Println("I'm goroutine 2")
            time.Sleep(time.Second)
            ch3 <- 1
        }

    }()

    go func() {
        for {
            <-ch3
            fmt.Println("I'm goroutine 3")
            time.Sleep(time.Second)
            ch4 <- 1
        }

    }()

    go func() {
        for {
            <-ch4
            fmt.Println("I'm goroutine 4")
            time.Sleep(time.Second)
            ch1 <- 1
        }

    }()

    select {}
}
