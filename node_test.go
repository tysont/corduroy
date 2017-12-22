package corduroy

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func TestNodePutGetEntity(t *testing.T) {
	node := createTestNode()
	key := "foo"
	payload := "bar"
	entity := newTestObject(payload)
	err := node.putEntityRemote(node.Address, key, entity)
	assert.NoError(t, err)
	storedEntity := &testObject{}
	err = node.getEntityRemote(node.Address, key, storedEntity)
	assert.NoError(t, err)
	assert.Equal(t, payload, storedEntity.Payload)
}

func TestNodeRegisterSync(t *testing.T) {
	n1 := createTestNode()
	n2 := createTestNode()
	n3 := createTestNode()
	err := n2.registerNodeRemote(n3.Address)
	assert.NoError(t, err)
	err = n1.syncNodesRemote(n3.Address)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(n1.nodes))
}

func createTestNode() *Node {
	store := NewMemoryStore()
	port := getNextTestPort()
	node := NewNode(port, "/" + strconv.Itoa(port), store)
	node.Start(port)
	time.Sleep(time.Millisecond * 10)
	return node
}

func createTestCluster(size int) []*Node {
	cluster := make([]*Node, size)
	firstNode := createTestNode()
	cluster[0] = firstNode
	for i := 1; i < size; i++ {
		node := createTestNode()
		node.registerNodeRemote(firstNode.Address)
		cluster[i] = node
	}
	for _, node := range cluster {
		node.syncNodesRemote(firstNode.Address)
	}
	return cluster
}