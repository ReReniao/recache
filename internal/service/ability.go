package service

// PeerPicker 和 PeerGetter 用于结点间通信

type Picker interface {
	Pick(key string) (Fetcher, bool)
}

type Fetcher interface {
	Fetch(group string, key string) ([]byte, error)
}

type Retriever interface {
	retrieve(key string) ([]byte, error)
}

// RetrieverFunc 接口回调函数
// 用于缓存未命中时获取数据
type RetrieverFunc func(key string) ([]byte, error)

func (f RetrieverFunc) retrieve(key string) ([]byte, error) {
	return f(key)
}

var _ Retriever = RetrieverFunc(nil)
