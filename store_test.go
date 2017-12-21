package corduroy

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"encoding/json"
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
	storedValue := store.Get(key)
	storedBytes := []byte(storedValue)
	storedEntity := &testObject{}
	err = json.Unmarshal(storedBytes, storedEntity)
	assert.NoError(t, err)
	assert.Equal(t, payload, storedEntity.Payload)
}
