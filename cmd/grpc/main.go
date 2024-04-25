package main

import (
	"flag"
	"fmt"
	"recache/conf"
	etcdservice "recache/internal/middleware/etcd"
	"recache/internal/middleware/logger"
	grpcserver "recache/internal/server/grpc"
	"recache/internal/service"
)

var (
	port        = flag.Int("port", 9999, "service node port")
	serviceName = "recache"
)

func main() {
	conf.Init()
	flag.Parse()
	logger.Init()
	recache := service.NewGroupInstance("scores")
	// grpc node local service address
	serviceAddr := fmt.Sprintf("localhost:%d", *port)

	// get a grpc service instance
	ch := make(chan bool)
	svr, err := grpcserver.NewServer(ch, serviceAddr)
	if err != nil {
		fmt.Printf("acquire grpc server instance failed, %v\n", err)
		return
	}

	go etcdservice.DynamicServices(ch, serviceName)
	// Server implemented Pick interface, register a node selector for ggcache
	svr.UpdatePeers(etcdservice.ListServicePeers(serviceName))
	recache.RegisterPickerForGroup(svr)

	// start grpc service
	svr.Start()
}
