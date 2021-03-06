### map的基本使用方法。
``` go
map[K]V
```
- key类型的K必须是可比较的（comparable），也就是可以通过==和!=操作符进行比较；value的值和类型无所谓，可以是任意的类型，或者为nil。
- 在Go语言中，bool、整数、浮点数、复数、字符串、指针、Channel、接口都是可比较的，包含可比较元素的struct和数组，这俩也是可比较的，而slice、map、函数值都是不可比较的。
- 那么，上面这些可比较的数据类型都可以作为map的key吗？显然不是。通常情况下，我们会选择内建的基本类型，比如整数、字符串做key的类型，因为这样最方便。这里有一点需要注意，如果使用struct类型做key其实是有坑的，因为如果struct的某个字段值修改了，查询map时无法获取它add进去的值，如下面的例子：
``` go
type mapKey struct {
    key int
}

func main() {
    var m = make(map[mapKey]string)
    var key = mapKey{10}
    m[key] = "hello"
    fmt.Printf("m[key]=%s\n", m[key])
    // 修改key的字段的值后再次查询map，无法获取刚才add进去的值
    key.key = 100
    fmt.Printf("再次查询m[key]=%s\n", m[key])
}
```
- 如果非要使用struct作为key，我们要保证struct对象在逻辑上是不可变的，这样才会保证map的逻辑没有问题。
- map是无序的，如果我们想要保证元素有序，比如按照元素插入的顺序进行遍历，可以使用辅助的数据结构，比如 [orderedmap](https://github.com/elliotchance/orderedmap)。
### 使用map的2种常见错误。
- 常见错误一：未初始化，和slice或者Mutex、RWmutex等struct类型不同，map对象必须在使用之前初始化。如果不初始化就直接赋值的话，会出现panic异常。从一个nil的map对象中获取值不会panic，而是会得到零值。有时候map作为一个struct字段的时候，就很容易忘记初始化。
- 常见错误二：并发读写，如果没有注意到并发问题，程序在运行的时候就有可能出现并发读写导致的panic。Go内建的map对象不是线程（goroutine）安全的，并发读写的时候运行时会有检查，遇到并发问题就会导致panic。
### 如何实现线程安全的map类型。
- 避免map并发读写panic的方式之一就是加锁，考虑到读写性能，可以使用读写锁提供性能。
- 分片加锁，更高效的并发map，虽然使用读写锁可以提供线程安全的map，但是在大量并发读写的情况下，锁的竞争会非常激烈。锁是性能下降的万恶之源之一。在并发编程中，我们的一条原则就是尽量减少锁的使用。一些单线程单进程的应用（比如Redis等），基本上不需要使用锁去解决并发线程访问的问题，所以可以取得很高的性能。但是对于Go开发的应用程序来说，并发是常用的一个特性，在这种情况下，我们能做的就是，尽量减少锁的粒度和锁的持有时间。减少锁的粒度常用的方法就是分片（Shard），将一把锁分成几把锁，每个锁控制一个分片。Go比较知名的分片并发map的实现是 [orcaman/concurrent-map](https://github.com/orcaman/concurrent-map)。
### sync.Map的实现
- 空间换时间。通过冗余的两个数据结构（只读的read字段、可写的dirty），来减少加锁对性能的影响。对只读字段（read）的操作不需要加锁。
- 优先从read字段读取、更新、删除，因为对read字段的读取不需要锁。
- 动态调整。miss次数多了之后，将dirty数据提升为read，避免总是从dirty中加锁读取。
- double-checking。加锁之后先还要再检查read字段，确定真的不存在才操作dirty字段。
- 延迟删除。删除一个键值只是打标记，只有在提升dirty字段为read字段的时候才清理删除的数据。
### 在以下两个场景中使用sync.Map，会比使用map+RWMutex的方式，性能要好得多。
- 只会增长的缓存系统中，一个key只写入一次而被读很多次。
- 多个goroutine为不相交的键集读、写和重写键值对。
### 扩展其它功能的map实现。
- 带有过期功能的 [timedmap](https://github.com/zekroTJA/timedmap)。
- 使用红黑树实现的key有序的 [treemap](https://pkg.go.dev/github.com/emirpasic/gods/maps/treemap?utm_source=godoc)。
### Map知识地图。
<img src="https://github.com/liusuxian/StudyGo/blob/master/img/Map.jpg" width = "100%" height = "100%" alt="image-name"/>

### 为什么sync.Map中的集合核心方法的实现中，如果read中项目不存在，加锁后还要双检查，再检查一次read？
- 加锁之后先还要再检查read字段，确定真的不存在才操作dirty字段。
### sync.map元素删除的时候只是把它的值设置为nil，那么什么时候这个key才会真正从map对象中删除？
- 在提升dirty字段为read字段的时候才清理删除的数据。