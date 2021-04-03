### readers-writers问题一般有三类，基于对读和写操作的优先级，读写锁的设计和实现也分成三类。
- Read-preferring：读优先的设计可以提供很高的并发性，但是在竞争激烈的情况下可能会导致写饥饿。这是因为如果有大量的读，这种设计会导致只有所有的读都释放了锁之后，写才可能获取到锁。
- Write-preferring：写优先的设计意味着，如果已经有一个writer在等待请求锁的话，它会阻止新来的请求锁的reader获取到锁，所以优先保障writer。当然，如果有一些reader已经请求了锁的话，新请求的writer也会等待已经存在的reader都释放锁之后才能获取。所以写优先级设计中的优先权是针对新来的请求而言的。这种设计主要避免了writer的饥饿问题。
- 不指定优先级：这种设计比较简单，不区分reader和writer优先级，某些场景下这种不指定优先级的设计反而更有效，因为第一类优先级会导致写饥饿，第二类优先级可能会导致读饥饿，这种不指定优先级的访问不再区分读写，大家都是同一个优先级，解决了饥饿的问题。
### 什么是RWMutex？
- 标准库中的RWMutex是一个reader/writer互斥锁。RWMutex在某一时刻只能由任意数量的reader持有，或者是只被单个的writer持有。只要有一个线程在执行写操作，其它的线程都不能执行读写操作。
- Lock/Unlock：写操作时调用的方法。如果锁已经被reader或者writer持有，那么Lock方法会一直阻塞，直到能获取到锁；Unlock则是配对的释放锁的方法。
- RLock/RUnlock：读操作时调用的方法。如果锁已经被writer持有的话，RLock方法会一直阻塞，直到能获取到锁，否则就直接返回；而RUnlock是reader释放锁的方法。
- RLocker：这个方法的作用是为读操作返回一个Locker接口的对象。它的Lock方法会调用RWMutex的RLock方法，它的Unlock方法会调用RWMutex的RUnlock方法。
- 遇到可以明确区分reader和writer goroutine的场景，且有大量的并发读、少量的并发写，并且有强烈的性能需求，就可以考虑使用读写锁RWMutex替换Mutex。
- Go标准库中的RWMutex设计是Write-preferring方案。一个正在阻塞的Lock调用会排除新的reader请求到锁。
### RWMutex包含一个Mutex，以及四个辅助字段writerSem、readerSem、readerCount和readerWait。
``` go
type RWMutex struct {
  w           Mutex   // 互斥锁解决多个writer的竞争
  writerSem   uint32  // writer信号量
  readerSem   uint32  // reader信号量
  readerCount int32   // reader的数量
  readerWait  int32   // writer等待完成的reader的数量
}

const rwmutexMaxReaders = 1 << 30
```
- 字段w：为writer的竞争锁而设计。
- 字段readerCount：记录当前reader的数量（以及是否有writer竞争锁）。没有writer竞争或持有锁时，readerCount和我们正常理解的reader的计数是一样的。如果有writer竞争锁或者持有锁时，那么readerCount不仅仅承担着reader的计数功能，还能够标识当前是否有writer竞争或持有锁。
- 字段readerWait：记录writer请求锁时需要等待read完成的reader的数量。
- 字段writerSem和字段readerSem：都是为了阻塞设计的信号量。
- 常量rwmutexMaxReaders，定义了最大的reader数量。
### RWMutex的3个踩坑点。
- 坑点 1：不可复制，同Mutex。
- 坑点 2：重入导致死锁，同Mutex。第二种死锁的场景有点隐蔽。我们知道，有活跃reader的时候，writer会等待，如果我们在reader的读操作时调用writer的写操作（它会调用Lock方法），那么这个reader和writer就会形成互相依赖的死锁状态。Reader想等待writer完成后再释放锁，而writer需要这个reader释放锁之后，才能不阻塞地继续执行。这是一个读写锁常见的死锁场景。第三种死锁的场景更加隐蔽。当一个writer请求锁的时候，如果已经有一些活跃的reader，它会等待这些活跃的reader完成，才有可能获取到锁，但是如果之后活跃的reader再依赖新的reader的话，这些新的reader就会等待writer释放锁之后才能继续执行，这就形成了一个环形依赖： writer依赖活跃的reader->活跃的reader依赖新来的reader->新来的reader依赖writer。
- 坑点 3：释放未加锁的RWMutex，同Mutex。
### RWMutex的知识地图。
![avatar](https://static001.geekbang.org/resource/image/69/42/695b9aa6027b5d3a61e92cbcbba10042.jpg)