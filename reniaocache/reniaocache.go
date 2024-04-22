package reniaocache

import (
	pb "ReniaoCache/reniaocache/reniaocachepb"
	"ReniaoCache/reniaocache/singleflight"
	"fmt"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter // 用于获取数据库数据
	mainCache *cache
	peers     PeerPicker
	// 使用 singleflight.Group 确保每个key只被请求一次
	loader *singleflight.SingleFlight
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 创建一个新的Group实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("Group Getter Must be existed!")
	}

	group := GetGroup(name)
	if group != nil {
		return group
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: newCache(cacheBytes),
		loader:    &singleflight.SingleFlight{},
	}
	groups[name] = g
	return g
}

// GetGroup 返回之间创建的命名组；如果返回nil，则表示没有该group
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		//log.Logger.Infof("[GeeCaChe hit]")
		return v, nil
	}
	return g.load(key)
}

func (g *Group) LoadLocally(key string) (ByteView, error) {
	// 从DB源获取值
	bytes, err := g.getter.Get(key)
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

// RegisterPeers 注册一个结点选择器以选择远程结点
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 本地和远程的相同key请求都只请求一次
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				//log.Logger.Infof("[GeeCache] Failed to get from peer %s", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return ByteView{}, err
}

func (g *Group) getLocally(key string) (ByteView, error) {
	return g.LoadLocally(key)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}

	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}
