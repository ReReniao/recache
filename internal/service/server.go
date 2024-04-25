package service

//
//import (
//	services "ReniaoCache/etcd"
//	log "ReniaoCache/internal/middleware/logger"
//	"ReniaoCache/reniaocache/consistenthash"
//	pb "ReniaoCache/reniaocache/reniaocachepb"
//	"ReniaoCache/utils"
//	"context"
//	"fmt"
//	clientv3 "go.etcd.io/etcd/client/v3"
//	"google.golang.org/grpc"
//	"net"
//	"strings"
//	"sync"
//	"time"
//)
//
//const (
//	defaultAddr     = "127.0.0.1:6324"
//	defaultReplicas = 50
//)
//
//var (
//	defaultEtcdConfig = clientv3.Config{
//		Endpoints:   []string{"localhost；2379"},
//		DialTimeout: 5 * time.Second,
//	}
//)
//
//var _ pb.ReGroupCacheServer = (*Server)(nil)
//
//type Server struct {
//	pb.UnimplementedReGroupCacheServer
//	Addr        string     // format ip:port
//	Status      bool       // true:running false:stop
//	stopsSignal chan error // 通知 registery revoke服务
//	mu          sync.Mutex // Group 和 Server 是解耦合的，需要自己控制并发
//	consHash    *consistenthash.Map
//	clients     map[string]*client
//}
//
//// NewServer 创建 cache 的 server
//func NewServer(addr string) (*Server, error) {
//	if addr == "" {
//		addr = defaultAddr
//	}
//	if !utils.ValidPerrAddr(addr) {
//		return nil, fmt.Errorf("invalid addr %s", addr)
//	}
//	return &Server{Addr: addr}, nil
//}
//
//func (s *Server) Get(ctx context.Context, req *pb.Request) (*pb.Response, error) {
//	group, key := req.GetGroup(), req.GetKey()
//	resp := &pb.Response{}
//	log.Logger.Infof("[groupcache server %s] Recv RPC Request - (%s)/(%s)", s.Addr, group, key)
//	if key == "" || group == "" {
//		return resp, fmt.Errorf("key and group name is reqiured")
//	}
//
//	g := GetGroup(group)
//	if g == nil {
//		return resp, fmt.Errorf("group %s not found", group)
//	}
//	view, err := g.Get(key)
//	if err != nil {
//		return resp, err
//	}
//	resp.Value = view.ByteSlice()
//	return resp, nil
//}
//
//// Start 启动 Cache 服务
//func (s *Server) Start() error {
//	s.mu.Lock()
//
//	if s.Status {
//		s.mu.Unlock()
//		return fmt.Errorf("server %s is already started", s.Addr)
//	}
//
//	s.Status = true
//	s.stopsSignal = make(chan error)
//
//	port := strings.Split(s.Addr, ":")[1]
//	lis, err := net.Listen("tcp", ":"+port)
//
//	if err != nil {
//		return fmt.Errorf("failed to listen %s, error: %v", s.Addr, err)
//	}
//
//	grpcServer := grpc.NewServer()
//	pb.RegisterReGroupCacheServer(grpcServer, s)
//
//	go func() {
//		// Register never return unless stop signal received (blocked)
//		err := services.Register("reniaocache", s.Addr, s.stopsSignal)
//		if err != nil {
//			log.Logger.Error(err.Error())
//		}
//		// close channel
//		close(s.stopsSignal)
//		// close tcp listen
//		err = lis.Close()
//		if err != nil {
//			log.Logger.Error(err.Error())
//		}
//		log.Logger.Infof("[%s] Revoke service and close tcp socket ok.", s.Addr)
//	}()
//
//	log.Logger.Infof("[%s] register service ok\n", s.Addr)
//	s.mu.Unlock()
//
//	if err = grpcServer.Serve(lis); s.Status && err != nil {
//		return fmt.Errorf("failed to serve %s, error: %v", s.Addr, err)
//	}
//	return nil
//}
//
//// SetPeers 将各个远端主机 IP 配置到 Server 里
//// 这样 Server 就可以 Pick 它们了
//// 注意：此操作是覆写操作，peersIP 必须满足 x.x.x.x:port 的格式
//func (s *Server) SetPeers(peersAddr []string) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	s.consHash = consistenthash.New(defaultReplicas, nil)
//	s.consHash.Add(peersAddr...)
//	s.clients = make(map[string]*client)
//
//	for _, peersAddr := range peersAddr {
//		if !utils.ValidPerrAddr(peersAddr) {
//			panic(fmt.Sprintf("[peer %s] invalid address format, it shoulb be x.x.x.x:port", peersAddr))
//		}
//		// reniaocache/localhost:8000
//		service := fmt.Sprintf("reniaocache/%s", peersAddr)
//		// client {name string}  (c *client) Fetch(key string) ([]byte, error)
//		s.clients[peersAddr] = NewClient(service)
//	}
//}
//
//func (s *Server) PickPeer(key string) (PeerGetter, bool) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	peerAddr := s.consHash.Get(key)
//
//	if peerAddr == s.Addr {
//		log.Logger.Infof("oohhh! pick myself, i am %s\n", s.Addr)
//		return nil, false
//	}
//	log.Logger.Infof("[cache %s] pick remote peer: %s\n", s.Addr, peerAddr)
//	return s.clients[peerAddr], true
//}
//
//func (s *Server) Stop() {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	if !s.Status {
//		return
//	}
//	// 发送停止 keepAlive 的信号，因为该节点要退出了，不需要再发送心跳探测了
//	s.stopsSignal <- nil
//	s.Status = false
//	s.clients = nil // 清空一致性哈希信息，帮助 GC 进行垃圾回收
//	s.consHash = nil
//}
//
//var _ PeerPicker = (*Server)(nil)
