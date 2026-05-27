package aep

import "testing"

// TestInferLayerType_ByteDispatch covers the ldta @0x83 enum dispatch in
// [inferLayerType] — the byte that distinguishes AE's built-in 3D layer
// kinds (Light=1 / Camera=2 / 3DModel=5). Text / Shape are detected via
// the property tree elsewhere, so they're not exercised here.
func TestInferLayerType_ByteDispatch(t *testing.T) {
	mkLdta := func(subtype byte) []byte {
		b := make([]byte, 0x84)
		b[0x83] = subtype
		return b
	}
	cases := []struct {
		name    string
		subtype byte
		want    LayerType
	}{
		{"Light=1", 0x01, LayerTypeLight},
		{"Camera=2", 0x02, LayerTypeCamera},
		{"3DModel=5", 0x05, LayerType3DModel},
		{"AV=0_falls_through", 0x00, LayerTypeAV},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// SourceID > 0 so the byte=0 fallback path doesn't drop into
			// LayerTypeNull (see inferLayerType post-switch logic).
			got := inferLayerType(mkLdta(c.subtype), &Layer{SourceID: 1})
			if got != c.want {
				t.Errorf("inferLayerType(subtype=%#x) = %q, want %q", c.subtype, got, c.want)
			}
		})
	}
}
