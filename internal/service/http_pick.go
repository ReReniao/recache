package service

import (
	"fmt"
	"net/http"
	"recache/internal/service/consistenthash"
	"recache/utils/logger"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_recache/"
	apiServerAddr   = "127.0.0.1:9999"
)

var _ Picker = (*HTTPPool)(nil)

type HTTPPool struct {
	address      string
	basePath     string
	mu           sync.Mutex              // 保护 peers 和 httpFetchers
	peers        *consistenthash.Map     // 用于选择结点
	httpFetchers map[string]*httpFetcher // key的格式如： "http://10.0.0.1:8080"
}

func NewHTTPPool(address string) *HTTPPool {
	return &HTTPPool{
		address:  address,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	logger.LogrusObj.Infof("[Server %s] %s", p.address, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	// print the requested method and requested resource path
	// 打印请求的方法和请求的路径
	p.Log("%s %s", r.Method, r.URL.Path)

	// prefix/group/key
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request format, expected prefix/group/key", http.StatusBadRequest)
		return
	}
	groupName, key := parts[0], parts[1]

	g := GetGroup(groupName)
	if g == nil {
		http.Error(w, "no such group"+groupName, http.StatusBadRequest)
		return
	}

	view, err := g.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	// write value's deep copy
	w.Write(view.Bytes())
}

// Pick 根据key返回承载缓存的真实结点
func (p *HTTPPool) Pick(key string) (Fetcher, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	peerAddress := p.peers.GetTruthNode(key)
	if peerAddress == p.address {
		// upper layer get the value of the key locally after receiving false
		return nil, false
	}

	logger.LogrusObj.Infof("[dispatcher peer %s] pick remote peer: %s", apiServerAddr, peerAddress)
	return p.httpFetchers[peerAddress], true
}

// 重建 hash ring
func (p *HTTPPool) UpdatePeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.NewMap(defaultReplicas, nil)
	p.peers.AddNode(peers...)
	p.httpFetchers = make(map[string]*httpFetcher, len(peers))

	for _, peer := range peers {
		p.httpFetchers[peer] = &httpFetcher{
			baseURL: peer + p.basePath, // 例如 "http://10.0.0.1:9999/_recache/"
		}
	}
}
