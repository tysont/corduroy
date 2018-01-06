package main

import (
	"github.com/tysont/corduroy/core"
	"time"
	"github.com/jessevdk/go-flags"
	"os"
	"log"
)

type Options struct {
	Port int `short:"p" long:"port" description:"Port to listen on"`
	Path string `short:"a" long:"path" description:"Path to host endpoints"`
	RemoteUri string `short:"u" long:"remote" description:"Remote uri of a seed node"`
	StoreType string `short:"s" long:"store" description:"Type of store to hold data"`
	RegistryType string `short:"r" long:"registry" description:"Type of registry to track nodes"`
}

func NewOptions() *Options {
	return &Options {
		Port: 8080,
		Path: "/",
		RemoteUri: "",
		StoreType: "memory",
		RegistryType: "memory",
	}
}

func main() {
	options := NewOptions()
	_, err := flags.Parse(options)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	store := corduroy.StoreFromShorthand(options.StoreType)
	registry := corduroy.RegistryFromShorthand(options.RegistryType)
	node := corduroy.NewNode(options.Port, options.Path, store, registry)
	node.Start()
	if options.RemoteUri != "" {
		node.Connect(options.RemoteUri)
	}

	for {
		time.Sleep(time.Second)
	}
}