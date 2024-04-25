package http

import (
	"log"
	"net/http"
	"recache/internal/middleware/logger"
	"recache/internal/service"
)

func StartHTTPCacheServer(addr string, addrs []string, recache *service.Group) {
	peers := NewHTTPPool(addr)
	peers.UpdatePeers(addrs...)
	recache.RegisterPickerForGroup(peers)
	logger.Logger.Infof("service is running at %v", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func StartHTTPAPIServer(apiAddr string, recache *service.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logger.Logger.Infof("recache")
			key := r.URL.Query().Get("key")
			view, err := recache.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.Bytes())
		}))
	logger.Logger.Infof("frontend server is running at %v", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}
