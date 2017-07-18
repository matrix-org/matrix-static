package tests

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"testing"

	"github.com/valyala/quicktemplate"
	"github.com/valyala/quicktemplate/testdata/templates"
)

func init() {
	// Make sure both encoding/json and templates generate identical output.
	d := newTemplatesData(3)
	bb := quicktemplate.AcquireByteBuffer()

	expectedData, err := json.Marshal(d)
	if err != nil {
		log.Fatalf("unexpected error: %s", err)
	}

	e := json.NewEncoder(bb)
	if err := e.Encode(d); err != nil {
		log.Fatalf("unexpected error: %s", err)
	}
	bb.B = bytes.TrimSpace(bb.B)
	if !bytes.Equal(bb.B, expectedData) {
		log.Fatalf("unexpected data generated with encoding/json:\n%q\n. Expecting\n%q\n", bb.B, expectedData)
	}

	bb.Reset()
	d.WriteJSON(bb)
	if !bytes.Equal(bb.B, expectedData) {
		log.Fatalf("unexpected data generated with quicktemplate:\n%q\n. Expecting\n%q\n", bb.B, expectedData)
	}

	// make sure both encoding/xml and templates generate identical output.
	expectedData, err = xml.Marshal(d)
	if err != nil {
		log.Fatalf("unexpected error: %s", err)
	}

	bb.Reset()
	xe := xml.NewEncoder(bb)
	if err := xe.Encode(d); err != nil {
		log.Fatalf("unexpected error: %s", err)
	}
	if !bytes.Equal(bb.B, expectedData) {
		log.Fatalf("unexpected data generated with encoding/xml:\n%q\n. Expecting\n%q\n", bb.B, expectedData)
	}

	bb.Reset()
	d.WriteXML(bb)
	if !bytes.Equal(bb.B, expectedData) {
		log.Fatalf("unexpected data generated with quicktemplate:\n%q\n. Expecting\n%q\n", bb.B, expectedData)
	}

	quicktemplate.ReleaseByteBuffer(bb)
}

func BenchmarkMarshalJSONStd1(b *testing.B) {
	benchmarkMarshalJSONStd(b, 1)
}

func BenchmarkMarshalJSONStd10(b *testing.B) {
	benchmarkMarshalJSONStd(b, 10)
}

func BenchmarkMarshalJSONStd100(b *testing.B) {
	benchmarkMarshalJSONStd(b, 100)
}

func BenchmarkMarshalJSONStd1000(b *testing.B) {
	benchmarkMarshalJSONStd(b, 1000)
}

func benchmarkMarshalJSONStd(b *testing.B, n int) {
	d := newTemplatesData(n)
	b.RunParallel(func(pb *testing.PB) {
		bb := quicktemplate.AcquireByteBuffer()
		e := json.NewEncoder(bb)
		for pb.Next() {
			if err := e.Encode(d); err != nil {
				b.Fatalf("unexpected error: %s", err)
			}
			bb.Reset()
		}
		quicktemplate.ReleaseByteBuffer(bb)
	})
}

func BenchmarkMarshalJSONQuickTemplate1(b *testing.B) {
	benchmarkMarshalJSONQuickTemplate(b, 1)
}

func BenchmarkMarshalJSONQuickTemplate10(b *testing.B) {
	benchmarkMarshalJSONQuickTemplate(b, 10)
}

func BenchmarkMarshalJSONQuickTemplate100(b *testing.B) {
	benchmarkMarshalJSONQuickTemplate(b, 100)
}

func BenchmarkMarshalJSONQuickTemplate1000(b *testing.B) {
	benchmarkMarshalJSONQuickTemplate(b, 1000)
}

func benchmarkMarshalJSONQuickTemplate(b *testing.B, n int) {
	d := newTemplatesData(n)
	b.RunParallel(func(pb *testing.PB) {
		bb := quicktemplate.AcquireByteBuffer()
		for pb.Next() {
			d.WriteJSON(bb)
			bb.Reset()
		}
		quicktemplate.ReleaseByteBuffer(bb)
	})
}

func BenchmarkMarshalXMLStd1(b *testing.B) {
	benchmarkMarshalXMLStd(b, 1)
}

func BenchmarkMarshalXMLStd10(b *testing.B) {
	benchmarkMarshalXMLStd(b, 10)
}

func BenchmarkMarshalXMLStd100(b *testing.B) {
	benchmarkMarshalXMLStd(b, 100)
}

func BenchmarkMarshalXMLStd1000(b *testing.B) {
	benchmarkMarshalXMLStd(b, 1000)
}

func benchmarkMarshalXMLStd(b *testing.B, n int) {
	d := newTemplatesData(n)
	b.RunParallel(func(pb *testing.PB) {
		bb := quicktemplate.AcquireByteBuffer()
		e := xml.NewEncoder(bb)
		for pb.Next() {
			if err := e.Encode(d); err != nil {
				b.Fatalf("unexpected error: %s", err)
			}
			bb.Reset()
		}
		quicktemplate.ReleaseByteBuffer(bb)
	})
}

func BenchmarkMarshalXMLQuickTemplate1(b *testing.B) {
	benchmarkMarshalXMLQuickTemplate(b, 1)
}

func BenchmarkMarshalXMLQuickTemplate10(b *testing.B) {
	benchmarkMarshalXMLQuickTemplate(b, 10)
}

func BenchmarkMarshalXMLQuickTemplate100(b *testing.B) {
	benchmarkMarshalXMLQuickTemplate(b, 100)
}

func BenchmarkMarshalXMLQuickTemplate1000(b *testing.B) {
	benchmarkMarshalXMLQuickTemplate(b, 1000)
}

func benchmarkMarshalXMLQuickTemplate(b *testing.B, n int) {
	d := newTemplatesData(n)
	b.RunParallel(func(pb *testing.PB) {
		bb := quicktemplate.AcquireByteBuffer()
		for pb.Next() {
			d.WriteXML(bb)
			bb.Reset()
		}
		quicktemplate.ReleaseByteBuffer(bb)
	})
}

func newTemplatesData(n int) *templates.MarshalData {
	var rows []templates.MarshalRow
	for i := 0; i < n; i++ {
		rows = append(rows, templates.MarshalRow{
			Msg: fmt.Sprintf("тест %d", i),
			N:   i,
		})
	}
	return &templates.MarshalData{
		Foo:  1,
		Bar:  "foobar",
		Rows: rows,
	}
}
