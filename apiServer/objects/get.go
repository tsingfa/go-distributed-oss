package objects

import (
	"fmt"
	"go-distributed-oss/apiServer/locate"
	"go-distributed-oss/src/lib/objectstream"
	"io"
	"log"
	"net/http"
	"strings"
)

func get(w http.ResponseWriter, r *http.Request) {
	object := strings.Split(r.URL.EscapedPath(), "/")[2]
	stream, err := getStream(object)
	if err != nil { //定位对象失败
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	io.Copy(w, stream) //将读取到的数据复制到写入流writer
}

// getStream 查找指定数据对象object，返回对其get请求的读取流reader。
func getStream(object string) (io.Reader, error) {
	server := locate.Locate(object) //查找指定文件对象所在的数据结点
	if server == "" {               //找不到对象所在的数据结点
		return nil, fmt.Errorf("object %s locate fail", object)
	}
	return objectstream.NewGetStream(server, object)
}
