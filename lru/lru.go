package lru

import "container/list"

// Cache 不是并发安全的缓存
type Cache struct {
	maxBytes int64      // 允许使用的最大内存
	nbytes   int64      // 已经使用的内存
	ll       *list.List // 实现 lru 的队列，双向循环链表
	cache    map[string]*list.Element
	// 可选，在删除kv时执行
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

// Value 的 Len() 方法用于返回它的字节数
type Value interface {
	Len() int
}

// New 构建缓存实例
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 移除最近最少使用
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		// 移除队头元素
		c.ll.Remove(ele)

		kv := ele.Value.(entry)
		// 删除字典对应key
		delete(c.cache, kv.key)
		// 更新内存占用
		c.nbytes -= int64(kv.value.Len()) + int64(len(kv.key))
		// 对移除kv执行回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// 已有键值对，更新值
		c.ll.MoveToBack(ele)
		kv := ele.Value.(entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		// 添加键值对
		kv := entry{
			key:   key,
			value: value,
		}
		ele := c.ll.PushFront(kv)
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(kv.value.Len())

	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// 测试用 返回存储的kv数量
func (c *Cache) Len() int {
	return c.ll.Len()
}
