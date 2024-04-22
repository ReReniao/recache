package reniaocache

import (
	"ReniaoCache/reniaocache/consistenthash"
	pb "ReniaoCache/reniaocache/reniaocachepb"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// HTTPPool HTTP池的结点定义
// HTTP 服务端
type HTTPPool struct {
	self       string // 结点自身地址
	basePath   string
	mu         sync.Mutex
	peers      *consistenthash.Map
	httpGetter map[string]*HTTPGetter
}

// HTTPGetter 客户端
type HTTPGetter struct {
	baseURL string
}

// 验证是否实现 PeerGetter 接口
var _ PeerGetter = (*HTTPGetter)(nil)

func (h *HTTPGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%s%s/%s",
		h.baseURL,
		url.QueryEscape(in.Group),
		url.QueryEscape(in.Key),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %s", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading respnse body: %v", err)
	}
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

// NewHTTPPool 初始化一个HTTP的结点
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log 打印 服务名称 和 信息
func (p *HTTPPool) Log(format string, v ...interface{}) {
	//log.Logger.Infof("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 实现HTTP Handler接口
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	// 正确的请求api /<basepath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
	}
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 将值解析成 protobuf 的响应
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// Set 更新 HTTPPool 的结点列表
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetter = make(map[string]*HTTPGetter, len(peers))

	for _, peer := range peers {
		p.httpGetter[peer] = &HTTPGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer 根据 key 找到对应的结点
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetter[peer], true
	}
	return nil, false
}

// 验证是否实现 PeerPicker 接口
var _ PeerPicker = (*HTTPPool)(nil)
