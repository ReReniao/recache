package service

import (
	"recache/internal/service/policy"
	"recache/internal/service/policy/interfaces"
	"recache/utils/logger"
	"sync"
)

// cache 模块负责提供对lru模块的并发控制

type cache struct {
	mu         sync.Mutex // 在  interfaces.CacheStrategy 上层上锁
	strategy   interfaces.CacheStrategy
	cacheBytes int64
}

func newCache(strategy string, cacheBytes int64) *cache {
	onEvicted := func(key string, val interfaces.Value) {
		logger.LogrusObj.Infof("缓存条目 [%s:%s] 被淘汰", key, val)
	}

	return &cache{
		strategy:   policy.New(strategy, cacheBytes, onEvicted),
		cacheBytes: cacheBytes,
	}
}

// set 并发安全 by sync.Mutex
// set 和 put 指令在实际效果上都等于 strategy.Put 方法
func (c *cache) set(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.strategy.Put(key, value)
}

// get 并发安全的get方法 by sync.Mutex
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, _, ok := c.strategy.Get(key); ok {
		return v.(ByteView), true
	}

	return ByteView{}, false
}

// put 并发安全 by sync.Mutex
func (c *cache) put(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	logger.LogrusObj.Infof("存入数据库之后压入缓存, (key, value)=(%s, %s)", key, val)
	c.strategy.Put(key, val)
}
