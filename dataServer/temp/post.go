//"POST", "http://"+server+"/temp/"+hash,获得一个带有uuid的写入流

package temp

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

//用于存储临时文件信息的文件 "POST", "http://"+server+"/temp/"+hash
//os.Getenv("STORAGE_ROOT") + "/temp/" + t.Uuid

//用于存储临时对象的.dat文件 "PATCH", "http://"+w.Server+"/temp/"+w.Uuid
//os.Getenv("STORAGE_ROOT") + "/temp/" + t.Uuid + ".dat"

//临时对象转正的命名 "PUT"/"DELETE", "http://"+w.Server+"/temp/"+w.Uuid
//os.Getenv("STORAGE_ROOT")+"/objects/"+tempinfo.Name

type tempInfo struct {
	Uuid string
	Name string //用于临时文件转正时的重命名，具体查看对应NewXXXStream的中的hash参数
	Size int64
}

func post(w http.ResponseWriter, r *http.Request) {
	output, _ := exec.Command("uuidgen").Output()
	uuid := strings.TrimSuffix(string(output), "\n") //生成一个uuid并做格式处理
	hash := strings.Split(r.URL.EscapedPath(), "/")[2]
	size, err := strconv.ParseInt(r.Header.Get("size"), 0, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	t := tempInfo{
		Uuid: uuid,
		Name: hash, //文件对象的哈希，此处用于临时文件命名（实现分片后为"hash.id"）
		Size: size,
	}
	err = t.writeToFile()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = os.Create(os.Getenv("STORAGE_ROOT") + "/temp/" + t.Uuid + ".dat") //创建".dat"临时文件
	_, _ = w.Write([]byte(uuid))                                             //写回响应（生成的uuid）
}

// writeToFile 将t的信息（uuid、hash、size）写入<uuid>临时文件。
// 该文件用于保存【临时对象信息】，而".dat"文件用于保存对象内容，注意区分。
func (t *tempInfo) writeToFile() error {
	file, err := os.Create(os.Getenv("STORAGE_ROOT") + "/temp/" + t.Uuid)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(file)
	b, _ := json.Marshal(t)
	_, _ = file.Write(b)
	return nil
}
