package service

import (
	"log"
	"net/http"

	"recache/utils/logger"
)

/*
本文件用于 启动http缓存服务 和 启动 http api 服务
*/

func StartHTTPCacheServer(addr string, addrs []string, recache *Group) {
	peers := NewHTTPPool(addr)
	peers.UpdatePeers(addrs...)
	recache.RegisterServer(peers)
	logger.LogrusObj.Infof("service is running at %v", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// todo: gin 路由拆分请求负载
func StartHTTPAPIServer(apiAddr string, recache *Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := recache.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.Bytes())
		}))
	logger.LogrusObj.Infof("fontend server is running at %v", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}
