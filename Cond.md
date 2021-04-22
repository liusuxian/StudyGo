### Go标准库的Cond。
- Go标准库提供Cond原语的目的是，为等待/通知场景下的并发问题提供支持。Cond通常应用于等待某个条件的一组goroutine，等条件变为true的时候，其中一个goroutine或者所有的goroutine都会被唤醒执行。顾名思义，Cond是和某个条件相关，这个条件需要一组goroutine协作共同完成，在条件还没有满足的时候，所有等待这个条件的goroutine都会被阻塞住，只有这一组goroutine通过协作达到了这个条件，等待的goroutine才可能继续进行下去。
### Cond的基本用法。
- 标准库中的Cond并发原语初始化的时候，需要关联一个Locker接口的实例，一般我们使用Mutex或者RWMutex。Cond关联的Locker实例可以通过c.L访问，它内部维护着一个先入先出的等待队列。
``` go
type Cond struct {
    noCopy noCopy // noCopy是一个辅助结构，用来帮助vet检查用的类型，nocpoy是静态检查。
    // L is held while observing or changing the condition
    L Locker
    notify  notifyList
    checker copyChecker // copyChecker是一个辅助结构，可以在运行时检查Cond是否被复制使用。
}
func NeWCond(l Locker) *Cond
func (c *Cond) Broadcast()
func (c *Cond) Signal()
func (c *Cond) Wait()
```
- Signal方法，允许调用者Caller唤醒一个等待此Cond的goroutine。如果此时没有等待的goroutine，显然无需通知waiter；如果Cond等待队列中有一个或者多个等待的goroutine，则需要从等待队列中移除第一个goroutine并把它唤醒。在其他编程语言中，比如Java语言中，Signal方法也被叫做notify方法。调用Signal方法时，不强求你一定要持有c.L的锁。
- Broadcast方法，允许调用者Caller唤醒所有等待此Cond的goroutine。如果此时没有等待的goroutine，显然无需通知waiter；如果Cond等待队列中有一个或者多个等待的goroutine，则清空所有等待的goroutine，并全部唤醒。在其他编程语言中，比如Java语言中，Broadcast方法也被叫做notifyAll方法。同样地，调用Broadcast方法时，也不强求你一定持有c.L的锁。
- Wait方法，会把调用者Caller放入Cond的等待队列中并阻塞，直到被Signal或者Broadcast的方法从等待队列中移除并唤醒。调用Wait方法时必须要持有c.L的锁。
### 使用Cond的2个常见错误。
- 调用Wait的时候没有加锁。运行程序就会报释放未加锁的panic。出现这个问题的原因在于Wait方法的实现是把当前调用者加入到notify队列之中后会释放锁（如果不释放锁，其他Wait的调用者就没有机会加入到notify队列中了），然后一直等待；等调用者被唤醒之后，又会去争抢这把锁。如果调用Wait之前不加锁的话，就有可能Unlock一个未加锁的Locker。所以切记，调用Wait方法之前一定要加锁。
- 只调用了一次Wait，没有检查等待条件是否满足，结果条件没满足，程序就继续执行了。出现这个问题的原因在于误以为Cond的使用，就像WaitGroup那样调用一下Wait方法等待那么简单。waiter goroutine被唤醒不等于等待条件被满足，只是有goroutine把它唤醒了而已，等待条件有可能已经满足了，也有可能不满足，我们需要进一步检查。你也可以理解为，等待者被唤醒，只是得到了一次检查的机会而已。
### Cond有三点特性是Channel无法替代的。
- Cond和一个Locker关联，可以利用这个Locker对相关的依赖条件更改提供保护。
- Cond可以同时支持Signal和Broadcast方法，而Channel只能同时支持其中一种。
- Cond的Broadcast方法可以被重复调用（Signal也可以被重复调用）。等待条件再次变成不满足的状态后，我们又可以调用Broadcast再次唤醒等待的goroutine。这也是Channel不能支持的，Channel被close掉了之后不支持再open。
### Cond的知识地图。
<img src="https://github.com/liusuxian/StudyGo/blob/master/img/Cond.jpg" width = "60%" height = "60%" alt="image-name"/>
