### 上下文是啥。
- 在API之间或者方法调用之间，所传递的除了业务参数之外的额外信息。
- 比如，服务端接收到客户端的HTTP请求之后，可以把客户端的IP地址和端口、客户端的身份信息、请求接收的时间、Trace ID等信息放入到上下文中，这个上下文可以在后端的方法调用中传递，后端的业务方法除了利用正常的参数做一些业务处理（如订单处理）之外，还可以从上下文读取到消息请求的时间、Trace ID等信息，把服务处理的时间推送到Trace服务中。Trace服务可以把同一Trace ID的不同方法的调用顺序和调用时间展示成流程图，方便跟踪。不过Go标准库中的Context功能还不止于此，它还提供了超时（Timeout）和取消（Cancel）的机制。
### 当前Context的问题。
- Context包名导致使用的时候重复ctx context.Context。
- Context.WithValue可以接受任何类型的值，非类型安全。
- Context包名容易误导人，实际上Context最主要的功能是取消goroutine的执行。
- Context漫天飞，函数污染。
### Context的应用场景。
- 上下文信息传递（request-scoped），比如处理http请求、在请求处理链路上传递信息。
- 控制子goroutine的运行。
- 超时控制的方法调用。
- 可以取消的方法调用。
### Context基本使用方法。
- 包context定义了Context接口，Context的具体实现包括4个方法，分别是Deadline、Done、Err和Value，如下所示：
``` go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key interface{}) interface{}
}
```
- Deadline方法会返回这个Context被取消的截止日期。如果没有设置截止日期，ok的值是false。后续每次调用这个对象的Deadline方法时，都会返回和第一次调用相同的结果。
- Done方法返回一个Channel对象。在Context被取消时，此Channel会被close，如果没被取消，可能会返回nil。后续的Done调用总是返回相同的结果。当Done被close的时候，你可以通过ctx.Err获取错误信息。Done这个方法名其实起得并不好，因为名字太过笼统，不能明确反映Done被close的原因，因为cancel、timeout、deadline都可能导致Done被close，不过目前还没有一个更合适的方法名称。关于Done方法，你必须要记住的知识点就是：如果Done没有被close，Err方法返回nil；如果Done被close，Err方法会返回Done被close的原因。
- Value返回此ctx中和指定的key相关联的value。
### Context中实现了2个常用的生成顶层Context的方法。
- context.Background()：返回一个非nil的、空的Context，没有任何值，不会被cancel，不会超时，没有截止日期。一般用在主函数、初始化、测试以及创建根Context的时候。
- context.TODO()：返回一个非nil的、空的Context，没有任何值，不会被cancel，不会超时，没有截止日期。当你不清楚是否该用Context，或者目前还不知道要传递一些什么上下文信息的时候，就可以使用这个方法。
- 两个方法底层的实现是一模一样的。
### 在使用Context的时候，有一些约定俗成的规则。
- 一般函数使用Context的时候，会把这个参数放在第一个参数的位置。
- 从来不把nil当做Context类型的参数值，可以使用context.Background()创建一个空的上下文对象，也不要使用nil。
- Context只用来临时做函数之间的上下文透传，不能持久化Context或者把Context长久保存。把Context持久化到数据库、本地文件或者全局变量、缓存中都是错误的用法。
- key的类型不应该是字符串类型或者其它内建类型，否则容易在包之间使用Context时候产生冲突。使用WithValue时，key的类型应该是自己定义的类型。
- 常常使用struct{}作为底层类型定义key的类型。对于exported key的静态类型，常常是接口或者指针。这样可以尽量减少内存分配。
- 如果你能保证别人使用你的Context时不会和你定义的key冲突，那么key的类型就比较随意，因为你自己保证了不同包的key不会冲突，否则建议你尽量采用保守的unexported的类型。
### 创建特殊用途Context的方法。
- WithValue基于parent Context生成一个新的Context，保存了一个key-value键值对。它常常用来传递上下文。WithValue方法其实是创建了一个类型为valueCtx的Context，它的类型定义如下。它持有一个key-value键值对，还持有parent的Context。它覆盖了Value方法，优先从自己的存储中检查这个key，不存在的话会从parent中继续检查。
``` go
type valueCtx struct {
    Context
    key, val interface{}
}
```
- WithCancel方法返回parent的副本，只是副本中的Done Channel是新建的对象，它的类型是cancelCtx。我们常常在一些需要主动取消长时间的任务时，创建这种类型的Context，然后把这个Context传给长时间执行任务的goroutine。当需要中止任务时，我们就可以cancel这个Context，这样长时间执行任务的goroutine，就可以通过检查这个Context，知道Context已经被取消了。WithCancel返回值中的第二个值是一个cancel函数。其实这个返回值的名称（cancel）和类型（Cancel）也非常迷惑人。记住不是只有你想中途放弃，才去调用cancel，只要你的任务正常完成了，就需要调用cancel，这样这个Context才能释放它的资源（通知它的children处理cancel，从它的parent中把自己移除，甚至释放相关的goroutine）。很多人在使用这个方法的时候，都会忘记调用cancel，切记切记，而且一定尽早释放。当这个cancelCtx的cancel函数被调用的时候，或者parent的Done被close的时候，这个cancelCtx的Done才会被close。cancel是向下传递的，如果一个WithCancel生成的Context被cancel时，如果它的子Context（也有可能是孙，或者更低，依赖子的类型）也是cancelCtx类型的，就会被cancel，但是不会向上传递。parent Context不会因为子Context被cancel而cancel。
- WithTimeout其实是和WithDeadline一样，只不过一个参数是超时时间，一个参数是截止时间。超时时间加上当前时间，其实就是截止时间。
- WithDeadline会返回一个parent的副本，并且设置了一个不晚于参数d的截止时间，类型为timerCtx（或者是cancelCtx）。如果它的截止时间晚于parent的截止时间，那么就以parent的截止时间为准，并返回一个类型为cancelCtx的Context，因为parent的截止时间到了，就会取消这个cancelCtx。如果当前时间已经超过了截止时间，就直接返回一个已经被cancel的timerCtx。否则就会启动一个定时器，到截止时间取消这个timerCtx。综合起来timerCtx的Done被Close掉，主要是由以下的某个事件触发的：1、截止时间到了；2、cancel函数被调用；3、parent的Done被close。和cancelCtx一样，WithDeadline（WithTimeout）返回的cancel一定要调用，并且要尽可能早地被调用，这样才能尽早释放资源，不要单纯地依赖截止时间被动取消。
### 总结。
- 我们经常使用Context来取消一个goroutine的运行，这是Context最常用的场景之一，Context也被称为goroutine生命周期范围（goroutine-scoped）的Context，把Context传递给goroutine。但是goroutine需要尝试检查Context的Done是否关闭了。
- 如果要为Context实现一个带超时功能的调用，比如访问远程的一个微服务，超时并不意味着你会通知远程微服务已经取消了这次调用，大概率的实现只是避免客户端的长时间等待，远程的服务器依然还执行着你的请求。所以有时候Context并不会减少对服务器的请求负担。如果在Context被cancel的时候，你能关闭和服务器的连接，中断和数据库服务器的通讯、停止对本地文件的读写，那么这样的超时处理，同时能减少对服务调用的压力，但是这依赖于你对超时的底层处理机制。
### Context知识地图。
<img src="https://github.com/liusuxian/learning_golang/blob/master/img/Context.jpg" width = "60%" height = "60%" alt="image-name"/>