package corduroy

import (
	"sync"
)

type MemoryStore struct {
	values       map[string]string
	index        []string
	reverseIndex map[string]int
	indexMux     sync.Mutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		values:       make(map[string]string),
		index:        make([]string, 0),
		reverseIndex: make(map[string]int),
	}
}

func (ms *MemoryStore) Put(key string, value string) {
	ms.indexMux.Lock()
	ms.values[key] = value
	ms.reverseIndex[key] = len(ms.index)
	ms.index = append(ms.index, key)
	ms.indexMux.Unlock()
}

func (ms *MemoryStore) Get(key string) string {
	return ms.values[key]
}

func (ms *MemoryStore) GetKeys(first int, length int) []string {
	f := first
	if first < 0 {
		f = 0
	}

	l := first + length
	s := ms.Size()
	ms.indexMux.Lock()
	if l > s {
		l = s
	}

	keys := ms.index[f:l]
	ms.indexMux.Unlock()
	return keys
}

func (ms *MemoryStore) Contains(key string) bool {
	if _, found := ms.values[key]; found {
		return true
	}
	return false
}

func (ms *MemoryStore) Delete(key string) {
	ms.indexMux.Lock()
	delete(ms.values, key)
	n := ms.reverseIndex[key]
	ms.index = append(ms.index[:n], ms.index[n+1:]...)
	delete(ms.reverseIndex, key)
	ms.indexMux.Unlock()
}

func (ms *MemoryStore) Size() int {
	return len(ms.index)
}
