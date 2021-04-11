package iface

import "crontab/src/common"

type IJobManager interface {

	SaveJob(job *common.Job) (oldJob *common.Job , err error )
}