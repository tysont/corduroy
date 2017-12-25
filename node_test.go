package corduroy

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"testing"
)

func TestPing(t *testing.T) {
	node := createTestNode()
	uri := node.Address + pingPath
	response, err := http.Get(uri)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestNodePutGetEntity(t *testing.T) {
	node := createTestNode()
	key := "foo"
	payload := "bar"
	entity := newTestObject(payload)
	b, err := json.Marshal(entity)
	_, _, err = node.putValueRemote(node.Address, key, string(b), []int{node.ID}, redundantCopies)
	assert.NoError(t, err)
	_, body, err := node.getValueRemote(node.Address, key, []int{node.ID}, redundantCopies)
	storedEntity := &testObject{}
	err = json.Unmarshal([]byte(body), storedEntity)
	assert.NoError(t, err)
	assert.Equal(t, payload, storedEntity.Payload)
}

func TestNodeGetNotFound(t *testing.T) {
	node := createTestNode()
	key := "foo"
	statusCode, body, err := node.getValueRemote(node.Address, key, []int{node.ID}, redundantCopies)
	assert.NoError(t, err)
	assert.Equal(t, "", body)
	assert.Equal(t, http.StatusNotFound, statusCode)
}

func TestNodeRegisterSync(t *testing.T) {
	n1 := createTestNode()
	n2 := createTestNode()
	n3 := createTestNode()
	err := n2.registerNodeRemote(n3.Address)
	assert.NoError(t, err)
	err = n1.syncNodeRegistryRemote(n3.Address)
	assert.NoError(t, err)
	assert.Equal(t, 3, n1.registry.Size())
}

func TestClusterPutGetEntity(t *testing.T) {
	cluster := createTestCluster(20)
	key := "foo"
	payload := "bar"
	entity := newTestObject(payload)
	b, err := json.Marshal(entity)
	_, _, err = cluster[0].putValueRemote(cluster[1].Address, key, string(b), []int{cluster[0].ID}, redundantCopies)
	assert.NoError(t, err)
	_, body, err := cluster[3].getValueRemote(cluster[4].Address, key, []int{cluster[3].ID}, redundantCopies)
	storedEntity := &testObject{}
	err = json.Unmarshal([]byte(body), storedEntity)
	assert.NoError(t, err)
	assert.Equal(t, payload, storedEntity.Payload)
}

func TestClusterDetectStoppedNode(t *testing.T) {
	cluster := createTestCluster(3)
	assert.True(t, cluster[0].registry.Contains(cluster[1].ID))
	cluster[1].Stop()
	cluster[0].syncNodeRemote(cluster[1].ID)
	assert.False(t, cluster[0].registry.Contains(cluster[1].ID))
}

func createTestNode() *Node {
	port := getNextTestPort()
	store := NewMemoryStore()
	registry := NewMemoryRegistry()
	node := NewNode(port, "/" + strconv.Itoa(port), store, registry)
	node.Start(port)
	node.waitStart()
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
		node.syncNodeRegistryRemote(firstNode.Address)
	}
	return cluster
}
