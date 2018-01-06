# Corduroy
A simple HTTP peer-to-peer distributed hash table inspired by [Chord](https://en.wikipedia.org/wiki/Chord_(peer-to-peer).

## Overview
Corduroy started as a lightweight implementation of the Chord peer-to-peer hashtable that was written in Java. It has since been migrated to [Go](https://golang.org/) and some of the signature elements of Chord have been removed. Nodes in Corduroy still form a logical ring where elements are hashed to a particular node and duplicated in the prior N nodes for redudancy, but each node keeps a directory of all other known nodes and can reach the best node directly instead of relying on a finger table to forward queries. Corduroy can be embedded into another Go application or can be run as a standalone process and queried via HTTP.

## Building Corduroy
To build Corduroy natively:
```
make all
```

To build Corduroy in a Docker container:
```
make all-container
```

## Running Corduroy
To run Corduroy:
```
make run
```
To run without Make with a few fictious sample flags:
```
bin/corduroy -p 8081 -u http://localhost:8080
```
To run in a Docker container:
```
make run-container
```

## Embedding Corduroy
To embed Corduroy into an application:
```
port := 8081
path := "/"
store := NewMemoryStore()
registry := NewMemoryRegistry()
node := NewNode(port, path, store, registry)
node.Start()

seed := "http://localhost:8080"
node.Connect(seed)

node.Put("foo", "bar")
s := node.Get("foo")
```