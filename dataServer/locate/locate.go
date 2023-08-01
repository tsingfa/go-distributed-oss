//定位服务：监听来自接口服务的定位请求，返回定位对象所属的数据结点地址

package locate

import (
	"go-distributed-oss/src/lib/rabbitmq"
	"log"
	"os"
	"strconv"
)

// StartLocate 用于监听定位请求（由dataServers消息队列指定要定位的对象）
func StartLocate() {
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	defer q.Close()
	q.Bind("dataServers")
	ch := q.Consume()
	for msg := range ch { //遍历channel，接收（消费）来自dataServers的消息
		object, err := strconv.Unquote(string(msg.Body))
		if err != nil {
			log.Println(err)
			panic(err)
		}
		if Locate(os.Getenv("STORAGE_ROOT") + "/objects/" + object) {
			q.Send(msg.ReplyTo, os.Getenv("LISTEN_ADDRESS")) //将数据对象所在的数据结点地址，返回给指定地址（临时消息队列）
		}
	}
}

// Locate 查看是否能定位到指定的文件名路径
func Locate(name string) bool {
	_, err := os.Stat(name)    //访问指定文件路径
	return !os.IsNotExist(err) //返回是否存在访问错误（找不找得到该文件）
}
