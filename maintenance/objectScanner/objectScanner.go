package main

import (
	"go-distributed-oss/apiServer/objects"
	"go-distributed-oss/src/lib/es"
	"go-distributed-oss/src/lib/mylogger"
	"go-distributed-oss/src/lib/utils"
	"os"
	"path/filepath"
	"strings"
)

// 定期运行，用于检查并修复数据
func main() {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/*")
	for i := range files {
		hash := strings.Split(filepath.Base(files[i]), ".")[0]
		verify(hash)
	}
}

// 验证并修复对象分片
func verify(hash string) {
	mylogger.L().Println("verify:", hash)
	size, err := es.SearchHashSize(hash)
	if err != nil {
		mylogger.L().Println(err)
		return
	}
	stream, err := objects.GetStream(hash, size)
	if err != nil {
		mylogger.L().Println(err)
		return
	}
	d := utils.CalculateHash(stream) //生成的读取流，在读取时会自动修复对象分片
	if d != hash {                   //如果（修复后）对象hash与记录的不一致，则报告
		mylogger.L().Printf("object hash mismatch,calculated=%s,requested=%s\n", d, hash)
	}
	stream.Close()
}
