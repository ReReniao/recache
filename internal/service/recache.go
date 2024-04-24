package service

import (
	"fmt"
	log "recache/internal/middleware/logger"
	"recache/internal/service/singleflight"
	"sync"
)

const (
	apiServer  = "http://127.0.0.1:9999"
	bindServer = "http://127.0.0.1:8001"
)

type Group struct {
	name      string
	retriever Retriever // 用于获取数据库数据
	mainCache *cache    // 缓存实例接口
	locator   Picker    // 缓存服务接口
	// 使用 singleflight.Group 确保每个key只被请求一次
	flight *singleflight.SingleFlight
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// 新 Group 的 Picker 没有注册，因此您可以为其指定 peer Picker
func (g *Group) RegisterPickerForGroup(p Picker) {
	if g.locator != nil {
		panic("group has been registered node locator")
	}
	g.locator = p
}

// NewGroup 创建一个新的Group实例
func NewGroup(name string, strategy string, cacheBytes int64, retriever Retriever) *Group {
	if retriever == nil {
		panic("Group fetcher Must be existed!")
	}
	g := &Group{
		name:      name,
		retriever: retriever,
		mainCache: newCache(strategy, cacheBytes),
		flight:    &singleflight.SingleFlight{},
	}
	if _, ok := groups[name]; ok {
		return groups[name]
	}
	mu.Lock()
	groups[name] = g
	log.Logger.Infof(groups[name].name)
	mu.Unlock()
	return g
}

// GetGroup 返回之间创建的命名组；如果返回nil，则表示没有该group
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]

}

// TODO
func DestroyGroup(name string) {
	mu.Lock()
	defer mu.Unlock()
	g := groups[name]
	if g != nil {
		delete(groups, name)
	}
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key must be required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Logger.Infof("[GeeCaChe hit]")
		return v, nil
	}
	return g.load(key)
}

func (g *Group) LoadLocally(key string) (ByteView, error) {
	// 从DB源获取值
	bytes, err := g.retriever.retrieve(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: bytes}
	// 更新缓存
	g.populateCache(key, value)
	return value, nil
}

// populateCaChe 将某数据写入缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.set(key, value)
}

// RegisterPeers 为 Group 注册 server
func (g *Group) RegisterPeers(locator Picker) {
	if g.locator != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.locator = locator
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 本地和远程的相同key请求都只请求一次
	view, err := g.flight.Do(key, func() (interface{}, error) {
		if g.locator != nil {
			if fetcher, ok := g.locator.Pick(key); ok {
				if value, err = g.getFromPeer(fetcher, key); err == nil {
					return value, nil
				}
				log.Logger.Infof("[GeeCache] Failed to get from peer %s", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}
	return ByteView{}, err
}

func (g *Group) getLocally(key string) (ByteView, error) {
	return g.LoadLocally(key)
}

// 向其他结点请求
func (g *Group) getFromPeer(peer Fetcher, key string) (ByteView, error) {

	bytes, err := peer.Fetch(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
