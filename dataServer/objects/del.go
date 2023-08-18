package objects

import (
	"go-distributed-oss/dataServer/locate"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// del 查找并删除（移动到garbage）指定hash的对象，并清除缓存
func del(w http.ResponseWriter, r *http.Request) {
	hash := strings.Split(r.URL.EscapedPath(), "/")[2]
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/" + hash + ".*")
	if len(files) != 1 {
		return
	}
	locate.Del(hash)
	os.Rename(files[0], os.Getenv("STORAGE_ROOT")+"/garbage/"+filepath.Base(files[0]))
}
