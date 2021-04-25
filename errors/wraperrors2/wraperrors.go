package main

import (
    "fmt"
    "github.com/pkg/errors"
)

func test0() error {
    return errors.New("main: test0 error")
}

func test1() error {
    err := test0()
    if err != nil {
        return err
    }
    return nil
}

func test00() error {
    return errors.Errorf("main: %s error\n", "test00")
}

func test11() error {
    err := test00()
    if err != nil {
        return err
    }
    return nil
}

func main() {
    err := test1()
    if err != nil {
        fmt.Printf("fatal: %+v\n", err)
    }

    err = test11()
    if err != nil {
        fmt.Printf("fatal: %+v\n", err)
    }
}
