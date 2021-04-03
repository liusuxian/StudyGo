### Go标准库中的WaitGroup提供了三个方法
``` go
func (wg *WaitGroup) Add(delta int)
func (wg *WaitGroup) Done()
func (wg *WaitGroup) Wait()
```
- Add: 用来设置WaitGroup的计数值。
- Done: 用来将WaitGroup的计数值减1，其实就是调用了Add(-1)。
- Wait: 调用这个方法的goroutine会一直阻塞，直到WaitGroup的计数值变为0。