package work

import (
	"context"
	"crontab/src/common"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// 用于服务注册
// 开启一个协程，把服务器注册到etcd里面，并且自动续约
// 注册的地址 common.SERVER_REGISTER_DIR + IpAddress

type ServerRegister struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
	leaseId clientv3.LeaseID
	watcher clientv3.Watcher
	localIp string // 本机IP地址
}

// 把服务器注册到etcd里面，并且自动续约
func (this *ServerRegister) KeepOnLine() (err error){

	var (
		ServeRegKey string
		leaseGrantResponse *clientv3.LeaseGrantResponse
		leaseKeepAliveChan <- chan *clientv3.LeaseKeepAliveResponse
		//putResponse *clientv3.PutResponse
		keepAliveResponse *clientv3.LeaseKeepAliveResponse
		cancelCtx context.Context
		cancelFunc context.CancelFunc
	)

	for {
		ServeRegKey = common.SERVER_REGISTER_DIR + this.localIp
		cancelFunc = nil
		// 创建一个5s的租约
		if leaseGrantResponse , err = this.lease.Grant( context.TODO() , 5 ); nil != err {
			goto TRYAGAIN
		}
		this.leaseId = leaseGrantResponse.ID


		// 自动续租 , context.TODO 无线续租
		// ctx , _ := context.WithTimeout(context.TODO() , 5 * time.Second) ctx续租5s
		if leaseKeepAliveChan , err = this.lease.KeepAlive( context.TODO() , leaseGrantResponse.ID ) ; nil != err {
			goto TRYAGAIN
		}

		cancelCtx , cancelFunc = context.WithCancel( context.TODO() )
		// 注册到etcd
		if _ , err = this.kv.Put(cancelCtx , ServeRegKey , "" ,
			clientv3.WithLease(this.leaseId)) ; nil != err {
			goto TRYAGAIN
		}

		// 处理续租应答
		for {
			select {
				case keepAliveResponse = <- leaseKeepAliveChan :
					// 续租失败
					if nil == keepAliveResponse {
						goto TRYAGAIN
					}
			}
		}

		TRYAGAIN:
			time.Sleep( 1 * time.Second )
			if nil != cancelFunc {
				this.lease.Revoke( context.TODO() , this.leaseId )
				cancelFunc()
			}

	}



}

var (
	GlobalServerRegister *ServerRegister
)

func InitRegister () (err error){
	var (
		config clientv3.Config
		client *clientv3.Client
		lease clientv3.Lease
		kv clientv3.KV
		watcher clientv3.Watcher
		localIp string
	)
	config = clientv3.Config{
		Endpoints: common.GlobalConfig.EtcdEndPoint , // 服务器集群地址
		DialTimeout: time.Duration(common.GlobalConfig.EtcdDiaTimeOut) * time.Millisecond , // 连接超时时间
	}

	// 建立连接
	if client , err = clientv3.New(config); nil != err {
		return
	}
	// 得到KV对象
	kv = clientv3.NewKV( client )
	// 获取租约对象
	lease = clientv3.NewLease( client )
	// 监听对象
	watcher = clientv3.NewWatcher( client )

	if localIp , err = common.GetLocalIp() ; nil != err {
		return
	}

	GlobalServerRegister = &ServerRegister{
		client : client,
		kv : kv ,
		lease : lease,
		watcher : watcher,
		localIp: localIp ,
	}

	go GlobalServerRegister.KeepOnLine()
	return
}
