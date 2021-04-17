package common

import "strings"

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

/**
通过etcd里面存储的KEY截取里面的名称
比如：
	从 /cron/jobs/job1里面获取job1
 */
func ExtraJobName ( jobKey string ) string {
	return strings.Trim(jobKey , JOB_SAVE_DIR)
}
