package reniaocache

import (
	"ReniaoCache/conf"
	log "ReniaoCache/logger"
	"ReniaoCache/policy"
	"sync"
)

type cache struct {
	mu         sync.Mutex // 在 policy.Cache 上层上锁
	policy     policy.Cache
	cacheBytes int64
}

func newCache(cacheBytes int64) *cache {
	return &cache{
		cacheBytes: cacheBytes,
	}
}

// set 并发安全 by sync.Mutex
func (c *cache) set(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 延迟初始化
	if c.policy == nil {
		c.policy = policy.New(conf.Policy, c.cacheBytes, nil)
	}
	c.policy.Add(key, value)
}

// get 并发安全的get方法 by sync.Mutex
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.policy == nil {
		c.policy = policy.New(conf.Policy, c.cacheBytes, nil)
	}

	if v, _, ok := c.policy.Get(key); ok {
		return v.(ByteView), true
	}

	return ByteView{}, false
}

// put 并发安全 by sync.Mutex
func (c *cache) put(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.policy == nil { // 策略类模式
		c.policy = policy.New(conf.Policy, c.cacheBytes, nil)
	}
	log.Logger.Info("cache.put(key, val)")
	c.policy.Add(key, val)
}
