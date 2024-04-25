package fifo

import (
	"container/list"
	"recache/internal/service/cachepurge/interfaces"
	"time"
	"unsafe"
)

type FifoCache struct {
	maxBytes int64
	nbytes   int64
	ll       *list.List
	cache    map[string]*list.Element
	// optional and executed when an entry is purged.
	// 回调函数，采用依赖注入的方式，该函数用于处理从缓存中淘汰的数据
	OnEvicted func(key string, value interfaces.Value)
}

func (f *FifoCache) Get(key string) (value interfaces.Value, updateAt *time.Time, ok bool) {
	if ele, ok := f.cache[key]; ok {
		e := ele.Value.(*interfaces.Entry)
		value = e.Value
		updateAt = e.UpdateAt
		return e.Value, e.UpdateAt, true
	}
	return
}

func (f *FifoCache) Put(key string, value interfaces.Value) {
	if ele, ok := f.cache[key]; ok {
		f.nbytes += int64(value.Len()) - int64(ele.Value.(*interfaces.Entry).Value.Len())
		e := ele.Value.(*interfaces.Entry)
		e.Touch()
		e.Value = value
	} else {
		kv := &interfaces.Entry{
			Key:   key,
			Value: value,
		}
		kv.Touch()
		ele := f.ll.PushBack(kv)
		f.cache[key] = ele
		f.nbytes += int64(len(kv.Key)) + int64(kv.Value.Len()) + int64(unsafe.Sizeof(kv.UpdateAt))
	}
	for f.maxBytes != 0 && f.maxBytes < f.nbytes {
		f.RemoveFront()
	}
}

func (f *FifoCache) Len() int {
	return f.ll.Len()
}

func (f *FifoCache) RemoveFront() {
	ele := f.ll.Front()
	if ele != nil {
		kv := f.ll.Remove(ele).(*interfaces.Entry)
		delete(f.cache, kv.Key)
		f.nbytes -= int64(len(kv.Key)) + int64(kv.Value.Len()) + int64(unsafe.Sizeof(kv.UpdateAt))
		if f.OnEvicted != nil {
			f.OnEvicted(kv.Key, kv.Value)
		}
	}
}

func (f *FifoCache) CleanUp(ttl time.Duration) {
	for e := f.ll.Front(); e != nil; e = e.Next() {
		if e.Value.(*interfaces.Entry).Expired(ttl) {
			kv := f.ll.Remove(e).(*interfaces.Entry)
			delete(f.cache, kv.Key)
			f.nbytes -= int64(len(kv.Key)) + int64(kv.Value.Len()) + int64(unsafe.Sizeof(kv.UpdateAt))
			if f.OnEvicted != nil {
				f.OnEvicted(kv.Key, kv.Value)
			}
		} else {
			break
		}
	}
}

func NewFIFOCache(maxBytes int64, onEvicted func(string, interfaces.Value)) *FifoCache {
	return &FifoCache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}
