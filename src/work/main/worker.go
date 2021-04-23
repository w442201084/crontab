package main

import (
	"crontab/src/common"
	"crontab/src/work"
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
	if err = common.InitConfig("./client.json") ; nil != err {
		fmt.Println("加载配置文件失败..." , err)
		return
	}

	// 服务注册
	if err = work.InitRegister(); nil != err  {
		fmt.Println("服务注册失败..." , err)
		return
	}

	// 启动执行器
	work.InitExecutor()

	// 启动任务调度器
	work.InitScheduler()

	// 启动etcd连接
	if err = work.InitEtcdManager() ; nil != err {
		fmt.Println("启动etcd连接失败..." , err)
		return
	}

	select {

	}

}
