/*
 * @Author: liusuxian 382185882@qq.com
 * @Date: 2023-03-07 19:21:05
 * @LastEditors: liusuxian 382185882@qq.com
 * @LastEditTime: 2023-03-07 20:21:51
 * @FilePath: /playlet-server/Users/liusuxian/Desktop/project-code/golang-project/StudyGo/goroutine/debug_gpm/trace.go
 * @Description:
 *
 * Copyright (c) 2023 by ${git_name_email}, All Rights Reserved.
 */
package main

import (
	"fmt"
	"os"
	"runtime/trace"
)

func main() {
	var f *os.File
	var err error
	// 创建一个trace文件
	if f, err = os.Create("trace.out"); err != nil {
		panic(err)
	}
	defer f.Close()
	// 启动trace
	if err = trace.Start(f); err != nil {
		panic(err)
	}
	// 正常要调试的业务
	fmt.Println("Hello GPM")
	// 停止trace
	trace.Stop()
}
