package main

import "fmt"

func chanFun1() {
    ch := make(chan int, 10)
    for i := 0; i < 10; i++ {
        select {
        case ch <- i:
        case v := <-ch:
            fmt.Println("chanFun1: ", v)
        }
    }
    close(ch)
}

func chanFun2() {
    ch := make(chan int, 10)
    for i := 0; i < 10; i++ {
        ch <- i
    }
    close(ch)

    for v := range ch {
        fmt.Println("chanFun2: ", v)
    }
}

func chanFun3() {
    ch := make(chan int, 10)
    for i := 0; i < 10; i++ {
        ch <- i
    }
    close(ch)

    // 清空chan
    for range ch {
    }
}

func chanFun4() {
    ch := make(chan int, 10)
    for i := 0; i < 10; i++ {
        ch <- i
    }
    close(ch)

    for i := 0; i < 10; i++ {
        v := <-ch
        fmt.Println("chanFun4: ", v)
    }
}

func chanFun5() {
    ch := make(chan int, 10)
    for i := 0; i < 10; i++ {
        ch <- i
    }
    close(ch)

    // 清空chan
    for range ch {
    }

    for i := 0; i < 10; i++ {
        v, ok := <-ch
        fmt.Println("chanFun5: ", v, ok)
    }
}

func main() {
    chanFun1()
    chanFun2()
    chanFun3()
    chanFun4()
    chanFun5()
}
