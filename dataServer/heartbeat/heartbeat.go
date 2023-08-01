//心跳服务：数据服务节点向接口发送心跳，接口持续监控并更新心跳。
//此处实现：数据服务节点向接口发送心跳。

//数据节点每隔5s向apiServers exchange（消息队列）发送一次心跳信息。
//心跳内容为本结点的监听地址。

package heartbeat

import (
	"go-distributed-oss/src/lib/rabbitmq"
	"os"
	"time"
)

func StartHeartbeat() {
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER")) //新建rabbitmq结构体
	defer q.Close()
	for {
		q.Publish("apiServers", os.Getenv("LISTEN_ADDRESS")) //向apiServers exchange发送本节点监听地址
		time.Sleep(5 * time.Second)                          //每隔5s发送一次心跳信息
	}
}
