### 原子操作的基础知识。
- Package sync/atomic实现了同步算法底层的原子的内存操作原语，我们把它叫做原子操作原语，它提供了一些实现原子操作的方法。之所以叫原子操作，是因为一个原子在执行的时候，其它线程不会看到执行一半的操作结果。在其它线程看来，原子操作要么执行完了，要么还没有执行，就像一个最小的粒子-原子一样，不可分割。
### atomic原子操作的应用场景。
- 不涉及对资源复杂的竞争逻辑。
- 实现配置对象的更新和加载。
- 可以使用atomic实现自己定义的基本并发原语。
- 实现lock-free数据结构的基石。
### atomic提供的方法。
- atomic为了支持int32、int64、uint32、uint64、uintptr、Pointer（Add 方法不支持）类型，分别提供了AddXXX、CompareAndSwapXXX、SwapXXX、LoadXXX、StoreXXX等方法。
- atomic操作的对象是一个地址，你需要把可寻址的变量的地址作为参数传递给方法，而不是把变量的值传递给方法。
- Add方法就是给第一个参数地址中的值增加一个delta值。对于有符号的整数来说，delta可以是一个负数，相当于减去一个值。对于无符号的整数和uintptr类型来说，可以利用计算机补码的规则，把减法变成加法。以uint32类型为例：AddUint32(&x, ^uint32(c-1))。尤其是减1这种特殊的操作，我们可以简化为：AddUint32(&x, ^uint32(0))。
- CAS（CompareAndSwap）在CAS的方法签名中，需要提供要操作的地址、原数据值、新值，以int32为例，这个方法会比较当前addr地址里的值是不是old，如果不等于old，就返回false；如果等于old，就把此地址的值替换成new值，返回true。这就相当于“判断相等才替换”。
``` go
func CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool)
```