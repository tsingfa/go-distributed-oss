//PUT 或 DELETE
//method, "http://"+w.Server+"/temp/"+w.Uuid,缓存转正 或 缓存删除

package temp

import (
	"go-distributed-oss/dataServer/locate"
	"log"
	"net/http"
	"os"
	"strings"
)

func put(w http.ResponseWriter, r *http.Request) {
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	tempinfo, err := readFromFile(uuid) //获取uuid以及临时文件信息
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	datFile := infoFile + ".dat"
	file, err := os.Open(datFile)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(file)
	info, err := file.Stat()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	actual := info.Size()
	os.Remove(infoFile)
	if actual != tempinfo.Size { //传输的文件大小与渔区不符
		os.Remove(datFile)
		log.Println("actual size mismatch,expect:", tempinfo.Size, ",but actual:", actual)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	commitTempObject(datFile, tempinfo) //临时文件转正
}

// commitTempObject 将临时对象转正
func commitTempObject(datFile string, tempinfo *tempInfo) {
	os.Rename(datFile, os.Getenv("STORAGE_ROOT")+"/objects/"+tempinfo.Name)
	locate.Add(tempinfo.Name) //加入hash缓存记录
}
