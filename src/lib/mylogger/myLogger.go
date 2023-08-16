package mylogger

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
)

var colorCodeRegex = regexp.MustCompile("\x1b\\[[0-9;]+m")

type MyLogger struct {
	logger *log.Logger
}

func NewMyLogger(outputFile string) (*MyLogger, error) {
	// 检查文件是否存在
	_, err := os.Stat(outputFile)
	if os.IsNotExist(err) {
		// 文件不存在，创建新文件
		file, createErr := os.Create(outputFile)
		if createErr != nil {
			return nil, createErr
		}
		return &MyLogger{
			logger: log.New(file, "", log.LstdFlags),
		}, nil
	}

	// 文件已存在，追加日志到现有文件
	file, appendErr := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if appendErr != nil {
		return nil, appendErr
	}
	return &MyLogger{
		logger: log.New(file, "", log.LstdFlags),
	}, nil
}

func (ml *MyLogger) Println(v ...any) {
	output := ml.logWithInfo(v...)
	// 将带颜色的输出字符串写入文件（去除颜色信息）
	ml.logger.Println(removeColorCodes(output))
	// 在终端中输出带颜色的信息
	fmt.Println(output)
}

func (ml *MyLogger) Printf(format string, v ...any) {
	output := ml.logWithInfo(fmt.Sprintf(format, v...))
	// 将带颜色的输出字符串写入文件（去除颜色信息）
	ml.logger.Printf(removeColorCodes(output))
	// 在终端中输出带颜色的信息
	fmt.Printf(output)
}

func (ml *MyLogger) Fatal(v ...interface{}) {
	output := ml.logWithInfo(v...)
	// 将带颜色的输出字符串写入文件（去除颜色信息）
	ml.logger.Fatal(removeColorCodes(output))
	// 在终端中输出带颜色的信息
	log.Fatal(output)
}

func (ml *MyLogger) logWithInfo(v ...any) string {
	server := os.Getenv("LISTEN_ADDRESS")
	pc, file, line, _ := runtime.Caller(2)
	funcName := runtime.FuncForPC(pc).Name()
	// 获取文件名的相对路径
	_, fileName := path.Split(file)
	// 创建缩进空格
	indent := strings.Repeat(" ", 0)
	now := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s %s:%s(%s):%d]\n%s", now, server, funcName, fileName, line, indent)

	// 使用颜色为 prefix 添加样式
	coloredPrefix := colorBlue + prefix + colorReset

	// 创建带颜色的输出字符串
	coloredOutput := fmt.Sprintf("%s%s", coloredPrefix, fmt.Sprint(v...))
	return coloredOutput
}

func removeColorCodes(s string) string {
	// 移除 ANSI 转义序列
	return colorCodeRegex.ReplaceAllString(s, "")
}
