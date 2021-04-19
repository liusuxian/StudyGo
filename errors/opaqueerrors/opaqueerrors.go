package main

import (
    "fmt"
    "opaqueerrors/myerrors"
    "os"
)

func test() error {
    return myerrors.New("main: test error")
}

func main() {
    err := test()
    if err != nil && myerrors.IsReconnect(err) {
        fmt.Println("IsReconnect!!!")
        fmt.Println(err.Error())
        os.Exit(-1)
    }
    fmt.Println("nothing!!!")
}
