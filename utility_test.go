package corduroy

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