package lru

import (
	"container/list"
	"recache/internal/service/policy/interfaces"
	"time"
	"unsafe"
)

// Cache 不是并发安全的缓存
type LruCache struct {
	maxBytes int64      // 允许使用的最大内存
	nbytes   int64      // 已经使用的内存
	ll       *list.List // 实现 policy 的队列，双向循环链表
	cache    map[string]*list.Element
	// optional and executed when an entry is purged.
	// 回调函数，采用依赖注入的方式，该函数用于处理从缓存中淘汰的数据
	OnEvicted func(key string, value interfaces.Value)
}

func (c *LruCache) Get(key string) (value interfaces.Value, updateAt *time.Time, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*interfaces.Entry)
		kv.Touch()
		return kv.Value, kv.UpdateAt, true
	}
	return
}

// RemoveOldest 移除最近最少使用
func (c *LruCache) RemoveOldest() {
	ele := c.ll.Front()
	if ele != nil {
		// 移除队头元素
		c.ll.Remove(ele)

		kv := ele.Value.(*interfaces.Entry)
		// 删除字典对应key
		delete(c.cache, kv.Key)
		// 更新内存占用
		c.nbytes -= int64(kv.Value.Len()) + int64(len(kv.Key)) + int64(unsafe.Sizeof(kv.UpdateAt))
		// 对移除kv执行回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.Key, kv.Value)
		}
	}
}

func (c *LruCache) Put(key string, value interfaces.Value) {
	if ele, ok := c.cache[key]; ok {
		// 已有键值对，更新值
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*interfaces.Entry)
		kv.Touch()
		c.nbytes += int64(value.Len()) - int64(kv.Value.Len())
		kv.Value = value
	} else {
		// 添加键值对
		kv := &interfaces.Entry{
			Key:   key,
			Value: value,
		}
		kv.Touch()
		ele := c.ll.PushBack(kv)
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(kv.Value.Len()) + int64(unsafe.Sizeof(kv.UpdateAt))

	}
	// 超过最大内存占用，删除队头元素
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// 测试用 返回存储的kv数量
func (c *LruCache) Len() int {
	return c.ll.Len()
}

func (c *LruCache) CleanUp(ttl time.Duration) {
	for ele := c.ll.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(*interfaces.Entry).Expired(ttl) {
			kv := c.ll.Remove(ele).(*interfaces.Entry)
			delete(c.cache, kv.Key)
			c.nbytes -= int64(len(kv.Key)) + int64(kv.Value.Len()) + int64(unsafe.Sizeof(kv.UpdateAt))
			if c.OnEvicted != nil {
				c.OnEvicted(kv.Key, kv.Value)
			}
		} else {
			break
		}
	}
}

func NewLRUCache(maxBytes int64, onEvicted func(string, interfaces.Value)) *LruCache {
	return &LruCache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}
