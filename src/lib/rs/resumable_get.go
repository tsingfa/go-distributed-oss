package rs

import (
	"go-distributed-oss/src/lib/objectstream"
	"io"
)

type RSResumableGetStream struct {
	*decoder
}

func NewRSResumableGetStream(dataServers []string, uuids []string, size int64) (*RSResumableGetStream, error) {
	readers := make([]io.Reader, AllShards)
	var e error
	for i := 0; i < AllShards; i++ {
		readers[i], e = objectstream.NewTempGetStream(dataServers[i], uuids[i])
		if e != nil {
			return nil, e
		}
	}
	writers := make([]io.Writer, AllShards)
	dec := NewDecoder(readers, writers, size)
	return &RSResumableGetStream{dec}, nil
}
