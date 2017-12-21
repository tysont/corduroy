package corduroy

import (
	"testing"
	"net/http"
	"encoding/json"
	"bytes"
	"time"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func TestNodeGet(t *testing.T) {
	node := createStartTestNode()
	key := "foo"
	payload := "bar"
	entity := newTestObject("bar")
	bytes, err := json.Marshal(entity)
	assert.NoError(t, err)
	value := string(bytes)
	err = testPutEntity(node, key, value)
	node.store.Put("foo", value)

	storedEntity := &testObject{}
	err = testGetEntity(node, key, storedEntity)
	assert.NoError(t, err)
	assert.Equal(t, payload, storedEntity.Payload)
}

func TestNodePut(t *testing.T) {
	node := createStartTestNode()
	key := "foo"
	payload := "bar"
	entity := newTestObject(payload)
	err := testPutEntity(node, key, entity)
	assert.NoError(t, err)
	storedValue := node.store.Get(key)
	storedBytes := []byte(storedValue)
	storedEntity := &testObject{}
	err = json.Unmarshal(storedBytes, storedEntity)
	assert.NoError(t, err)
	assert.Equal(t, payload, storedEntity.Payload)
}

func TestNodePutGet(t *testing.T) {
	node := createStartTestNode()
	key := "foo"
	payload := "bar"
	entity := newTestObject(payload)
	err := testPutEntity(node, key, entity)
	assert.NoError(t, err)
	storedEntity := &testObject{}
	err = testGetEntity(node, key, storedEntity)
	assert.NoError(t, err)
	assert.Equal(t, payload, storedEntity.Payload)
}

func createStartTestNode() Node {
	store := NewMemoryStore()
	node := NewNode(store)
	port := getNextTestPort()
	node.RootPath = "/v1/" + strconv.Itoa(port) + "/entities"
	node.Start(port)
	time.Sleep(time.Millisecond * 100)
	return node
}

func testGetEntity(n Node, key string, entity interface{}) error {
	uri := "http://" + n.Address + n.RootPath + "/" + key
	response, err := http.Get(uri)
	if err != nil {
		return err
	}
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

func testPutEntity(n Node, key string, entity interface{}) error {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	err := encoder.Encode(entity)
	if err != nil {
		return err
	}

	uri := "http://" + n.Address + n.RootPath + "/" + key
	_, err = http.Post(uri, "application/json; charset=utf-8", b)
	if err != nil {
		return err
	}
	return nil
}
