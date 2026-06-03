package aep

import (
	"encoding/binary"
	"fmt"
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

// Remove deletes this marker from its owning composition / layer marker set.
//
// It splices the marker's 16-byte ldat keyframe block, decrements the kfl
// count, removes the marker's Nmrd from mrky, shifts the trailing markers'
// ldat offsets down, and drops the marker from the public Markers slice. The
// receiver is detached afterward — a second Remove (or any Set*) errors.
//
// Errors (project untouched): the marker was built outside the parser, is
// already detached, or its chunk references are inconsistent.
func (m *Marker) Remove() error {
	ml := m.list
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

	off := m.ldatOffset
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
		if x.ldatOffset > off {
			x.ldatOffset -= 16
		}
	}

	// 3. lhd3: decrement the keyframe count.
	binary.BigEndian.PutUint32(ml.lhd3.Data[0x08:0x0C], count-1)

	// 4. mrky: remove this marker's Nmrd LIST.
	if ml.mrky != nil && m.nmrd != nil {
		if ni := indexOfChunk(ml.mrky.Children, m.nmrd); ni >= 0 {
			ch := ml.mrky.Children
			ml.mrky.Children = append(ch[:ni:ni], ch[ni+1:]...)
		}
	}

	// 5. scene: drop from the public Markers slice and detach.
	*ml.owner = append(markers[:idx:idx], markers[idx+1:]...)
	m.list = nil
	return nil
}
