package corduroy

import (
	"testing"
	"net/http"
	"encoding/json"
	"bytes"
	"time"
	"github.com/stretchr/testify/assert"
	"strconv"
	"net/url"
)

func TestNodeGetEntity(t *testing.T) {
	node := createTestNode()
	key := "foo"
	payload := "bar"
	entity := newTestObject("bar")
	bytes, err := json.Marshal(entity)
	assert.NoError(t, err)
	value := string(bytes)
	err = tryPutEntity(node, key, value)
	node.store.Put("foo", value)

	storedEntity := &testObject{}
	err = tryGetEntity(node, key, storedEntity)
	assert.NoError(t, err)
	assert.Equal(t, payload, storedEntity.Payload)
}

func TestNodePutEntity(t *testing.T) {
	node := createTestNode()
	key := "foo"
	payload := "bar"
	entity := newTestObject(payload)
	err := tryPutEntity(node, key, entity)
	assert.NoError(t, err)
	storedValue := node.store.Get(key)
	storedBytes := []byte(storedValue)
	storedEntity := &testObject{}
	err = json.Unmarshal(storedBytes, storedEntity)
	assert.NoError(t, err)
	assert.Equal(t, payload, storedEntity.Payload)
}

func TestNodePutGetEntity(t *testing.T) {
	node := createTestNode()
	key := "foo"
	payload := "bar"
	entity := newTestObject(payload)
	err := tryPutEntity(node, key, entity)
	assert.NoError(t, err)
	storedEntity := &testObject{}
	err = tryGetEntity(node, key, storedEntity)
	assert.NoError(t, err)
	assert.Equal(t, payload, storedEntity.Payload)
}

func TestNodeRegisterNode(t *testing.T) {
	n := createTestNode()
	o := createTestNode()
	id := url.QueryEscape(strconv.Itoa(o.ID))
	address := url.QueryEscape(o.Address)
	uri := "http://" + n.Address + n.RootPath + registerPath + "?" + idParam + "=" + id + "&" + addressParam + "=" + address
	_, err := http.Get(uri)
	assert.NoError(t, err)
	a := n.nodes[o.ID]
	assert.Equal(t, o.Address, a)
}

func createTestNode() Node {
	store := NewMemoryStore()
	node := NewNode(store)
	port := getNextTestPort()
	node.RootPath = "/v1/" + strconv.Itoa(port) + "/entities"
	node.Start(port)
	time.Sleep(time.Millisecond * 100)
	return node
}

func tryPutEntity(n Node, key string, entity interface{}) error {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	err := encoder.Encode(entity)
	if err != nil {
		return err
	}

	uri := "http://" + n.Address + n.RootPath + entitiesPath + "/" + key
	_, err = http.Post(uri, "application/json; charset=utf-8", b)
	if err != nil {
		return err
	}
	return nil
}

func tryGetEntity(n Node, key string, entity interface{}) error {
	uri := "http://" + n.Address + n.RootPath + entitiesPath + "/" + key
	response, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(entity)
	if err != nil {
		return err
	}
	return nil
}