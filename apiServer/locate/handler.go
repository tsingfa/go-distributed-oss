package locate

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// Handler 用于向【数据服务结点】查询定位请求并接收回复。
func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	object := strings.Split(r.URL.EscapedPath(), "/")[2] //增加元数据功能后，改为用哈希值做locate
	dataServer := Locate(object)                         //定位对象资源所在的数据结点
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
