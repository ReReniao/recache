package service

import (
	"recache/conf"
	"recache/internal/service/cachepurge"
	"recache/internal/service/cachepurge/interfaces"
	"sync"
)

type cache struct {
	mu         sync.Mutex // 在  interfaces.CacheStrategy 上层上锁
	strategy   interfaces.CacheStrategy
	cacheBytes int64
}

func newCache(strategy string, cacheBytes int64) *cache {
	// 默认使用 lru 作为缓存淘汰策略
	if strategy == "" {
		strategy = "lru"
	}
	// 默认缓存最大体积 2<<10 bytes
	if cacheBytes == 0 {
		cacheBytes = 2 << 10
	}
	return &cache{
		strategy:   cachepurge.New(strategy, cacheBytes, nil),
		cacheBytes: cacheBytes,
	}
}

// set 并发安全 by sync.Mutex
// set 和 put 指令在实际效果上都等于 strategy.Put 方法
func (c *cache) set(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 延迟初始化
	if c.strategy == nil {
		c.strategy = cachepurge.New(conf.Policy, c.cacheBytes, nil)
	}
	c.strategy.Put(key, value)
}

// get 并发安全的get方法 by sync.Mutex
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.strategy == nil {
		c.strategy = cachepurge.New(conf.Policy, c.cacheBytes, nil)
	}

	if v, _, ok := c.strategy.Get(key); ok {
		return v.(ByteView), true
	}

	return ByteView{}, false
}

// put 并发安全 by sync.Mutex
func (c *cache) put(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.strategy == nil { // 策略类模式
		c.strategy = cachepurge.New(conf.Policy, c.cacheBytes, nil)
	}
	//log.Logger.Info("cache.put(key, val)")
	c.strategy.Put(key, val)
}
