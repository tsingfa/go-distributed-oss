// 心跳服务：接收和处理来自数据服务结点的心跳信息

package heartbeat

import (
	"go-distributed-oss/src/lib/rabbitmq"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

var dataServers = make(map[string]time.Time) //记录所有【数据结点-->上次心跳时间】
var mutex sync.Mutex                         //更新dataServers表时要上锁，保证并发安全

// ListenHeartbeat 监听并更新来自数据结点的心跳
func ListenHeartbeat() {
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	defer q.Close()
	q.Bind("apiServers")
	ch := q.Consume()            //接收（消费）来自所绑定的exchange的消息
	go removeExpiredDataServer() //并发删除未及时更新心跳的数据结点
	for msg := range ch {        //遍历来自apiServers exchange的消息，msg记录了是哪个数据结点请求更新心跳
		dataServer, err := strconv.Unquote(string(msg.Body))
		if err != nil {
			log.Println(err)
			panic(err)
		}
		mutex.Lock()
		dataServers[dataServer] = time.Now() //更新数据结点的心跳
		mutex.Unlock()
	}
}

// removeExpiredDataServer 并发删除【心跳过期】的数据结点
func removeExpiredDataServer() {
	for {
		time.Sleep(5 * time.Second)
		mutex.Lock() //删除结点，操作表要上锁
		for server, lastTime := range dataServers {
			if lastTime.Add(10 * time.Second).Before(time.Now()) {
				delete(dataServers, server) //删除【超过10s未更新心跳】的数据结点
			}
		}
		mutex.Unlock()
	}
}

// GetDataServers 返回当前所有的数据服务结点集合
func GetDataServers() []string {
	mutex.Lock()
	defer mutex.Unlock()
	ds := make([]string, 0)
	for server := range dataServers {
		ds = append(ds, server)
	}
	return ds
}

// ChooseRandomDataServer 返回随机选择的数据服务结点
func ChooseRandomDataServer() string {
	ds := GetDataServers()
	n := len(ds)
	if n == 0 {
		return "" //没有数据服务结点
	}
	return ds[rand.Intn(n)] //随机[0,n)整数，指定数据节点
}
