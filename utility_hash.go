package corduroy

import (
	"crypto/sha1"
	"encoding/binary"
	"strconv"
	"log"
	"fmt"
)

func hash(s string) int {
	b := sha1.Sum([]byte(s))
	u := binary.LittleEndian.Uint32(b[:4])
	t := fmt.Sprint(u)
	n, err := strconv.Atoi(t)
	if err != nil {
		log.Fatal(err)
		return -1
	}
	return n
}