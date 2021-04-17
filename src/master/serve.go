package master

import (
	"crontab/src/common"
	"crontab/src/helper"
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

/**
@desc 接受所有handler的panic
 */
func safeHandler ( fn http.HandlerFunc ) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				msg , err := helper.JsonResponse( -1 , err.(string), nil )
				if nil != err {
					return
				}
				w.Write( msg )
			}
		}()
		fn(w, r)
	}
}

func InitApiServe( jobManager iface.IJobManager ) ( error ) {
	var (
		muxHandler *http.ServeMux
		listener net.Listener
		err error
		httpServer *http.Server
		handler iface.IServerHandler
		staticDir http.Dir // 静态文件目录
		staticHandler http.Handler // 静态文件handler
	)
	handler = router.NewServeHandler(jobManager)
	// 路由对象
	muxHandler = http.NewServeMux()
	muxHandler.HandleFunc( "/job/save" , safeHandler(handler.JobSaveHandler) )
	muxHandler.HandleFunc( "/job/delete" , safeHandler(handler.JobDeleteHandler) )
	muxHandler.HandleFunc( "/job/lists" , safeHandler(handler.JobListsHandler) )
	muxHandler.HandleFunc( "/job/kill" , safeHandler(handler.JobKillHandler) )

	// 处理静态资源、设置静态路径
	staticDir = http.Dir(common.GlobalConfig.WebRoot)
	staticHandler = http.FileServer(staticDir)
	// 1、当访问路径/index.html 遵循最大匹配规则
	// 2、通过StripPrefix过滤掉/之后剩下的部分index.html交给staticHandler处理
	// 3、最后在staticHandler取之前配置的路径变成 ../view/index.html
	muxHandler.Handle("/" , http.StripPrefix("/" , staticHandler))


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

