package objects

import (
	"net/http"
)

// Handler 整体的控制函数
// w (http.ResponseWriter)用于写入HTTP响应，
// r (*http.Request)代表当前处理的HTTP请求。
func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method //根据具体的请求方式，选择对应的处理函数
	if m == http.MethodPut {
		put(w, r)
		return
	}
	if m == http.MethodGet {
		get(w, r)
		return
	}
	if m == http.MethodDelete {
		del(w, r)
		return
	}
	//找不到请求方式对应的处理函数 --> 请求方式不合规
	w.WriteHeader(http.StatusMethodNotAllowed) //WriteHeader写HTTP响应的错误代码
}
