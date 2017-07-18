package tests

import (
	"io/ioutil"
	"testing"

	"github.com/valyala/quicktemplate/testdata/templates"
)

func TestIntegration(t *testing.T) {
	expectedS, err := ioutil.ReadFile("../testdata/templates/integration.qtpl.out")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	s := templates.Integration()
	if s != string(expectedS) {
		t.Fatalf("unexpected output\n%q\nExpecting\n%q\n", s, expectedS)
	}
}
