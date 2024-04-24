package reniaocache

// PeerPicker 和 PeerGetter 用于结点间通信

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Fetch(group string, name string) ([]byte, error)
}
