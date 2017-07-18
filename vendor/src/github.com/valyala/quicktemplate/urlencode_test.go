package quicktemplate

import (
	"net/url"
	"testing"
)

func TestAppendURLEncode(t *testing.T) {
	testAppendURLEncode(t, "")
	testAppendURLEncode(t, "f")
	testAppendURLEncode(t, " ")
	testAppendURLEncode(t, ".-_")
	testAppendURLEncode(t, "тест+this/&=;?\n\t\rabc")
}

func testAppendURLEncode(t *testing.T, s string) {
	expectedResult := url.QueryEscape(s)
	result := appendURLEncode(nil, s)
	if string(result) != expectedResult {
		t.Fatalf("unexpected result %q. Expecting %q. str=%q", result, expectedResult, s)
	}
}
