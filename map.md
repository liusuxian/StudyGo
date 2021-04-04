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
- map是无序的，如果我们想要保证元素有序，比如按照元素插入的顺序进行遍历，可以使用辅助的数据结构，比如: https://github.com/elliotchance/orderedmap
### 使用map的2种常见错误
- 常见错误一：未初始化，和slice或者Mutex、RWmutex等struct类型不同，map对象必须在使用之前初始化。如果不初始化就直接赋值的话，会出现panic异常。从一个nil的map对象中获取值不会panic，而是会得到零值。有时候map作为一个struct字段的时候，就很容易忘记初始化。
