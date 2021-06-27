### Leader选举
- Leader选举常常用在主从架构的系统中。主从架构中的服务节点分为主（Leader、Master）和从（Follower、Slave）两种角色，实际节点包括1主n从，一共是n+1个节点。
- 主节点常常执行写操作，从节点常常执行读操作，如果读写都在主节点，从节点只是提供一个备份功能的话，那么主从架构就会退化成主备模式架构。
- 在同一时刻，系统中不能有两个主节点，否则，如果两个节点都是主，都执行写操作的话，就有可能出现数据不一致的情况，所以，我们需要一个选主机制，选择一个节点作为主节点，这个过程就是Leader选举。
- 当主节点宕机或者是不可用时，就需要新一轮的选举，从其它的从节点中选择出一个节点，让它作为新主节点，宕机的原主节点恢复后，可以变为从节点，或者被摘掉。
### 选举
- 如果你的业务集群还没有主节点，或者主节点宕机了，你就需要发起新一轮的选主操作，主要会用到Campaign和Proclaim。如果你需要主节点放弃主的角色，让其它从节点有机会成为主节点，就可以调用Resign方法。
- 第一个方法是Campaign。它的作用是，把一个节点选举为主节点，并且会设置一个值。它的签名如下所示：
``` go
func (e *Election) Campaign(ctx context.Context, val string) error
```
  - 需要注意的是，这是一个阻塞方法，在调用它的时候会被阻塞，直到满足下面的三个条件之一，才会取消阻塞。
    - 成功当选为主；
    - 此方法返回错误；
    - ctx被取消。
- 第二个方法是Proclaim。它的作用是，重新设置Leader的值，但是不会重新选主，这个方法会返回新值设置成功或者失败的信息。方法签名如下所示：
``` go
func (e *Election) Proclaim(ctx context.Context, val string) error
```
- 第三个方法是 Resign：开始新一次选举。这个方法会返回新的选举成功或者失败的信息。它的签名如下所示：
``` go
func (e *Election) Resign(ctx context.Context) (err error)
```
### 查询
- etcd提供了查询当前Leader的方法Leader，如果当前还没有Leader，就返回一个错误，你可以使用这个方法来查询主节点信息。这个方法的签名如下：
``` go
func (e *Election) Leader(ctx context.Context) (*v3.GetResponse, error)
```
- 每次主节点的变动都会生成一个新的版本号，你还可以查询版本号信息（Rev方法），了解主节点变动情况：
``` go
func (e *Election) Rev() int64
```
### 监控
- 我们可以通过Observe来监控主的变化，它的签名如下：
``` go
func (e *Election) Observe(ctx context.Context) <-chan v3.GetResponse
```
- 它会返回一个chan，显示主节点的变动信息。需要注意的是，它不会返回主节点的全部历史变动信息，而是只返回最近的一条变动信息以及之后的变动信息。
### 互斥锁
- 互斥锁的应用场景和主从架构的应用场景不太一样。使用互斥锁的不同节点是没有主从这样的角色的，所有的节点都是一样的，只不过在同一时刻，只允许其中的一个节点持有锁。
### Locker
- etcd提供了一个简单的Locker原语，它类似于Go标准库中的sync.Locker接口，也提供了Lock/UnLock的机制：
``` go
func NewLocker(s *Session, pfx string) sync.Locker
```
- 获得锁是有先后顺序的，一个节点释放了锁之后，另外一个节点才能获取到这个分布式锁。
### Mutex
- Locker是基于Mutex实现的，只不过Mutex提供了查询Mutex的key的信息的功能。
- Mutex并没有实现sync.Locker接口，它的Lock/Unlock方法需要提供一个context.Context实例做参数，这也就意味着，在请求锁的时候，你可以设置超时时间，或者主动取消请求。
### 读写锁
- etcd提供的分布式读写锁的功能和标准库的读写锁的功能是一样的。只不过etcd提供的读写锁，可以在分布式环境中的不同的节点使用。它提供的方法也和标准库中的读写锁的方法一致，分别提供了RLock/RUnlock、Lock/Unlock方法。
### Etcd的知识地图。
<img src="https://github.com/liusuxian/StudyGo/blob/master/img/Etcd.jpg" width = "100%" height = "100%" alt="image-name"/>

### 分布式队列和优先级队列
- etcd通过github.com/coreos/etcd/contrib/recipes包提供了分布式队列这种数据结构。
- 创建分布式队列的方法非常简单，只有一个，即NewQueue，你只需要传入etcd的client和这个队列的名字，就可以了。代码如下：
``` go
func NewQueue(client *v3.Client, keyPrefix string) *Queue
```
- 这个队列只有两个方法，分别是出队和入队，队列中的元素是字符串类型。这两个方法的签名如下所示：
``` go
// 入队
func (q *Queue) Enqueue(val string) error
// 出队
func (q *Queue) Dequeue() (string, error)
```
- 需要注意的是，如果这个分布式队列当前为空，调用Dequeue方法的话，会被阻塞，直到有元素可以出队才返回。
- 既然是分布式的队列，那就意味着，我们可以在一个节点将元素放入队列，在另外一个节点把它取出。
- etcd的分布式队列是一种多读多写的队列，所以你也可以启动多个写节点和多个读节点。
- etcd还提供了优先级队列（PriorityQueue）。它的用法和队列类似，也提供了出队和入队的操作，只不过在入队的时候，除了需要把一个值加入到队列，我们还需要提供uint16类型的一个整数，作为此值的优先级，优先级高的元素会优先出队。
### 分布式栅栏
- Barrier：分布式栅栏。如果持有Barrier的节点释放了它，所有等待这个Barrier的节点就不会被阻塞，而是会继续执行。
- DoubleBarrier：计数型栅栏。在初始化计数型栅栏的时候，我们就必须提供参与节点的数量，当这些数量的节点都Enter或者Leave的时候，这个栅栏就会放开。所以我们把它称为计数型栅栏。
### Barrier：分布式栅栏
- 分布式Barrier的创建很简单，你只需要提供etcd的Client和Barrier的名字就可以了，如下所示：
``` go
func NewBarrier(client *v3.Client, key string) *Barrier
```
- Barrier提供了三个方法，分别是Hold、Release和Wait，代码如下：
``` go
func (b *Barrier) Hold() error
func (b *Barrier) Release() error
func (b *Barrier) Wait() error
```
- Hold方法是创建一个Barrier。如果Barrier已经创建好了，有节点调用它的Wait方法，就会被阻塞。
- Release方法是释放这个Barrier，也就是打开栅栏。如果使用了这个方法，所有被阻塞的节点都会被放行，继续执行。
- Wait方法会阻塞当前的调用者，直到这个Barrier被release。如果这个栅栏不存在，调用者不会被阻塞，而是会继续执行。 
### DoubleBarrier：计数型栅栏
- etcd还提供了另外一种栅栏，叫做DoubleBarrier，这也是一种非常有用的栅栏。这个栅栏初始化的时候需要提供一个计数count，如下所示：
``` go
func NewDoubleBarrier(s *concurrency.Session, key string, count int) *DoubleBarrier
```
- 同时，它还提供了两个方法，分别是Enter和Leave，代码如下：
``` go
func (b *DoubleBarrier) Enter() error
func (b *DoubleBarrier) Leave() error
```
- 当调用者调用Enter时，会被阻塞住，直到一共有count（初始化这个栅栏的时候设定的值）个节点调用了Enter，这count个被阻塞的节点才能继续执行。所以你可以利用它编排一组节点，让这些节点在同一个时刻开始执行任务。
- 同理，如果你想让一组节点在同一个时刻完成任务，就可以调用Leave方法。节点调用Leave方法的时候，会被阻塞，直到有count个节点，都调用了Leave方法，这些节点才能继续执行。
### STM
- 