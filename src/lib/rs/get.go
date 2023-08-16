package rs

import (
	"fmt"
	"go-distributed-oss/src/lib/objectstream"
	"io"
)

type RSGetStream struct {
	*decoder
}

// NewRSGetStream
//
// locateInfo是个哈希表，输入时仅记录分片完好的节点，丢失的分片对应节点记录为空。
//
// dataSevers是随机选择出来用于存放修复分片的数据节点，若无丢失、无需修复，则为空切片。
func NewRSGetStream(locateInfo map[int]string, dataServers []string, hash string, size int64) (*RSGetStream, error) {
	if len(locateInfo)+len(dataServers) != AllShards {
		return nil, fmt.Errorf("dataServers number mismatch")
	}
	readers := make([]io.Reader, AllShards)
	for i := 0; i < AllShards; i++ {
		server := locateInfo[i] //查看每个分片id对应的server
		if server == "" {       //若某分片id对应的server为空 --> 该分片丢失
			locateInfo[i] = dataServers[0]
			dataServers = dataServers[1:]
			continue
		}
		//从数据节点get指定分片文件("hash.id")，返回对应读取流（读取到分片文件）
		reader, err := objectstream.NewGetStream(server, fmt.Sprintf("%s.%d", hash, i))
		if err == nil { //没有读取错误
			readers[i] = reader
		}
	}

	writers := make([]io.Writer, AllShards)
	perShard := (size + DataShards - 1) / DataShards
	var err error
	for i := range readers {
		if readers[i] == nil { //某节点读不到指定分片，读取结果为nil，则创建对应的分片写入流用于恢复分片
			writers[i], err = objectstream.NewTempPutStream(locateInfo[i], fmt.Sprintf("%s.%d", hash, i), perShard)
			if err != nil {
				return nil, err
			}
		}
	}
	//readers和writers数组二者互补，对于某个分片id，
	//要么readers中存在对应的读取流 --> 分片存在且未损坏，可读取 ；
	//要么writers中存在相应的写入流 --> 分片不存在或因损坏而被清除，待写入修复分片
	dec := NewDecoder(readers, writers, size)
	return &RSGetStream{dec}, nil
}

// Close 在流关闭时，将临时对象转正
func (s *RSGetStream) Close() {
	for i := range s.writers {
		if s.writers[i] != nil {
			s.writers[i].(*objectstream.TempPutStream).Commit(true)
		}
	}
}

// Seek 用于调整读取流中的位置，以便能够从指定位置开始读取数据
func (s *RSGetStream) Seek(offset int64, whence int) (int64, error) {
	if whence != io.SeekCurrent { //只能从当前位置起跳
		panic("only support SeekCurrent")
	}
	if offset < 0 { //跳过的字节数不能为负
		panic("only support forward seek")
	}
	for offset != 0 {
		//如果offset小于BlockSize，则直接读取offset即可
		//如果offset大于BlockSize，则每轮往后读BlockSize长度，直至读到offset
		length := int64(BlockSize)
		if offset < length {
			length = offset
		}
		buf := make([]byte, length)
		_, _ = io.ReadFull(s, buf) //读取并丢弃
		offset -= length           //本次读取了length长度
	}
	return offset, nil
}
