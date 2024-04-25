package lfu

import (
	"container/heap"
	"recache/internal/service/cachepurge/interfaces"
	"time"
	"unsafe"
)

type LfuCache struct {
	nbytes    int64
	maxBytes  int64
	cache     map[string]*lfuEntry
	pq        *priorityqueue
	OnEvicted func(key string, value interfaces.Value)
}

func (p *LfuCache) Get(key string) (Value interfaces.Value, updateAt *time.Time, ok bool) {
	if e, ok := p.cache[key]; ok {
		e.referenced()
		heap.Fix(p.pq, e.index)
		return e.entry.Value, e.entry.UpdateAt, ok
	}
	return
}

func (p *LfuCache) Put(key string, value interfaces.Value) {
	if e, ok := p.cache[key]; ok {
		p.nbytes += int64(value.Len()) - int64(e.entry.Value.Len())
		e.entry.Value = value
		e.referenced()
		heap.Fix(p.pq, e.index)
	} else {
		e := &lfuEntry{entry: interfaces.Entry{Key: key, Value: value}}
		e.referenced()
		heap.Push(p.pq, e)
		p.cache[key] = e
		p.nbytes += int64(len(e.entry.Key)) + int64(e.entry.Value.Len()) + int64(unsafe.Sizeof(e.entry.UpdateAt)) + int64(unsafe.Sizeof(e.count)) + int64(unsafe.Sizeof(e.index))

	}
	for p.maxBytes != 0 && p.maxBytes < p.nbytes {
		p.Remove()
	}
}

func (p *LfuCache) Remove() {
	e := heap.Pop(p.pq).(*lfuEntry)
	delete(p.cache, e.entry.Key)
	p.nbytes -= int64(len(e.entry.Key)) + int64(e.entry.Value.Len()) + int64(unsafe.Sizeof(e.entry.UpdateAt)) + int64(unsafe.Sizeof(e.count)) + int64(unsafe.Sizeof(e.index))
	if p.OnEvicted != nil {
		p.OnEvicted(e.entry.Key, e.entry.Value)
	}
}

func (p *LfuCache) Len() int {
	return p.pq.Len()
}

func (p *LfuCache) CleanUp(ttl time.Duration) {
	for _, e := range *p.pq {
		if e.entry.Expired(ttl) {
			kv := heap.Remove(p.pq, e.index).(*lfuEntry).entry
			delete(p.cache, kv.Key)
			p.nbytes -= int64(len(kv.Key)) + int64(kv.Value.Len()) + int64(unsafe.Sizeof(e.entry.UpdateAt)) + int64(unsafe.Sizeof(e.count)) + int64(unsafe.Sizeof(e.index))
			if p.OnEvicted != nil {
				p.OnEvicted(kv.Key, kv.Value)
			}
		}
	}
}

func NewLFUCache(maxBytes int64, onEvicted func(string, interfaces.Value)) *LfuCache {
	queue := priorityqueue(make([]*lfuEntry, 0))
	return &LfuCache{
		maxBytes:  maxBytes,
		cache:     make(map[string]*lfuEntry),
		pq:        &queue,
		OnEvicted: onEvicted,
	}
}
