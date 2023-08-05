// 数据服务的缓存相关操作的整体控制函数

//"POST", "http://"+server+"/temp/"+hash,请求获得uuid

//"PATCH", "http://"+w.Server+"/temp/"+w.Uuid, strings.NewReader(string(p))，
//根据server和uuid，以PATCH方法访问数据服务的temp接口，将数据（字节流p）上传至缓存区

//PUT 或 DELETE
//method, "http://"+w.Server+"/temp/"+w.Uuid,缓存转正 或 缓存删除

package temp

import "net/http"

func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m == http.MethodPut {
		put(w, r)
		return
	}
	if m == http.MethodPatch {
		patch(w, r)
		return
	}
	if m == http.MethodPost {
		post(w, r)
		return
	}
	if m == http.MethodDelete {
		del(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}
