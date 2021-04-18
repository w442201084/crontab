package work

import (
	"context"
	"crontab/src/common"
	"crontab/src/helper"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

/**
监听etcd里面的任务，然后同步到内存里面
 */
type JobManager struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
	watcher clientv3.Watcher
}

func(this *JobManager) WatchEtcdJobs() ( err error ){

	var (
		getResponse *clientv3.GetResponse
		kvPair *mvccpb.KeyValue
		job *common.Job
		watchStartReversion int64
		watchChan clientv3.WatchChan
		watchResponse clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobName string
		jobEvent *common.JobEvent
	)

	// 监听JOB_SAVE_DIR路径所有的任务
	if getResponse , err = this.kv.Get(context.TODO() , common.JOB_SAVE_DIR , clientv3.WithPrefix() ); nil != err {
		return
	}

	// 当前有哪些任务。第一次进来的时候也需要推送到调度的协程
	for _ , kvPair = range getResponse.Kvs {
		if job , err = helper.UnPackJob(kvPair.Value); nil == err {
			jobEvent = common.BuildEventJob( common.JOB_EVENT_SAVE , job )
			GlobalScheduler.PushJobPlan( jobEvent )
		}
	}

	// 启动监听协程
	go func() {
		// 从下一个版本开始监听
		watchStartReversion = getResponse.Header.Revision + 1

		// 启动监听-监听/cron/jobs/路径下所有任务的变化
		watchChan = this.watcher.Watch( context.TODO() , common.JOB_SAVE_DIR ,
			clientv3.WithRev(watchStartReversion) , clientv3.WithPrefix() )

		// 处理监听
		for watchResponse = range watchChan {
			for _ , watchEvent = range watchResponse.Events {
				switch watchEvent.Type {
				case mvccpb.PUT : // 任务保存
					// 解析获得job
					if job , err = helper.UnPackJob( watchEvent.Kv.Value ); nil != err {
						continue
					}
					// 构建一个EVENT事件
					jobEvent = common.BuildEventJob( common.JOB_EVENT_SAVE , job )

				case mvccpb.DELETE : // 任务删除
					// 推送删除事件
					jobName = string(watchEvent.Kv.Key)
					job = &common.Job{
						Name: jobName,
					}
					// 构建一个EVENT事件
					jobEvent = common.BuildEventJob( common.JOB_EVENT_DELETE , job )
				}
				GlobalScheduler.PushJobPlan( jobEvent )
			}
		}

	}()

	return
}

var (
	GlobalJobManager *JobManager
)

/**
@desc 创建一个分布式锁、基于etcd
 */
func(this *JobManager) CreateJobLock( jobName string ) (jobLock *JobLock){
	jobLock = InitJobLock( jobName , this.kv , this.lease )
	return
}


// 初始化ETCD对象
func InitEtcdManager () error {
	var (
		config clientv3.Config
		client *clientv3.Client
		err error
		lease clientv3.Lease
		kv clientv3.KV
		watcher clientv3.Watcher
	)
	config = clientv3.Config{
		Endpoints: common.GlobalConfig.EtcdEndPoint , // 服务器集群地址
		DialTimeout: time.Duration(common.GlobalConfig.EtcdDiaTimeOut) * time.Millisecond , // 连接超时时间
	}

	// 建立连接
	if client , err = clientv3.New(config); nil != err {
		return err
	}
	// 得到KV对象
	kv = clientv3.NewKV( client )
	// 获取租约对象
	lease = clientv3.NewLease( client )

	watcher = clientv3.NewWatcher( client )

	GlobalJobManager = &JobManager{
		client : client ,
		kv : kv ,
		lease: lease,
		watcher: watcher ,
	}


	// 开启任务监听
	GlobalJobManager.WatchEtcdJobs()

	return nil
}