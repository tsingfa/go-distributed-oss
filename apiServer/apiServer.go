// apiServer 接口服务

package main

import (
	"fmt"
	"go-distributed-oss/apiServer/heartbeat"
	"go-distributed-oss/apiServer/locate"
	"go-distributed-oss/apiServer/objects"

	"log"
	"net/http"
	"os"
)

func main() {
	go heartbeat.ListenHeartbeat() //心跳服务：由接口服务监控数据服务节点的心跳
	//路由管理
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "apiServer:"+os.Getenv("LISTEN_ADDRESS")+" is connected...\n") //测试路由
	})
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/locate/", locate.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}
