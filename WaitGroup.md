### Go标准库中的WaitGroup提供了三个方法
``` go
func (wg *WaitGroup) Add(delta int)
func (wg *WaitGroup) Done()
func (wg *WaitGroup) Wait()
```
- Add: 用来设置WaitGroup的计数值。
- Done: 用来将WaitGroup的计数值减1，其实就是调用了Add(-1)。
- Wait: 调用这个方法的goroutine会一直阻塞，直到WaitGroup的计数值变为0。
### WaitGroup的数据结构。
``` go
type WaitGroup struct {
    // 避免复制使用的一个技巧，可以告诉vet工具违反了复制使用的规则
    noCopy noCopy
    // 64bit(8bytes)的值分成两段，高32bit是计数值，低32bit是waiter的计数
    // 另外32bit是用作信号量的
    // 因为64bit值的原子操作需要64bit对齐，但是32bit编译器不支持，所以数组中的元素在不同的架构中不一样，具体处理看下面的方法
    // 总之，会找到对齐的那64bit作为state，其余的32bit做信号量
    state1 [3]uint32
}

// 得到state的地址和信号量的地址
func (wg *WaitGroup) state() (statep *uint64, semap *uint32) {
    if uintptr(unsafe.Pointer(&wg.state1))%8 == 0 {
        // 如果地址是64bit对齐的，数组前两个元素做state，后一个元素做信号量
        return (*uint64)(unsafe.Pointer(&wg.state1)), &wg.state1[2]
    } else {
        // 如果地址是32bit对齐的，数组后两个元素用来做state，它可以用来做64bit的原子操作，第一个元素32bit用来做信号量
        return (*uint64)(unsafe.Pointer(&wg.state1[1])), &wg.state1[0]
    }
}
```
- noCopy的辅助字段，主要就是辅助vet工具检查是否通过copy赋值这个WaitGroup实例。
- state1，一个具有复合意义的字段，包含WaitGroup的计数、阻塞在检查点的waiter数和信号量。