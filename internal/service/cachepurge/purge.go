package cachepurge

import (
	log "recache/internal/middleware/logger"
	"recache/internal/service/cachepurge/fifo"
	"recache/internal/service/cachepurge/interfaces"
	"recache/internal/service/cachepurge/lfu"
	"recache/internal/service/cachepurge/lru"
	"strings"
)

// New 构建缓存实例
func New(name string, maxBytes int64, onEvicted func(string, interfaces.Value)) interfaces.CacheStrategy {
	name = strings.ToLower(name)
	log.Logger.Infof("select policy %s", name)
	switch name {
	case "lru":
		return lru.NewLRUCache(maxBytes, onEvicted)
	case "fifo":
		return fifo.NewFIFOCache(maxBytes, onEvicted)
	case "lfu":
		return lfu.NewLFUCache(maxBytes, onEvicted)
	default:
		return lfu.NewLFUCache(maxBytes, onEvicted)
	}
}
