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

func (this *ServeHandle) JobListsHandler( res http.ResponseWriter , req * http.Request ){
	var (
		err error
		jobList []*common.Job
		results []byte
	)
	defer func() {
		if r := recover(); r != nil {
			if results , err = helper.JsonResponse( -1 , r.(string) , nil ); nil == err {
				res.Write( results )
				return
			}
		}
	}()
	if jobList , err = this.jobManager.ListsJob(); nil != err {
		panic(err.Error())
	}
	if results , err = helper.JsonResponse( 0 , "" , jobList ); nil == err {
		res.Write( results )
		return
	}
	// 返回错误
	if results , err = helper.JsonResponse( -1 , err.Error() , nil ); nil == err {
		panic(err.Error())
	}
	return
}

func (this *ServeHandle) JobDeleteHandler (res http.ResponseWriter , req * http.Request) {

	var (
		err error
		results []byte
		jobName string
		oldJob *common.Job
	)
	defer func() {
		if r := recover(); r != nil {
			if results , err = helper.JsonResponse( -1 , r.(string) , nil ); nil == err {
				res.Write( results )
				return
			}
		}
	}()

	if err = req.ParseForm() ; nil != err {
		panic("表单解析失败")
	}
	// 任务名称
	jobName = req.PostForm.Get("jobName")
	if "" == jobName {
		panic("获取任务名称失败")
	}

	if oldJob , err = this.jobManager.DeleteJob( jobName ); nil != err {
		panic(err.Error())
	}
	// 返回正常应答
	if results , err = helper.JsonResponse( 0 , "" , oldJob ); nil == err {
		res.Write( results )
		return
	}

	// 返回错误
	if results , err = helper.JsonResponse( -1 , err.Error() , nil ); nil == err {
		panic(err.Error())
		return
	}
	return
}

/**
api-保存任务的具体操作
POST jobs:{"name":"xxx" , "command":"echo 123;" , "cronExpr":"* * * * ..."}
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
		if results , err = helper.JsonResponse( -1 , "表单解析失败", nil ); nil == err {
			res.Write( results )
			return
		}
	}
	// 取表单的数据
	postJob = req.PostForm.Get("job")
	if "" == postJob {
		if results , err = helper.JsonResponse( -1 , "job不能为空", nil ); nil == err {
			res.Write( results )
			return
		}
	}

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