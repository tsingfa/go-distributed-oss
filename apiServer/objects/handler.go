package objects

import (
	"net/http"
)

const MethodGarbage = "GARBAGE" //（维护时）调用dataServer将无引用对象移到garbage目录

// Handler 整体的控制函数
// w (http.ResponseWriter)用于写入HTTP响应，
// r (*http.Request)代表当前处理的HTTP请求。
func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method //根据具体的请求方式，选择对应的处理函数
	if m == http.MethodPost {
		post(w, r)
		return
	}
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
	if m == MethodGarbage { //（维护时）调用dataServer将无引用对象移到garbage目录
		garbage(w, r)
	}
	//找不到请求方式对应的处理函数 --> 请求方式不合规
	w.WriteHeader(http.StatusMethodNotAllowed) //WriteHeader写HTTP响应的错误代码
}
