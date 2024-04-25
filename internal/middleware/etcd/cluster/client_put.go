package main

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"recache/internal/middleware/logger"
)

func main() {
	//初始化
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		fmt.Println("new clientv3 failed,err:", err)
		return
	}

	fmt.Println("connect to etcd success!")
	defer cli.Close()

	//put
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = cli.Put(ctx, "clusters/localhost:9999", "localhost:9999")
	if err != nil {
		logger.Logger.Error("put recache service to etcd failed")
		return
	}

	fmt.Println("put recache service to etcd success!")
}
