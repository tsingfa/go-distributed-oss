package objectstream

import (
	"fmt"
	"io"
	"net/http"
)

type GetStream struct {
	reader io.Reader
}

// newGetStream
func newGetStream(url string) (*GetStream, error) {
	resp, err := http.Get(url) //向指定URL发起get请求，响应结果保存到resp（含响应头、响应主体）
	if err != nil {            //调用错误
		return nil, err
	}
	if resp.StatusCode != http.StatusOK { //其他错误码
		return nil, fmt.Errorf("dataServer return http code %d", resp.StatusCode)
	}
	return &GetStream{resp.Body}, nil
}

// NewGetStream	新建一个面向指定【数据服务结点和数据对象-->URL】的读取流。
// 该读取流中的内容为对该URL的GET请求的响应主体。
func NewGetStream(server, object string) (*GetStream, error) {
	if server == "" || object == "" { //数据结点地址和数据对象都不应为空
		return nil, fmt.Errorf("invalid server %s object %s", server, object)
	}
	return newGetStream("http://" + server + "/objects/" + object)
}

// GetStream结构体实现了io.Reader接口
func (r *GetStream) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}
