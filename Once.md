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