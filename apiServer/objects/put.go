//将HTTP操作请求转发给数据服务
//PUT操作相关

package objects

import (
	"fmt"
	"go-distributed-oss/apiServer/heartbeat"
	"go-distributed-oss/apiServer/locate"
	"go-distributed-oss/src/lib/es"
	"go-distributed-oss/src/lib/mylogger"
	"go-distributed-oss/src/lib/rs"
	"go-distributed-oss/src/lib/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// put 上传（更新）
func put(w http.ResponseWriter, r *http.Request) {
	hash := utils.GetHashFromHeader(r.Header) //获取http头部中的hash值（客户端提供的哈希）
	if hash == "" {
		mylogger.L().Println("missing object hash in digest header...")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	size := utils.GetSizeFromHeader(r.Header) //（临时）对象大小
	code, err := storeObject(r.Body, hash, size)
	if err != nil { //服务器错误 或 数据校验失败
		mylogger.L().Println(err)
		w.WriteHeader(code) //返回HTTP错误码
		return
	}
	if code != http.StatusOK { //其他网络错误
		w.WriteHeader(code)
		return
	}
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	err = es.AddVersion(name, hash, size) //更新版本记录（单例检查跳过 或者 缓存提交成功）
	if err != nil {
		mylogger.L().Println("addVersion error:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// storeObject 将指定数据内容（输入流）写入指定数据对象对应的写入流，
// 该写入流中的内容会被包装成对应的HTTP请求并交给数据服务执行（由NewPutStream在后台异步执行）。
// 注：初始化NewTempPutStream时做了POST操作，对stream做写入操作时做了PATCH操作
func storeObject(r io.Reader, hash string, size int64) (int, error) {
	if locate.Exist(url.PathEscape(hash)) { //单例检查：如果在数据服务中已存在，可跳过存储（另仍需版本更新）
		return http.StatusOK, nil
	}
	stream, err := putStream(url.PathEscape(hash), size) //指定对象的写入流（随机分配了数据结点）
	if err != nil {
		mylogger.L().Println(err)
		return http.StatusInternalServerError, err
	}
	reader := io.TeeReader(r, stream) //对r的读取，都会写入stream（stream实现了io.Writer接口）
	d := utils.CalculateHash(reader)
	if d != hash { //数据校验（不匹配）
		stream.Commit(false) //删除缓存
		return http.StatusBadRequest, fmt.Errorf("object hash mismatch,calculated=%s,but requested=%s\n", d, hash)
	}
	stream.Commit(true) //匹配成功，缓存转正
	return http.StatusOK, nil
}

// putStream 随机指定一个数据服务结点，为该结点以及指定数据对象的put请求构建写入流。
// 返回的TempPutStream结构体，支持数据服务的缓存功能（转正、删除）
//
// feat: 随机选择【一组】数据节点构建写入流，写入分片文件。
func putStream(hash string, size int64) (*rs.RSPutStream, error) {
	servers := heartbeat.ChooseRandomDataServer(rs.AllShards, nil) //随机选择的数据服务结点结果
	mylogger.L().Println("choose servers:", servers)
	if len(servers) != rs.AllShards {
		return nil, fmt.Errorf("cannot find enough dataServer")
	}
	return rs.NewRSPutStream(servers, hash, size)
}
