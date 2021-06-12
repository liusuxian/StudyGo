package main

import (
    "fmt"
    "net"
)

func main() {
    // 创建udp地址
    udpAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8081")
    // 服务端建立监听
    serverConn, _ := net.ListenUDP("udp", udpAddr)
    defer func() {
        _ = serverConn.Close()
        fmt.Println("main over!")
    }()

    // 与客户端IO
    buffer := make([]byte, 1024)
    n, remoteAddress, _ := serverConn.ReadFromUDP(buffer)
    contents := buffer[:n]
    fmt.Println("客户端：" + string(contents))

    // 回复客户端消息
    _, _ = serverConn.WriteToUDP([]byte("孽障！"), remoteAddress)
}
