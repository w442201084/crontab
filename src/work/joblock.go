package work

import (
	"context"
	"crontab/src/common"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type JobLock struct {
	kv clientv3.KV
	lease clientv3.Lease
	JobName string // 任务名称
	cancelFunc context.CancelFunc
	leaseId clientv3.LeaseID
	isLocked bool // 是否上锁成功
}

func (this *JobLock) TryLock () (err error) {

	var (
		leaseGrantResponse *clientv3.LeaseGrantResponse
		cancelCtx context.Context
		cancelFunc context.CancelFunc
		leaseId clientv3.LeaseID
		leaseKeepResponseChan <-chan *clientv3.LeaseKeepAliveResponse
		txn clientv3.Txn
		lockKey string
		txnResponse *clientv3.TxnResponse
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
	txn = this.kv.Txn(context.TODO())
	// 锁目录
	lockKey = common.JOB_LOCK_DIR + this.JobName
	// 事务抢锁
	// 如果KEY不存在
	txn.If( clientv3.Compare( clientv3.CreateRevision( lockKey ) , "=" , 0 ) ).Then(
		clientv3.OpPut(lockKey , "" , clientv3.WithLease(leaseId))).Else(clientv3.OpGet(lockKey))
	// 提交事务
	if txnResponse , err = txn.Commit() ; nil != err {
		goto FAIL
	}

	// 判断commit是否成功。
	if !txnResponse.Succeeded {
		err = common.ERR_LOCK_ALREADY_REQUIRED
		goto FAIL
	}

	// 上锁成功
	this.leaseId = leaseId // 记录租约ID
	this.cancelFunc = cancelFunc // 取消上下文的回调函数
	this.isLocked = true

	return

	FAIL:
		cancelFunc() // 取消自动续租
		this.lease.Revoke( context.TODO() , leaseId ) // 释放租约
	return
}

func (this *JobLock) UnLock () {
	// 上锁成功了才进行释放
	if this.isLocked {
		this.cancelFunc() // 取消自动续租的协程
		this.lease.Revoke(context.TODO() , this.leaseId) // 释放租约
	}
}

func InitJobLock( jobName string , kv clientv3.KV , lease clientv3.Lease ) *JobLock {
	return &JobLock{
		kv: kv,
		lease: lease,
		JobName: jobName,
	}
}