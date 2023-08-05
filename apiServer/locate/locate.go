package locate

import (
	"go-distributed-oss/src/lib/rabbitmq"
	"os"
	"strconv"
	"time"
)

// Locate 指定对象文件名，查找并返回所在的数据结点地址。
// 在新增元数据功能后，改用哈希值定位（关联到数据服务的Locate函数）。
func Locate(object string) string {
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	q.Publish("dataServers", object) //消息队列作为接口服务与数据服务之间的桥梁
	ch := q.Consume()
	go func() { //临时消息队列（超时1s关闭）
		time.Sleep(time.Second)
		q.Close()
	}()
	msg := <-ch //阻塞等待查找结果
	server, _ := strconv.Unquote(string(msg.Body))
	return server
}

// Exist 指定的对象文件在各数据服务结点中是否存在
func Exist(name string) bool {
	return Locate(name) != ""
}
