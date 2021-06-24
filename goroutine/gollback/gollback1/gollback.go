package main

import (
    "context"
    "errors"
    "fmt"
    "github.com/vardius/gollback"
    "time"
)

func main() {
    rs, errs := gollback.All( // 调用All方法
        context.Background(),
        func(ctx context.Context) (interface{}, error) {
            time.Sleep(3 * time.Second)
            return 1, nil // 第一个任务没有错误，返回1
        },
        func(ctx context.Context) (interface{}, error) {
            return nil, errors.New("failed") // 第二个任务返回一个错误
        },
        func(ctx context.Context) (interface{}, error) {
            return 3, nil // 第三个任务没有错误，返回3
        },
    )

    fmt.Println(rs)   // 输出子任务的结果
    fmt.Println(errs) // 输出子任务的错误信息
}
