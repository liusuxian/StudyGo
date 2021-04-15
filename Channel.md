### CSP模型。
- CSP是Communicating Sequential Process的简称，中文直译为通信顺序进程，或者叫做交换信息的循序进程，是用来描述并发系统中进行交互的一种模式。
- CSP允许使用进程组件来描述系统，它们独立运行，并且只通过消息传递的方式通信。
### Channel的应用场景。
- 数据交流：当作并发的buffer或者queue，解决生产者-消费者问题。多个goroutine可以并发当作生产者（Producer）和消费者（Consumer）。
- 数据传递：一个goroutine将数据交给另一个goroutine，相当于把数据的拥有权 (引用) 托付出去。
- 信号通知：一个goroutine可以将信号 (closing、closed、data ready 等) 传递给另一个或者另一组goroutine。
- 任务编排：可以让一组goroutine按照一定的顺序并发或者串行的执行，这就是编排的功能。
- 锁：利用Channel也可以实现互斥锁的机制。
### Channel基本用法。
- Channel类型分为只能接收、只能发送、既可以接收又可以发送三种类型，如下。我们把既能接收又能发送的chan叫做双向的chan，把只能发送和只能接收的chan叫做单向的chan。其中“<-”表示单向的chan，这个箭头总是射向左边的，元素类型总在最右边。如果箭头指向chan，就表示可以往chan中塞数据；如果箭头远离chan，就表示chan会往外吐数据。
``` go
chan string          // 可以发送接收string
chan<- struct{}      // 只能发送struct{}
<-chan int           // 只能从chan接收int
```
- chan中的元素是任意的类型，所以也可能是chan类型，比如下面的chan类型也是合法的：
``` go
chan<- chan int   
chan<- <-chan int  
<-chan <-chan int
chan (<-chan int)
```
- 如何判定箭头符号属于哪个chan，其实“<-”有个规则，总是尽量和左边的chan结合，如下：
``` go
chan<- （chan int） // <- 和第一个chan结合
chan<- （<-chan int） // 第一个<-和最左边的chan结合，第二个<-和左边第二个chan结合
<-chan （<-chan int） // 第一个<-和最左边的chan结合，第二个<-和左边第二个chan结合 
chan (<-chan int) // 因为括号的原因，<-和括号内第一个chan结合
```
- 通过make，我们可以初始化一个chan，未初始化的chan的零值是nil。你可以设置它的容量，我们把这样的chan叫做buffered chan；如果没有设置，它的容量是0，我们把这样的chan叫做unbuffered chan。
- 如果chan中还有数据，那么从这个chan接收数据的时候就不会阻塞，如果chan还未满（“满”指达到其容量），给它发送数据也不会阻塞，否则就会阻塞。unbuffered chan只有读写都准备好之后才不会阻塞。
- nil是chan的零值，是一种特殊的chan，对值是nil的chan的发送接收调用者总是会阻塞。
### 发送数据。
- 往chan中发送一个数据使用“ch<-”，发送数据是一条语句，这里的ch是chan int类型或者是chan <-int。
``` go
ch <- 2000
```
### 接收数据。
- 从chan中接收一条数据使用“<-ch”，接收数据也是一条语句，这里的ch类型是chan T或者<-chan T。接收数据时，还可以返回两个值。第一个值是返回的chan中的元素，第二个值是bool类型，代表是否成功地从chan中读取到一个值，如果第二个参数是false，chan已经被close而且chan中没有缓存的数据，这个时候，第一个值是零值。所以如果从chan读取到一个零值，可能是sender真正发送的零值，也可能是closed的并且没有缓存元素产生的零值。
``` go
x := <-ch // 把接收的一条数据赋值给变量x
foo(<-ch) // 把接收的一个的数据作为参数传给函数
<-ch // 丢弃接收的一条数据
```
### 其它操作。
- Go内建的函数close、cap、len都可以操作chan类型：close会把chan关闭掉，cap返回chan的容量，len返回chan中缓存的还未被取走的元素数量。send和recv都可以作为select语句的case clause，如下面的例子：
``` go
func main() {
    ch := make(chan int, 10)
    for i := 0; i < 10; i++ {
        select {
        case ch <- i:
        case v := <-ch:
            fmt.Println(v)
        }
    }
}
```
- chan还可以应用于for-range语句中，比如：
``` go
for v := range ch {
    fmt.Println(v)
}
```
- 或者是忽略读取的值，只是清空chan：
``` go
for range ch {
}
```
### chan数据结构。
- chan类型的数据结构如下图所示，它的数据类型是[runtime.hchan](https://github.com/golang/go/blob/master/src/runtime/chan.go#L32)。
![avatar](https://github.com/liusuxian/learning_golang/blob/master/img/Channel.jpg)