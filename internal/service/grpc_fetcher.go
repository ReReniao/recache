package service

import (
	"context"
	pb "recache/api/recachepb"
	"recache/internal/middleware/etcd/discovery/discovery3"
	"recache/utils/logger"

	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// 测试 Client 是否实现了 Fetcher 接口
var _ Fetcher = (*Client)(nil)

// The client module implements groupcache's ability to access other remote nodes to fetch caches.
type Client struct {
	serviceName string // 服务名称 recache/ip:addr
}

// Fetch gets the corresponding cache value from remote peer
func (c *Client) Fetch(group string, key string) ([]byte, error) {
	cli, err := clientv3.NewFromURL("http://localhost:2379")
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	// Discover services and obtain connection to services
	start := time.Now()
	conn, err := discovery3.Discovery(cli, c.serviceName)
	logger.LogrusObj.Warnf("本次 grpc dial 的耗时为: %v ms", time.Since(start).Milliseconds())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	grpcClient := pb.NewReCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	// 使用带有超时自动取消的上下文和指定请求调用客户端的 Get 方法发起 rpc 请求调用
	start = time.Now()
	resp, err := grpcClient.Get(ctx, &pb.GetRequest{
		Group: group,
		Key:   key,
	})
	logger.LogrusObj.Warnf("本次 grpc Call 的耗时为: %v ms", time.Since(start).Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("could not get %s/%s from peer %s", group, key, c.serviceName)
	}

	return resp.Value, nil
}

func NewClient(service string) *Client {
	return &Client{serviceName: service}
}
