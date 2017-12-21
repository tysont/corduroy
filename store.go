package corduroy

type Store interface {
	Put(key string, value string)
	Get(key string) string
}