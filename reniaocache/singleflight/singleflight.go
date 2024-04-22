package singleflight

import (
	"sync"
)

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex // 保护 m
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 加锁，避免 g.m 被并发读写
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		// 已有相同请求，释放互斥锁
		g.mu.Unlock()
		// 已有相同请求，等待请求完成
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	// 请求已经创建，释放锁
	g.mu.Unlock()

	c.val, c.err = fn()
	// 请求完成，释放锁
	c.wg.Done()

	// 获取锁，因为可能有其他请求正在获取相同结果
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
