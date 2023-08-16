package temp

import (
	"errors"
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

// 客户端
// "PUT", /temp/<token>
// range: bytes=<first>-<last>
//
// <partial content of object>

// 调用接口服务，将<partial content of object>写入 STORAGE_ROOT/temp/<uuid>.dat
// "PATCH", /temp/<uuid>
//
// <partial content of object>

func put(w http.ResponseWriter, r *http.Request) {
	token := strings.Split(r.URL.EscapedPath(), "/")[2]
	stream, err := rs.NewRSResumablePutStreamFromToken(token)
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	current := stream.CurrentSize() //数据服务节点中已上传的大小
	if current == -1 {              //如果token不存在
		w.WriteHeader(http.StatusNotFound)
		return
	}
	offset := utils.GetOffsetFromHeader(r.Header)
	if current != offset { //当前已上传的大小 vs 客户端指定的续传起点offset
		mylogger.L().Printf("current %d mismatch with offset %d...\n", current, offset)
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	bytes := make([]byte, rs.BlockSize)
	for {
		n, err := io.ReadFull(r.Body, bytes)
		if err != nil && err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
			mylogger.L().Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		current += int64(n)
		if current > stream.Size { //读到的总长度超过了对象的大小，则客户端上传的数据有误
			stream.Commit(false)
			mylogger.L().Println("resumable put exceed size...")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if n != rs.BlockSize && current != stream.Size {
			//【某次读取的长度n不到BlockSize】（已经读到r.Body的尾巴，所以没将BlockSize填满） --> 本次客户端上传结束
			//【读到的总长度current不等于对象的大小】--> 对象仍没上传完，还有后续数据需要上传
			//此时接口服务会【丢弃】最后那次读取不到BlockSize的数据，因为这是本次上传的尾巴而不是整个数据对象的尾巴，不能容忍其【不成块】
			//所以，只会保存已上传数据的【成块】部分
			return
		}
		_, _ = stream.Write(bytes[:n])
		if current == stream.Size { //如果读到
			stream.Flush()
			getStream, err := rs.NewRSResumableGetStream(stream.Servers, stream.Uuids, stream.Size)
			hash := utils.CalculateHash(getStream)
			if url.PathEscape(hash) != stream.Hash { //hash不一致，说明客户端上传的数据有误
				stream.Commit(false)
				mylogger.L().Println("resumable put done but hash mismatch")
				//mylogger.L().Printf("the hash specified by the client is %s, while the received object hash is %s\n", stream.Hash, hash)
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if locate.Exist(url.PathEscape(hash)) { //该对象已存在
				stream.Commit(false)
			} else {
				stream.Commit(true) //全新对象，保存
			}
			err = es.AddVersion(stream.Name, stream.Hash, stream.Size)
			if err != nil {
				mylogger.L().Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}
	}
}
