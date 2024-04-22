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

// New 返回一个 Map 实例;
// replicas 虚拟结点倍数； fn 指定哈希算法
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}

	// fn 为 nil 采用默认 hash 算法
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加 key 到 hash环
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(key) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// 第一个 大于等于hash 的下标,没有则返回切片长度
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
