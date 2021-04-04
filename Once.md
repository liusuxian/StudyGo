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