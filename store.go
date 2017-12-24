package corduroy

type Store interface {
	Put(key string, value string)
	Get(key string) string
	GetKeys(first int, count int) []string
	Delete(key string)
	Contains(key string) bool
	Size() int
}
