package aep_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	aep "github.com/yueli-fx/aep-parser"
)

func BenchmarkParseSolidNull(b *testing.B) {
	data := benchmarkFixture(b)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for range b.N {
		if _, err := aep.Parse(bytes.NewReader(data)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProfileSolidNull(b *testing.B) {
	data := benchmarkFixture(b)
	document, err := aep.Parse(bytes.NewReader(data))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := document.ProfileJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkFixture(b *testing.B) []byte {
	b.Helper()
	path := filepath.Join("test_data", "fixtures", "re_solidnull.aep")
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skipf("fixture unavailable: %v", err)
	}
	return data
}
