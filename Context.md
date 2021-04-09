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
- WithValue，WithValue基于parent Context生成一个新的Context，保存了一个key-value键值对。它常常用来传递上下文。WithValue方法其实是创建了一个类型为valueCtx的Context，它的类型定义如下。它持有一个key-value键值对，还持有parent的Context。它覆盖了Value方法，优先从自己的存储中检查这个key，不存在的话会从parent中继续检查。
``` go
type valueCtx struct {
    Context
    key, val interface{}
}
```