package temp

import (
	"fmt"
	"go-distributed-oss/src/lib/mylogger"
	"net/http"
	"os"
	"strings"
)

// head 将 /temp/<uuid>.dat文件大小放在content-length响应头部返回
func head(w http.ResponseWriter, r *http.Request) {
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	file, err := os.Open(os.Getenv("STORAGE_ROOT") + "/temp/" + uuid + ".dat")
	if err != nil {
		mylogger.L().Println(err)
		w.WriteHeader(http.StatusNotFound)
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
	w.Header().Set("content-length", fmt.Sprintf("%d", info.Size()))
}
