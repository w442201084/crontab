package helper

import (
	"crontab/src/common"
	"encoding/json"
)

/**
http请求格式化输出
 */
func JsonResponse ( code int , msg string , data interface{} ) (res []byte , err error) {
	var (
		response *common.ResponseFormat
	)
	response = &common.ResponseFormat{
		Code: code ,
		Msg: msg,
		Data: data,
	}
	res , err = json.Marshal( response )
	return
}