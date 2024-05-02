package main

import (
	"flag"
	"fmt"
	"recache/config"
	etcdservice "recache/internal/middleware/etcd"
	grpcserver "recache/internal/server/grpc"
	"recache/internal/service"
)

var (
	port        = flag.Int("port", 9999, "service node port")
	serviceName = "recachepb"
)

func main() {
	config.InitConfig()
	flag.Parse()
	// grpc node local service address
	// grpc 结点本地服务地址
	serviceAddr := fmt.Sprintf("localhost:%d", *port)
	gm := service.NewGroupManager([]string{"scores", "website"}, serviceAddr)

	// get a grpc service instance
	// 获得一个 gprc 服务实例
	updateChan := make(chan bool)
	svr, err := grpcserver.NewServer(updateChan, serviceAddr)
	if err != nil {
		fmt.Printf("acquire grpc server instance failed, %v\n", err)
		return
	}

	go etcdservice.DynamicServices(updateChan, serviceName)

	// Server implemented Pick interface, register a node selector for recachepb
	svr.UpdatePeers(etcdservice.ListServicePeers(serviceName))
	gm["scores"].RegisterServer(svr)
	// start grpc service
	svr.Start()
}
