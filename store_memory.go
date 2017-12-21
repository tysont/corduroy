package corduroy

type MemoryStore struct {
	values map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		values: make(map[string]string),
	}
}

func (ms *MemoryStore) Put(key string, value string) {
	ms.values[key] = value
}

func (ms *MemoryStore) Get(key string) string {
	return ms.values[key]
}
