package main

import (
	"ReniaoCache/conf"
	services "ReniaoCache/etcd"
	log "ReniaoCache/logger"
	group "ReniaoCache/reniaocache"
	"flag"
	"fmt"
)

var (
	port = flag.Int("port", 9999, "port")
)

func main() {
	log.Init()
	conf.Init()
	flag.Parse()
	// 新建 cache 实例
	g := group.CreateGroup("scores")
	for _, n := range group.Groups {
		fmt.Println(n)
	}
	// New 一个自己实现的服务实例
	addr := fmt.Sprintf("localhost:%d", *port)
	svr, err := group.NewServer(addr)
	if err != nil {
		log.Logger.Fatal(err)
	}

	// 设置同伴节点包括自己（同伴的地址从 etcd 中获取）
	addrs, err := services.GetPeers("clusters")
	if err != nil { // 如果查询失败使用默认的地址
		addrs = []string{"localhost:9999"}
	}

	fmt.Println("从 etcd 处获取的 server 地址", addrs)
	// 将节点打到哈希环上
	svr.SetPeers(addrs)
	// 为 Group 注册服务 Picker
	g.RegisterPeers(svr)
	log.Logger.Printf("reniaocache is running at %s", addr)

	// 启动服务（注册服务至 etcd、计算一致性 hash）
	err = svr.Start()
	if err != nil {
		log.Logger.Fatal(err)
	}
}
