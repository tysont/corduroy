package corduroy

import (
	"strings"
)

type Registry interface {
	Put(id int, address string)
	Get(id int) string
	GetIDs(start int, length int) []int
	GetRandomID() int
	GetAll() map[int]string
	Delete(id int)
	Contains(id int) bool
	Size() int
}

func RegistryFromShorthand(s string) Registry {
	if strings.EqualFold(strings.ToLower(s), "memory") {
		return NewMemoryRegistry()
	}
	return nil
}