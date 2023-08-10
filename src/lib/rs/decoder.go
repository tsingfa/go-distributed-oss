package rs

import (
	"errors"
	"github.com/klauspost/reedsolomon"
	"io"
)

type decoder struct {
	readers   []io.Reader
	writers   []io.Writer
	enc       reedsolomon.Encoder
	size      int64
	cache     []byte
	cacheSize int
	total     int64
}

func NewDecoder(readers []io.Reader, writers []io.Writer, size int64) *decoder {
	enc, _ := reedsolomon.New(DataShards, ParityShards)
	return &decoder{readers, writers, enc, size, nil, 0, 0}
}

// Read decoder结构体实现了io.Reader，将读取结果保存到 p 中。
// 数据从readers和writers经修复后读取到cache，再将cache写入给 p 。
func (d *decoder) Read(p []byte) (n int, err error) {
	if d.cacheSize == 0 { //如果当cache中没有更多数据（缓存占用为0）
		err = d.getData() //则调用getData()获取数据
		if err != nil {   //没能获得更多数据
			return 0, err
		}
	}
	length := len(p)
	if d.cacheSize < length {
		length = d.cacheSize //要读取的长度不超过p的空间大小，也不超过缓存大小
	}
	d.cacheSize -= length
	copy(p, d.cache[:length])  //缓存被读取，保存到p中，缓存占用减小
	d.cache = d.cache[length:] //调整缓存
	return length, nil
}

// getData 将正常的（修复好的也算）分片读取到cache中。
// 读取过程中会检查readers和writers，从中获取正常的分片，并尝试恢复出所有的分片，
// 修复好后会将所有正常分片写入到cache中。
func (d *decoder) getData() error {
	if d.total == d.size { //已读取大小是否等于对象原始大小
		return io.EOF //若相等，则所有数据已被读取
	}
	shards := make([][]byte, AllShards) //若还有数据要读取，shards每个元素都用于保存相应分片中读取的数据
	repairIds := make([]int, 0)
	for i := range shards {
		if d.readers[i] == nil { //读不到的，则分片丢失，记录待修复分片id，shards[i]为nil
			repairIds = append(repairIds, i)
		} else { //若读到的reader不为nil
			shards[i] = make([]byte, BlockPerShard)        //初始化每个shard为8000字节的数组
			n, err := io.ReadFull(d.readers[i], shards[i]) //将readers上的数据读到shards
			if err != nil && err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
				shards[i] = nil //若发生非EOF失败，则分片读取错误--> 将shards[i]置为nil
			} else if n != BlockPerShard {
				shards[i] = shards[i][:n] //若读取数据长度不到8000字节，将该shard实际长度缩减为n
			}
		}
	}
	//若分片丢失或读取错误，则shards[i]为nil，调用enc的Reconstruct()方法修复分片
	err := d.enc.Reconstruct(shards)
	if err != nil { //修复失败
		return err
	}
	for i := range repairIds {
		id := repairIds[i]                     //若修复好后，shards[id]的数据已恢复为正常分片数据
		_, _ = d.writers[id].Write(shards[id]) //将修复好的分片数据，重新写回对应writers
	}
	//将所有的分片读到缓存中
	for i := 0; i < DataShards; i++ {
		shardSize := int64(len(shards[i]))
		if d.total+shardSize > d.size {
			shardSize -= d.total + shardSize - d.size //从该分片中要读取的大小（与total总和不超过size）
		}
		d.cache = append(d.cache, shards[i][:shardSize]...) //将每个分片数据添加到缓存cache中
		d.cacheSize += int(shardSize)                       //缓存占用大小
		d.total += shardSize                                //已读取到的全部数据大小
	}
	return nil
}
