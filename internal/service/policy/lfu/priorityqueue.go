package lfu

import (
	"recache/internal/service/policy/interfaces"
)

type priorityqueue []*lfuEntry

type lfuEntry struct {
	index int
	entry interfaces.Entry
	count int
}

func (l *lfuEntry) referenced() {
	l.count++
	l.entry.Touch()
}

// Less compare ttl and referenced count
func (pq *priorityqueue) Less(i, j int) bool {
	if (*pq)[i].count == (*pq)[j].count {
		return (*pq)[i].entry.UpdateAt.Before(*(*pq)[j].entry.UpdateAt)
	}
	return (*pq)[i].count < (*pq)[j].count
}

func (pq *priorityqueue) Len() int {
	return len(*pq)
}

func (pq *priorityqueue) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
	(*pq)[i].index = i
	(*pq)[j].index = j
}

func (pq *priorityqueue) Pop() interface{} {
	oldpq := *pq
	n := len(oldpq)
	entry := oldpq[n-1]
	oldpq[n-1] = nil
	*pq = oldpq[:n-1]
	for i := 0; i < len(*pq); i++ {
		(*pq)[i].index = i
	}
	return entry
}

// Push 插入新元素
func (pq *priorityqueue) Push(x interface{}) {
	entry := x.(*lfuEntry)
	entry.index = len(*pq)
	*pq = append(*pq, x.(*lfuEntry))
}
