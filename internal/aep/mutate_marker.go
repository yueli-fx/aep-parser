package aep

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// Marker structural ops (P3 §3G). Comp/layer markers are a keyframe-backed
// set: the keyframe times live in one ldat (count × 16 bytes), the count in
// the kfl's lhd3 @0x08, and each marker's text/metadata in a sibling Nmrd
// under mrky. Removing one marker therefore splices three places at once —
// the 16-byte ldat block, the lhd3 count, and the Nmrd — then drops the
// marker from its public Markers slice.
//
// length-variable: the ldat and mrky LISTs shrink; WriteAEP recomputes all
// ancestor LIST sizes (the same path SetComment already exercises on the mrst
// chain). Alpha until the AE 2020 + 2025 ship-gate passes (CLAUDE.md #6).
//
// Atomicity: every precondition is checked before the first byte is written,
// so a rejected Remove leaves the project untouched without a rollback path.

// RemoveMarker deletes this marker from its owning composition / layer marker set.
//
// It splices the marker's 16-byte ldat keyframe block, decrements the kfl
// count, removes the marker's Nmrd from mrky, shifts the trailing markers'
// ldat offsets down, and drops the marker from the public Markers slice. The
// receiver is detached afterward — a second RemoveMarker (or any Set*) errors.
//
// Errors (project untouched): the marker was built outside the parser, is
// already detached, or its chunk references are inconsistent.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. Renamed + BREAKING vs the former Marker.Remove method form.
func RemoveMarker(m *Marker) error {
	ml := markerSet(m)
	if ml == nil || ml.owner == nil {
		return fmt.Errorf("marker: Remove unsupported (built outside parser or owner unbound)")
	}
	if ml.ldat == nil || ml.lhd3 == nil {
		return fmt.Errorf("marker: Remove missing ldat/lhd3 reference")
	}
	if len(ml.lhd3.Data) < 0x0C {
		return fmt.Errorf("marker: lhd3 too short for count (len=%d)", len(ml.lhd3.Data))
	}

	markers := *ml.owner
	idx := -1
	for i, x := range markers {
		if x == m {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("marker: not found in owning set (already removed?)")
	}

	off := scene.MarkerLdatOffset(m)
	if off < 0 || off+16 > len(ml.ldat.Data) {
		return fmt.Errorf("marker: ldat offset %d out of range (len=%d)", off, len(ml.ldat.Data))
	}
	count := binary.BigEndian.Uint32(ml.lhd3.Data[0x08:0x0C])
	if count == 0 {
		return fmt.Errorf("marker: lhd3 count already 0")
	}

	// --- All preconditions passed; commit (no failure points below). ---

	// 1. ldat: drop this marker's 16-byte keyframe block.
	newLdat := make([]byte, 0, len(ml.ldat.Data)-16)
	newLdat = append(newLdat, ml.ldat.Data[:off]...)
	newLdat = append(newLdat, ml.ldat.Data[off+16:]...)
	ml.ldat.Data = newLdat

	// 2. Shift every later marker's block offset down by one block.
	for _, x := range markers {
		if scene.MarkerLdatOffset(x) > off {
			scene.SetMarkerLdatOffset(x, scene.MarkerLdatOffset(x)-16)
		}
	}

	// 3. lhd3: decrement the keyframe count.
	binary.BigEndian.PutUint32(ml.lhd3.Data[0x08:0x0C], count-1)

	// 4. mrky: remove this marker's Nmrd LIST.
	// P3-tracked stopgap: type-assert to access chunk field not in MarkerWriter.
	if ml.mrky != nil {
		if mb := markerBack(m); mb != nil && mb.nmrd != nil {
			if ni := indexOfChunk(ml.mrky.Children, mb.nmrd); ni >= 0 {
				ch := ml.mrky.Children
				ml.mrky.Children = append(ch[:ni:ni], ch[ni+1:]...)
			}
		}
	}

	// 5. scene: drop from the public Markers slice and detach.
	*ml.owner = append(markers[:idx:idx], markers[idx+1:]...)
	scene.SetMarkerSetList(m, nil)
	return nil
}

// AddMarker appends a new composition marker at the given time (seconds) and
// returns it for further Set* calls. The new marker is a clean point marker:
// no duration, no label color, empty text fields.
//
// Mechanics (clone-template): to avoid reverse-engineering the canonical
// defaults of the ldat block's opaque metadata (0x04-0x0F) and the NmHd's
// reserved/flag bytes, the new marker clones an existing marker's ldat block
// and NmHd verbatim (opaque preservation, CLAUDE.md #5), then resets the time
// plus the known semantic NmHd fields (duration @0x08, label @0x10) to zero.
// The Nmrd gets five empty Utf8 slots, matching AE's always-five layout.
//
// length-variable — the ldat and mrky LISTs grow; WriteAEP recomputes the
// mrst-chain LIST sizes. Alpha until the AE 2020 + 2025 ship-gate passes.
//
// Restriction: requires the comp to already have ≥1 marker (the clone
// template). Seeding the entire "Markers" pseudo-layer for an empty comp is a
// separate slice (needs a canonical seed); AddMarker returns an error there.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Composition.AddMarker method form.
func AddMarker(c *Composition, seconds float64) (*Marker, error) {
	if seconds < 0 {
		return nil, fmt.Errorf("marker: negative time %g not supported", seconds)
	}
	if len(c.Markers) == 0 {
		return nil, fmt.Errorf("marker: AddMarker into an empty comp marker set is unsupported (no template to clone; needs a canonical seed)")
	}
	tmpl := c.Markers[len(c.Markers)-1]
	ml := markerSet(tmpl)
	if ml == nil || ml.owner == nil || ml.ldat == nil || ml.lhd3 == nil {
		return nil, fmt.Errorf("marker: AddMarker missing marker-set references")
	}
	if ml.mrky == nil {
		return nil, fmt.Errorf("marker: AddMarker requires an mrky branch (none in this set)")
	}
	tmplMb := markerBack(tmpl)
	if tmplMb == nil || tmplMb.nmHd == nil {
		return nil, fmt.Errorf("marker: AddMarker template marker has no NmHd to clone")
	}
	if len(ml.lhd3.Data) < 0x0C {
		return nil, fmt.Errorf("marker: lhd3 too short for count (len=%d)", len(ml.lhd3.Data))
	}
	tmplOff := scene.MarkerLdatOffset(tmpl)
	if tmplOff < 0 || tmplOff+16 > len(ml.ldat.Data) {
		return nil, fmt.Errorf("marker: template ldat block out of range")
	}
	tmplTick := tmplMb.tickRate
	rate := tmplTick
	if rate == 0 {
		rate = aeLegacyTimeBase
	}

	// --- All preconditions passed; commit (no failure points below). ---

	// 1. ldat: append a clone of the template's 16-byte block; set the time.
	newOff := len(ml.ldat.Data)
	block := make([]byte, 16)
	copy(block, ml.ldat.Data[tmplOff:tmplOff+16])
	ticks := uint32(math.Round(seconds * rate))
	binary.BigEndian.PutUint32(block[0:4], ticks)
	ml.ldat.Data = append(ml.ldat.Data, block...)

	// 2. lhd3: increment the keyframe count.
	count := binary.BigEndian.Uint32(ml.lhd3.Data[0x08:0x0C])
	binary.BigEndian.PutUint32(ml.lhd3.Data[0x08:0x0C], count+1)

	// 3. mrky: new Nmrd { NmHd(clone, reset to point marker) + 5 empty Utf8 }.
	nmHdClone := deepCloneChunk(tmplMb.nmHd)
	if len(nmHdClone.Data) >= 0x0C {
		binary.BigEndian.PutUint32(nmHdClone.Data[0x08:0x0C], 0) // duration → 0
	}
	if len(nmHdClone.Data) >= 0x11 {
		nmHdClone.Data[0x10] = 0 // label → default
	}
	nmrd := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDNmrd, Children: []*rifx.Chunk{nmHdClone}}
	for i := 0; i < 5; i++ {
		nmrd.Children = append(nmrd.Children, &rifx.Chunk{ID: rifx.IDUtf8})
	}
	ml.mrky.Children = append(ml.mrky.Children, nmrd)

	// 4. scene: the new Marker, fully back-referenced.
	newMb := &markerBackrefs{ldat: ml.ldat, ldatOffset: newOff, tickRate: tmplTick, nmHd: nmHdClone, nmrd: nmrd}
	nm := &Marker{
		Time: float64(ticks) / rate,
	}
	scene.SetMarkerLdatOffset(nm, newOff)
	scene.SetMarkerRates(nm, tmplTick, scene.MarkerCompFps(tmpl))
	scene.SetMarkerBack(nm, newMb)
	scene.SetMarkerSetList(nm, ml)
	*ml.owner = append(*ml.owner, nm)
	return nm, nil
}
