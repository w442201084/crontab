package master

import (
	"crontab/src/common"
	"crontab/src/iface"
	"crontab/src/router"
	"net"
	"net/http"
	"strconv"
	"time"
)

var (
	// 一个单例的对象
	GlobalApiServe *ApiServe
)

/**
http接口
 */
type ApiServe struct {
	httpServer *http.Server
	listener net.Listener
	muxHandler *http.ServeMux
}

func InitApiServe( jobManager iface.IJobManager ) ( error ) {
	var (
		muxHandler *http.ServeMux
		listener net.Listener
		err error
		httpServer *http.Server
		handler iface.IServerHandler
	)
	handler = router.NewServeHandler(jobManager)
	// 路由对象
	muxHandler = http.NewServeMux()
	muxHandler.HandleFunc( "/job/save" , handler.JobSaveHandler )
	muxHandler.HandleFunc( "/job/delete" , handler.JobDeleteHandler )
	if listener , err = net.Listen("tcp" , ":" + strconv.Itoa( common.GlobalConfig.ApiPort )) ; nil != err {
		return err
	}
	httpServer = &http.Server{
		ReadTimeout: time.Duration(common.GlobalConfig.ApiReadTimeOut) * time.Second,
		WriteTimeout: time.Duration(common.GlobalConfig.ApiWriteTimeOut) * time.Second,
		Handler : muxHandler ,
	}
	GlobalApiServe = &ApiServe{
		httpServer: httpServer ,
		listener : listener ,
		muxHandler : muxHandler ,
	}
	httpServer.Serve( listener )
	return nil
}

