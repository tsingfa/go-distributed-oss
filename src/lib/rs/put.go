package rs

import (
	"fmt"
	"go-distributed-oss/src/lib/objectstream"
	"io"
)

type RSPutStream struct {
	*encoder
}

func NewRSPutStream(dataServers []string, hash string, size int64) (*RSPutStream, error) {
	if len(dataServers) != AllShards {
		return nil, fmt.Errorf("dataServers number mismatch")
	}
	perShard := (size + DataShards - 1) / DataShards //计算每个分片的大小，即[size/DataShards]，有小数点的向上取整，没小数点则不变
	writers := make([]io.Writer, AllShards)
	var err error
	for i := range writers { //长度为：数据分片数+纠错分片数
		writers[i], err = objectstream.NewTempPutStream(dataServers[i],
			fmt.Sprintf("%s.%d", hash, i), perShard)
		if err != nil {
			return nil, err
		}
	}
	newEncoder := NewEncoder(writers) //encoder结构体
	return &RSPutStream{newEncoder}, nil
}

func (s *RSPutStream) Commit(success bool) {
	s.Flush() //将缓存最后的数据写入writers
	for i := range s.writers {
		s.writers[i].(*objectstream.TempPutStream).Commit(success) //由bool决定是否提交缓存
	}
}
