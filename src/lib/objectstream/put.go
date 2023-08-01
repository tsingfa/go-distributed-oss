package objectstream

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

type PutStream struct {
	writer *io.PipeWriter //用于实现Write方法
	ch     chan error     //用于将goroutine传输过程中的error传回主线程
}

// NewPutStream 生成一个PutStream结构体，内置io.Writer和记录error的channel。
// 参数server、object用于构建HTTP请求目标。
// 将面向指定【数据服务结点、数据对象】的put请求的函数调用，转换成读写流的形式。
func NewPutStream(server, object string) *PutStream {
	reader, writer := io.Pipe() //一对管道互联的读写器：写入writer的内容，可以从reader读出来
	ch := make(chan error)
	go func() { //异步执行，不会随函数返回而终止
		//创建一个http request请求，该请求为发给指定URL的PUT请求
		//该请求的主体内容可以从reader中读取出来（要等待writer有写入）
		req, _ := http.NewRequest("PUT", "http://"+server+"/objects/"+object, reader)
		client := http.Client{}
		resp, err := client.Do(req)                         //通过一个客户端执行该请求
		if err == nil && resp.StatusCode != http.StatusOK { //可执行但结果不为200
			err = fmt.Errorf("dataServer return http code %d", resp.StatusCode)
		}
		ch <- err //err!=nil
	}()
	return &PutStream{writer, ch}
}

// Write 实现了io.Writer接口，用于写入writer
func (w *PutStream) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

// Close 实现了io.Closer接口，用于关闭writer。
// 让管道另一边的reader读到io.EOF，否则在goroutine中运行的Client.Do将一直阻塞。
func (w *PutStream) Close() error {
	err := w.writer.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	return <-w.ch
}
