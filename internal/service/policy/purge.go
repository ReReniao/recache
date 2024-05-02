package policy

import (
	"recache/internal/service/policy/fifo"
	"recache/internal/service/policy/interfaces"
	"recache/internal/service/policy/lfu"
	"recache/internal/service/policy/lru"
	log "recache/utils/logger"
	"strings"
)

// New 构建缓存实例
func New(name string, maxBytes int64, onEvicted func(string, interfaces.Value)) interfaces.CacheStrategy {
	name = strings.ToLower(name)
	log.LogrusObj.Infof("select policy %s", name)
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
