package locate

import (
	"encoding/json"
	"go-distributed-oss/src/lib/rabbitmq"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Handler 用于向【数据服务结点】查询定位请求并接收回复。
func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	object := strings.Split(r.URL.EscapedPath(), "/")[2]
	dataServer := Locate(object) //定位对象资源所在的数据结点
	if len(dataServer) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	b, _ := json.Marshal(dataServer)
	_, err := w.Write(b)
	if err != nil {
		log.Println(err)
		return
	}
}

// Locate 指定对象文件名，查找并返回所在的数据结点地址
func Locate(name string) string {
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	q.Publish("dataServers", name) //消息队列作为接口服务与数据服务之间的桥梁
	ch := q.Consume()
	go func() { //临时消息队列（超时1s关闭）
		time.Sleep(time.Second)
		q.Close()
	}()
	msg := <-ch //阻塞等待查找结果
	server, _ := strconv.Unquote(string(msg.Body))
	return server
}

// Exist 指定的对象文件是否在各数据服务结点中存在
func Exist(name string) bool {
	return Locate(name) != ""
}
