package aep

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
)

// SetDimensionsSeparated toggles AE's "Separate Dimensions" on a Position
// leader. Alpha / structural. Both directions are implemented for a static
// Position; animated Position is refused pending its own RE + ship-gate.
//
// Byte mechanics REd from AE 2020 controlled before/after pairs (see
// test_data/re_separate_dims*.jsx + incidents/separate-dimensions-write-mechanics.md):
//
//   - separate (merge→separate): leader flips tdsb byte2→0x08 + byte3 bit1 and
//     resets to its default ([w/2,h/2,0]); the real value migrates into the
//     per-axis Position_0/1 (+ Position_2 for 3D layers) followers, each
//     clearing its own bit1. AE pre-allocates Position_0/1 even while merged;
//     the Z follower Position_2 is synthesized (clone of Position_1's
//     tdmn+tdbs) only for 3D layers — 2D layers separate into X/Y only.
//   - merge (separate→merged): leader clears tdsb byte2→0x00 + byte3 bit1 and
//     takes back the migrated [X,Y,Z] value; ALL Position_0/1/2 followers are
//     removed (AE's merged-after-separate form is leader-only).
//
// Atomicity: the only fallible step (re-parsing a synthesized Position_2)
// runs before any in-place mutation, so a failure leaves the project
// untouched and there is nothing to roll back.
func (p *Property) SetDimensionsSeparated(separated bool) error {
	if p.MatchName != MatchNamePosition {
		return fmt.Errorf("SetDimensionsSeparated: only %q can be separated (got %q)", MatchNamePosition, p.MatchName)
	}
	if p.back == nil || p.back.tdsb == nil || p.back.cdat == nil {
		return fmt.Errorf("SetDimensionsSeparated: Position built outside parser (no tdsb/cdat back-ref)")
	}
	if p.Components != 3 {
		return fmt.Errorf("SetDimensionsSeparated: only 3-component Position supported (Components=%d)", p.Components)
	}
	if len(p.back.tdsb.Data) < 4 || len(p.back.cdat.Data) < 24 {
		return fmt.Errorf("SetDimensionsSeparated: Position tdsb/cdat too short (tdsb=%d cdat=%d)", len(p.back.tdsb.Data), len(p.back.cdat.Data))
	}
	grp := p.parentTreeGroup
	if grp == nil || grp.chunk == nil {
		return fmt.Errorf("SetDimensionsSeparated: Position has no owning tdgp group chunk")
	}
	layer := p.ownerLayer()
	if layer == nil {
		return fmt.Errorf("SetDimensionsSeparated: cannot reach owning layer")
	}

	if separated {
		return p.separatePosition(grp, layer)
	}
	return p.mergePosition(grp)
}

// separatePosition splits the merged Position leader into per-axis followers.
// 3D layers get Position_0/1/2; 2D layers get Position_0/1 (no Z follower).
func (p *Property) separatePosition(grp *AEPropertyGroup, layer *Layer) error {
	if p.DimensionsSeparated() {
		return fmt.Errorf("SetDimensionsSeparated: Position already separated")
	}
	xyz, ok := p.StaticValue.([]float64)
	if !ok || len(xyz) < 3 {
		return fmt.Errorf("SetDimensionsSeparated: Position is animated or non-static; only static Position supported")
	}
	def, ok := p.DefaultValue.([]float64)
	if !ok || len(def) < 3 {
		return fmt.Errorf("SetDimensionsSeparated: Position DefaultValue unavailable")
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

	// Synthesize the Z follower up front (3D only) — the lone fallible step.
	var newTdmn, newTdbs *rifx.Chunk
	var pos2 *Property
	var insertAt int
	if layer.Is3D {
		groupChildren := grp.chunk.Children
		pos1TdbsIdx := indexOfChunk(groupChildren, pos1.back.tdbs)
		if pos1TdbsIdx < 1 {
			return fmt.Errorf("SetDimensionsSeparated: Position_1 tdbs not located in group LIST")
		}
		pos1Tdmn := groupChildren[pos1TdbsIdx-1]
		if pos1Tdmn.ID != rifx.IDTdmn || len(pos1Tdmn.Data) == 0 {
			return fmt.Errorf("SetDimensionsSeparated: expected tdmn before Position_1 tdbs, found %s", pos1Tdmn.ID)
		}
		newTdmn = deepCloneChunk(pos1Tdmn)
		writeTdmnName(newTdmn, MatchNamePosition2)
		newTdbs = deepCloneChunk(pos1.back.tdbs)
		newCdat := newTdbs.FindFirst(rifx.IDCdat)
		newTdsb := newTdbs.FindFirst(rifx.IDTdsb)
		if newCdat == nil || newTdsb == nil || len(newCdat.Data) < 8 || len(newTdsb.Data) < 4 {
			return fmt.Errorf("SetDimensionsSeparated: cloned Position_2 block missing cdat/tdsb")
		}
		binary.BigEndian.PutUint64(newCdat.Data[0:8], math.Float64bits(xyz[2]))
		newTdsb.Data[3] &^= 0x02

		var localWarnings []string
		ctx := newParseCtx(0, layer.Name, &localWarnings)
		pos2 = parseLeafProperty(MatchNamePosition2, newTdbs, ctx)
		if pos2 == nil {
			return fmt.Errorf("SetDimensionsSeparated: synthesized Position_2 failed to parse")
		}
		if len(localWarnings) > 0 {
			return fmt.Errorf("SetDimensionsSeparated: synthesized Position_2 produced parser warnings: %v", localWarnings)
		}
		pos2.DefaultValue = 0.0 // ADBE Position_2 fixed default
		pos2.parentTreeGroup = grp
		insertAt = pos1TdbsIdx + 1
	}

	// === Commit: in-place byte mutations (all bounds pre-validated) ===
	p.back.tdsb.Data[2] = 0x08
	p.back.tdsb.Data[3] |= 0x02
	binary.BigEndian.PutUint64(p.back.cdat.Data[0:8], math.Float64bits(def[0]))
	binary.BigEndian.PutUint64(p.back.cdat.Data[8:16], math.Float64bits(def[1]))
	binary.BigEndian.PutUint64(p.back.cdat.Data[16:24], math.Float64bits(def[2]))
	p.StaticValue = []float64{def[0], def[1], def[2]}

	pos0.back.tdsb.Data[3] &^= 0x02
	binary.BigEndian.PutUint64(pos0.back.cdat.Data[0:8], math.Float64bits(xyz[0]))
	pos0.StaticValue = xyz[0]
	pos1.back.tdsb.Data[3] &^= 0x02
	binary.BigEndian.PutUint64(pos1.back.cdat.Data[0:8], math.Float64bits(xyz[1]))
	pos1.StaticValue = xyz[1]

	if pos2 != nil {
		groupChildren := grp.chunk.Children
		spliced := make([]*rifx.Chunk, 0, len(groupChildren)+2)
		spliced = append(spliced, groupChildren[:insertAt]...)
		spliced = append(spliced, newTdmn, newTdbs)
		spliced = append(spliced, groupChildren[insertAt:]...)
		grp.chunk.Children = spliced
		insertChildAfter(grp, pos1, pos2)
		layer.Properties = append(layer.Properties, pos2)
	}

	return nil
}

// mergePosition collapses separated per-axis followers back into the Position
// leader: the leader takes the [X,Y,Z] value and every Position_0/1/2 follower
// chunk is removed (AE's merged-after-separate form is leader-only).
func (p *Property) mergePosition(grp *AEPropertyGroup) error {
	if !p.DimensionsSeparated() {
		return fmt.Errorf("SetDimensionsSeparated: Position already merged")
	}

	var followers []*Property
	axisVal := [3]float64{} // Z stays 0 when no Position_2 (2D)
	for axis, mn := range []string{MatchNamePosition0, MatchNamePosition1, MatchNamePosition2} {
		f := grp.Property(mn)
		if f == nil {
			continue
		}
		if f.back == nil || f.back.tdbs == nil {
			return fmt.Errorf("SetDimensionsSeparated: follower %q missing tdbs back-ref", mn)
		}
		v, ok := f.StaticValue.(float64)
		if !ok {
			return fmt.Errorf("SetDimensionsSeparated: follower %q is animated or non-scalar; only static Position supported", mn)
		}
		axisVal[axis] = v
		followers = append(followers, f)
	}
	if len(followers) < 2 {
		return fmt.Errorf("SetDimensionsSeparated: separated Position has %d followers, expected >= 2", len(followers))
	}

	// === Commit: leader takes the value + clears separated flags ===
	p.back.tdsb.Data[2] = 0x00
	p.back.tdsb.Data[3] &^= 0x02
	binary.BigEndian.PutUint64(p.back.cdat.Data[0:8], math.Float64bits(axisVal[0]))
	binary.BigEndian.PutUint64(p.back.cdat.Data[8:16], math.Float64bits(axisVal[1]))
	binary.BigEndian.PutUint64(p.back.cdat.Data[16:24], math.Float64bits(axisVal[2]))
	p.StaticValue = []float64{axisVal[0], axisVal[1], axisVal[2]}

	// Remove every follower's tdmn+tdbs pair from the group LIST + scene tree.
	remove := make(map[*rifx.Chunk]bool, len(followers)*2)
	for _, f := range followers {
		idx := indexOfChunk(grp.chunk.Children, f.back.tdbs)
		if idx < 0 {
			continue
		}
		remove[f.back.tdbs] = true
		if idx >= 1 && grp.chunk.Children[idx-1].ID == rifx.IDTdmn {
			remove[grp.chunk.Children[idx-1]] = true
		}
	}
	keep := grp.chunk.Children[:0:0]
	for _, ch := range grp.chunk.Children {
		if !remove[ch] {
			keep = append(keep, ch)
		}
	}
	grp.chunk.Children = keep

	removeFollowers := make(map[PropertyBase]bool, len(followers))
	for _, f := range followers {
		removeFollowers[f] = true
	}
	grp.Children = filterPropertyBase(grp.Children, removeFollowers)
	filterLayerProperties(grp, followers)

	return nil
}

// filterLayerProperties drops the given followers from the owning layer's flat
// Property slice (by pointer identity).
func filterLayerProperties(g *AEPropertyGroup, drop []*Property) {
	var layer *Layer
	for cur := g; cur != nil; cur = cur.parent {
		if cur.layer != nil {
			layer = cur.layer
			break
		}
	}
	if layer == nil {
		return
	}
	dropSet := make(map[*Property]bool, len(drop))
	for _, d := range drop {
		dropSet[d] = true
	}
	kept := layer.Properties[:0:0]
	for _, pr := range layer.Properties {
		if !dropSet[pr] {
			kept = append(kept, pr)
		}
	}
	layer.Properties = kept
}

// filterPropertyBase returns children with the dropped nodes removed.
func filterPropertyBase(children []PropertyBase, drop map[PropertyBase]bool) []PropertyBase {
	out := children[:0:0]
	for _, c := range children {
		if !drop[c] {
			out = append(out, c)
		}
	}
	return out
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
