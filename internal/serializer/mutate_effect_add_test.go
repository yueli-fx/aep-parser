// White-box regression for the effect-template host-layer retarget: every tdpi
// inside the spliced sspc must carry the destination layer's ID, not the
// extraction fixture's host id (15) — AE rejects dangling tdpi with "cannot
// find layer ID=N in composition" (caught by the parade auto-create ship-gate).
package serializer

import (
	"encoding/binary"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func TestAddEffect_RetargetsTdpiHostLayer(t *testing.T) {
	p := NewProject()
	comp, err := NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewShapeLayer(comp, "S"); err != nil {
		t.Fatal(err)
	}
	rp, err := Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	var l *Layer
	for _, c := range rp.Compositions {
		for _, cl := range c.Layers {
			if cl.Name == "S" {
				l = cl
			}
		}
	}
	if l == nil {
		t.Fatal("layer S not found after Reopen")
	}
	if _, err := AddEffect(l, EffectGaussianBlur); err != nil {
		t.Fatalf("AddEffect: %v", err)
	}

	lb := layerBack(l)
	if lb == nil || lb.layrList == nil {
		t.Fatal("layer has no Layr back-ref")
	}
	var tdpis []uint32
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if c.ID == rifx.IDTdpi && len(c.Data) >= 4 {
			tdpis = append(tdpis, binary.BigEndian.Uint32(c.Data[0:4]))
		}
		for _, ch := range c.Children {
			walk(ch)
		}
	}
	walk(lb.layrList)
	if len(tdpis) == 0 {
		t.Fatal("no tdpi chunks found in spliced effect payload")
	}
	for i, v := range tdpis {
		if v != l.ID {
			t.Errorf("tdpi[%d] = %d, want host layer ID %d", i, v, l.ID)
		}
	}
}
