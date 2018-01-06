package corduroy

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemoryStorePutGet(t *testing.T) {
	store := NewMemoryStore()
	key := "foo"
	payload := "bar"
	entity := newTestObject(payload)
	bytes, err := json.Marshal(entity)
	assert.NoError(t, err)
	value := string(bytes)
	store.Put(key, value)
	assert.True(t, store.Contains(key))
	storedValue := store.Get(key)
	storedBytes := []byte(storedValue)
	storedEntity := &testObject{}
	err = json.Unmarshal(storedBytes, storedEntity)
	assert.NoError(t, err)
	assert.Equal(t, payload, storedEntity.Payload)
}

func TestMemoryStoreSize(t *testing.T) {
	store := NewMemoryStore()
	key1 := "luke"
	value1 := "skywalker"
	store.Put(key1, value1)
	key2 := "han"
	value2 := "solo"
	store.Put(key2, value2)
	assert.Equal(t, 2, store.Size())
}

func TestMemoryStorePutGetKeys(t *testing.T) {
	store := NewMemoryStore()
	key1 := "marilyn"
	value1 := "monroe"
	store.Put(key1, value1)
	key2 := "audrey"
	value2 := "hepburn"
	store.Put(key2, value2)
	keys := store.GetKeys(0, 2)
	assert.Equal(t, 2, len(keys))
	assert.Equal(t, key1, keys[0])
	keys = store.GetKeys(1, 1)
	assert.Equal(t, 1, len(keys))
	assert.Equal(t, key2, keys[0])
	keys = store.GetKeys(0, 3)
	assert.Equal(t, 2, len(keys))
}
