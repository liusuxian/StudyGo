package main

import (
    "fmt"
    "time"
)

type Token struct{}

func newWorker(id int, ch chan Token, nextCh chan Token) {
    for {
        token := <-ch         // 取得令牌
        fmt.Println((id + 1)) // id从1开始
        time.Sleep(time.Second)
        nextCh <- token
    }
}

func main() {
    chs := []chan Token{make(chan Token), make(chan Token), make(chan Token), make(chan Token)}

    // 创建4个worker
    for i := 0; i < 4; i++ {
        go newWorker(i, chs[i], chs[(i+1)%4])
    }

    // 首先把令牌交给第一个worker
    chs[0] <- struct{}{}

    select {}
}
