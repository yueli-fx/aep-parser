package serializer

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// markerBackrefs holds the per-marker rifx.Chunk references that power a
// parsed Marker's write paths (SetTime / SetDuration / SetLabel / the five
// SetComment-family text setters). Lives in a separate shard so the Marker
// scene type stays chunk-free.
//
// Lifecycle:
//   - Populated by parseMarkers (and AddMarker) when a Marker is built from
//     a parsed .aep file.
//   - nil for markers built outside the parser; every Set* refuses with an
//     error when the reference it needs is missing.
type markerBackrefs struct {
	// ldat is the shared keyframe-block chunk; this marker's 16-byte block
	// starts at Marker.ldatOffset. SetTime patches its time slot in place.
	ldat *rifx.Chunk
	// nmHd is the per-marker NmHd chunk (Duration @0x08 / Label @0x10).
	nmHd *rifx.Chunk
	// nmrd is the per-marker Nmrd LIST holding the Utf8 text children.
	nmrd *rifx.Chunk
	// ldatOffset is this marker's 16-byte block start in ldat.Data.
	ldatOffset int
	// tickRate is the owning composition's TickRate (for SetTime).
	tickRate float64
}

var _ MarkerWriter = (*markerBackrefs)(nil)

func (b *markerBackrefs) CompTickRate() float64 {
	if b == nil {
		return 0
	}
	return b.tickRate
}

// markerBack returns the concrete back-refs behind a Marker's writer interface
// for serializer-stage raw chunk access. Returns nil when built outside the
// parser. Free function (the receiver is a scene type post package-split).
func markerBack(m *Marker) *markerBackrefs {
	if mb, ok := scene.MarkerBack(m).(*markerBackrefs); ok {
		return mb
	}
	return nil
}

// markerSet returns the concrete owning marker-set container behind a Marker's
// MarkerSetRef. Returns nil when built outside the parser.
func markerSet(m *Marker) *markerList {
	if ml, ok := scene.MarkerSetList(m).(*markerList); ok {
		return ml
	}
	return nil
}

func (b *markerBackrefs) SetTime(seconds float64) error {
	if b.ldat == nil {
		return fmt.Errorf("marker: no ldat reference (built outside parser?)")
	}
	if seconds < 0 {
		return fmt.Errorf("marker: negative time %g not supported", seconds)
	}
	if b.ldatOffset+4 > len(b.ldat.Data) {
		return fmt.Errorf("marker: ldat offset %d out of range (len=%d)", b.ldatOffset, len(b.ldat.Data))
	}
	rate := b.tickRate
	if rate == 0 {
		rate = aeLegacyTimeBase
	}
	ticks := uint32(math.Round(seconds * rate))
	binary.BigEndian.PutUint32(b.ldat.Data[b.ldatOffset:b.ldatOffset+4], ticks)
	return nil
}

func (b *markerBackrefs) SetDuration(seconds float64) error {
	if b.nmHd == nil {
		return fmt.Errorf("marker: no NmHd reference (built outside parser?)")
	}
	if len(b.nmHd.Data) < 0x0C {
		return fmt.Errorf("marker: NmHd too short for Duration write (len=%d)", len(b.nmHd.Data))
	}
	if seconds < 0 {
		seconds = 0
	}
	const nmHdDurationBase = 600.0
	ticks := uint32(math.Round(seconds * nmHdDurationBase))
	binary.BigEndian.PutUint32(b.nmHd.Data[0x08:0x0C], ticks)
	return nil
}

func (b *markerBackrefs) SetLabel(index uint8) error {
	if b.nmHd == nil {
		return fmt.Errorf("marker: no NmHd reference (built outside parser?)")
	}
	if len(b.nmHd.Data) < 0x11 {
		return fmt.Errorf("marker: NmHd too short for Label write (len=%d)", len(b.nmHd.Data))
	}
	b.nmHd.Data[0x10] = index
	return nil
}

func (b *markerBackrefs) SetComment(s string) error {
	return b.setNmrdUtf8(0, s)
}

func (b *markerBackrefs) SetChapter(s string) error {
	return b.setNmrdUtf8(1, s)
}

func (b *markerBackrefs) SetURL(s string) error {
	return b.setNmrdUtf8(2, s)
}

func (b *markerBackrefs) SetFrameTarget(s string) error {
	return b.setNmrdUtf8(3, s)
}

func (b *markerBackrefs) SetCuePointName(s string) error {
	return b.setNmrdUtf8(4, s)
}

// setNmrdUtf8 mutates the n-th Utf8 child of the marker's Nmrd block.
// If fewer Utf8 children exist than slot+1, missing slots are filled with
// empty Utf8 chunks so the requested slot reaches the right index in
// declaration order (AE writes all 5 slots even when empty).
func (b *markerBackrefs) setNmrdUtf8(slot int, s string) error {
	if b.nmrd == nil {
		return fmt.Errorf("marker: no Nmrd reference (built outside parser?)")
	}
	nmrd := b.nmrd
	var utf8Idx []int
	for i, ch := range nmrd.Children {
		if ch.ID == rifx.IDUtf8 {
			utf8Idx = append(utf8Idx, i)
		}
	}
	for len(utf8Idx) <= slot {
		empty := &rifx.Chunk{ID: rifx.IDUtf8, Data: nil}
		nmrd.Children = append(nmrd.Children, empty)
		utf8Idx = append(utf8Idx, len(nmrd.Children)-1)
	}
	target := nmrd.Children[utf8Idx[slot]]
	target.Data = []byte(s)
	return nil
}

// markerList is the internal container behind one "ADBE Marker" set (a comp's
// or a layer's markers). It holds the chunk references the structural ops
// splice — the keyframe count (lhd3), the keyframe blocks (ldat), and the
// Nmrd-bearing mrky LIST — plus owner, a pointer to the public Markers field
// these markers live in, so Remove/Add can keep that slice in sync. All
// markers in one set share a single *markerList.
type markerList struct {
	lhd3  *rifx.Chunk // kfl count chunk: count @0x08, bpk @0x10 = 16
	ldat  *rifx.Chunk // keyframe blocks: count × 16 bytes
	mrky  *rifx.Chunk // Nmrd container (nil when the set has no mrky branch)
	owner *[]*Marker  // the public Composition.Markers / Layer.Markers field
}

var _ scene.MarkerSetRef = (*markerList)(nil)

func (b *markerList) IsMarkerSetRef() {}
