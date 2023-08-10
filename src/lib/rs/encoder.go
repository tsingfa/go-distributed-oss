package rs

import (
	"github.com/klauspost/reedsolomon"
	"io"
)

type encoder struct {
	writers []io.Writer
	enc     reedsolomon.Encoder
	cache   []byte
}

func NewEncoder(writers []io.Writer) *encoder {
	enc, _ := reedsolomon.New(DataShards, ParityShards)
	return &encoder{writers, enc, nil}
}

// Write 将数据先写入缓存，缓存满再调用Flush()方法实际写入writers
// （有调用TempPutStream的Write方法），然后清空缓存接着写。
func (e *encoder) Write(p []byte) (n int, err error) {
	length := len(p)  //还需要写入的长度
	current := 0      //每轮写入的起点
	for length != 0 { //将p待写入的部分以块的形式放入缓存
		next := BlockSize - len(e.cache) //本轮要写入的长度（上限）
		if next > length {
			next = length
		}
		e.cache = append(e.cache, p[current:current+next]...)
		if len(e.cache) == BlockSize { //若缓存已满，则将缓存实际写入writers并清空缓存
			e.Flush()
		}
		current += next //更新每轮写入的起点、仍待写入的长度
		length -= next
	}
	return len(p), nil
}

// Flush 将缓存数据分片后写入writers并清空缓存。
func (e *encoder) Flush() {
	if len(e.cache) == 0 {
		return
	}
	shards, _ := e.enc.Split(e.cache) //将缓存数据切成4个数据片
	_ = e.enc.Encode(shards)          //根据现有的数据分片，生成2个校验片
	for i := range shards {           //将6个片的数据依次写入writers并清空缓存
		_, _ = e.writers[i].Write(shards[i])
	}
	e.cache = []byte{}
}

func (e *encoder) Close() {

}
