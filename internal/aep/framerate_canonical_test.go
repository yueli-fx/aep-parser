package aep

import (
	"math"
	"testing"

	"github.com/example/aep-parser/internal/codec"
)

func TestNTSCCanonicalRoundtrip(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want float64
	}{
		{"29.97 exact", 29.97, 29.97},
		{"30000/1001 NTSC", 30000.0 / 1001.0, 29.97},
		{"23.976 exact", 23.976, 23.976},
		{"24000/1001 NTSC", 24000.0 / 1001.0, 23.976},
		{"59.94 exact", 59.94, 59.94},
		{"30 fps integer", 30.0, 30.0},
		{"24 fps integer", 24.0, 24.0},
		{"25 fps PAL", 25.0, 25.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := codec.EncodeFrameRate(tc.in)
			got := codec.DecodeFrameRate(enc)
			if math.Abs(got-tc.want) > 1e-4 {
				t.Errorf("encode→decode %g: got %g, want %g (enc=%+v)", tc.in, got, tc.want, enc)
			}
		})
	}
}

func TestNTSCCanonicalIdempotent(t *testing.T) {
	v1 := codec.EncodeFrameRate(29.97)
	v2 := codec.EncodeFrameRate(30000.0 / 1001.0)
	if v1 != v2 {
		t.Errorf("29.97 (%+v) != 30000/1001 (%+v): canonicalization failed", v1, v2)
	}
}
