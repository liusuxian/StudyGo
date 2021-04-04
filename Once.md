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
    done uint32
    m    Mutex
}

func (o *Once) Do(f func()) {
    if atomic.LoadUint32(&o.done) == 0 {
        o.doSlow(f)
    }
}

func (o *Once) doSlow(f func()) {
    o.m.Lock()
    defer o.m.Unlock()
    // 双检查
    if o.done == 0 {
        defer atomic.StoreUint32(&o.done, 1)
        f()
    }
}
```