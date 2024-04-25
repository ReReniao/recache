package service

//
//import (
//	service "ReniaoCache/etcd"
//	pb "ReniaoCache/reniaocache/reniaocachepb"
//	"context"
//	"fmt"
//	clientv3 "go.etcd.io/etcd/client/v3"
//	"time"
//)
//
//type client struct {
//	name string // 服务名称 reniaocache/ip:addr
//}
//
//// Fetch 从 remote peer 获取对应的缓存值
//func (c *client) Fetch(group string, key string) ([]byte, error) {
//	// 创建一个 etcd client
//	cli, err := clientv3.New(service.DefaultEtcdConfig)
//	if err != nil {
//		return nil, err
//	}
//	defer cli.Close()
//
//	// 发现服务，取得与服务的链接
//	conn, err := service.EtcdDial(cli, c.name)
//	if err != nil {
//		return nil, err
//	}
//	defer conn.Close()
//	grpcClient := pb.NewReGroupCacheClient(conn)
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	resp, err := grpcClient.Get(ctx, &pb.Request{
//		Group: group,
//		Key:   key,
//	})
//	if err != nil {
//		return nil, fmt.Errorf("could not get %s/%s from perr %s", group, key, c.name)
//	}
//	return resp.Value, nil
//}
//
//func NewClient(name string) *client {
//	return &client{name: name}
//}
//
//var _ PeerGetter = (*client)(nil)
