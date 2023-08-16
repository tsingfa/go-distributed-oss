package versions

import (
	"encoding/json"
	"go-distributed-oss/src/lib/es"
	"go-distributed-oss/src/lib/mylogger"
	"net/http"
	"strings"
)

// Handler 用于
func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m != http.MethodGet { //versions相关均为GET请求
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	from := 0 //每页起始点、页长
	size := 1000
	name := strings.Split(r.URL.EscapedPath(), "/")[2] //拿到object_name
	for {
		metas, err := es.SearchAllVersions(name, from, size) //返回元数据数组
		if err != nil {
			mylogger.L().Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for i := range metas { //对于每个元数据结构，将其一一写入HTTP响应正文
			body, _ := json.Marshal(metas[i])
			w.Write(body)
			w.Write([]byte("\n"))
		}
		if len(metas) != size {
			return
		}
		from += size //下一页
	}
}
