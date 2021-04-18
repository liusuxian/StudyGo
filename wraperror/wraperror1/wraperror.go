package main

import (
    "fmt"
    xerrors "github.com/pkg/errors"
    "os"
    "path/filepath"
)

func readFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, xerrors.Wrap(err, "open failed")
    }
    defer f.Close()
    return nil, nil
}

func readConfig() ([]byte, error) {
    home := os.Getenv("HOME")
    config, err := readFile(filepath.Join(home, ".settings.xml"))
    return config, xerrors.WithMessage(err, "could not read config")
}

func main() {
    _, err := readConfig()
    if err != nil {
        fmt.Printf("original error: %T %v\n", xerrors.Cause(err), xerrors.Cause(err))
        fmt.Printf("stack trace: %+v\n", err)
    }
}
