package aep

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
)

// SetDimensionsSeparated toggles AE's "Separate Dimensions" on a
// multidimensional Position leader. Alpha / structural — only the
// merge→separate direction on a static 3D Position is implemented; the
// reverse (separate→merge) and the 2D / animated cases are refused pending
// their own RE + ship-gate.
//
// Byte mechanics REd from AE 2020 (test_data/re_separate_dims_{before,after}.aep):
//
//   - The leader keeps its slot but flips tdsb byte2→0x08 + byte3 bit1
//     (dimensions_separated) and its value resets to the property default
//     ([compW/2, compH/2, 0] for Position). The leader value is no longer
//     authoritative once separated — AE reads the per-axis followers.
//   - The real X/Y/Z value migrates into the Position_0 / Position_1 /
//     Position_2 followers, each of which clears its own tdsb bit1.
//   - AE pre-allocates Position_0 / Position_1 (zeroed) even while merged; the
//     Z follower Position_2 does not exist until separation, so we synthesize
//     it by cloning the Position_1 tdmn+tdbs block (identical tdb4 / tdsn /
//     tdum / tduM) and writing the migrated Z into its cdat.
//
// Atomic: the only fallible step (re-parsing the synthesized Position_2 block)
// runs before any in-place mutation, so a failure leaves the project
// untouched and there is nothing to roll back.
func (p *Property) SetDimensionsSeparated(separated bool) error {
	if !separated {
		return fmt.Errorf("SetDimensionsSeparated(false): merge (separated→merged) not yet supported")
	}
	if p.MatchName != MatchNamePosition {
		return fmt.Errorf("SetDimensionsSeparated: only %q can be separated (got %q)", MatchNamePosition, p.MatchName)
	}
	if p.back == nil || p.back.tdsb == nil || p.back.cdat == nil {
		return fmt.Errorf("SetDimensionsSeparated: Position built outside parser (no tdsb/cdat back-ref)")
	}
	if p.DimensionsSeparated() {
		return fmt.Errorf("SetDimensionsSeparated: Position already separated")
	}
	if p.Components != 3 {
		return fmt.Errorf("SetDimensionsSeparated: only 3D Position supported (Components=%d)", p.Components)
	}
	if len(p.back.tdsb.Data) < 4 || len(p.back.cdat.Data) < 24 {
		return fmt.Errorf("SetDimensionsSeparated: Position tdsb/cdat too short (tdsb=%d cdat=%d)", len(p.back.tdsb.Data), len(p.back.cdat.Data))
	}
	xyz, ok := p.StaticValue.([]float64)
	if !ok || len(xyz) < 3 {
		return fmt.Errorf("SetDimensionsSeparated: Position is animated or non-static; only static 3D Position supported")
	}
	def, ok := p.DefaultValue.([]float64)
	if !ok || len(def) < 3 {
		return fmt.Errorf("SetDimensionsSeparated: Position DefaultValue unavailable")
	}

	grp := p.parentTreeGroup
	if grp == nil || grp.chunk == nil {
		return fmt.Errorf("SetDimensionsSeparated: Position has no owning tdgp group chunk")
	}
	layer := p.ownerLayer()
	if layer == nil {
		return fmt.Errorf("SetDimensionsSeparated: cannot reach owning layer")
	}

	pos0 := grp.Property(MatchNamePosition0)
	pos1 := grp.Property(MatchNamePosition1)
	if pos0 == nil || pos1 == nil {
		return fmt.Errorf("SetDimensionsSeparated: merged Position missing pre-allocated Position_0/_1 followers")
	}
	if grp.Property(MatchNamePosition2) != nil {
		return fmt.Errorf("SetDimensionsSeparated: Position_2 already present")
	}
	for _, f := range []*Property{pos0, pos1} {
		if f.back == nil || f.back.tdsb == nil || f.back.cdat == nil || f.back.tdbs == nil {
			return fmt.Errorf("SetDimensionsSeparated: follower %q missing back-refs", f.MatchName)
		}
		if len(f.back.tdsb.Data) < 4 || len(f.back.cdat.Data) < 8 {
			return fmt.Errorf("SetDimensionsSeparated: follower %q tdsb/cdat too short", f.MatchName)
		}
	}

	// Locate Position_1's tdmn+tdbs pair in the group LIST — clone source for
	// the synthesized Position_2 and anchor for the splice.
	groupChildren := grp.chunk.Children
	pos1TdbsIdx := indexOfChunk(groupChildren, pos1.back.tdbs)
	if pos1TdbsIdx < 1 {
		return fmt.Errorf("SetDimensionsSeparated: Position_1 tdbs not located in group LIST")
	}
	pos1Tdmn := groupChildren[pos1TdbsIdx-1]
	if pos1Tdmn.ID != rifx.IDTdmn || len(pos1Tdmn.Data) == 0 {
		return fmt.Errorf("SetDimensionsSeparated: expected tdmn before Position_1 tdbs, found %s", pos1Tdmn.ID)
	}

	// === Build the Position_2 block (clone of Position_1's pair) ===
	z := xyz[2]
	newTdmn := deepCloneChunk(pos1Tdmn)
	writeTdmnName(newTdmn, MatchNamePosition2)
	newTdbs := deepCloneChunk(pos1.back.tdbs)
	newCdat := newTdbs.FindFirst(rifx.IDCdat)
	newTdsb := newTdbs.FindFirst(rifx.IDTdsb)
	if newCdat == nil || newTdsb == nil || len(newCdat.Data) < 8 || len(newTdsb.Data) < 4 {
		return fmt.Errorf("SetDimensionsSeparated: cloned Position_2 block missing cdat/tdsb")
	}
	binary.BigEndian.PutUint64(newCdat.Data[0:8], math.Float64bits(z))
	newTdsb.Data[3] &^= 0x02 // clear dimensions_separated on the follower

	// === Re-parse the synthesized follower (only fallible step) ===
	var localWarnings []string
	ctx := newParseCtx(0, layer.Name, &localWarnings)
	pos2 := parseLeafProperty(MatchNamePosition2, newTdbs, ctx)
	if pos2 == nil {
		return fmt.Errorf("SetDimensionsSeparated: synthesized Position_2 failed to parse")
	}
	if len(localWarnings) > 0 {
		return fmt.Errorf("SetDimensionsSeparated: synthesized Position_2 produced parser warnings: %v", localWarnings)
	}
	pos2.DefaultValue = 0.0 // ADBE Position_2 fixed default
	pos2.parentTreeGroup = grp

	// === Commit: in-place byte mutations (all bounds pre-validated) ===
	// Leader: set separated flags + reset value to default.
	p.back.tdsb.Data[2] = 0x08
	p.back.tdsb.Data[3] |= 0x02
	binary.BigEndian.PutUint64(p.back.cdat.Data[0:8], math.Float64bits(def[0]))
	binary.BigEndian.PutUint64(p.back.cdat.Data[8:16], math.Float64bits(def[1]))
	binary.BigEndian.PutUint64(p.back.cdat.Data[16:24], math.Float64bits(def[2]))
	p.StaticValue = []float64{def[0], def[1], def[2]}

	// Followers X / Y: clear separated bit + migrate value.
	pos0.back.tdsb.Data[3] &^= 0x02
	binary.BigEndian.PutUint64(pos0.back.cdat.Data[0:8], math.Float64bits(xyz[0]))
	pos0.StaticValue = xyz[0]
	pos1.back.tdsb.Data[3] &^= 0x02
	binary.BigEndian.PutUint64(pos1.back.cdat.Data[0:8], math.Float64bits(xyz[1]))
	pos1.StaticValue = xyz[1]

	// Splice the Position_2 tdmn+tdbs pair into the group LIST right after
	// Position_1's tdbs (matches AE's on-disk ordering).
	insertAt := pos1TdbsIdx + 1
	spliced := make([]*rifx.Chunk, 0, len(groupChildren)+2)
	spliced = append(spliced, groupChildren[:insertAt]...)
	spliced = append(spliced, newTdmn, newTdbs)
	spliced = append(spliced, groupChildren[insertAt:]...)
	grp.chunk.Children = spliced

	// Mirror into the scene property tree + flat layer property list.
	insertChildAfter(grp, pos1, pos2)
	layer.Properties = append(layer.Properties, pos2)

	return nil
}

// writeTdmnName overwrites a tdmn chunk's NUL-padded match-name in place,
// preserving its original byte length.
func writeTdmnName(tdmn *rifx.Chunk, name string) {
	for i := range tdmn.Data {
		tdmn.Data[i] = 0
	}
	copy(tdmn.Data, name)
}

// insertChildAfter inserts child into grp.Children immediately after the
// existing direct child `after`; appends to the end if `after` is not found.
func insertChildAfter(grp *AEPropertyGroup, after PropertyBase, child PropertyBase) {
	idx := grp.PropertyIndex(after)
	if idx < 0 {
		grp.Children = append(grp.Children, child)
		return
	}
	at := idx + 1
	out := make([]PropertyBase, 0, len(grp.Children)+1)
	out = append(out, grp.Children[:at]...)
	out = append(out, child)
	out = append(out, grp.Children[at:]...)
	grp.Children = out
}
