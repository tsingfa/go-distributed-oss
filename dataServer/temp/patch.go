//"PATCH", "http://"+w.Server+"/temp/"+w.Uuid, strings.NewReader(string(p))，
//根据server和uuid，以PATCH方法访问数据服务的temp接口，将数据（字节流p）上传至缓存区

package temp

import (
	"encoding/json"
	"go-distributed-oss/src/lib/mylogger"
	"io"
	"net/http"
	"os"
	"strings"
)

func patch(w http.ResponseWriter, r *http.Request) {
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	tempinfo, err := readFromFile(uuid) //获取uuid以及临时文件信息
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	datFile := infoFile + ".dat"
	file, err := os.OpenFile(datFile, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(file)
	_, err = io.Copy(file, r.Body) //将请求正文写入dat文件中
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	info, err := file.Stat() //获取dat的文件信息
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	actual := info.Size()       //实际写入的dat文件大小
	if actual > tempinfo.Size { //实际写入大小与预期不符-->文件有误（删除）
		_ = os.Remove(datFile)
		_ = os.Remove(infoFile)
		mylogger.L().Println("actual size", actual, "exceeds", tempinfo.Size)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// readFromFile 根据uuid，查询临时对象的信息，返回含有信息的结构体指针。
func readFromFile(uuid string) (*tempInfo, error) {
	file, err := os.Open(os.Getenv("STORAGE_ROOT") + "/temp/" + uuid)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(file)
	b, _ := io.ReadAll(file)
	var info tempInfo
	_ = json.Unmarshal(b, &info) //将b的内容反序列化到结构体中
	return &info, nil
}
