package rs

const (
	DataShards    = 4                          //数据分片
	ParityShards  = 2                          //纠错分片（冗余）
	AllShards     = DataShards + ParityShards  //总分片数
	BlockPerShard = 8000                       //缓存上限（单个分片）
	BlockSize     = BlockPerShard * DataShards //缓存上限（所有数据分片总计）
)
