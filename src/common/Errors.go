package common

import "errors"

var (
	ERR_LOCK_ALREADY_REQUIRED = errors.New("锁以被占用")

	ERR_CANT_GET_LOCAL_IP = errors.New("获取本机IPv4地址失败")
)
