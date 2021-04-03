# learning_golang
learning golang

### go run -race 文件名.go 检测并发访问共享资源是否有问题的命令。
- 会输出警告信息，这个警告不但会告诉你有并发问题，而且还会告诉你哪个goroutine在哪一行对哪个变量有写操作，同时，哪个goroutine在哪一行对哪个变量有读操作，就是这些并发的读写访问，引起了data race。虽然这个工具使用起来很方便，但是，因为它的实现方式，只能通过真正对实际地址进行读写访问的时候才能探测，所以它并不能在编译的时候发现data race的问题。而且，在运行的时候，只有在触发了data race之后，才能检测到，如果碰巧没有触发，是检测不出来的。而且，把开启了race的程序部署在线上，还是比较影响性能的。
### go tool compile -race -S 文件名.go 查看计数器命令。
- 在编译的代码中，增加了runtime.racefuncenter、runtime.raceread、runtime.racewrite、runtime.racefuncexit等检测data race的方法。通过这些插入的指令，Go race detector工具就能够成功地检测出data race问题了。
### go tool compile -S 文件名.go 查看汇编代码命令。
### Unlock坑点。
- Unlock方法可以被任意的goroutine调用释放锁，即使是没持有这个互斥锁的goroutine，也可以进行这个操作。这是因为Mutex本身并没有包含持有这把锁的goroutine的信息，所以Unlock也不会对此进行检查。Mutex的这个设计一直保持至今。
- 所以我们在使用Mutex的时候，必须要保证goroutine尽可能不去释放自己未持有的锁，一定要遵循“谁申请，谁释放”的原则。在真实的实践中，我们使用互斥锁的时候，很少在一个方法中单独申请锁，而在另外一个方法中单独释放锁，一般都会在同一个方法中获取锁和释放锁。
### 如果Mutex已经被一个goroutine获取了锁，其它等待中的goroutine们只能一直等待。那么，等这个锁释放后，等待中的goroutine中哪一个会优先获取Mutex呢？
- 互斥锁有两种状态：正常状态和饥饿状态。
- 在正常状态下，所有等待锁的goroutine按照FIFO顺序等待。唤醒的goroutine不会直接拥有锁，而是会和新请求锁的goroutine竞争锁的拥有。新请求锁的goroutine具有优势：它正在CPU上执行，而且可能有好几个，所以刚刚唤醒的goroutine有很大可能在锁竞争中失败。在这种情况下，这个被唤醒的goroutine会加入到等待队列的前面。 如果一个等待的goroutine超过1ms没有获取锁，那么它将会把锁转变为饥饿模式。
- 在饥饿模式下，锁的所有权将从unlock的gorutine直接交给交给等待队列中的第一个。新来的goroutine将不会尝试去获得锁，即使锁看起来是unlock状态, 也不会去尝试自旋操作，而是放在等待队列的尾部。
- 如果一个等待的goroutine获取了锁，并且满足一以下其中的任何一个条件：(1)它是队列中的最后一个；(2)它等待的时候小于1ms。它会将锁的状态转换为正常状态。
- 正常状态有很好的性能表现，饥饿模式也是非常重要的，因为它能阻止尾部延迟的现象。
### 目前Mutex的state字段有几个意义，这几个意义分别是由哪些字段表示的？
- 前三个bit分别为mutexLocked（持有锁的标记）、mutexWoken（唤醒标记）、mutexStarving（饥饿标记），剩余bit表示mutexWaiter（阻塞等待的waiter数量）
### 等待一个Mutex的goroutine数最大是多少？是否能满足现实的需求？
- 单从程序来看，可以支持1<<(32-3) -1 ，约0.5Billion个，其中32为state的类型int32，3位waiter字段的shift，考虑到实际goroutine初始化的空间为2K，0.5Billin * 2K达到了1TB，单从内存空间来说已经要求极高了，当前的设计肯定可以满足了。
### 使用Mutex常见的错误场景有4类。
- Lock/Unlock不是成对出现，就意味着会出现死锁的情况，或者是因为Unlock一个未加锁的Mutex而导致panic，注意：未加锁的mutex导致的panic，无法被recover()捕获。
- Copy已使用的Mutex，Package sync的同步原语在使用后是不能复制的。Mutex是一个有状态的对象，它的state字段记录这个锁的状态。如果你要复制一个已经加锁的Mutex给一个新的变量，那么新的刚初始化的变量居然被加锁了，这显然不符合你的期望。go vet 文件名.go 检测Mutex复制问题的命令。
- 重入，Mutex不是可重入的锁。
- 死锁，两个或两个以上的进程（或线程，goroutine）在执行过程中，因争夺共享资源而处于一种互相等待的状态，如果没有外部干涉，它们都将无法推进下去，此时，我们称系统处于死锁状态或系统产生了死锁。Go 运行时，有死锁探测的功能，能够检查出是否出现了死锁的情况。
### 避免死锁，只要破坏这四个条件中的一个或者几个，就可以了。
- 互斥： 至少一个资源是被排他性独享的，其他线程必须处于等待状态，直到资源被释放。
- 持有和等待：goroutine 持有一个资源，并且还在请求其它 goroutine 持有的资源，也就是咱们常说的“吃着碗里，看着锅里”的意思。
- 不可剥夺：资源只能由持有它的 goroutine 来释放。
- 环路等待：一般来说，存在一组等待进程，P={P1，P2，…，PN}，P1 等待 P2 持有的资源，P2 等待 P3 持有的资源，依此类推，最后是 PN 等待 P1 持有的资源，这就形成了一个环路等待的死结。
### 锁的性能。
- 锁是性能下降的“罪魁祸首”之一，所以，有效地降低锁的竞争，就能够很好地提高性能。因此，监控关键互斥锁上等待的 goroutine 的数量，是我们分析锁竞争的激烈程度的一个重要指标。
### 源码分析。
- sync.mutex源代码分析：https://colobu.com/2018/12/18/dive-into-sync-mutex/
- golang源码分析sync.Mutex概述：https://studygolang.com/articles/17017
### Mutex知识地图。
![avatar](https://static001.geekbang.org/resource/image/5a/0b/5ayy6cd9ec9fe0bcc13113302056ac0b.jpg)
### 什么是RWMutex？
- 标准库中的RWMutex是一个reader/writer互斥锁。RWMutex在某一时刻只能由任意数量的reader持有，或者是只被单个的writer持有。只要有一个线程在执行写操作，其它的线程都不能执行读写操作。
- Lock/Unlock：写操作时调用的方法。如果锁已经被reader或者writer持有，那么Lock方法会一直阻塞，直到能获取到锁；Unlock则是配对的释放锁的方法。
- RLock/RUnlock：读操作时调用的方法。如果锁已经被writer持有的话，RLock方法会一直阻塞，直到能获取到锁，否则就直接返回；而RUnlock是reader释放锁的方法。
- RLocker：这个方法的作用是为读操作返回一个Locker接口的对象。它的Lock方法会调用RWMutex的RLock方法，它的Unlock方法会调用RWMutex的RUnlock方法。
- 遇到可以明确区分reader和writer goroutine的场景，且有大量的并发读、少量的并发写，并且有强烈的性能需求，就可以考虑使用读写锁RWMutex替换Mutex。