package myerrors

import (
    "math/rand"
    "time"
)

type myError interface {
    error
    Reconnect() bool
}

func IsReconnect(err error) bool {
    v, ok := err.(myError)
    return ok && v.Reconnect()
}

type errorString struct {
    s string
}

func (e *errorString) Error() string {
    return e.s
}

func (e *errorString) Reconnect() bool {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    if r.Intn(5) >= 3 {
        return true
    }

    return false
}

func New(text string) error {
    return &errorString{s: text}
}
