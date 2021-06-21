package main

import "reflect"

func fanInReflect(chans ...<-chan interface{}) <-chan interface{} {
    out := make(chan interface{})
    go func() {
        defer close(out)
        // 构造SelectCase slice
        var cases []reflect.SelectCase
        for _, c := range chans {
            cases = append(cases, reflect.SelectCase{
                Dir:  reflect.SelectRecv,
                Chan: reflect.ValueOf(c),
            })
        }

        // 循环，从cases中选择一个可用的
        for len(cases) > 0 {
            i, v, ok := reflect.Select(cases)
            if !ok { // 此channel已经close
                cases = append(cases[:i], cases[i+1:]...)
                continue
            }
            out <- v.Interface()
        }
    }()

    return out
}
