package iface

import "crontab/src/common"

type IJobManager interface {

	SaveJob(job *common.Job) (oldJob *common.Job , err error )

	DeleteJob(jobName string) (oldJob *common.Job , err error )

	ListsJob() (jobLists []*common.Job , err error )
}