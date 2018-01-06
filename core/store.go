package corduroy

import (
	"strings"
)

type Store interface {
	Put(key string, value string)
	Get(key string) string
	GetRandomKey() string
	GetKeys(first int, count int) []string
	Delete(key string)
	Contains(key string) bool
	Size() int
}

func StoreFromShorthand(s string) Store {
	if strings.EqualFold(strings.ToLower(s), "memory") {
		return NewMemoryStore()
	}
	return nil
}
