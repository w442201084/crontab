package router

import (
	"crontab/src/common"
	"crontab/src/helper"
	"crontab/src/iface"
	"encoding/json"
	"net/http"
)

type ServeHandle struct {
	jobManager iface.IJobManager
}

func NewServeHandler( jobManager iface.IJobManager ) *ServeHandle{
	return &ServeHandle{
		jobManager : jobManager ,
	}
}

/**
api-保存任务的具体操作
POST jobs:{"name":"xxx" , "commond":"echo 123;" , "cronExpr":"* * * * ..."}
 */
func(this *ServeHandle) JobSaveHandler( res http.ResponseWriter , req * http.Request ) {
	var (
		err error
		postJob string
		job *common.Job
		oldJob *common.Job
		results []byte
	)
	job = &common.Job{}
	// 解析表单
	if err = req.ParseForm() ; nil != err {
		// 解析表单失败
		return
	}
	// 取表单的数据
	postJob = req.PostForm.Get("job")

	// 反序列Job
	if err = json.Unmarshal([]byte(postJob) , job); nil != err {
		if results , err = helper.JsonResponse( -1 , err.Error() , nil ); nil == err {
			res.Write( results )
			return
		}
	}

	if oldJob , err = this.jobManager.SaveJob( job ); nil != err {
		if results , err = helper.JsonResponse( -1 , err.Error() , nil ); nil == err {
			res.Write( results )
			return
		}
	}
	// 返回正常应答
	if results , err = helper.JsonResponse( 0 , "" , oldJob ); nil == err {
		res.Write( results )
		return
	}

	// 返回错误
	if results , err = helper.JsonResponse( -1 , err.Error() , nil ); nil == err {
		res.Write( results )
		return
	}
	return
}