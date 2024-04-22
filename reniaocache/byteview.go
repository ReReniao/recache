package reniaocache

// ByteView 只读的缓存值
type ByteView struct {
	b []byte
}

// 实现Value接口
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回一个副本
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String 返回值的字符串形式
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
