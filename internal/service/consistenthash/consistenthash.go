package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 计算数据的32位hash值
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int            // 虚拟结点倍数
	keys     []int          // 排序后的哈希环
	hashMap  map[int]string // 虚拟结点和真实结点的映射表
}

// NewMap 返回一个 Map 实例;
// replicas 虚拟结点倍数； fn 指定哈希算法
func NewMap(replicas int, fn Hash) *Map {
	// fn 为 nil 采用默认 hash 算法
	if fn == nil {
		fn = crc32.ChecksumIEEE
	}
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	return m
}

// Add 添加 key 到 hash环
func (m *Map) AddNode(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get 返回承载 key 对应缓存的真实结点
func (m *Map) GetTruthNode(key string) string {
	if len(key) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// 第一个大于等于 hash 的下标,没有则返回切片长度
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// 取模后得到真实结点的映射
	realKey := m.hashMap[m.keys[idx%len(m.keys)]]
	//log.Logger.Infof("计算出 key 的 hash: %d, 顺时针选择的节点下标 idx: %d", hash, idx)
	//log.Logger.Infof("选择的真实节点：%s", realKey)

	return realKey
}

func (m *Map) RemovePeer(peer string) {
	virtualKeys := []int{}
	for key, v := range m.hashMap {
		if peer == v {
			// 删除结点映射
			delete(m.hashMap, key)
			virtualKeys = append(virtualKeys, key)
		}
	}
	for i := 0; i < len(virtualKeys); i++ {
		for idx, value := range m.keys {
			if value == virtualKeys[i] {
				// 在哈希环移除对应结点
				m.keys = append(m.keys[:idx], m.keys[idx+1:]...)
			}
		}
	}
	//log.Logger.Infof("结点移除成功，缓存被顺位结点继承")
}
