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
<img src="https://github.com/liusuxian/StudyGo/blob/master/img/Channel.jpg" width = "60%" height = "60%" alt="image-name"/>

- qcount：代表chan中已经接收但还没被取走的元素的个数。内建函数len可以返回这个字段的值。
- dataqsiz：队列的大小。chan使用一个循环队列来存放元素，循环队列很适合这种生产者-消费者的场景。
- buf：存放元素的循环队列的buffer。
- elemtype和elemsize：chan中元素的类型和size。因为chan一旦声明，它的元素类型是固定的，即普通类型或者指针类型，所以元素大小也是固定的。
- sendx：处理发送数据的指针在buf中的位置。一旦接收了新的数据，指针就会加上elemsize移向下一个位置。buf的总大小是elemsize的整数倍，而且buf是一个循环列表。
- recvx：处理接收请求时的指针在buf中的位置。一旦取出数据，此指针会移动到下一个位置。
- recvq：chan是多生产者多消费者的模式，如果消费者因为没有数据可读而被阻塞了，就会被加入到recvq队列中。
- sendq：如果生产者因为buf满了而阻塞，会被加入到sendq队列中。
### 初始化。
- Go在编译的时候，会根据容量的大小选择调用makechan64还是makechan。makechan64只是做了size检查，底层还是调用makechan实现的。makechan的目标就是生成hchan对象。
### send。
- Go在编译发送数据给chan的时候，会把send语句转换成chansend1函数，chansend1函数会调用chansend。1、如果chan是nil的话，就把调用者永远阻塞。2、如果往一个已经满了的chan实例发送数据时，并且想不阻塞当前调用，那么直接返回。chansend1方法在调用chansend的时候设置了阻塞参数。3、如果chan已经被close了，再往里面发送数据的话会panic。4、如果等待队列中有等待的receiver，那么就把它从队列中弹出，然后直接把数据交给它，而不需要放入到buf中，速度可以更快一些。5、如果当前没有receiver，需要把数据放入到buf中，放入之后就成功返回了。6、如果buf满了，发送者的goroutine就会加入到发送者的等待队列中，直到被唤醒。这个时候数据或者被取走了，或者chan被close了。
### recv。
- 在处理从chan中接收数据时，Go会把代码转换成chanrecv1函数，如果要返回两个返回值，会转换成chanrecv2，chanrecv1函数和chanrecv2会调用chanrecv。chanrecv1和chanrecv2传入的block参数的值是true，都是阻塞方式。1、chan为nil的情况和send一样，从nil chan中接收（读取、获取）数据时，调用者会被永远阻塞。2、如果chan已经被close了，并且队列中没有缓存的元素，那么将得到零值。3、如果buf满了。这个时候如果是unbuffer的chan，就直接将sender的数据复制给receiver，否则就从队列头部读取一个值，并把这个sender的值加入到队列尾部。4、如果没有等待的sender的情况，这个是和chansend共用一把大锁，所以不会有并发的问题，如果buf有元素，就取出一个元素给receiver。5、如果buf中没有元素，那么当前的receiver就会被阻塞，直到它从sender中接收了数据，或者是chan被close才返回。
### close。
- 通过close函数，可以把chan关闭，编译器会替换成closechan方法的调用。1、如果chan为nil，close会panic；2、如果chan已经closed，再次close也会panic。3、如果chan不为nil，chan也没有closed，就把等待队列中的sender（writer）和 receiver（reader）从队列中全部移除并唤醒。
### 使用Channel最常见的错误是panic和goroutine泄漏。
- close为nil的chan，会panic。
- close已经close的chan，会panic。
- send已经close的chan，会panic。
### 选择Channel还是选择并发原语的方法。
- 共享资源的并发访问使用传统并发原语。
- 复杂的任务编排和消息传递使用Channel。
- 消息通知机制使用Channel，除非只想signal一个goroutine才使用Cond。
- 简单等待所有任务的完成用WaitGroup，也有Channel的推崇者用Channel，都可以。
- 需要和Select语句结合，使用Channel。 
- 需要和超时配合时，使用Channel和Context。
### Channel不同状态下各种操作的结果。
<img src="https://github.com/liusuxian/StudyGo/blob/master/img/Channel1.jpg" width = "60%" height = "60%" alt="image-name"/>

### 使用反射操作Channel。
- 通过反射的方式执行select语句，在处理很多的case clause，尤其是不定长的case clause的时候，非常有用。任务编排的实现，也可以用这种方法。
### 典型的应用场景。
- 消息交流。从chan的内部实现看，它是以一个循环队列的方式存放数据，所以它有时候也会被当成线程安全的队列和buffer使用。一个goroutine可以安全地往Channel中塞数据，另外一个goroutine可以安全地从Channel中读取数据，goroutine就可以安全地实现信息交流了。比如worker池的例子，Marcio Castilho [使用Go每分钟处理百万请求](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/) 这篇文章中，就介绍了他们应对大并发请求的设计。
- 数据传递。这类场景有一个特点，就是当前持有数据的goroutine都有一个信箱，信箱使用chan实现，goroutine只需要关注自己的信箱中的数据，处理完毕后，就把结果发送到下一家的信箱中。
- 信号通知。chan类型有这样一个特点：chan如果为空，那么receiver接收数据的时候就会阻塞等待，直到chan被关闭或者有新的数据到来。利用这个机制，我们可以实现wait/notify的设计模式。传统的并发原语Cond也能实现这个功能。但是Cond使用起来比较复杂，容易出错，而使用chan实现wait/notify模式，就方便多了。除了正常的业务处理时的wait/notify，我们经常碰到的一个场景，就是程序关闭的时候，我们需要在退出之前做一些清理的动作。这个时候，我们经常要使用chan。比如使用chan实现程序的graceful shutdown，在退出之前执行一些连接关闭、文件close、缓存落盘等一些动作。有时候清理可能是一个很耗时的操作，比如十几分钟才能完成，如果程序退出需要等待这么长时间，用户是不能接受的，所以在实践中，我们需要设置一个最长的等待时间。只要超过了这个时间，程序就不再等待，可以直接退出。所以退出的时候分为两个阶段：closing代表程序退出，但是清理工作还没做；closed代表清理工作已经做完。
- 锁。在chan的内部实现中，就有一把互斥锁保护着它的所有字段。从外在表现上，chan的发送和接收之间也存在着happens-before的关系，保证元素放进去之后，receiver才能读取到（关于happends-before的关系，是指事件发生的先后顺序关系）。要想使用chan实现互斥锁，至少有两种方式。一种方式是先初始化一个capacity等于1的Channel，然后再放入一个元素。这个元素就代表锁，谁取得了这个元素，就相当于获取了这把锁。另一种方式是，先初始化一个capacity等于1的Channel，它的“空槽”代表锁，谁能成功地把元素发送到这个Channel谁就获取了这把锁。
- 任务编排。1、Or-Done模式，Or-Done模式是信号通知模式中更宽泛的一种模式。这里提到了“信号通知模式”。我们会使用“信号通知”实现某个任务执行完成后的通知机制，在实现时，我们为这个任务定义一个类型为chan struct{}类型的done变量，等任务结束后，我们就可以close这个变量，然后其它receiver就会收到这个通知。这是有一个任务的情况，如果有多个任务，只要有任意一个任务执行完，我们就想获得这个信号，这就是Or-Done模式。比如你发送同一个请求到多个微服务节点，只要任意一个微服务节点返回结果，就算成功。可以使用递归、反射，或者是用最笨的每个goroutine处理一个Channel的方式来实现。2、扇入模式。在软件工程中，模块的扇入是指有多少个上级模块调用它。而对于我们这里的Channel扇入模式来说，就是指有多个源Channel输入、一个目的Channel输出的情况。扇入比就是源Channel数量比1。每个源Channel的元素都会发送给目标Channel，相当于目标Channel的receiver只需要监听目标Channel，就可以接收所有发送给源Channel的数据。扇入模式也可以使用反射、递归，或者是用最笨的每个goroutine处理一个Channel的方式来实现。3、扇出模式。扇出模式只有一个输入源Channel，有多个目标Channel，扇出比就是1比目标Channel数的值，经常用在设计模式中的观察者模式中（观察者设计模式定义了对象间的一种一对多的组合关系。这样一来一个对象的状态发生变化时，所有依赖于它的对象都会得到通知并自动刷新）。在观察者模式中，数据变动后，多个观察者都会收到这个变更信号。从源Channel取出一个数据后，依次发送给目标Channel。在发送给目标Channel的时候，可以同步发送，也可以异步发送。4、Stream。一种把Channel当作流式管道使用的方式，也就是把Channel看作流（Stream），提供跳过几个元素，或者是只取其中的几个元素等方法。首先我们提供创建流的方法。这个方法把一个数据slice转换成流。流创建好以后，该咋处理呢？下面实现流的方法：takeN只取流中的前n个数据；takeFn筛选流中的数据，只保留满足条件的数据；takeWhile只取前面满足条件的数据，一旦不满足条件，就不再取；skipN跳过流中前几个数据；skipFn跳过满足条件的数据；skipWhile跳过前面满足条件的数据，一旦不满足条件，当前这个元素和以后的元素都会输出给Channel的receiver。