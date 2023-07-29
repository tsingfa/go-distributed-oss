//一些HTTP的操作函数
//put、get操作

//*os.File同时实现了io.Writer和io.Reader，可读可写。

package objects

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// put 上传（更新）文件的操作函数
func put(w http.ResponseWriter, r *http.Request) {
	//1.创建文件（根据文件名，拿到一个io.Writer）
	//os.Create()如果创建的文件名已存在，则原文件的内容会被清空。
	//"/objects/<object_name>" --> ["", "objects", "<object_name>"] --> "object_name"
	file, err := os.Create(os.Getenv("STORAGE_ROOT") + "/objects/" + strings.Split(r.URL.EscapedPath(), "/")[2])
	if err != nil { //文件创建失败
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func(file *os.File) {
		err := file.Close() //如果文件创建成功，记得延迟关闭
		if err != nil {
			log.Println(err)
		}
	}(file)
	//2.拷贝文件内容
	_, err = io.Copy(file, r.Body) //读取r.Body，写入到服务器中的file
	if err != nil {
		log.Println(err)
		return
	}
}

// get 下载文件的操作函数
func get(w http.ResponseWriter, r *http.Request) {
	//1.打开文件
	file, err := os.Open(os.Getenv("STORAGE_ROOT") + "/objects/" + strings.Split(r.URL.EscapedPath(), "/")[2])
	if err != nil { //文件打开失败 --> 找不到该文件
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer func(file *os.File) {
		err := file.Close() //文件打开成功，记得关闭
		if err != nil {
			log.Println(err)
		}
	}(file)
	//2.写回客户端
	_, err = io.Copy(w, file) //读取服务器中的文件，从输出流写回给客户端
	if err != nil {
		log.Println(err)
		return
	}
}
