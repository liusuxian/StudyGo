package main

import (
    "context"
    "errors"
    "fmt"
    "github.com/vardius/gollback"
    "time"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 尝试5次，或者超时返回
    res, err := gollback.Retry(ctx, 5, func(ctx context.Context) (interface{}, error) {
        return nil, errors.New("failed")
    })

    fmt.Println(res) // 输出结果
    fmt.Println(err) // 输出错误信息
}
