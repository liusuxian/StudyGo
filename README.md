# learning_golang
learning golang

#### go run -race 文件名.go 检测并发访问共享资源是否有问题的命令。
- 会输出警告信息，这个警告不但会告诉你有并发问题，而且还会告诉你哪个 goroutine 在哪一行对哪个变量有写操作，同时，哪个 goroutine 在哪一行对哪个变量有读操作，就是这些并发的读写访问，引起了 data race。虽然这个工具使用起来很方便，但是，因为它的实现方式，只能通过真正对实际地址进行读写访问的时候才能探测，所以它并不能在编译的时候发现 data race 的问题。而且，在运行的时候，只有在触发了 data race 之后，才能检测到，如果碰巧没有触发，是检测不出来的。而且，把开启了 race 的程序部署在线上，还是比较影响性能的。
#### go tool compile -race -S 文件名.go 查看计数器命令。
- 在编译的代码中，增加了 runtime.racefuncenter、runtime.raceread、runtime.racewrite、runtime.racefuncexit 等检测 data race 的方法。通过这些插入的指令，Go race detector 工具就能够成功地检测出 data race 问题了。
#### go tool compile -S 文件名.go 查看汇编代码命令。
