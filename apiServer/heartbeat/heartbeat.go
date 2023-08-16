// 心跳服务：接收和处理来自数据服务结点的心跳信息

package heartbeat

import (
	"go-distributed-oss/src/lib/mylogger"
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

// ChooseRandomDataServer 返回随机选择的数据服务结点（指定数量的选择结果）
//
// feat: 原先为随机返回一个数据节点用于存储对象文件，现新增分片及RS冗余纠错功能，
// 需要一次性随机返回若干个数据节点（用于存储数据分片和纠错分片）。
//
// n int 指定随机数据节点数量。
//
// exclude map 用于记录本次随机选择过程中要排除的数据节点名单，若没有排除则为nil。
// 【作用】：对原函数的一种复用，通过排除节点的方式，进一步划定选择范围。
//
// 【排除数据节点的场景】：当定位完成后（知道哪些节点有分片数据），而实际收到的分片可能并不足，此时需要进行数据修复，
// 根据已有的分片，将丢失的分片复原出来并再次上传到对应的数据节点【所以此时不再随机，而相当于要指定某些节点】
// 通过跳过正常节点的方式，挑选出异常节点，上传复原分片。
func ChooseRandomDataServer(n int, exclude map[int]string) (res []string) {
	candidates := make([]string, 0) //从所有节点中，过滤掉排除名单，留下候选名单
	reverseExcludeMap := make(map[string]int)
	for id, addr := range exclude { //排除名单（键值转换，方便查找）
		reverseExcludeMap[addr] = id
	}
	servers := GetDataServers() //所有数据节点（地址）的列表
	for i := range servers {
		s := servers[i]
		_, excluded := reverseExcludeMap[s] //节点是否在排除名单内
		if !excluded {                      //若未被排除
			candidates = append(candidates, s)
		}
	}
	length := len(candidates)
	if length < n { //候选节点数量不足
		mylogger.L().Printf("insufficient number of candidate servers,expect %d but found %d.\n", n, length)
		return nil
	}
	seq := rand.Perm(length) //[0,length-1]的所有整数组成的乱序序列
	for i := 0; i < n; i++ { //取出前n个
		res = append(res, candidates[seq[i]])
	}
	return res
}
