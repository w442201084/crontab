package work

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type JobLock struct {
	kv clientv3.KV
	lease clientv3.Lease
	JobName string // 锁任务
	cancelFunc context.CancelFunc
}

func (this *JobLock) TryLock () (err error) {

	var (
		leaseGrantResponse *clientv3.LeaseGrantResponse
		cancelCtx context.Context
		cancelFunc context.CancelFunc
		leaseId clientv3.LeaseID
		leaseKeepResponseChan <-chan *clientv3.LeaseKeepAliveResponse
	)

	// 创建租约 5s
	if leaseGrantResponse , err = this.lease.Grant( context.TODO() , 5 ); nil != err {
		return
	}
	leaseId = leaseGrantResponse.ID
	cancelCtx , cancelFunc = context.WithCancel(context.TODO())
	// 自动续租
	if leaseKeepResponseChan , err = this.lease.KeepAlive( cancelCtx , leaseId ); nil != err {
		goto FAIL
	}

	// 处理自动续租的应答
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)

		for {
			select {
				case keepResp = <- leaseKeepResponseChan : // 自动续租应答
					if nil == keepResp {
						goto END
					}
			}
		}
		END:

	}()

	// 创建事务

	// 事务抢锁


	FAIL:
		cancelFunc() // 取消自动续租
		this.lease.Revoke( context.TODO() , leaseId ) // 释放租约
	return
}

func (this *JobLock) UnLock () {

}

func InitJobLock( jobName string , kv clientv3.KV , lease clientv3.Lease ) *JobLock {
	return &JobLock{
		kv: kv,
		lease: lease,
		JobName: jobName,
	}
}