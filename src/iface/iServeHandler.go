package iface

import "net/http"

type IServerHandler interface {

	/** 处理etcd-Save */
	JobSaveHandler( res http.ResponseWriter , req * http.Request )

	/** 处理etcd-Delete */
	JobDeleteHandler( res http.ResponseWriter , req * http.Request )
}