//接收并处理来自接口服务转发来的HTTP操作

package objects

/*
//实现数据校验功能后，上传都是使用temp接口的put进行缓存转正来实现
//此处的put不再使用

// put 上传（更新）文件的操作函数
func put(w http.ResponseWriter, r *http.Request) {
	//1.创建文件（根据文件名，拿到一个io.Writer）
	//os.Create()如果创建的文件名已存在，则原文件的内容会被清空。
	//"/objects/<object_name>" --> ["", "objects", "<object_name>"] --> "object_name"
	file, err := os.Create(os.Getenv("STORAGE_ROOT") + "/objects/" + strings.Split(r.URL.EscapedPath(), "/")[2])
	if err != nil { //文件创建失败
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func(file *os.File) {
		err := file.Close() //如果文件创建成功，记得延迟关闭
		if err != nil {
			log.Println(err)
		}
	}(file)
	//2.拷贝文件内容
	_, err = io.Copy(file, r.Body) //读取r.Body，写入到服务器中的file
	if err != nil {
		log.Println(err)
		return
	}
}
*/
