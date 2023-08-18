package objects

import (
	"go-distributed-oss/src/lib/es"
	"go-distributed-oss/src/lib/mylogger"
	"net/http"
	"strings"
)

// del 来自客户端的“删除”行为：删除元数据信息，使对象对客户端不可见，但没有实际删除对象
func del(w http.ResponseWriter, r *http.Request) {
	name := strings.Split(r.URL.EscapedPath(), "/")[2] //拿到object_name
	version, err := es.SearchLatestVersion(name)       //搜索最新版本号
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = es.PutMetadata(name, version.Version+1, 0, "") //删除操作：版本+1，size置零，hash置空
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
