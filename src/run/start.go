package main

import (
	"crontab/src/common"
	"crontab/src/master"
	"fmt"
	"runtime"
)

func initEnv() {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	// 机器的CPU数量
	runtime.GOMAXPROCS(2)
}

func main() {
	var (
		err error
	)
	// 初始化线程
	initEnv()
	// 加载配置文件
	if err = common.InitConfig("./src/config/master.json") ; nil != err {
		goto ERR
	}
	// 启动etcd连接
	if err = master.InitEtcdManager() ; nil != err {
		goto ERR
	}
	// 启动http服务-API
	if err = master.InitApiServe( master.GlobalJobManager ) ; nil != err {
		goto ERR
	}
	select {}

ERR:
	fmt.Println( err )
}