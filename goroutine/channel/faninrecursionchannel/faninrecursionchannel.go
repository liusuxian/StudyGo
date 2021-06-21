package main

func fanInRec(chans ...<-chan interface{}) <-chan interface{} {
    switch len(chans) {
    case 0:
        c := make(chan interface{})
        close(c)
        return c
    case 1:
        return chans[0]
    case 2:
        return mergeTwo(chans[0], chans[1])
    default:
        m := len(chans) / 2
        return mergeTwo(
            fanInRec(chans[:m]...),
            fanInRec(chans[m:]...))
    }
}

func mergeTwo(a, b <-chan interface{}) <-chan interface{} {
    c := make(chan interface{})
    go func() {
        defer close(c)
        for a != nil || b != nil { //只要还有可读的chan
            select {
            case v, ok := <-a:
                if !ok { // a 已关闭，设置为nil
                    a = nil
                    continue
                }
                c <- v
            case v, ok := <-b:
                if !ok { // b 已关闭，设置为nil
                    b = nil
                    continue
                }
                c <- v
            }
        }
    }()

    return c
}
