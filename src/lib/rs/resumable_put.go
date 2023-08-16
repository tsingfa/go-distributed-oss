package rs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-distributed-oss/src/lib/mylogger"
	"go-distributed-oss/src/lib/objectstream"
	"go-distributed-oss/src/lib/utils"
	"io"
	"net/http"
)

// resumableToken 保存对象的名字、大小、hash以及各分片所在的数据节点地址和uuid
type resumableToken struct {
	Name    string
	Size    int64
	Hash    string
	Servers []string
	Uuids   []string
}

//回顾uuid
//临时对象分片信息文件 dataSever地址/STORAGE_ROOT/temp/uuid
//临时对象分片 dataSever地址/STORAGE_ROOT/temp/uuid.dat
//临时对象转正
//正式对象分片 dataSever地址/STORAGE_ROOT/objects/对象hash.分片id.分片hash

type RSResumablePutStream struct {
	*RSPutStream
	*resumableToken
}

// NewRSResumablePutStream 新建一个RSResumablePutStream结构体指针，
// 其中内嵌了RSPutStream和resumableToken。
func NewRSResumablePutStream(dataServers []string, name, hash string, size int64) (*RSResumablePutStream, error) {
	putStream, err := NewRSPutStream(dataServers, hash, size)
	if err != nil {
		return nil, err
	}
	uuids := make([]string, AllShards)
	for i := range uuids { //获取uuid（初始化的结构体中有分配）
		uuids[i] = putStream.writers[i].(*objectstream.TempPutStream).Uuid
	}
	token := &resumableToken{name, size, hash, dataServers, uuids}
	return &RSResumablePutStream{putStream, token}, nil
}

func NewRSResumablePutStreamFromToken(token string) (*RSResumablePutStream, error) {
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}

	var t resumableToken
	err = json.Unmarshal(b, &t)
	mylogger.L().Println("t.hash:", t.Hash)
	if err != nil {
		return nil, err
	}

	writers := make([]io.Writer, AllShards)
	for i := range writers {
		writers[i] = &objectstream.TempPutStream{t.Servers[i], t.Uuids[i]}
		//mylogger.L().Printf("server:%s, uuid:%s \n", t.Servers[i], t.Uuids[i])
	}
	enc := NewEncoder(writers)
	return &RSResumablePutStream{&RSPutStream{enc}, &t}, nil
}

// ToToken 将自身数据以JSON格式输入，返回base64编码的字符串，用于接口服务返回的响应头部。
//
// **待优化**：尚未对该base64做相应的加密解密操作，传输的文件相关信息有可能被窃取。
func (s *RSResumablePutStream) ToToken() string {
	b, _ := json.Marshal(s)
	return base64.StdEncoding.EncodeToString(b)
}

// CurrentSize 获取某对象在数据服务节点中已上传的数据总大小
func (s *RSResumablePutStream) CurrentSize() int64 {
	//以http.HEAD方法获取第一个临时对象分片的大小，乘以4作为size返回
	r, err := http.Head(fmt.Sprintf("http://%s/temp/%s", s.Servers[0], s.Uuids[0]))
	if err != nil {
		mylogger.L().Println(err)
		return -1 //token不存在
	}
	if r.StatusCode != http.StatusOK {
		mylogger.L().Println(r.StatusCode)
		return -1
	}
	size := utils.GetSizeFromHeader(r.Header) * DataShards
	if size > s.Size { //若size（四倍分片）超过对象的大小，则返回对象大小
		size = s.Size
	}
	return size
}
