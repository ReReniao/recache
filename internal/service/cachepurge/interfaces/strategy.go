package interfaces

import (
	"time"
)

type CacheStrategy interface {
	Get(string) (Value, *time.Time, bool)
	Put(string, Value)
	CleanUp(ttl time.Duration)
	Len() int
}

// Entry 数据实例
type Entry struct {
	Key      string
	Value    Value
	UpdateAt *time.Time
}

// Value 的 Len() 方法用于返回它的字节数
type Value interface {
	Len() int
}

func (ele *Entry) Expired(duration time.Duration) (ok bool) {
	if ele.UpdateAt == nil {
		ok = false
	} else {
		ok = ele.UpdateAt.Add(duration).Before(time.Now())
	}
	return
}

// touch set updateAt
func (ele *Entry) Touch() {
	nowTime := time.Now()
	ele.UpdateAt = &nowTime
}
