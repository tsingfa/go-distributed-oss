package objects

import (
	"fmt"
	"go-distributed-oss/apiServer/heartbeat"
	"go-distributed-oss/apiServer/locate"
	"go-distributed-oss/src/lib/es"
	"go-distributed-oss/src/lib/rs"
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
	if len(versionArr) != 0 { //如果version（数组）有值 --> 有指定版本
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
	if meta.Hash == "" { //若哈希值为空（删除标记） --> 已删除
		w.WriteHeader(http.StatusNotFound)
		return
	}
	hash := url.PathEscape(meta.Hash)         //转义编码
	stream, err := GetStream(hash, meta.Size) //凭借转义后的hash值，去数据服务中获取文件的读取流
	if err != nil {                           //定位对象失败
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_, err = io.Copy(w, stream) //将读取到的数据复制到写入流writer
	if err != nil {             //RS解码过程中发生错误
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	stream.Close() //用于在流关闭时，将临时对象转正
}

// GetStream 查找指定数据对象（对应各个分片及其所属的数据节点），返回对其get请求的读取流reader。
//
// size参数是由于RS码的实现要求每一个数据片的长度完全一样，若编码时对象长度无法被4整除，
// 函数会对最后一个数据片进行填充。而在解码时，需根据对象的准确长度，防止填充数据被当做原始数据返回
func GetStream(hash string, size int64) (*rs.RSGetStream, error) {
	locateInfo := locate.Locate(hash)    //查找指定文件对象所分布的数据结点（含数据+纠错分片）
	if len(locateInfo) < rs.DataShards { //若分片数量不足以拼成一个完整对象
		return nil, fmt.Errorf("object %s locate fail，need %d severs,but result %v", hash, rs.DataShards, locateInfo)
	}
	dataSevers := make([]string, 0)      //接收【恢复分片】的节点
	if len(locateInfo) != rs.AllShards { //有部分分片丢失
		dataSevers = heartbeat.ChooseRandomDataServer(rs.AllShards-len(locateInfo), locateInfo) //随机选取用于接收【恢复分片】的数据节点
	}
	return rs.NewRSGetStream(locateInfo, dataSevers, hash, size)
}
