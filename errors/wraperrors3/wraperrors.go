package main

import (
    "errors"
    "fmt"
    xerrors "github.com/pkg/errors"
)

type QueryError struct {
    Query string
    Err   error
}

func (e *QueryError) Error() string {
    return e.Query
}

func New(text string) error {
    return &QueryError{Query: text}
}

var errVal = errors.New("main: test error")

func test0() error {
    return errVal
}

func test1() error {
    err := test0()
    if err != nil {
        return xerrors.Wrap(err, "test1 error")
    }
    return nil
}

func test2() error {
    err := test1()
    if err != nil {
        return xerrors.WithMessage(err, "test2 error")
    }
    return nil
}

func main() {
    err := test2()
    if err != nil {
        if errors.Is(errVal, xerrors.Cause(err)) {
            fmt.Printf("stack error: %+v\n", err)
        }
    }

    err = New("main: test QueryError")
    if err != nil {
        var e *QueryError
        if errors.As(err, &e) {
            fmt.Println(e.Query, e.Err)
        }
    }
}
