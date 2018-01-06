package main

import (
	"github.com/tysont/corduroy/core"
	"time"
)

func main() {
	port := 8080
	store := corduroy.NewMemoryStore()
	registry := corduroy.NewMemoryRegistry()
	node := corduroy.NewNode(port, "/", store, registry)
	node.Start(port)
	for {
		time.Sleep(time.Second)
	}
}