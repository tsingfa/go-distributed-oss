package main

import (
	"go-distributed-oss/src/lib/es"
	"go-distributed-oss/src/lib/mylogger"
)

const MinVersionCount = 5 //版本数量容忍度

func main() {
	buckets, err := es.SearchVersionStatus(MinVersionCount + 1) //查找元数据服务（ES）中，所有版本数量超限的对象
	if err != nil {
		mylogger.L().Println(err)
		return
	}
	for i := range buckets { //遍历buckets，从该对象当前的最小版本开始一一删除，直到还剩5个
		bucket := buckets[i]
		for v := 0; v < bucket.DocCount-MinVersionCount; v++ {
			es.DelMetadata(bucket.Key, v+int(bucket.MinVersion.Value))
		}
	}
}
