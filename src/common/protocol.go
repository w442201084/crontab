package common

import (
	"github.com/gorhill/cronexpr"
	"net"
	"strings"
	"time"
)

// 定时任务
type Job struct {
	// 任务名称
	Name string `json:"name"`
	// 执行的shell命令
	Command string `json:"command"`
    // cron表达式
	CronExpr string `json:"cronExpr"`
}

// HTTP接口返回JSON格式
type ResponseFormat struct {
	Code int `json:"code"`

	Msg string `json:"msg"`

	Data interface{} `json:"data"`
}

// 变化的事件
type JobEvent struct {
	EventType int // SAVE 、 DELETE
	Job *Job
}

// 结合第三方包的下次调度计划
type JobSchedulerPlan struct {
	Job *Job // 调度的任务
	Expr *cronexpr.Expression // 第三方包解析好的一个表达式
	NextTime time.Time // 下次调度的时间
}

// 任务的执行状态
type JobExecuteStatus struct {
	Job *Job
	PlanTime time.Time //理论上的执行时间
	RealTime time.Time // 实际的调度时间
}

// 任务执行结果
type JobExecuteResult struct {
	JobExecuteStatus *JobExecuteStatus // 执行的状态
	OutPut []byte // 执行结果
	Err error // 错误原因
	StartTime time.Time // 脚本开始时间
	EndTime time.Time // 脚本结束时间
}

/**
通过etcd里面存储的KEY截取里面的名称
比如：
	从 /cron/jobs/job1里面获取job1
 */
func ExtraJobName ( jobKey string ) string {
	return strings.Trim(jobKey , JOB_SAVE_DIR)
}


/**
	构建一个任务变化的事件
	主要是 1、更新任务  2、删除任务
 */
func BuildEventJob( eventType int , job *Job ) (jobEvent *JobEvent){
	return &JobEvent{
		EventType: eventType,
		Job: job,
	}
}

// 根据创建的任务生成一个任务执行计划
func BuildJobSchedulerPlan( job *Job ) ( *JobSchedulerPlan , error ){
	var (
		expr *cronexpr.Expression
		err error
	)

	if expr , err = cronexpr.Parse( job.CronExpr ); nil != err {
		return nil , err
	} else {
		// 返回任务调度计划
		return &JobSchedulerPlan{
			Job: job,
			Expr: expr,
			NextTime: expr.Next(time.Now()),
		} , nil
	}
}

/**
创建一个任务状态
 */
func BuildJobExecuteStatus( plan *JobSchedulerPlan ) *JobExecuteStatus{
	return &JobExecuteStatus{
		Job: plan.Job ,
		PlanTime: plan.NextTime , //计划调度时间
		RealTime: time.Now(), // 真实调度时间
	}
}

func GetLocalIp () (ipv4 string , err error){
	var (
		addrs []net.Addr
		addr net.Addr
		ipNet *net.IPNet // IP地址
		isIpNet bool
	)

	if addrs , err = net.InterfaceAddrs() ; nil != err {
		return
	}

	// 取第一个非local的host
	for _ , addr = range addrs {
		// 这个网络地址是IP地址、ipv4、ipv6
		if ipNet , isIpNet = addr.(*net.IPNet) ; isIpNet && !(ipNet.IP.IsLoopback())  {
			// 跳过ipv6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}
	err = ERR_CANT_GET_LOCAL_IP
	return
}
