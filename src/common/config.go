package common

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {

	ApiPort int `json:"apiPort"` // API接口服务端口

	ApiReadTimeOut int `json:"apiReadTimeOut"`// API读超时 秒

	ApiWriteTimeOut int `json:"apiWriteTimeOut"` // API写超时 秒

	EtcdEndPoint []string `json:"etcdEndPoint"` // etcd集群IP

	EtcdDiaTimeOut int `json:"etcdDiaTimeOut"` // etcd连接超时时间

	WebRoot string `json:"webRoot"` // web静态文件的路径

}

var GlobalConfig *Config

func InitConfig( fileName string ) error {
	var (
		content []byte
		err error
		config Config
	)
	if content , err = ioutil.ReadFile(fileName) ; nil != err {
		return  err
	}

	if err = json.Unmarshal( content , &config ) ; nil != err {
		return  err
	}
	GlobalConfig = &config
	return nil
}