package typical

import (
	"bytes"
	"io"
	"testing"
)

func BenchmarkTypical(b *testing.B) {
	buf := &bytes.Buffer{}

	for i := 0; i < b.N; i++ {
		harness(b, buf, writeJsonToStream)
	}
}

func BenchmarkConventional(b *testing.B) {
	buf := &bytes.Buffer{}

	for i := 0; i < b.N; i++ {
		harness(b, buf, normalJsonWriteFunction)
	}
}

const expect = `{"Field":"hello world"}` + "\n"

func harness(b *testing.B, buf *bytes.Buffer, fn func(interface{}, io.Writer) error) {
	buf.Reset()

	if err := fn(&someObject{"hello world"}, buf); err != nil {
		b.Fatalf("unexpected error: %s", err)
	}
	if buf.String() != expect {
		b.Fatalf("unexpected result: %q", buf.String())
	}
	if err := fn(complex(1, 2), buf); err == nil || err.Error() != "json: unsupported type: complex128" {
		b.Fatalf("unexpected error: %s", err)
	}
}
