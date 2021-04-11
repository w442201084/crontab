package master

import (
	"context"
	"crontab/src/common"
	"crontab/src/iface"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type JobManager struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}

var (
	GlobalJobManager iface.IJobManager
)

/**
删除某一个job
 */
func (this *JobManager) DeleteJob( jobName string ) (oldJob *common.Job , err error ) {
	var (
		jobKey string
		deleteResponse *clientv3.DeleteResponse
		oldJobObj *common.Job
	)
	oldJobObj = &common.Job{}
	jobKey = "/cron/jobs/" + jobName
	if deleteResponse , err = this.kv.Delete( context.TODO() , jobKey , clientv3.WithPrevKV() ); nil != err {
		return nil , err
	}
	if 0 != len( deleteResponse.PrevKvs ) {
		err = json.Unmarshal(  deleteResponse.PrevKvs[0].Value  ,  oldJobObj)
		oldJob = oldJobObj
	}
	return
}

/**
保存任务到etcd节点中
 */
func (this *JobManager) SaveJob(job *common.Job) (oldJob *common.Job , err error ) {
	// 把任务保存到/cron/jobs/目录中
	var (
		jobKey string
		jobValue []byte
		putResponse *clientv3.PutResponse
		oldJobObj *common.Job
	)
	jobKey = "/cron/jobs/" + job.Name
	// 任务信息JSON
	if jobValue , err = json.Marshal( job ); nil != err {
		return
	}
	if putResponse , err = this.kv.Put( context.TODO() , jobKey , string(jobValue) , clientv3.WithPrevKV() );
		nil != err {
		return
	}
	// 如果是更新返回旧值
	if nil != putResponse.PrevKv {
		oldJobObj = &common.Job{}
		// 对旧值反序列化
		if err = json.Unmarshal( putResponse.PrevKv.Value , oldJobObj ); nil != err {
			fmt.Println("获取旧值异常" , err)
			// etcd已经操作成功了，如果旧值有异常，可以忽略
			err = nil
			return
		}
		oldJob = oldJobObj
	}
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

	GlobalJobManager = &JobManager{
		client : client ,
		kv : kv ,
		lease: lease,
	}
	return nil
}