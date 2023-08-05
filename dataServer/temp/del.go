//PUT 或 DELETE
//method, "http://"+w.Server+"/temp/"+w.Uuid,缓存转正 或 缓存删除

package temp

import (
	"net/http"
	"os"
	"strings"
)

func del(w http.ResponseWriter, r *http.Request) {
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	datFile := infoFile + ".dat"
	//清除临时文件
	os.Remove(infoFile)
	os.Remove(datFile)
}
