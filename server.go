package main

import (
	"fmt"
	"go-distributed-oss/apiServer/objects"
	"log"
	"net/http"
	"os"
)

func main() {
	// 定义路由处理函数
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "pong") //测试路由
	})
	http.HandleFunc("/objects/", objects.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}
