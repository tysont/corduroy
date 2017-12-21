package corduroy

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func newTestObject(payload string) *testObject{
	return &testObject{
		Payload: payload,
	}
}

type testObject struct {
	Payload string
}

var testPort = 8080

func getNextTestPort() int {
	testPort++
	return testPort
}

func TestGetLocalAddresses(t *testing.T) {
	a, err := getLocalAddresses()
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.True(t, len(a) > 0)
}

func TestHash(t *testing.T) {
	s := "ok computer"
	i := hash(s)

	i2 := hash(s)
	assert.Equal(t, i, i2)

	s3 := "nevermind"
	i3 := hash(s3)
	assert.NotEqual(t, i, i3)
}