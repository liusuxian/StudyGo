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
- 内存浪费。要做到物尽其用，尽可能不浪费的话，我们可以将buffer池分成几层。首先小于512byte的元素的buffer占一个池子；其次小于1K byte大小的元素占一个池子；再次小于4K byte大小的元素占一个池子。这样分成几个池子以后，就可以根据需要，到所需大小的池子中获取buffer了。在标准库net/http/server.go中的代码中，就提供了2K和4K两个writer的池子。YouTube开源的知名项目vitess中提供了[bucketpool](https://github.com/vitessio/vitess/blob/master/go/bucketpool/bucketpool.go)的实现，它提供了更加通用的多层buffer池。你在使用的时候，只需要指定池子的最大和最小尺寸，vitess就会自动计算出合适的池子数。而且当你调用Get方法的时候，只需要传入你要获取的buffer的大小，就可以了。
### 第三方库。
- [bytebufferpool](https://github.com/valyala/bytebufferpool)。这是fasthttp作者valyala提供的一个buffer池，基本功能和sync.Pool相同。它的底层也是使用sync.Pool实现的，包括会检测最大的buffer，超过最大尺寸的buffer，就会被丢弃。这个库提供了校准（calibrate，用来动态调整创建元素的权重）的机制，可以“智能”地调整Pool的defaultSize和maxSize。一般来说，我们使用buffer size的场景比较固定，所用buffer的大小会集中在某个范围里。有了校准的特性，bytebufferpool就能够偏重于创建这个范围大小的buffer，从而节省空间。
