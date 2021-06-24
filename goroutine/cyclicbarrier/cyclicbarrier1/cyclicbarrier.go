package main

import (
    "context"
    "github.com/marusama/cyclicbarrier"
    "golang.org/x/sync/semaphore"
    "log"
    "math/rand"
    "sort"
    "sync"
    "time"
)

// H2O 定义水分子合成的辅助数据结构
type H2O struct {
    semaH *semaphore.Weighted         // 氢原子的信号量
    semaO *semaphore.Weighted         // 氧原子的信号量
    b     cyclicbarrier.CyclicBarrier // 循环栅栏，用来控制合成
}

func New() *H2O {
    return &H2O{
        semaH: semaphore.NewWeighted(2), // 氢原子需要两个
        semaO: semaphore.NewWeighted(1), // 氧原子需要一个
        b:     cyclicbarrier.New(3),     // 需要三个原子才能合成
    }
}

func (h2o *H2O) hydrogen(releaseHydrogen func()) {
    _ = h2o.semaH.Acquire(context.Background(), 1)

    releaseHydrogen()                     // 输出H
    _ = h2o.b.Await(context.Background()) // 等待栅栏放行
    h2o.semaH.Release(1)                  // 释放氢原子空槽
}

func (h2o *H2O) oxygen(releaseOxygen func()) {
    _ = h2o.semaO.Acquire(context.Background(), 1)

    releaseOxygen()                       // 输出O
    _ = h2o.b.Await(context.Background()) // 等待栅栏放行
    h2o.semaO.Release(1)                  // 释放氧原子空槽
}

func main() {
    // 用来存放水分子结果的channel
    var ch chan string
    releaseHydrogen := func() {
        ch <- "H"
    }
    releaseOxygen := func() {
        ch <- "O"
    }

    // 300个原子，300个goroutine，每个goroutine并发的产生一个原子
    var N = 100
    ch = make(chan string, N*3)
    h2o := New()

    // 用来等待所有的goroutine完成
    var wg sync.WaitGroup
    wg.Add(N * 3)

    // 200个氢原子goroutine
    for i := 0; i < 2*N; i++ {
        go func() {
            time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
            h2o.hydrogen(releaseHydrogen)
            wg.Done()
        }()
    }
    // 100个氧原子goroutine
    for i := 0; i < N; i++ {
        go func() {
            time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
            h2o.oxygen(releaseOxygen)
            wg.Done()
        }()
    }

    // 等待所有的goroutine执行完
    wg.Wait()

    // 结果中肯定是300个原子
    if len(ch) != N*3 {
        log.Fatalf("expect %d atom but got %d", N*3, len(ch))
    }

    // 每三个原子一组，分别进行检查。要求这一组原子中必须包含两个氢原子和一个氧原子，这样才能正确组成一个水分子。
    var s = make([]string, 3)
    for i := 0; i < N; i++ {
        s[0] = <-ch
        s[1] = <-ch
        s[2] = <-ch
        sort.Strings(s)

        water := s[0] + s[1] + s[2]
        if water != "HHO" {
            log.Fatalf("expect a water molecule but got %s", water)
        }
        log.Println(water)
    }
}
