//定位服务：监听来自接口服务的定位请求，返回定位对象所属的数据结点地址

package locate

import (
	"go-distributed-oss/src/lib/mylogger"
	"go-distributed-oss/src/lib/rabbitmq"
	"go-distributed-oss/src/lib/types"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var objects = make(map[string]int) //定位缓存--检查对象是否存在磁盘（维护一个缓存中的表，而不用实际读盘）
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
			mylogger.L().Println(err)
			panic(err)
		}
		id := Locate(hash) //查询该对象在本节点所分到的分片id
		if id != -1 {      //查得到
			q.Send(msg.ReplyTo, types.LocateMessage{
				Addr: os.Getenv("LISTEN_ADDRESS"),
				Id:   id,
			}) //将数据对象所在的数据结点地址（以及分片id），返回给指定地址（临时消息队列）
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
// feat: 实现RS分片后，不仅要检查是否在map中，还要知道该分片为几号分片
func Locate(hash string) int {
	mutex.Lock()            //互斥锁保护对objects的读写操作
	id, ok := objects[hash] //对象hash--分片id
	mutex.Unlock()
	if !ok { //如果没找到
		return -1
	}
	return id //几号分片
}

// Add 将分片hash加入缓存记录，并记录分片id
func Add(hash string, id int) {
	mutex.Lock()
	objects[hash] = id
	mutex.Unlock()
}

// Del 将hash移出缓存记录
func Del(hash string) {
	mutex.Lock()
	delete(objects, hash)
	mutex.Unlock()
}

// CollectObjects 初始化定位缓存，记录该节点/objects/目录下所有的文件，而不用再读磁盘IO。
//
// feat: 不仅要知道该对象（分片）在磁盘中，还要知道该分片是第几块分片（分片id）。
// 分片文件名：<对象hash.id.分片hash>; 借助该列表，可由文件对象的hash得出分片id
func CollectObjects() {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/*")
	for i := range files {
		filenameSplit := strings.Split(filepath.Base(files[i]), ".") //拆分分片文件名
		if len(filenameSplit) != 3 {
			mylogger.L().Printf("The RS shard file '%s' is named incorrectly.\n", files[i])
			panic(files[i])
		}
		hash := filenameSplit[0] //对象hash
		id, err := strconv.Atoi(filenameSplit[1])
		if err != nil {
			mylogger.L().Printf("The RS shard file '%s' is named incorrectly.\n", files[i])
			panic(err)
		}
		objects[hash] = id //对象hash--分片id（一个对象在一个节点下仅有一个分片，故可一一对应）
	}
}
