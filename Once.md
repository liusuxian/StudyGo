### 初始化单例资源的方法。
- 1、定义package级别的变量，这样程序在启动的时候就可以初始化：
``` go
package abc

import time

var startTime = time.Now()
```
- 2、在init函数中进行初始化：
``` go
package abc

var startTime time.Time

func init() {
    startTime = time.Now()
}
```
- 3、在main函数开始执行的时候，执行一个初始化的函数：
``` go
package abc

var startTime time.Tim

func initApp() {
    startTime = time.Now()
}

func main() {
    initApp()
}
```
- 4、Go标准库的Once。Once可以用来执行且仅仅执行一次动作，常常用于单例对象的初始化场景。
- 这些方法都是线程安全的，并且除第一种方法外，后面的方法还可以根据传入的参数实现定制化的初始化操作。
### Once的使用场景。
- sync.Once只暴露了一个方法Do，你可以多次调用Do方法，但是只有第一次调用Do方法时f参数才会执行，这里的f是一个无参数无返回值的函数。在实际的使用中，绝大多数情况下，会使用闭包的方式去初始化外部的一个资源，比如：
``` go
var addr = "baidu.com"
var conn net.Conn
var err error

once.Do(func() {
    conn, err = net.Dial("tcp", addr)
})
```
- Once常常用来初始化单例资源，或者并发访问只需初始化一次的共享资源，或者在测试的时候初始化一次测试资源。
### 很值得学习的math/big/sqrt.go中实现的一个数据结构，它通过Once封装了一个只初始化一次的值。
``` go
// 值是3.0或者0.0的一个数据结构
var threeOnce struct {
    sync.Once
    v *Float
}

// 返回此数据结构的值，如果还没有初始化为3.0，则初始化
func three() *Float {
    // 使用Once初始化
    threeOnce.Do(func() {
        threeOnce.v = NewFloat(3.0)
    })
    return threeOnce.v
}
```
- 当使用Once的时候，可以尝试采用这种结构，将值和Once封装成一个新的数据结构，提供只初始化一次的值。
### 如何实现一个Once？
- 第一种方法，只需使用一个flag标记是否初始化过即可，用atomic原子操作这个flag，比如下面的实现，但是这个实现有一个很大的问题，就是如果参数f执行很慢的话，后续调用Do方法的goroutine虽然看到done已经设置为执行过了，但是获取某些初始化资源的时候可能会得到空的资源，因为f还没有执行完。
``` go
type Once struct {
    done uint32
}

func (o *Once) Do(f func()) {
    if !atomic.CompareAndSwapUint32(&o.done, 0, 1) {
        return
    }
    f()
}
```
- 第二种方法，使用一个互斥锁，这样初始化的时候如果有并发的goroutine，就会进入doSlow方法。互斥锁的机制保证只有一个goroutine进行初始化，同时利用双检查的机制（double-checking），再次判断o.done是否为0，如果为0，则是第一次执行，执行完毕后，就将o.done设置为1，然后释放锁。即使此时有多个goroutine同时进入了doSlow方法，因为双检查的机制，后续的goroutine会看到o.done的值为1，也不会再次执行f。这样既保证了并发的goroutine会等待f完成，而且还不会多次执行f。
``` go
type Once struct {
    sync.Mutex
    done uint32
}

func (o *Once) Do(f func()) {
    if atomic.LoadUint32(&o.done) == 0 {
        o.doSlow(f)
    }
}

func (o *Once) doSlow(f func()) {
    o.Lock()
    defer o.Unlock()
    // 双检查
    if o.done == 0 {
        defer atomic.StoreUint32(&o.done, 1)
        f()
    }
}
```
### 使用Once可能出现的2种错误。
- 第一种错误：死锁，Do方法仅会执行一次f，但是如果f中再次调用这个Once的Do方法的话，就会导致死锁的情况出现。这还不是无限递归的情况，而是的的确确的Lock的递归调用导致的死锁。想要避免这种情况的出现，就不要在f参数中调用当前的这个Once，不管是直接的还是间接的。
- 第二种错误：未初始化，如果f方法执行的时候panic，或者f执行初始化资源的时候失败了，这个时候，Once还是会认为初次执行已经成功了，即使再次调用Do方法，也不会再次执行f。那么这种初始化未完成的问题该怎么解决呢？我们可以自己实现一个类似Once的并发原语，既可以返回当前调用Do方法是否正确完成，还可以在初始化失败后调用Do方法再次尝试初始化，直到初始化成功才不再初始化了。
``` go
// 一个功能更加强大的Once
type Once struct {
    sync.Mutex
    done uint32
}

// 传入的函数f有返回值error，如果初始化失败，需要返回失败的error
// Do方法会把这个error返回给调用者
func (o *Once) Do(f func() error) error {
    // fast path
    if atomic.LoadUint32(&o.done) == 1 {
        return nil
    }
    return o.slowDo(f)
}

// 如果还没有初始化
func (o *Once) slowDo(f func() error) error {
    o.Lock()
    defer o.Unlock()
    var err error
    // 双检查，还没有初始化
    if o.done == 0 {
        err = f()
        // 初始化成功才将标记置为已初始化
        if err == nil {
            atomic.StoreUint32(&o.done, 1)
        }
    }
    return err
}

// Done 返回此Once是否执行成功过
// 如果执行成功过则返回true
// 如果没有执行成功过或者正在执行，返回false
func (o *Once) Done() bool {
    return atomic.LoadUint32(&o.done) == 1
}
```
### 扩展官方的Once，提供一个Done方法，返回此Once是否执行过。
``` go
// Once 是一个扩展的sync.Once类型，提供了一个Done方法
type Once struct {
    sync.Once
}

// Done 返回此Once是否执行过
// 如果执行过则返回true
// 如果没有执行过或者正在执行，返回false
func (o *Once) Done() bool {
    return atomic.LoadUint32((*uint32)(unsafe.Pointer(&o.Once))) == 1
}
```
### Once的知识地图。
![avatar](https://github.com/liusuxian/learning_golang/blob/master/img/Once.jpg)