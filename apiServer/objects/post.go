package objects

import (
	"go-distributed-oss/apiServer/heartbeat"
	"go-distributed-oss/apiServer/locate"
	"go-distributed-oss/src/lib/es"
	"go-distributed-oss/src/lib/mylogger"
	"go-distributed-oss/src/lib/rs"
	"go-distributed-oss/src/lib/utils"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// 支持“断点续传”的上传服务
// "POST",/objects/<object_name>
// Digest: SHA-256=<对象hash>
// Size: <总长度size>

func post(w http.ResponseWriter, r *http.Request) {
	// 1.参数获取和解析
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	size, err := strconv.ParseInt(r.Header.Get("size"), 0, 64)
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	hash := utils.GetHashFromHeader(r.Header)
	if hash == "" { //没有填写hash，或者格式不合规（要求SHA-256）
		mylogger.L().Println("missing object hash in digest header...")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// 2.如果对象在数据节点中已存在（可跳过上传）
	if locate.Exist(url.PathEscape(hash)) {
		err = es.AddVersion(name, hash, size)
		if err != nil {
			mylogger.L().Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		return
	}
	// 3.上传相关操作
	//分配数据节点
	ds := heartbeat.ChooseRandomDataServer(rs.AllShards, nil)
	if len(ds) != rs.AllShards { //找不到足够的数据节点
		mylogger.L().Println("cannot find enough dataServer...")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	//生成数据写入流（可通过该流，将数据写入到数据节点）
	stream, err := rs.NewRSResumablePutStream(ds, name, url.PathEscape(hash), size)
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("location", "/temp/"+url.PathEscape(stream.ToToken())) //生成字符串token，放入location响应头部
	w.WriteHeader(http.StatusCreated)                                     //上传（创建）对象成功
}
