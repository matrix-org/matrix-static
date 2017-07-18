package quicktemplate

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteJSONString(t *testing.T) {
	testWriteJSONString(t, ``)
	testWriteJSONString(t, `f`)
	testWriteJSONString(t, `"`)
	testWriteJSONString(t, `<`)
	testWriteJSONString(t, "\x00\n\r\t\b\f"+`"\`)
	testWriteJSONString(t, `"foobar`)
	testWriteJSONString(t, `foobar"`)
	testWriteJSONString(t, `foo "bar"
		baz`)
	testWriteJSONString(t, `this is a "тест"`)
	testWriteJSONString(t, `привет test ыва`)

	testWriteJSONString(t, `</script><script>alert('evil')</script>`)
}

func testWriteJSONString(t *testing.T, s string) {
	expectedResult, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("unexpected error when encoding string %q: %s", s, err)
	}
	expectedResult = expectedResult[1 : len(expectedResult)-1]

	bb := AcquireByteBuffer()
	writeJSONString(bb, s)
	result := string(bb.B)
	ReleaseByteBuffer(bb)

	if strings.Contains(result, "'") {
		t.Fatalf("json string shouldn't contain single quote: %q, src %q", result, s)
	}
	result = strings.Replace(result, `\u0027`, "'", -1)
	result = strings.Replace(result, ">", `\u003e`, -1)
	if result != string(expectedResult) {
		t.Fatalf("unexpected result %q. Expecting %q. original string %q", result, expectedResult, s)
	}
}
