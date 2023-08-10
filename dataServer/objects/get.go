package objects

import (
	"crypto/sha256"
	"encoding/base64"
	"go-distributed-oss/dataServer/locate"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// get 下载文件的操作函数
//
// 请求链:（请求 对象hash） --> get[api层] --> NewRSGetStream -->
// NewGetStream(请求 对象hash.分片id） --> get[data层] --> 写回读取流，再将读取流写回response
func get(w http.ResponseWriter, r *http.Request) {
	filenamePath := getFile(strings.Split(r.URL.EscapedPath(), "/")[2]) //由对象哈希获取完整路径文件名
	if filenamePath == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	sendFile(w, filenamePath) //将该路径的文件传到写入流中
}

/* //适用于未分片的版本

// getFile 根据hash值，获取【带路径文件名】。
// 返回前需检查该文件能否正常打开、内容（哈希）是否与要求的一致。
func getFile(hash string) string {
	filepath := os.Getenv("STORAGE_ROOT") + "/objects/" + hash
	file, err := os.Open(filepath)
	if err != nil {
		log.Println("open file failed,filepath:", filepath)
		return ""
	}
	d := url.PathEscape(utils.CalculateHash(file)) //获取对象hash值
	_ = file.Close()
	if d != hash { //数据校验（防止数据降解--上传时正确的数据也有可能随着时间流逝而损坏）
		log.Println("object hash mismatch,remove", filepath) //所以获取数据时仍需校验一遍
		locate.Del(hash)
		_ = os.Remove(filepath) //若校验失败，则从缓存和磁盘中删除对象
		return ""
	}
	return filepath
}
*/

// getFile 实现RS分片后的版本，仅需传入"对象hash.分片id"，即可返回完整带路径的分片文件名
func getFile(name string) string {
	filenamePaths, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/" + name + ".*") //文件名匹配
	if len(filenamePaths) != 1 {
		log.Printf("filenamePaths not found,name %s, found %s\n", name, filenamePaths)
		return ""
	}
	filenamePath := filenamePaths[0] //分片的带路径文件名
	h := sha256.New()
	sendFile(h, filenamePath)
	d := url.PathEscape(base64.StdEncoding.EncodeToString(h.Sum(nil))) //真实的分片hash
	hash := strings.Split(filenamePath, ".")[2]                        //文件名中标记的分片hash
	if d != hash {                                                     //数据校验不通过 --> 内容已经损坏
		log.Println("object hash mismatch, remove", filenamePath)
		locate.Del(hash)
		os.Remove(filenamePath)
		return ""
	}
	return filenamePath
}

// sendFile 将指定路径的文件传输到指定写入流中
func sendFile(w io.Writer, filepath string) {
	file, _ := os.Open(filepath)
	defer func(f *os.File) {
		_ = f.Close()
	}(file)
	_, _ = io.Copy(w, file)
}
