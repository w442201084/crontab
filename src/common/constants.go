package common

const (
	// 任务保存路径
	JOB_SAVE_DIR = "/cron/jobs/"

	// 任务杀死路径
	JOB_KILLER_DIR = "/cron/killer/"

	// 分布式锁的路径
	JOB_LOCK_DIR = "/cron/lock/"

	// 保存任务事件
	JOB_EVENT_SAVE = 1

	// 删除任务事件
	JOB_EVENT_DELETE = 2
)