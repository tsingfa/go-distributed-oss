//注意！！！ 与apiServer/objects中的handler.go区分
//即便代码一模一样，但是所调用的put和get函数完全不一样！！！

package objects

import "net/http"

func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m == http.MethodPut {
		put(w, r)
		return
	}
	if m == http.MethodGet {
		get(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}
