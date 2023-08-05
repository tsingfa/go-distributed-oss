//定位服务：监听来自接口服务的定位请求，返回定位对象所属的数据结点地址

package locate

import (
	"go-distributed-oss/src/lib/rabbitmq"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

var objects = make(map[string]int) //检查对象是否存在磁盘（维护一个缓存中的表，而不用实际读盘）
var mutex sync.Mutex               //用于保护表的读写

// StartLocate 用于监听定位请求（由dataServers消息队列指定要定位的对象）
func StartLocate() {
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	defer q.Close()
	q.Bind("dataServers") //分布式系统，各个节点之间的操作是“平行”的，所有节点都会执行这些操作
	ch := q.Consume()
	for msg := range ch { //遍历channel，接收（消费）来自dataServers exchange的消息
		hash, err := strconv.Unquote(string(msg.Body)) //新增元数据功能后，改用hash定位
		if err != nil {
			log.Println(err)
			panic(err)
		}
		if Locate(hash) { //如果该hash能被定位到
			q.Send(msg.ReplyTo, os.Getenv("LISTEN_ADDRESS")) //将数据对象所在的数据结点地址，返回给指定地址（临时消息队列）
		}
	}
}

/*
// Locate 查看是否能定位到指定的文件名路径（磁盘检查-->太慢）
func Locate(object string) bool {
	_, err := os.Stat(object)    //访问指定文件路径
	return !os.IsNotExist(err) //返回是否存在访问错误（找不找得到该文件）
}
*/

// Locate 检查指定hash的文件是否在缓存map中（以内存代替磁盘检查）
func Locate(hash string) bool {
	mutex.Lock() //互斥锁保护对objects的读写操作
	_, ok := objects[hash]
	mutex.Unlock()
	return ok
}

// Add 将hash加入缓存记录
func Add(hash string) {
	mutex.Lock()
	objects[hash] = 1
	mutex.Unlock()
}

// Del 将hash移出缓存记录
func Del(hash string) {
	mutex.Lock()
	delete(objects, hash)
	mutex.Unlock()
}

// CollectObjects 初始化缓存列表，记录该节点/objects/目录下所有的文件。
func CollectObjects() {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/*")
	for idx := range files {
		hash := filepath.Base(files[idx]) //获取基本文件名（本身就以hash值命名）
		objects[hash] = 1
	}
}
