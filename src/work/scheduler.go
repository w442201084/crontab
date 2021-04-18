package work

import (
	"crontab/src/common"
	"fmt"
	"time"
)

// 任务调度协程
type Scheduler struct {

	// etcd任务事件的队列
	jobEventChan chan *common.JobEvent

	// 任务调度计划表
	JobPlanTable map[string] *common.JobSchedulerPlan


	// 任务执行表。用来记录任务执行状态
	JobExecuteTable map[string] *common.JobExecuteStatus

	// 接收任务执行完之后的回调，用于删除JobExecuteTable里面的任务
	JobExecuteResultsChan chan *common.JobExecuteResult
}

var (
	GlobalScheduler *Scheduler
)

// 检测所有任务 看看时间是否到期
func (this *Scheduler) scheduleLoop() {

	var (
		jobEvent *common.JobEvent
		schedulerAfter time.Duration
		schedulerTime *time.Timer
		jobExecuteResults *common.JobExecuteResult
	)

	// 初始化执行任务的调度器
	schedulerAfter = this.TryExecuteScheduler()
	schedulerTime = time.NewTimer( schedulerAfter )
	for {
		select {
			case jobEvent = <- this.jobEventChan : // 监听任务的变化
				this.HandlerOfJobEvent(jobEvent)
			case <- schedulerTime.C : // 最新的任务到期了
			case jobExecuteResults = <- this.JobExecuteResultsChan : // 监听任务执行完之后的回调
				this.HandlerExecuteResults(jobExecuteResults)

		}
		// 重新调度执行任务的调度器
		schedulerAfter = this.TryExecuteScheduler()
		// 重置调度间隔
		schedulerTime.Reset( schedulerAfter )
	}

}

// 删除JobExecuteTable状态表里面的任务
func (this *Scheduler) HandlerExecuteResults(jobExecuteResults *common.JobExecuteResult) {
	var (
		jobExecuteResultExist bool
	)
	if _ , jobExecuteResultExist = this.JobExecuteTable[jobExecuteResults.JobExecuteStatus.Job.Name];jobExecuteResultExist {
		delete( this.JobExecuteTable , jobExecuteResults.JobExecuteStatus.Job.Name )
	}
}

// 根绝任务的下次执行时间算出当前扫码所有任务的时间
func (this *Scheduler) TryExecuteScheduler() (schedulerAfter time.Duration) {

	var (
		jobPlan *common.JobSchedulerPlan
		nowTime time.Time
		nearTime *time.Time
	)
	// 没有任务执行
	if len(this.JobPlanTable) <= 0 {
		schedulerAfter = 1 * time.Second
		return
	}
	nowTime = time.Now()
	// 遍历所有任务
	for _ , jobPlan = range this.JobPlanTable {
		// 需要开始执行任务了
		if jobPlan.NextTime.Before( nowTime ) || jobPlan.NextTime.Equal( nowTime ) {
			/** 有可能上次这个任务还没有执行完 */
			// 执行任务
			this.ExecuteJob(jobPlan)
			// 更新下次执行时间
			jobPlan.NextTime = jobPlan.Expr.Next(nowTime)
		}
		// 统计最近一次任务的时间
		if nearTime == nil || jobPlan.NextTime.Before( *nearTime ){
			nearTime = &jobPlan.NextTime
		}
	}
	// 计算出下次执行任务的时间也就是下次遍历整个计划任务时需要休眠的时间
	schedulerAfter = (*nearTime).Sub(nowTime) // 下次执行的时间减去当前时间
	return
}

func ( this *Scheduler ) ExecuteJob(jobPlan *common.JobSchedulerPlan) {

	// 执行任务可能会很久，1分钟可能会调度60次，但是只能执行一次、防治并发

	var (
		jobExecuteExist bool
		jobExecuteStatus *common.JobExecuteStatus
	)

	// 如果任务已经被写过了
	if jobExecuteStatus , jobExecuteExist = this.JobExecuteTable[jobPlan.Job.Name]; jobExecuteExist {
		fmt.Println("任务正在执行中..." , jobPlan.Job.Name)
		// 说明正在运行
		return
	}
	// 保存任务状态
	jobExecuteStatus = common.BuildJobExecuteStatus( jobPlan )
	this.JobExecuteTable[jobPlan.Job.Name] = jobExecuteStatus

	// 执行任务
	fmt.Println("执行任务..." , jobPlan.Job.Name)
	GlobalExecutor.ExecuteJob( jobExecuteStatus )


}

/**
不停地执行监听任务的变化，保证ectd里面的任务的列表和map里面存储的一致
 */
func ( this *Scheduler ) HandlerOfJobEvent(jobEvent *common.JobEvent) {

	var (
		jobSchedulerPlan *common.JobSchedulerPlan
		err error
		jobExist bool
	)

	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE : // 更新任务
		if jobSchedulerPlan , err  = common.BuildJobSchedulerPlan( jobEvent.Job ); nil != err {
			return
		}
		fmt.Println("受到监听：..." , jobEvent.Job.Name)
		this.JobPlanTable[jobEvent.Job.Name] = jobSchedulerPlan
	case common.JOB_EVENT_DELETE : // 删除任务
		// 如果计划表里面存在这个任务就执行删除
		if jobSchedulerPlan , jobExist = this.JobPlanTable[jobEvent.Job.Name]; jobExist {
			delete( this.JobPlanTable , jobEvent.Job.Name )
		}
	}
}

// 当任务发生变化，就发起推送
func (this *Scheduler) PushJobPlan(jobEvent *common.JobEvent) {
	this.jobEventChan <- jobEvent
}

func InitScheduler() {
	GlobalScheduler = &Scheduler{
		jobEventChan: make( chan *common.JobEvent , 1000 ),
		JobPlanTable: make( map[string] *common.JobSchedulerPlan ),
		JobExecuteTable: make(map[string] *common.JobExecuteStatus),
		JobExecuteResultsChan: make(chan *common.JobExecuteResult , 1000),
	}
	go GlobalScheduler.scheduleLoop()
}

// 接收任务执行结果到chan
func ( this *Scheduler ) PushExecuteResults (result *common.JobExecuteResult) {
	this.JobExecuteResultsChan <- result
}
