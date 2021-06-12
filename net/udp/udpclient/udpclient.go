package main

import (
    "fmt"
    "net"
)

func main() {
    conn, _ := net.Dial("udp", "127.0.0.1:8081")
    defer func() {
        _ = conn.Close()
        fmt.Println("客户端已退出")
    }()

    // 客户端发起交谈
    _, _ = conn.Write([]byte("你妹，今天天气不错"))

    // 接收服务端消息
    buffer := make([]byte, 1024)
    n, _ := conn.Read(buffer)

    fmt.Println("服务端：" + string(buffer[:n]))
}
