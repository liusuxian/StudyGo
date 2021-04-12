### 原子操作的基础知识。
- Package sync/atomic实现了同步算法底层的原子的内存操作原语，我们把它叫做原子操作原语，它提供了一些实现原子操作的方法。之所以叫原子操作，是因为一个原子在执行的时候，其它线程不会看到执行一半的操作结果。在其它线程看来，原子操作要么执行完了，要么还没有执行，就像一个最小的粒子-原子一样，不可分割。
- 需要注意的是，因为需要处理器之间保证数据的一致性，atomic的操作也是会降低性能的。
### Atomic原子操作的应用场景。
- 不涉及对资源复杂的竞争逻辑。
- 实现配置对象的更新和加载。
- 可以使用atomic实现自己定义的基本并发原语。
- 实现lock-free数据结构的基石。
### Atomic提供的方法。
- atomic为了支持int32、int64、uint32、uint64、uintptr、Pointer（Add 方法不支持）类型，分别提供了AddXXX、CompareAndSwapXXX、SwapXXX、LoadXXX、StoreXXX等方法。
- atomic操作的对象是一个地址，你需要把可寻址的变量的地址作为参数传递给方法，而不是把变量的值传递给方法。
- Add方法就是给第一个参数地址中的值增加一个delta值。对于有符号的整数来说，delta可以是一个负数，相当于减去一个值。对于无符号的整数和uintptr类型来说，可以利用计算机补码的规则，把减法变成加法。以uint32类型为例：AddUint32(&x, ^uint32(c-1))。尤其是减1这种特殊的操作，我们可以简化为：AddUint32(&x, ^uint32(0))。
- CAS（CompareAndSwap）在CAS的方法签名中，需要提供要操作的地址、原数据值、新值，以int32为例，这个方法会比较当前addr地址里的值是不是old，如果不等于old，就返回false；如果等于old，就把此地址的值替换成new值，返回true。这就相当于“判断相等才替换”。
``` go
func CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool)
```
- Swap如果不需要比较旧值，只是比较粗暴地替换的话，就可以使用Swap方法，它替换后还可以返回旧值。
- Load方法会取出addr地址中的值，即使在多处理器、多核、有CPU cache的情况下，这个操作也能保证Load是一个原子操作。
- Store方法会把一个值存入到指定的addr地址中，即使在多处理器、多核、有CPU cache的情况下，这个操作也能保证Store是一个原子操作。别的goroutine通过Load读取出来，不会看到存取了一半的值。
- atomic还提供了一个特殊的类型：Value。它可以原子地存取对象类型，但也只能存取，不能CAS和Swap，常常用在配置变更等场景中。
### 第三方库的扩展。
- [uber-go/atomic](https://github.com/uber-go/atomic) 它定义和封装了几种与常见类型相对应的原子操作类型，这些类型提供了原子操作的方法。这些类型包括Bool、Duration、Error、Float64、Int32、Int64、String、Uint32、Uint64等。比如Bool类型，提供了CAS、Store、Swap、Toggle等原子方法，还提供String、MarshalJSON、UnmarshalJSON等辅助方法。
### 使用Atomic实现Lock-Free queue。
``` go
package queue

import (
    "sync/atomic"
    "unsafe"
)

// lock-free的queue
type LKQueue struct {
    head unsafe.Pointer
    tail unsafe.Pointer
}

// 通过链表实现，这个数据结构代表链表中的节点
type node struct {
    value interface{}
    next  unsafe.Pointer
}

func NewLKQueue() *LKQueue {
    n := unsafe.Pointer(&node{})
    return &LKQueue{head: n, tail: n}
}

// 入队
func (q *LKQueue) Enqueue(v interface{}) {
    n := &node{value: v}
    for {
        tail := load(&q.tail)
        next := load(&tail.next)
        if tail == load(&q.tail) { // 尾还是尾
            if next == nil { // 还没有新数据入队
                if cas(&tail.next, next, n) { //增加到队尾
                    cas(&q.tail, tail, n) //入队成功，移动尾巴指针
                    return
                }
            } else { // 已有新数据加到队列后面，需要移动尾指针
                cas(&q.tail, tail, next)
            }
        }
    }
}

// 出队，没有元素则返回nil
func (q *LKQueue) Dequeue() interface{} {
    for {
        head := load(&q.head)
        tail := load(&q.tail)
        next := load(&head.next)
        if head == load(&q.head) { // head还是那个head
            if head == tail { // head和tail一样
                if next == nil { // 说明是空队列
                    return nil
                }
                // 只是尾指针还没有调整，尝试调整它指向下一个
                cas(&q.tail, tail, next)
            } else {
                // 读取出队的数据
                v := next.value
                // 既然要出队了，头指针移动到下一个
                if cas(&q.head, head, next) {
                    return v // Dequeue is done.  return
                }
            }
        }
    }
}

// 将unsafe.Pointer原子加载转换成node
func load(p *unsafe.Pointer) (n *node) {
    return (*node)(atomic.LoadPointer(p))
}

// 封装CAS,避免直接将*node转换成unsafe.Pointer
func cas(p *unsafe.Pointer, old, new *node) (ok bool) {
    return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
```
### Atomic的知识地图。
![avatar](https://github.com/liusuxian/learning_golang/blob/master/img/Atomic.jpg)