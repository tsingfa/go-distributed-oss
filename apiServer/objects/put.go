//将HTTP操作请求转发给数据服务
//PUT操作相关

package objects

import (
	"fmt"
	"go-distributed-oss/apiServer/heartbeat"
	"go-distributed-oss/src/lib/es"
	"go-distributed-oss/src/lib/objectstream"
	"go-distributed-oss/src/lib/utils"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// put 上传（更新）
func put(w http.ResponseWriter, r *http.Request) {
	hash := utils.GetHashFromHeader(r.Header) //获取http头部中的hash值
	if hash == "" {
		log.Println("missing object hash in digest header...")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	code, err := storeObject(r.Body, url.PathEscape(hash))
	if err != nil {
		log.Println(err)
		w.WriteHeader(code) //返回HTTP错误码
		return
	}
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	size := utils.GetSizeFromHeader(r.Header) //文件大小
	err = es.AddVersion(name, hash, size)
	log.Println("addVersion error:", err)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// storeObject 将指定数据内容（输入流）写入指定数据对象对应的写入流，
// 该写入流中的内容会被包装成对应的HTTP请求并执行（由NewPutStream在后台异步执行）。
func storeObject(r io.Reader, object string) (int, error) {
	stream, err := putStream(object) //指定对象的写入流（随机分配了数据结点）
	if err != nil {
		log.Println(err)
		return http.StatusServiceUnavailable, err
	}
	_, _ = io.Copy(stream, r) //将reader内容写入stream的writer（会生成相应的put请求并执行）
	err = stream.Close()
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// putStream 随机指定一个数据服务结点，为该结点以及指定数据对象的put请求构建写入流。
func putStream(object string) (*objectstream.PutStream, error) {
	server := heartbeat.ChooseRandomDataServer() //随机选数据服务结点
	log.Println("choose server:", server)
	if server == "" {
		return nil, fmt.Errorf("cannot find any dataServer")
	}
	return objectstream.NewPutStream(server, object), nil
}
