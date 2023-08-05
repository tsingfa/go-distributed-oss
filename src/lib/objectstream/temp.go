package objectstream

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// TempPutStream 有Write方法，实现了io.Writer接口
type TempPutStream struct {
	Server string
	Uuid   string
}

func NewTempPutStream(server, hash string, size int64) (*TempPutStream, error) {
	request, err := http.NewRequest("POST", "http://"+server+"/temp/"+hash, nil) //请求获得uuid
	if err != nil {
		return nil, err
	}
	request.Header.Set("size", fmt.Sprintf("%d", size))
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	uuid, err := io.ReadAll(response.Body) //从响应中读取uuid
	if err != nil {
		return nil, err
	}
	return &TempPutStream{server, string(uuid)}, nil
}

// Write 根据server和uuid，以PATCH方法访问数据服务的temp接口，将需要的数据上传至缓存区，返回上传的字节长度。
func (w *TempPutStream) Write(p []byte) (int, error) {
	request, err := http.NewRequest("PATCH", "http://"+w.Server+"/temp/"+w.Uuid, strings.NewReader(string(p)))
	if err != nil {
		return 0, err
	}
	client := http.Client{}
	r, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	if r.StatusCode != http.StatusOK { //将请求正文写入dat文件失败||获取dat文件信息失败||实际写入大小与预期不符
		return 0, fmt.Errorf("dataServer return http code %d", r.StatusCode)
	}
	return len(p), nil
}

// Commit 决定是否提交缓存
func (w *TempPutStream) Commit(good bool) {
	method := "DELETE"
	if good {
		method = "PUT"
	}
	request, _ := http.NewRequest(method, "http://"+w.Server+"/temp/"+w.Uuid, nil)
	client := http.Client{}
	_, _ = client.Do(request)
}

func NewTempGetStream(server, uuid string) (*GetStream, error) {
	return newGetStream("http://" + server + "/temp/" + uuid)
}
