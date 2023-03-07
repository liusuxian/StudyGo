/*
 * @Author: liusuxian 382185882@qq.com
 * @Date: 2023-03-07 20:33:45
 * @LastEditors: liusuxian 382185882@qq.com
 * @LastEditTime: 2023-03-07 20:33:50
 * @FilePath: /playlet-server/Users/liusuxian/Desktop/project-code/golang-project/StudyGo/goroutine/debug_gpm/trace2.go
 * @Description:
 *
 * Copyright (c) 2023 by ${git_name_email}, All Rights Reserved.
 */
package main

import (
	"fmt"
	"time"
)

func main() {
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		fmt.Println("Hello GPM")
	}
}
