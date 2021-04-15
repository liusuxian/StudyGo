### Go标准库中的WaitGroup提供了三个方法。
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
- noCopy的辅助字段，主要就是辅助vet工具检查是否通过copy赋值这个WaitGroup实例。如果你想要自己定义的数据结构不被复制使用，或者说不能通过vet工具检查出复制使用的报警，就可以通过嵌入noCopy这个数据类型来实现。noCopy是静态检查。
- state1，一个具有复合意义的字段，包含WaitGroup的计数、阻塞在检查点的waiter数和信号量。
- 因为对64位整数的原子操作要求整数的地址是64位对齐的，所以针对64位和32位环境的state字段的组成是不一样的。在64位环境下，state1的第一个元素是waiter数，第二个元素是WaitGroup的计数值，第三个元素是信号量。如图。
<img src="https://github.com/liusuxian/learning_golang/blob/master/img/WaitGroup1.jpg" width = "60%" height = "60%" alt="image-name"/>

- 在32位环境下，如果state1不是64位对齐的地址，那么state1的第一个元素是信号量，后两个元素分别是waiter数和计数值。如图。
<img src="https://github.com/liusuxian/learning_golang/blob/master/img/WaitGroup2.jpg" width = "60%" height = "60%" alt="image-name"/>

### 使用WaitGroup时的常见错误。
- 常见问题一：计数器设置为负值。WaitGroup的计数器的值必须大于等于0。我们在更改这个计数值的时候，WaitGroup会先做检查，如果计数值被设置为负数，就会导致panic。一般情况下，有两种方法会导致计数器设置为负数。第一种方法是：调用Add的时候传递一个负数。如果你能保证当前的计数器加上这个负数后还是大于等于0的话，也没有问题，否则就会导致panic。第二个方法是：调用Done方法的次数过多，超过了WaitGroup的计数值。使用WaitGroup的正确姿势是，预先确定好WaitGroup的计数值，然后调用相同次数的Done完成相应的任务。
- 常见问题二：不期望的Add时机。在使用WaitGroup的时候，你一定要遵循的原则就是，等所有的Add方法调用之后再调用Wait，否则就可能导致panic或者不期望的结果。
- 常见问题三：前一个Wait还没结束就重用WaitGroup。WaitGroup是可以重用的。只要WaitGroup的计数值恢复到零值的状态，那么它就可以被看作是新创建的WaitGroup，被重复使用。但是如果我们在WaitGroup的计数值还没有恢复到零值的时候就重用，就会导致程序panic。总结一下：WaitGroup虽然可以重用，但是是有一个前提的，那就是必须等到上一轮的Wait完成之后，才能重用WaitGroup执行下一轮的Add/Wait，如果你在Wait还没执行完的时候就调用下一轮Add方法，就有可能出现panic。
### 如何避免错误使用WaitGroup的情况。
- 不重用WaitGroup。新建一个WaitGroup不会带来多大的资源开销，重用反而更容易出错。
- 保证所有的Add方法调用都在Wait之前。
- 不传递负数给Add方法，只通过Done来给计数值减1。
- 不做多余的Done方法调用，保证Add的计数值和Done方法调用的数量是一样的。
- 不遗漏Done方法的调用，否则会导致Wait hang住无法返回。
### WaitGroup的知识地图。
<img src="https://github.com/liusuxian/learning_golang/blob/master/img/WaitGroup3.jpg" width = "60%" height = "60%" alt="image-name"/>
