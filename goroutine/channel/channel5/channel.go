package main

import (
    "fmt"
    "sync"
)

type Token struct{}

func newWorker(wg *sync.WaitGroup, word string, maxCount int, ch chan Token, nextCh chan Token) {
    defer wg.Done()
    count := 0 // 统计打印次数
    for {
        token := <-ch // 取得令牌
        fmt.Println(word)
        count++
        select {
        case _, ok := <-nextCh:
            if !ok {
                // nextCh 已经关闭
            }
        default:
            nextCh <- token
        }

        if count >= maxCount {
            close(ch)
            return
        }
    }
}

func main() {
    chs := []chan Token{make(chan Token), make(chan Token), make(chan Token)}
    words := []string{"cat", "dog", "fish"}

    // 创建3个worker
    var wg sync.WaitGroup
    wg.Add(len(words))

    for k, v := range words {
        go newWorker(&wg, v, 10, chs[k], chs[(k+1)%3])
    }

    // 首先把令牌交给第一个worker
    chs[0] <- struct{}{}
    wg.Wait()
}
