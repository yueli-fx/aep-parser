package aep_test

import (
	"bytes"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// FuzzFromReaderNeverPanics exercises both the generic RIFX reader and the AEP
// serializer decoders. Valid structured seeds are important because arbitrary
// bytes alone almost always stop at the outer RIFX header.
func FuzzFromReaderNeverPanics(f *testing.F) {
	f.Add([]byte{})
	f.Add(buildMinimalAEP())
	f.Add(buildKeyframedAEP())

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = aep.FromReader(bytes.NewReader(data))
	})
}
