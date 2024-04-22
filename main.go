package main

import (
	"ReniaoCache/conf"
	"ReniaoCache/db"
	log "ReniaoCache/logger"
	"ReniaoCache/reniaocache"
	"flag"
	"fmt"
	"net/http"
)

func createGroup() *reniaocache.Group {
	return reniaocache.NewGroup("scores", 2<<10, reniaocache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Logger.Infof("[SlowDB] search key %s", key)
			// 从模拟数据库获取数据
			var student db.Student
			err := db.DB.Model(&student).Where("name = ?", key).First(&student).Error
			if err == nil {
				return []byte(student.Score), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, gee *reniaocache.Group) {
	peers := reniaocache.NewHTTPPool(addr)
	peers.Set(addrs...)
	gee.RegisterPeers(peers)
	log.Logger.Infof("reniaocache is running at %s", addr)
	log.Logger.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gee *reniaocache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))
	log.Logger.Infof("frontend server is running at %s", apiAddr)
	log.Logger.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {

	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "reniaocache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()
	log.Init()
	conf.Init()
	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()
	//如果 api 为 true;启动 api 服务
	if api {
		go startAPIServer(apiAddr, gee)
	}
	// 根据 port 启动结点
	startCacheServer(addrMap[port], []string(addrs), gee)
}
