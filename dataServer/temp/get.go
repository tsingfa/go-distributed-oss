package temp

import (
	"go-distributed-oss/src/lib/mylogger"
	"io"
	"net/http"
	"os"
	"strings"
)

// get 打开/temp/<uuid>.dat文件，并将其内容作为HTTP响应正文输出
func get(w http.ResponseWriter, r *http.Request) {
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
	_, _ = io.Copy(w, file)
}
