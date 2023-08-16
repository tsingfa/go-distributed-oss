package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// GetOffsetFromHeader 获取HTTP头部中的偏置起点（对应"range"头部）
//
// 格式要求：range: "bytes=first-xxx"，结果返回first。
func GetOffsetFromHeader(h http.Header) int64 {
	byteRange := h.Get("range")
	if len(byteRange) < 7 {
		return 0
	}
	if byteRange[:6] != "bytes=" {
		return 0
	}
	bytePos := strings.Split(byteRange[6:], "-")
	offset, _ := strconv.ParseInt(bytePos[0], 0, 64)
	return offset
}

// GetHashFromHeader 获取HTTP头部中的hash值(对应"digest"头部，要求"SHA-256"格式)
func GetHashFromHeader(h http.Header) string {
	digest := h.Get("digest") //获取http头部的digest
	if len(digest) < 9 {
		return ""
	} //digest的值形如SHA-256=example_hash
	if digest[:8] != "SHA-256=" {
		return ""
	} //确保SHA-256格式，否则hash为空
	return digest[8:] //返回hash内容
}

// GetSizeFromHeader 获取HTTP头部中的对象size(对应"content-length"头部)
func GetSizeFromHeader(h http.Header) int64 {
	size, _ := strconv.ParseInt(h.Get("content-length"), 0, 64)
	return size
}

// CalculateHash 计算对象内容哈希值的Base64编码
func CalculateHash(r io.Reader) string {
	h := sha256.New()                                    //sha256.digest结构体变量，内含以sha256为哈希函数的io.Writer
	_, _ = io.Copy(h, r)                                 //对writer的写入，会以哈希形式保存，Sum()方法可读取该哈希值
	return base64.StdEncoding.EncodeToString(h.Sum(nil)) //将哈希值编码为base64
}
