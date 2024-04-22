package policy

import (
	"container/list"
	"time"
	"unsafe"
)

// Cache 不是并发安全的缓存
type LRUCache struct {
	maxBytes int64      // 允许使用的最大内存
	nbytes   int64      // 已经使用的内存
	ll       *list.List // 实现 policy 的队列，双向循环链表
	cache    map[string]*list.Element
	// optional and executed when an entry is purged.
	// 回调函数，采用依赖注入的方式，该函数用于处理从缓存中淘汰的数据
	OnEvicted func(key string, value Value)
}

func (c *LRUCache) Get(key string) (value Value, updateAt *time.Time, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		kv.touch()
		return kv.value, kv.updateAt, true
	}
	return
}

// RemoveOldest 移除最近最少使用
func (c *LRUCache) RemoveOldest() {
	ele := c.ll.Front()
	if ele != nil {
		// 移除队头元素
		c.ll.Remove(ele)

		kv := ele.Value.(*entry)
		// 删除字典对应key
		delete(c.cache, kv.key)
		// 更新内存占用
		c.nbytes -= int64(kv.value.Len()) + int64(len(kv.key)) + int64(unsafe.Sizeof(kv.updateAt))
		// 对移除kv执行回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *LRUCache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// 已有键值对，更新值
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*entry)
		kv.touch()
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		// 添加键值对
		kv := &entry{
			key:   key,
			value: value,
		}
		kv.touch()
		ele := c.ll.PushBack(kv)
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(kv.value.Len()) + int64(unsafe.Sizeof(kv.updateAt))

	}
	// 超过最大内存占用，删除队头元素
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// 测试用 返回存储的kv数量
func (c *LRUCache) Len() int {
	return c.ll.Len()
}

func (c *LRUCache) CleanUp(ttl time.Duration) {
	for ele := c.ll.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(*entry).expired(ttl) {
			kv := c.ll.Remove(ele).(*entry)
			delete(c.cache, kv.key)
			c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) + int64(unsafe.Sizeof(kv.updateAt))
			if c.OnEvicted != nil {
				c.OnEvicted(kv.key, kv.value)
			}
		} else {
			break
		}
	}
}
