package quicktemplate

import (
	"testing"
)

func BenchmarkHTMLEscapeWriterNoHTML(b *testing.B) {
	s := "foobarbazabcdefghjkl"
	benchmarkHTMLEscapeWriter(b, s)
}

func BenchmarkHTMLEscapeWriterWithHTML(b *testing.B) {
	s := "foo<a>baza</a>fghjkl"
	benchmarkHTMLEscapeWriter(b, s)
}

func benchmarkHTMLEscapeWriter(b *testing.B, s string) {
	sBytes := []byte(s)
	b.RunParallel(func(pb *testing.PB) {
		var err error
		bb := AcquireByteBuffer()
		w := &htmlEscapeWriter{w: bb}
		for pb.Next() {
			if _, err = w.Write(sBytes); err != nil {
				b.Fatalf("unexpected error: %s", err)
			}
			bb.Reset()
		}
		ReleaseByteBuffer(bb)
	})
}
