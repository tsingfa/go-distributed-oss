package main

import (
	"go-distributed-oss/src/lib/es"
	"go-distributed-oss/src/lib/mylogger"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//注意：管理员维护时的“清除”与客户端的“删除”，二者不同！！
// 维护时的清除：将无引用的对象移至garbage，等待定期真正删除
// 客户端的删除：删除元数据，使之对客户端不可见，而没有实际删除对象

// 数据节点定期执行，用于定期删除数据节点中无引用的对象
func main() {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/*")
	for i := range files {
		hash := strings.Split(filepath.Base(files[i]), ".")[0]
		hashInMetadata, err := es.HasHash(hash)
		if err != nil {
			mylogger.L().Println(err)
			return
		}
		if !hashInMetadata { //对象无引用（元数据中没有记录）
			delByAPI(hash)
		}
	}
}

// del 请求到数据节点，执行对指定hash对象（分片）的删除（移到garbage）
// 每次指定一个数据节点，删除分片文件（要执行多次，才能删完整个对象）
func del(hash string) {
	mylogger.L().Printf("delete the shard of %s in server %s\n", hash, os.Getenv("LISTEN_ADDRESS"))
	url := "http://" + os.Getenv("LISTEN_ADDRESS") + "/objects/" + hash
	request, _ := http.NewRequest("DELETE", url, nil)
	client := http.Client{}
	_, _ = client.Do(request)
}

// 通过apiServer进行无引用对象的清除，api先locate得到各数据节点，再依次对其发送garbage请求。
// 故管理员只需执行一次即可，不用多次执行
func delByAPI(hash string) {
	mylogger.L().Println("delete: ", hash)
	url := "http://" + os.Getenv("LISTEN_ADDRESS") + "/objects/" + hash
	request, _ := http.NewRequest("GARBAGE", url, nil)
	client := http.Client{}
	_, _ = client.Do(request)
}
