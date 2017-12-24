package corduroy

import "sync"

type MemoryRegistry struct {
	nodes    map[int]string
	index    []int
	reverseIndex map[int]int
	indexMux sync.Mutex
}

func NewMemoryRegistry() *MemoryRegistry{
	return &MemoryRegistry{
		nodes: make(map[int]string),
		index: make([]int, 0),
		reverseIndex: make(map[int]int),
	}
}

func (mr *MemoryRegistry) Put(id int, address string) {
	mr.indexMux.Lock()
	mr.nodes[id] = address
	mr.reverseIndex[id] = len(mr.index)
	mr.index = append(mr.index, id)
	mr.indexMux.Unlock()
}

func (mr *MemoryRegistry) Get(id int) string {
	return mr.nodes[id]
}

func (mr *MemoryRegistry) GetIDs(first int, length int) []int {
	f := first
	if first < 0 {
		f = 0
	}

	l := first + length
	s := mr.Size()
	mr.indexMux.Lock()
	if l > s {
		l = s
	}

	keys := mr.index[f:l]
	mr.indexMux.Unlock()
	return keys
}

func (mr *MemoryRegistry) GetAll() map[int]string {
	return mr.nodes
}

func (mr *MemoryRegistry) Delete(id int) {
	mr.indexMux.Lock()
	delete(mr.nodes, id)
	n := mr.reverseIndex[id]
	mr.index = append(mr.index[:n], mr.index[n+1:]...)
	delete(mr.reverseIndex, id)
	mr.indexMux.Unlock()
}

func (mr *MemoryRegistry) Contains(id int) bool {
	if _, found := mr.nodes[id]; found {
		return true
	}
	return false
}

func (mr *MemoryRegistry) Size() int {
	return len(mr.index)
}