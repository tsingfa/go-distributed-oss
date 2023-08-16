//PUT 或 DELETE
//method, "http://"+w.Server+"/temp/"+w.Uuid,缓存转正 或 缓存删除

package temp

import (
	"go-distributed-oss/dataServer/locate"
	"go-distributed-oss/src/lib/mylogger"
	"go-distributed-oss/src/lib/utils"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func put(w http.ResponseWriter, r *http.Request) {
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	tempinfo, err := readFromFile(uuid) //获取uuid以及临时文件信息
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	datFile := infoFile + ".dat"
	file, err := os.Open(datFile)
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(file)
	info, err := file.Stat()
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	actual := info.Size()
	_ = os.Remove(infoFile)
	if actual != tempinfo.Size { //传输的文件大小与渔区不符
		_ = os.Remove(datFile)
		mylogger.L().Println("actual size mismatch,expect:", tempinfo.Size, ",but actual:", actual)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	commitTempObject(datFile, tempinfo) //临时文件转正
}

// commitTempObject 将临时对象（分片）转正
func commitTempObject(datFile string, tempinfo *tempInfo) {
	file, _ := os.Open(datFile)
	d := url.PathEscape(utils.CalculateHash(file)) //得到分片hash
	_ = file.Close()
	_ = os.Rename(datFile, os.Getenv("STORAGE_ROOT")+"/objects/"+tempinfo.Name+"."+d) //重命名为："对象hash.分片id.分片hash"
	locate.Add(tempinfo.hash(), tempinfo.id())                                        //加入定位缓存（对象hash--分片id）
}

// hash 获取对象hash
func (t *tempInfo) hash() string {
	s := strings.Split(t.Name, ".")
	return s[0]
}

// id 获取分片id
func (t *tempInfo) id() int {
	s := strings.Split(t.Name, ".")
	id, _ := strconv.Atoi(s[1])
	return id
}
