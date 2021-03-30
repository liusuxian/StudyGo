# learning_golang
learning golang

### go run -race 文件名.go 检测并发访问共享资源是否有问题的命令。
- 会输出警告信息，这个警告不但会告诉你有并发问题，而且还会告诉你哪个 goroutine 在哪一行对哪个变量有写操作，同时，哪个 goroutine 在哪一行对哪个变量有读操作，就是这些并发的读写访问，引起了 data race。虽然这个工具使用起来很方便，但是，因为它的实现方式，只能通过真正对实际地址进行读写访问的时候才能探测，所以它并不能在编译的时候发现 data race 的问题。而且，在运行的时候，只有在触发了 data race 之后，才能检测到，如果碰巧没有触发，是检测不出来的。而且，把开启了 race 的程序部署在线上，还是比较影响性能的。
### go tool compile -race -S 文件名.go 查看计数器命令。
- 在编译的代码中，增加了 runtime.racefuncenter、runtime.raceread、runtime.racewrite、runtime.racefuncexit 等检测 data race 的方法。通过这些插入的指令，Go race detector 工具就能够成功地检测出 data race 问题了。
### go tool compile -S 文件名.go 查看汇编代码命令。
### Unlock 坑点。
- Unlock 方法可以被任意的 goroutine 调用释放锁，即使是没持有这个互斥锁的 goroutine，也可以进行这个操作。这是因为，Mutex 本身并没有包含持有这把锁的 goroutine 的信息，所以，Unlock 也不会对此进行检查。Mutex 的这个设计一直保持至今。
- 所以，我们在使用 Mutex 的时候，必须要保证 goroutine 尽可能不去释放自己未持有的锁，一定要遵循“谁申请，谁释放”的原则。在真实的实践中，我们使用互斥锁的时候，很少在一个方法中单独申请锁，而在另外一个方法中单独释放锁，一般都会在同一个方法中获取锁和释放锁。
### 如果 Mutex 已经被一个 goroutine 获取了锁，其它等待中的 goroutine 们只能一直等待。那么，等这个锁释放后，等待中的 goroutine 中哪一个会优先获取 Mutex 呢？
- 等待的 goroutine 们是以 FIFO 排队的
- 当 Mutex 处于正常模式时，若此时没有新 goroutine 与队头 goroutine 竞争，则队头 goroutine 获得。若有新 goroutine 竞争大概率新 goroutine 获得。
- 当队头 goroutine 竞争锁失败 1ms 后，它会将 Mutex 调整为饥饿模式。进入饥饿模式后，锁的所有权会直接从解锁 goroutine 移交给队头 goroutine，此时新来的 goroutine 直接放入队尾。
- 当一个 goroutine 获取锁后，如果发现自己满足以下条件中的任何一个，1.它是队列中最后一个，2.它等待锁的时间少于 1ms，则将锁切换回正常模式
### 目前 Mutex 的 state 字段有几个意义，这几个意义分别是由哪些字段表示的？
- 前三个 bit 分别为 mutexLocked（持有锁的标记）、mutexWoken（唤醒标记）、mutexStarving（饥饿标记），剩余 bit 表示 mutexWaiter（阻塞等待的 waiter 数量）
### 等待一个 Mutex 的 goroutine 数最大是多少？是否能满足现实的需求？
- 单从程序来看，可以支持 1<<(32-3) -1 ，约 0.5 Billion 个，其中 32 为 state 的类型 int32，3 位 waiter 字段的 shift，考虑到实际 goroutine 初始化的空间为 2K，0.5 Billin * 2K 达到了 1TB，单从内存空间来说已经要求极高了，当前的设计肯定可以满足了。
### 使用 Mutex 常见的错误场景有 4 类。
- Lock/Unlock 不是成对出现，就意味着会出现死锁的情况，或者是因为 Unlock 一个未加锁的 Mutex 而导致 panic。
- Copy 已使用的 Mutex，Package sync 的同步原语在使用后是不能复制的。Mutex 是一个有状态的对象，它的 state 字段记录这个锁的状态。如果你要复制一个已经加锁的 Mutex 给一个新的变量，那么新的刚初始化的变量居然被加锁了，这显然不符合你的期望。