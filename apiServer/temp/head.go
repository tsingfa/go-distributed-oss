package temp

import (
	"fmt"
	"go-distributed-oss/src/lib/mylogger"
	"go-distributed-oss/src/lib/rs"
	"net/http"
	"strings"
)

// "HEAD", /temp/<token>
// 返回当前上传进度（第一个临时分片大小*4）

//token是由objects的post操作得来，是对RSResumablePutStream结构体做JSON序列化和base64编码
//由token可以解析出数据服务的server和uuid信息等

func head(w http.ResponseWriter, r *http.Request) {
	token := strings.Split(r.URL.EscapedPath(), "/")[2]
	stream, err := rs.NewRSResumablePutStreamFromToken(token)
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	current := stream.CurrentSize() //调用到数据服务的head /temp/<uuid> 请求
	if current == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("content-length", fmt.Sprintf("%d", current))
}
