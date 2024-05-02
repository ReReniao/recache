package service

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"recache/internal/service/singleflight"
	"recache/utils/logger"
	"sync"
	"time"
)

var (
	mu           sync.RWMutex
	GroupManager = make(map[string]*Group)
)

type Group struct {
	name      string
	retriever Retriever // 用于获取数据库数据
	mainCache *cache    // 缓存实例接口
	server    Picker    // 缓存服务接口
	// 使用 singleflight.Group 确保每个key只被请求一次
	flight *singleflight.SingleFlight
}

// 新 Group 的 Picker 没有注册，因此您可以为其指定 peer Picker
func (g *Group) RegisterServer(p Picker) {
	if g.server != nil {
		panic("group has been registered node locator")
	}
	g.server = p
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
		flight:    singleflight.NewSingleFlight(10 * time.Second),
	}
	if _, ok := GroupManager[name]; ok {
		return GroupManager[name]
	}

	mu.Lock()
	GroupManager[name] = g
	logger.LogrusObj.Infof(GroupManager[name].name)
	mu.Unlock()
	return g
}

// GetGroup 返回之间创建的命名组；如果返回nil，则表示没有该group
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return GroupManager[name]

}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key must be required")
	}
	if v, ok := g.mainCache.get(key); ok {
		logger.LogrusObj.Infof("[ReCaChe hit]")
		return v, nil
	}
	return g.load(key)
}

// RegisterPeers 为 Group 注册 server
func (g *Group) RegisterPeers(server Picker) {
	if g.server != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.server = server
}

func (g *Group) load(key string) (ByteView, error) {
	// 本地和远程的相同key请求都只请求一次
	view, err := g.flight.Do(key, func() (interface{}, error) {
		if g.server != nil {
			if fetcher, ok := g.server.Pick(key); ok {
				bytes, err := fetcher.Fetch(g.name, key)
				if err == nil {
					return ByteView{b: cloneBytes(bytes)}, nil
				}
				logger.LogrusObj.Infof("[GeeCache] Failed to get from peer %s", err.Error())
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
	// 从DB源获取值
	bytes, err := g.retriever.retrieve(key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.LogrusObj.Warnf("对于不存在的 key, 为了防止缓存穿透, 先存入缓存中并设置合理过期时间")
			g.mainCache.put(key, ByteView{})
		}
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	// 更新缓存
	g.populateCache(key, value)
	return value, nil
}

// populateCaChe 将某数据写入缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.set(key, value)
}
