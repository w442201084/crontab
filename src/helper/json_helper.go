package helper

import (
	"crontab/src/common"
	"encoding/json"
)

// 反序列化Job
func UnPackJob (value []byte) ( job *common.Job , err error ){
	job = &common.Job{}
	if err = json.Unmarshal(value , job); nil != err {
		return
	}
	return
}
