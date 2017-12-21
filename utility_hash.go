package corduroy

import (
	"hash/fnv"
)

func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return (int)(h.Sum32())
}