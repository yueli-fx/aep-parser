package serializer

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestSetStretchUsesPreciseDivisorWhenTemplateIsCoarse(t *testing.T) {
	layer := newSyntheticLayer(LayerTypeText, "Text", 1)
	ldta := layerBack(layer).ldta.Data
	binary.BigEndian.PutUint32(ldta[0x08:0x0C], 1)
	binary.BigEndian.PutUint32(ldta[0x6C:0x70], 1)

	if err := layer.SetStretch(0.5); err != nil {
		t.Fatalf("SetStretch(0.5): %v", err)
	}

	got, ok := layerBack(layer).StretchFrac()
	if !ok {
		t.Fatal("StretchFrac ok = false")
	}
	if math.Abs(got-0.5) > 1e-6 {
		t.Fatalf("stretch = %g, want 0.5", got)
	}
	if div := binary.BigEndian.Uint32(ldta[0x6C:0x70]); div == 1 {
		t.Fatalf("stretch divisor stayed coarse: %d", div)
	}
}
