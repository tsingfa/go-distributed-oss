package objects

import (
	"go-distributed-oss/dataServer/locate"
	"go-distributed-oss/src/lib/utils"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// get 下载文件的操作函数
func get(w http.ResponseWriter, r *http.Request) {
	filepath := getFile(strings.Split(r.URL.EscapedPath(), "/")[2])
	if filepath == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	sendFile(w, filepath)
}

// getFile 根据hash值获取含路径文件名
func getFile(hash string) string {
	filepath := os.Getenv("STORAGE_ROOT") + "/objects/" + hash
	file, err := os.Open(filepath)
	if err != nil {
		log.Println("open file failed,filepath:", filepath)
		return ""
	}
	d := url.PathEscape(utils.CalculateHash(file))
	_ = file.Close()
	if d != hash { //数据校验（防止数据降解--上传时正确的数据也有可能随着时间流逝而损坏）
		log.Println("object hash mismatch,remove", filepath) //所以获取数据时仍需校验一遍
		locate.Del(hash)
		_ = os.Remove(filepath) //若校验失败，则从缓存和磁盘中删除对象
		return ""
	}
	return filepath
}

// sendFile 将指定路径的文件传输到指定写入流中
func sendFile(w io.Writer, filepath string) {
	file, _ := os.Open(filepath)
	defer func(f *os.File) {
		_ = f.Close()
	}(file)
	_, _ = io.Copy(w, file)
}
