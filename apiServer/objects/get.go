package objects

import (
	"fmt"
	"go-distributed-oss/apiServer/locate"
	"go-distributed-oss/src/lib/es"
	"go-distributed-oss/src/lib/objectstream"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func get(w http.ResponseWriter, r *http.Request) {
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	versionArr := r.URL.Query()["version"] //URL中"?"后的查询字段对应version的值（字符串数组）
	version := 0
	var err error
	if len(versionArr) != 0 { //如果version（数组）有值
		version, err = strconv.Atoi(versionArr[0]) //不考虑多参数情况，仅取首个元素
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	meta, err := es.GetMetadata(name, version) //得到元数据
	if err != nil {                            //找不到元数据
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if meta.Hash == "" { //若哈希值为空，即为删除标记
		w.WriteHeader(http.StatusNotFound)
		return
	}
	object := url.PathEscape(meta.Hash) //转义编码
	stream, err := getStream(object)    //凭借转义后的hash值，去数据服务中获取文件的读取流
	if err != nil {                     //定位对象失败
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_, _ = io.Copy(w, stream) //将读取到的数据复制到写入流writer
}

// getStream 查找指定数据对象object，返回对其get请求的读取流reader。
func getStream(object string) (io.Reader, error) {
	server := locate.Locate(object) //查找指定文件对象所在的数据结点
	if server == "" {               //找不到对象所在的数据结点
		return nil, fmt.Errorf("object %s locate fail", object)
	}
	return objectstream.NewGetStream(server, object)
}
