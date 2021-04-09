### 垃圾回收给性能带来的影响。
- 如果你想使用Go开发一个高性能的应用程序的话，就必须考虑垃圾回收给性能带来的影响，毕竟Go的自动垃圾回收机制还是有一个STW（stop-the-world，程序暂停）的时间，而且大量地创建在堆上的对象，也会影响垃圾回收标记的时间。所以一般我们做性能优化的时候，会采用对象池的方式，把不同的对象回收起来，避免被垃圾回收掉，这样使用的时候就不必在堆上重新创建了。不止如此，像数据库连接、TCP的长连接，这些连接在创建的时候是一个非常耗时的操作。如果每次都创建一个新的连接对象，耗时较长，很可能整个业务的大部分耗时都花在了创建连接上。所以如果我们能把这些连接保存下来，避免每次使用的时候都重新创建，不仅可以大大减少业务的耗时，还能提高应用程序的整体性能。
### sync.Pool数据类型。
- sync.Pool数据类型用来保存一组可独立访问的临时对象。临时这两个字说明了sync.Pool这个数据类型的特点，也就是说，它池化的对象会在未来的某个时候被毫无预兆地移除掉。而且如果没有别的对象引用这个被移除的对象的话，这个被移除的对象就会被垃圾回收掉。
- sync.Pool本身就是线程安全的，多个goroutine可以并发地调用它的方法存取对象。
- sync.Pool不可在使用之后再复制使用。
### sync.Pool的使用方法。
- New，Pool struct包含一个New字段，这个字段的类型是函数func() interface{}。当调用Pool的Get方法从池中获取元素，没有更多的空闲元素可返回时，就会调用这个New方法来创建新的元素。如果你没有设置New字段，没有更多的空闲元素可返回时，Get方法将返回nil，表明当前没有可用的元素。
- Get，如果调用这个方法，就会从Pool取走一个元素，这也就意味着，这个元素会从Pool中移除，返回给调用者。不过除了返回值是正常实例化的元素，Get方法的返回值还可能会是一个nil（Pool.New字段没有设置，又没有空闲元素可以返回），所以你在使用的时候，可能需要判断。
- Put，这个方法用于将一个元素返还给Pool，Pool会把这个元素保存到池中，并且可以复用。但如果Put一个nil值，Pool就会忽略这个值。
### Go1.13之前的sync.Pool的实现有2大问题。
- 每次GC都会回收创建的对象。如果缓存元素数量太多，就会导致STW耗时变长；缓存元素都被回收后，会导致Get命中率下降，Get方法不得不新创建很多对象。
- 底层实现使用了Mutex，对这个锁并发请求竞争激烈的时候，会导致性能的下降。
### Pool最重要的两个字段是local和victim。
- 每次垃圾回收的时候，Pool会把victim中的对象移除，然后把local的数据给victim，这样的话，local就会被清空，而victim就像一个垃圾分拣站，里面的东西可能会被当做垃圾丢弃了，但是里面有用的东西也可能被捡回来重新使用。
- victim中的元素如果被Get取走，那么这个元素就很幸运，因为它又“活”过来了。但是如果这个时候Get的并发不是很大，元素没有被Get取走，那么就会被移除掉，因为没有别人引用它的话，就会被垃圾回收掉。
### 垃圾回收时sync.Pool的处理逻辑。
- 所有当前主要的空闲可用的元素都存放在local字段中，请求元素时也是优先从local字段中查找可用的元素。local字段包含一个poolLocalInternal字段，并提供CPU缓存对齐，从而避免false sharing。而poolLocalInternal也包含两个字段：private和shared。
- private，代表一个缓存的元素，而且只能由相应的一个P存取。因为一个P同时只能执行一个goroutine，所以不会有并发的问题。
- shared，可以由任意的P访问，但是只有本地的P才能pushHead/popHead，其它P可以popTail，相当于只有一个本地的P作为生产者（Producer），多个P作为消费者（Consumer），它是使用一个local-free的queue实现的。
### sync.Pool Get方法的具体实现原理。
- 首先从本地的private字段中获取可用元素，因为没有锁，获取元素的过程会非常快，如果没有获取到，就尝试从本地的shared获取一个，如果还没有，会使用getSlow方法去其它的shared中“偷”一个。最后如果没有获取到，就尝试使用New函数创建一个新的。
- getSlow方法，看名字也就知道了，它的耗时可能比较长。它首先要遍历所有的local，尝试从它们的shared弹出一个元素。如果还没找到一个，那么就开始对victim下手了。在victim中查询可用元素的逻辑还是一样的，先从对应的victim的private查找，如果查不到，就再从其它victim的shared中查找。
### sync.Pool Put方法的具体实现原理。
- Put的逻辑相对简单，优先设置本地private，如果private字段已经有值了，那么就把此元素push到本地队列中。
### sync.Pool的坑。
- 内存泄漏。在使用sync.Pool回收buffer的时候，一定要检查回收的对象的大小。如果buffer太大，就不要回收了，否则会有内存泄漏问题。
- 内存浪费。要做到物尽其用，尽可能不浪费的话，我们可以将buffer池分成几层。首先小于512byte的元素的buffer占一个池子；其次小于1K byte大小的元素占一个池子；再次小于4K byte大小的元素占一个池子。这样分成几个池子以后，就可以根据需要，到所需大小的池子中获取buffer了。在标准库net/http/server.go中的代码中，就提供了2K和4K两个writer的池子。YouTube开源的知名项目vitess中提供了[bucketpool](https://github.com/vitessio/vitess/blob/master/go/bucketpool/bucketpool.go) 的实现，它提供了更加通用的多层buffer池。你在使用的时候，只需要指定池子的最大和最小尺寸，vitess就会自动计算出合适的池子数。而且当你调用Get方法的时候，只需要传入你要获取的buffer的大小，就可以了。
### buffer池第三方库。
- [bytebufferpool](https://github.com/valyala/bytebufferpool) 这是fasthttp作者valyala提供的一个buffer池，基本功能和sync.Pool相同。它的底层也是使用sync.Pool实现的，包括会检测最大的buffer，超过最大尺寸的buffer，就会被丢弃。这个库提供了校准（calibrate，用来动态调整创建元素的权重）的机制，可以“智能”地调整Pool的defaultSize和maxSize。一般来说，我们使用buffer size的场景比较固定，所用buffer的大小会集中在某个范围里。有了校准的特性，bytebufferpool就能够偏重于创建这个范围大小的buffer，从而节省空间。
- [oxtoacart/bpool](https://github.com/oxtoacart/bpool) 这也是比较常用的buffer池，它提供了以下几种类型的buffer。1、bpool.BufferPool：提供一个固定元素数量的buffer池，元素类型是bytes.Buffer，如果超过这个数量，Put的时候就丢弃，如果池中的元素都被取光了，会新建一个返回。Put回去的时候，不会检测buffer的大小。2、bpool.BytesPool：提供一个固定元素数量的byte slice池，元素类型是byte slice。Put回去的时候不检测slice的大小。3、bpool.SizedBufferPool：提供一个固定元素数量的buffer池，如果超过这个数量，Put的时候就丢弃，如果池中的元素都被取光了，会新建一个返回。Put回去的时候，会检测buffer的大小，超过指定的大小的话，就会创建一个新的满足条件的buffer放回去。bpool最大的特色就是能够保持池子中元素的数量，一旦Put的数量多于它的阈值，就会自动丢弃，而sync.Pool是一个没有限制的池子，只要Put就会收进去。bpool是基于Channel实现的，不像sync.Pool为了提高性能而做了很多优化，所以在性能上比不过sync.Pool。不过它提供了限制Pool容量的功能，所以如果你想控制Pool的容量的话，可以考虑这个库。
### 标准库中的http client池。
- 标准库的http.Client是一个http client的库，可以用它来访问web服务器。为了提高性能，这个Client的实现也是通过池的方法来缓存一定数量的连接，以便后续重用这些连接。http.Client实现连接池的代码是在Transport类型中，它使用idleConn保存持久化的可重用的长连接。
### TCP连接池。
- 最常用的一个TCP连接池是fatih开发的[fatih/pool](https://github.com/fatih/pool) 虽然这个项目已经被fatih归档（Archived），不再维护了，但是因为它相当稳定了，我们可以开箱即用。即使你有一些特殊的需求，也可以fork它，然后自己再做修改。它的使用套路如下：
``` go
// 工厂模式，提供创建连接的工厂方法
factory := func() (net.Conn, error) { return net.Dial("tcp", "127.0.0.1:4000") }
// 创建一个tcp池，提供初始容量和最大容量以及工厂方法
p, err := pool.NewChannelPool(5, 30, factory)
// 获取一个连接
conn, err := p.Get()
// Close并不会真正关闭这个连接，而是把它放回池子，所以你不必显式地Put这个对象到池子中
conn.Close()
// 通过调用MarkUnusable, Close的时候就会真正关闭底层的tcp的连接了
if pc, ok := conn.(*pool.PoolConn); ok {
    pc.MarkUnusable()
    pc.Close()
}
// 关闭池子就会关闭=池子中的所有的tcp连接
p.Close()
// 当前池子中的连接的数量
current := p.Len()
```
- 它管理的是更通用的net.Conn，不局限于TCP连接。它通过把net.Conn包装成PoolConn，实现了拦截net.Conn的Close方法，避免了真正地关闭底层连接，而是把这个连接放回到池中：
``` go
type PoolConn struct {
    net.Conn
    mu       sync.RWMutex
    c        *channelPool
    unusable bool
}

//拦截Close
func (p *PoolConn) Close() error {
    p.mu.RLock()
    defer p.mu.RUnlock()

    if p.unusable {
        if p.Conn != nil {
            return p.Conn.Close()
        }
        return nil
    }
    return p.c.put(p.Conn)
}
```
- 它的Pool是通过Channel实现的，空闲的连接放入到Channel中，这也是Channel的一个应用场景：
``` go
type channelPool struct {
    // 存储连接池的channel
    mu    sync.RWMutex
    conns chan net.Conn

    // net.Conn 的产生器
    factory Factory
}
```
### 数据库连接池。
- 标准库sql.DB还提供了一个通用的数据库的连接池，通过MaxOpenConns和MaxIdleConns控制最大的连接数和最大的idle的连接数。默认的MaxIdleConns是2，这个数对于数据库相关的应用来说太小了，我们一般都会调整它。DB的freeConn保存了idle的连接，这样当我们获取数据库连接的时候，它就会优先尝试从freeConn获取已有的连接。
### Memcached Client连接池。
- [gomemcache](https://github.com/bradfitz/gomemcache) Memchaced客户端，其中也用了连接池的方式池化Memcached的连接。gomemcache Client有一个freeconn的字段，用来保存空闲的连接。当一个请求使用完之后，它会调用putFreeConn放回到池子中，请求的时候，调用getFreeConn优先查询freeConn中是否有可用的连接。它采用Mutex+Slice实现Pool。
### Worker Pool。
- goroutine是一个很轻量级的“纤程”，在一个服务器上可以创建十几万甚至几十万的goroutine。但是“可以”和“合适”之间还是有区别的，你会在应用中让几十万的goroutine一直跑吗？基本上是不会的。一个goroutine初始的栈大小是2048个字节，并且在需要的时候可以扩展到1GB，所以大量的goroutine还是很耗资源的。同时大量的goroutine对于调度和垃圾回收的耗时还是会有影响的，因此goroutine并不是越多越好。有的时候，我们就会创建一个Worker Pool来减少goroutine的使用。比如我们实现一个TCP服务器，如果每一个连接都要由一个独立的goroutine去处理的话，在大量连接的情况下，就会创建大量的goroutine，这个时候，我们就可以创建一个固定数量的goroutine（Worker），由这一组Worker去处理连接，比如fasthttp中的 [Worker Pool](https://github.com/valyala/fasthttp/blob/9f11af296864153ee45341d3f2fe0f5178fd6210/workerpool.go#L16)。
- Worker的实现也是五花八门的：有些是在后台默默执行的，不需要等待返回结果；有些需要等待一批任务执行完；有些Worker Pool的生命周期和程序一样长；有些只是临时使用，执行完毕后，Pool就销毁了。
- 大部分的Worker Pool都是通过Channel来缓存任务的，因为Channel能够比较方便地实现并发的保护，有的是多个Worker共享同一个任务Channel，有些是每个Worker都有一个独立的Channel。
- [gammazero/workerpool](https://pkg.go.dev/github.com/gammazero/workerpool?utm_source=godoc) 可以无限制地提交任务，提供了更便利的Submit和SubmitWait方法提交任务，还可以提供当前的worker数和任务数以及关闭Pool的功能。
- [ivpusic/grpool](https://pkg.go.dev/github.com/ivpusic/grpool?utm_source=godoc) grpool创建Pool的时候需要提供Worker的数量和等待执行的任务的最大数量，任务的提交是直接往Channel放入任务。
- [dpaks/goworkers](https://pkg.go.dev/github.com/dpaks/goworkers?utm_source=godoc) 提供了更便利的Submit方法提交任务以及Worker数、任务数等查询方法、关闭Pool的方法。它的任务的执行结果需要在ResultChan和ErrChan中去获取，没有提供阻塞的方法，但是它可以在初始化的时候设置Worker的数量和任务数。
- 类似的Worker Pool的实现非常多，比如还有 [panjf2000/ants]()、[Jeffail/tunny]()、[benmanns/goworker]()、[go-playground/pool]()、[Sherifabdlnaby/gpool]() 等第三方库。[pond]() 也是一个非常不错的Worker Pool，关注度目前不是很高，但是功能非常齐全。