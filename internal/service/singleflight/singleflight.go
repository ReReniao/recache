package singleflight

import (
	"recache/utils/logger"
	"sync"
	"time"
)

type Call struct {
	wg    sync.WaitGroup
	value interface{}
	err   error
}

type SingleFlight struct {
	mu     sync.RWMutex // 保护 m
	cache  map[string]*cacheValue
	m      map[string]*Call
	ttl    time.Duration
	ticker *time.Ticker
}

type cacheValue struct {
	value   interface{}
	expires time.Time
}

func NewSingleFlight(ttl time.Duration) *SingleFlight {
	sf := &SingleFlight{
		cache: make(map[string]*cacheValue),
		m:     make(map[string]*Call),
		ttl:   ttl,
	}
	// 设置定期清除时间
	sf.ticker = time.NewTicker(ttl)
	// 确定定期清除 goroutine
	go sf.cacheCleaner()
	return sf
}

func (sf *SingleFlight) cacheCleaner() {
	// 监听心跳 并启动清除逻辑
	for range sf.ticker.C {
		sf.mu.Lock()
		for Key, cv := range sf.cache {
			if time.Now().After(cv.expires) {
				delete(sf.cache, Key)
			}
		}
		sf.mu.Unlock()
	}
}

func (sf *SingleFlight) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 加锁，并发安全
	sf.mu.RLock()
	// 检查缓存中是否有 有效缓存
	if cv, ok := sf.cache[key]; ok && time.Now().Before(cv.expires) {
		// 有缓存则释放锁
		sf.mu.RUnlock()
		return cv.value, nil
	}
	c, ok := sf.m[key]
	sf.mu.RUnlock()

	// 判断是否已经有相同请求正在查询
	if ok {
		logger.LogrusObj.Warnf("%s 已经在查询了，阻塞等待 goroutine 返回结果", key)
		c.wg.Wait()
		return c.value, c.err
	}

	c = new(Call)
	c.wg.Add(1)

	sf.mu.Lock()
	sf.m[key] = c
	// 请求创建完成，释放锁
	sf.mu.Unlock()

	c.value, c.err = fn()
	// 请求完成，释放锁
	c.wg.Done()

	// 获取锁，因为可能有其他请求正在获取相同结果
	sf.mu.Lock()
	delete(sf.m, key)
	// 缓存结果
	if c.err != nil {
		sf.cache[key] = &cacheValue{
			value:   c.value,
			expires: time.Now().Add(sf.ttl),
		}
	}
	sf.mu.Unlock()

	return c.value, c.err
}
