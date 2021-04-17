package work

import (
	"context"
	"crontab/src/common"
	"crontab/src/helper"
	"fmt"
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
		watchResponse *clientv3.WatchResponse
		wacthEvent *clientv3.Event
	)

	// 监听JOB_SAVE_DIR路径所有的任务
	if getResponse , err = this.kv.Get(context.TODO() , common.JOB_SAVE_DIR , clientv3.WithPrefix() ); nil != err {
		return
	}

	// 当前有哪些任务
	for _ , kvPair = range getResponse.Kvs {
		if job , err = helper.UnPackJob(kvPair.Value); nil == err {
			fmt.Println(job)
		}
	}


	// 启动监听协程
	go func() {
		// 从下一个版本开始监听
		watchStartReversion = getResponse.Header.Revision + 1

		// 启动监听-监听/cron/jobs/路径下所有任务的变化
		watchChan = this.watcher.Watch( context.TODO() , common.JOB_SAVE_DIR , clientv3.WithRev(watchStartReversion) )

		// 处理监听
		for watchResponse = range watchChan {
			for wacthEvent = range watchResponse.Events {
				switch wacthEvent.Type {
				case mvccpb.PUT : // 任务保存
					// TODO 反序列化、推送给调度协程

				case mvccpb.DELETE : // 任务删除
				}
			}
		}


	}()

	return
}

var (
	GlobalJobManager *JobManager
)


// 初始化ETCD对象
func InitEtcdManager () error {
	var (
		config clientv3.Config
		client *clientv3.Client
		err error
		lease clientv3.Lease
		kv clientv3.KV
		wacther clientv3.Watcher
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

	wacther = clientv3.NewWatcher( client )

	GlobalJobManager = &JobManager{
		client : client ,
		kv : kv ,
		lease: lease,
		watcher: wacther ,
	}
	return nil
}