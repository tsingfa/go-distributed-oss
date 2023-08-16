package locate

import (
	"encoding/json"
	"go-distributed-oss/src/lib/rabbitmq"
	"go-distributed-oss/src/lib/rs"
	"go-distributed-oss/src/lib/types"
	"os"
	"time"
)

// Locate 指定对象文件名，查找并返回所在的数据结点地址。
//
// feat1: 在新增元数据功能后，入参改用对象hash值定位（关联到数据服务的Locate函数）。
//
// feat2: 新增数据冗余功能（RS纠错分片），查找并返回各分片id及其所属数据节点地址。
func Locate(object string) map[int]string {
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	q.Publish("dataServers", object) //消息队列作为接口服务与数据服务之间的桥梁
	ch := q.Consume()
	go func() { //临时消息队列（超时1s关闭）
		time.Sleep(time.Second)
		q.Close() //当1s超时，无论当前收到多少条反馈消息都会立即返回
	}()
	locateInfo := make(map[int]string)
	for i := 0; i < rs.AllShards; i++ {
		msg := <-ch //阻塞等待查找结果
		if len(msg.Body) == 0 {
			return locateInfo
		}
		var info types.LocateMessage
		_ = json.Unmarshal(msg.Body, &info)
		locateInfo[info.Id] = info.Addr //记录各个分片所属数据节点地址
	}
	return locateInfo
}

// Exist 指定的对象文件(hash)在各数据服务结点中是否存在。
//
// feat: 实现数据冗余（RS分片）后，数据分散在若干个节点中，
// 则需检查各节点中的分片能否组成一个完整对象。
func Exist(object string) bool {
	return len(Locate(object)) >= rs.DataShards //可以拼成一个完整的对象
}
