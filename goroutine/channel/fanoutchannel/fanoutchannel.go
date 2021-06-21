package main

func fanOut(ch <-chan interface{}, out []chan interface{}, async bool) {
    go func() {
        defer func() { //退出时关闭所有的输出chan
            for i := 0; i < len(out); i++ {
                close(out[i])
            }
        }()

        for v := range ch { // 从输入chan中读取数据
            v := v
            for i := 0; i < len(out); i++ {
                i := i
                if async { //异步
                    go func() {
                        out[i] <- v // 放入到输出chan中,异步方式
                    }()
                } else {
                    out[i] <- v // 放入到输出chan中，同步方式
                }
            }
        }
    }()
}
