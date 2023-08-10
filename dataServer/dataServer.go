package main

import (
	"fmt"
	"go-distributed-oss/dataServer/heartbeat"
	"go-distributed-oss/dataServer/locate"
	"go-distributed-oss/dataServer/objects"
	"go-distributed-oss/dataServer/temp"
	"log"
	"net/http"
	"os"
)

func main() {
	locate.CollectObjects()       //初始化对象文件的缓存列表
	go heartbeat.StartHeartbeat() //心跳服务：由接口服务监控数据服务节点的心跳
	go locate.StartLocate()       //定位服务：定位资源所属的数据节点位置
	//路由管理
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "dataServer:"+os.Getenv("LISTEN_ADDRESS")+" is connected...\n") //测试路由
	})
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/temp/", temp.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}
