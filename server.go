////LISTEN_ADDRESS=:12345 STORAGE_ROOT=~/GolandProjects/go-distributed-oss/tmp go run server.go

package main

import (
	"fmt"
	"go-distributed-oss/apiServer/objects"
	"log"
	"net/http"
	"os"
)

// main 单机版的server端函数
// 注意：在分布式版本中实现了接口服务与数据服务的解耦，
// 则二者有各自的main函数，不再使用单机版的main函数。
func main() {
	// 定义路由处理函数
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "pong") //测试路由
	})
	http.HandleFunc("/objects/", objects.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}
