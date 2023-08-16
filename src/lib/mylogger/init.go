package mylogger

import (
	"log"
)

var myLogger *MyLogger

func init() {
	var err error
	path := "/home/tsingfa/GolandProject/go-distributed-oss/"
	myLogger, err = NewMyLogger(path + "logfile.log")
	if err != nil {
		log.Fatal("无法创建日志文件:", err)
	}
}

func L() *MyLogger {
	return myLogger
}
