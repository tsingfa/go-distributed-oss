package objects

import (
	"go-distributed-oss/apiServer/locate"
	"go-distributed-oss/src/lib/mylogger"
	"net/http"
	"strings"
)

// garbage 管理员维护时的“清除”行为：将指定hash的对象移到garbage目录（等待管理员删除或其他处理）
func garbage(w http.ResponseWriter, r *http.Request) {
	hash := strings.Split(r.URL.EscapedPath(), "/")[2]
	locateInfo := locate.Locate(hash)       //查找指定文件对象所分布的数据结点
	for _, dataServer := range locateInfo { //[分片id]-->数据节点地址
		url := "http://" + dataServer + "/objects/" + hash
		request, _ := http.NewRequest("DELETE", url, nil)
		client := http.Client{}
		_, err := client.Do(request)
		if err != nil {
			mylogger.L().Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	w.WriteHeader(http.StatusOK)
}
