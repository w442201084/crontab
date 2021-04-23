package work

import (
	"context"
	"crontab/src/common"
	"fmt"
	"os/exec"
	"time"
)

// 任务执行器
type Executor struct {

}

var (
	GlobalExecutor *Executor
)

// 执行某个任务
func ( this *Executor ) ExecuteJob ( jobExecuteStatus *common.JobExecuteStatus ) {
	go func() {
		var (
			cmd *exec.Cmd
			err error
			outByte []byte
			jobExeResult *common.JobExecuteResult
			jobLock *JobLock
		)

		// 创建分布式锁
		jobLock = GlobalJobManager.CreateJobLock( jobExecuteStatus.Job.Name )

		//返回的执行结果
		jobExeResult = &common.JobExecuteResult{
			JobExecuteStatus:jobExecuteStatus,
			OutPut: make([]byte , 0),
		}
		jobExeResult.StartTime = time.Now()

		err = jobLock.TryLock()
		defer jobLock.UnLock()

		// 上锁如果失败
		if nil != err {
			jobExeResult.Err = err
			jobExeResult.EndTime = time.Now()
		} else {
			// 上锁成功后重新设置时间
			jobExeResult.StartTime = time.Now()

			cmd = exec.CommandContext(context.TODO() , "/bin/bash" , "-c" ,
				jobExecuteStatus.Job.Command)

			if outByte , err = cmd.CombinedOutput() ; nil != err {
				fmt.Println("执行command失败..." , err)
				jobExeResult.Err = err
			}
			jobExeResult.OutPut = outByte
			fmt.Println("执行结果...." , string(outByte))
			jobExeResult.EndTime = time.Now()


		}
		// 返回执行结果，把这个command从执行状态表里面去掉（JobExecuteTable）
		GlobalScheduler.PushExecuteResults( jobExeResult )
	}()
}

func InitExecutor () {
	GlobalExecutor = &Executor{

	}
}