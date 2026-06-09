package aep

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
)

// SetDimensionsSeparated toggles AE's "Separate Dimensions" on a Position
// leader. Structural; both directions (separate↔merge) are double-version
// ship-gated (AE 2020 + 2025) for static Position (2D + 3D) and animated
// Position (3D layers, near-linear leader path-ease). An animated leader
// routes to the keyframe-stream migration paths (separatePositionAnimated /
// mergePositionAnimated); animated cases outside that shipped subset — a 2D
// layer, or a leader carrying custom spatial-path temporal ease — are refused
// with an error rather than written.
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
//
// Free function (not a method) so the impl can live in internal/serializer after
// the M8 split (CLAUDE.md #2 lists SetDimensionsSeparated as a structural write path
// despite the Set prefix — it adds/removes follower Property nodes); the aep facade
// re-exports it. BREAKING vs the former Property.SetDimensionsSeparated method form.
func SetDimensionsSeparated(p *Property, separated bool) error {
	if p.MatchName != MatchNamePosition {
		return fmt.Errorf("SetDimensionsSeparated: only %q can be separated (got %q)", MatchNamePosition, p.MatchName)
	}
	pb := p.propertyBack()
	if pb == nil || pb.tdsb == nil {
		return fmt.Errorf("SetDimensionsSeparated: Position built outside parser (no tdsb back-ref)")
	}
	if p.Components != 3 {
		return fmt.Errorf("SetDimensionsSeparated: only 3-component Position supported (Components=%d)", p.Components)
	}
	if len(pb.tdsb.Data) < 4 {
		return fmt.Errorf("SetDimensionsSeparated: Position tdsb too short (tdsb=%d)", len(pb.tdsb.Data))
	}
	grp := p.parentTreeGroup
	if grp == nil || grp.propertyGroupBack().chunk == nil {
		return fmt.Errorf("SetDimensionsSeparated: Position has no owning tdgp group chunk")
	}
	layer := p.ownerLayer()
	if layer == nil {
		return fmt.Errorf("SetDimensionsSeparated: cannot reach owning layer")
	}

	// An animated leader carries a keyframe stream (no cdat) instead of a
	// static value; it routes to separatePositionAnimated. The merge direction
	// of an animated-separated Position (leader is static-default with a cdat,
	// followers animated) falls through to mergePosition, which detects the
	// animated followers and routes to mergePositionAnimated.
	if separated && pb.cdat == nil && len(p.Keyframes) > 0 {
		return separatePositionAnimated(p, grp, layer)
	}
	if pb.cdat == nil || len(pb.cdat.Data) < 24 {
		return fmt.Errorf("SetDimensionsSeparated: Position cdat too short or absent (cdat=%v)", pb.cdat != nil)
	}

	if separated {
		return separatePosition(p, grp, layer)
	}
	return mergePosition(p, grp)
}

// separatePosition splits the merged Position leader into per-axis followers.
// 3D layers get Position_0/1/2; 2D layers get Position_0/1 (no Z follower).
func separatePosition(p *Property, grp *AEPropertyGroup, layer *Layer) error {
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
		fb := f.propertyBack()
		if fb == nil || fb.tdsb == nil || fb.cdat == nil || fb.tdbs == nil {
			return fmt.Errorf("SetDimensionsSeparated: follower %q missing back-refs", f.MatchName)
		}
		if len(fb.tdsb.Data) < 4 || len(fb.cdat.Data) < 8 {
			return fmt.Errorf("SetDimensionsSeparated: follower %q tdsb/cdat too short", f.MatchName)
		}
	}
	pb, pos0b, pos1b := p.propertyBack(), pos0.propertyBack(), pos1.propertyBack()

	// Synthesize the Z follower up front (3D only) — the lone fallible step.
	var newTdmn, newTdbs *rifx.Chunk
	var pos2 *Property
	var insertAt int
	if layer.Is3D {
		groupChildren := grp.propertyGroupBack().chunk.Children
		pos1TdbsIdx := indexOfChunk(groupChildren, pos1b.tdbs)
		if pos1TdbsIdx < 1 {
			return fmt.Errorf("SetDimensionsSeparated: Position_1 tdbs not located in group LIST")
		}
		pos1Tdmn := groupChildren[pos1TdbsIdx-1]
		if pos1Tdmn.ID != rifx.IDTdmn || len(pos1Tdmn.Data) == 0 {
			return fmt.Errorf("SetDimensionsSeparated: expected tdmn before Position_1 tdbs, found %s", pos1Tdmn.ID)
		}
		newTdmn = deepCloneChunk(pos1Tdmn)
		writeTdmnName(newTdmn, MatchNamePosition2)
		newTdbs = deepCloneChunk(pos1b.tdbs)
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
	pb.tdsb.Data[2] = 0x08
	pb.tdsb.Data[3] |= 0x02
	binary.BigEndian.PutUint64(pb.cdat.Data[0:8], math.Float64bits(def[0]))
	binary.BigEndian.PutUint64(pb.cdat.Data[8:16], math.Float64bits(def[1]))
	binary.BigEndian.PutUint64(pb.cdat.Data[16:24], math.Float64bits(def[2]))
	p.StaticValue = []float64{def[0], def[1], def[2]}

	pos0b.tdsb.Data[3] &^= 0x02
	binary.BigEndian.PutUint64(pos0b.cdat.Data[0:8], math.Float64bits(xyz[0]))
	pos0.StaticValue = xyz[0]
	pos1b.tdsb.Data[3] &^= 0x02
	binary.BigEndian.PutUint64(pos1b.cdat.Data[0:8], math.Float64bits(xyz[1]))
	pos1.StaticValue = xyz[1]

	if pos2 != nil {
		groupChildren := grp.propertyGroupBack().chunk.Children
		spliced := make([]*rifx.Chunk, 0, len(groupChildren)+2)
		spliced = append(spliced, groupChildren[:insertAt]...)
		spliced = append(spliced, newTdmn, newTdbs)
		spliced = append(spliced, groupChildren[insertAt:]...)
		grp.propertyGroupBack().chunk.Children = spliced
		insertChildAfter(grp, pos1, pos2)
		layer.Properties = append(layer.Properties, pos2)
	}

	return nil
}

// separatePositionAnimated splits an ANIMATED merged Position leader (a 3D
// spatial motion-path keyframe stream) into three per-axis animated 1D
// temporal followers, then collapses the leader to its static default — the
// stream-migration analogue of separatePosition.
//
// Byte mechanics REd from AE 2020 re_sepdim_anim_{before,after}.aep
// (incidents/separate-dimensions-write-mechanics.md §animated):
//
//   - per follower keyframe i, axis a:
//       value     = leader.kf[i].Value[a]
//       out_speed = leader.kf[i].OutSpatialTangent[a] × 100
//       in_speed  = −leader.kf[i].InSpatialTangent[a] × 100
//       influence = 0.01 on a side that has an adjacent segment, 0 at the
//                   first-kf in-side / last-kf out-side boundary
//       in/out interp = bezier  (block header07 = 0x08, bpk = 48)
//   - followers convert static→animated: tdb4 @0x05 clears bit0 + @0x44=0x01,
//     tdsb clears bit1, cdat → LIST(kfl)(lhd3+ldat).
//   - leader collapses animated→static: tdb4 @0x05 sets bit0, @0x44=0x00,
//     @0x4f sets bit0; tdsb → separated (byte2=0x08, byte3 bit1); the kf
//     stream is replaced by a 72-byte cdat = [default(3), kf0.inSpatTan(3),
//     kf0.outSpatTan(3)].
//
// First slice limited to 3D layers with ~linear leader path-ease (no custom
// temporal ease on the leader's spatial keyframes); other cases refuse.
//
// Atomicity: all fallible work (validation + Position_2 synthesis/re-parse)
// runs before any in-place byte mutation, so failure leaves the project
// untouched and there is nothing to roll back.
func separatePositionAnimated(p *Property, grp *AEPropertyGroup, layer *Layer) error {
	if p.DimensionsSeparated() {
		return fmt.Errorf("SetDimensionsSeparated: Position already separated")
	}
	if !layer.Is3D {
		return fmt.Errorf("SetDimensionsSeparated: animated Position separate currently supports 3D layers only")
	}
	pb := p.propertyBack()
	if pb == nil || pb.tdbs == nil || pb.tdb4 == nil || len(pb.tdb4.Data) <= 0x4f {
		return fmt.Errorf("SetDimensionsSeparated: animated leader missing tdbs/tdb4 back-refs")
	}
	def, ok := p.DefaultValue.([]float64)
	if !ok || len(def) < 3 {
		return fmt.Errorf("SetDimensionsSeparated: Position DefaultValue unavailable")
	}

	kfs := p.Keyframes
	tickRate := 0.0
	if kfs[0].back != nil {
		if kb, ok := kfs[0].back.(*keyframeBackrefs); ok {
			tickRate = kb.tickRate
		}
	}
	if tickRate <= 0 {
		return fmt.Errorf("SetDimensionsSeparated: animated leader tickRate unavailable")
	}
	// Every leader keyframe must be a 3D spatial sample with ~linear path-ease.
	for i, kf := range kfs {
		v, ok := kf.Value.([]float64)
		if !ok || len(v) < 3 {
			return fmt.Errorf("SetDimensionsSeparated: leader kf%d value not 3D: %v", i, kf.Value)
		}
		if len(kf.InSpatialTangent) < 3 || len(kf.OutSpatialTangent) < 3 {
			return fmt.Errorf("SetDimensionsSeparated: leader kf%d missing spatial tangents", i)
		}
		if !temporalEaseLinear(kf.InTemporalEase) || !temporalEaseLinear(kf.OutTemporalEase) {
			return fmt.Errorf("SetDimensionsSeparated: leader kf%d has non-default path temporal ease; animated separate limited to linear path-ease (first slice)", i)
		}
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
		fb := f.propertyBack()
		if fb == nil || fb.tdsb == nil || fb.cdat == nil || fb.tdbs == nil || fb.tdb4 == nil {
			return fmt.Errorf("SetDimensionsSeparated: follower %q missing back-refs", f.MatchName)
		}
		if len(fb.tdb4.Data) <= 0x44 || len(fb.tdsb.Data) < 4 {
			return fmt.Errorf("SetDimensionsSeparated: follower %q tdb4/tdsb too short", f.MatchName)
		}
	}
	pos1b := pos1.propertyBack()

	// Locate the leader's kf stream LIST + the Position_1 tdmn/tdbs splice
	// point up front (pre-commit; refuse rather than half-mutate).
	leaderKflIdx := -1
	for j, ch := range pb.tdbs.Children {
		if ch.IsList() && ch.FormType == rifx.IDkfl {
			leaderKflIdx = j
			break
		}
	}
	if leaderKflIdx < 0 {
		return fmt.Errorf("SetDimensionsSeparated: animated leader has no kf stream to collapse")
	}
	groupChildren := grp.propertyGroupBack().chunk.Children
	pos1TdbsIdx := indexOfChunk(groupChildren, pos1b.tdbs)
	if pos1TdbsIdx < 1 {
		return fmt.Errorf("SetDimensionsSeparated: Position_1 tdbs not located in group LIST")
	}
	pos1Tdmn := groupChildren[pos1TdbsIdx-1]
	if pos1Tdmn.ID != rifx.IDTdmn || len(pos1Tdmn.Data) == 0 {
		return fmt.Errorf("SetDimensionsSeparated: expected tdmn before Position_1 tdbs, found %s", pos1Tdmn.ID)
	}

	// Build per-axis keyframe streams (pure construction, no mutation yet).
	kfl0 := buildSeparatedAxisKfl(kfs, 0, tickRate)
	kfl1 := buildSeparatedAxisKfl(kfs, 1, tickRate)
	kfl2 := buildSeparatedAxisKfl(kfs, 2, tickRate)

	// Synthesize the animated Z follower (the lone fallible step): clone
	// Position_1's still-static tdbs, rename, convert to animated, re-parse.
	newTdmn := deepCloneChunk(pos1Tdmn)
	writeTdmnName(newTdmn, MatchNamePosition2)
	newTdbs := deepCloneChunk(pos1b.tdbs)
	if err := convertFollowerTdbsToAnimated(newTdbs, kfl2); err != nil {
		return fmt.Errorf("SetDimensionsSeparated: synthesize Position_2: %w", err)
	}
	var localWarnings []string
	ctx := newParseCtx(tickRate, layer.Name, &localWarnings)
	pos2 := parseLeafProperty(MatchNamePosition2, newTdbs, ctx)
	if pos2 == nil {
		return fmt.Errorf("SetDimensionsSeparated: synthesized Position_2 failed to parse")
	}
	if len(localWarnings) > 0 {
		return fmt.Errorf("SetDimensionsSeparated: synthesized Position_2 produced parser warnings: %v", localWarnings)
	}
	if len(pos2.Keyframes) != len(kfs) {
		return fmt.Errorf("SetDimensionsSeparated: synthesized Position_2 has %d kf, want %d", len(pos2.Keyframes), len(kfs))
	}
	pos2.DefaultValue = 0.0 // ADBE Position_2 fixed default
	pos2.parentTreeGroup = grp
	insertAt := pos1TdbsIdx + 1

	// Leader's replacement static cdat (72B): default + kf0 spatial tangents.
	leaderCdat := buildAnimatedLeaderStaticCdat(def, kfs[0])

	// === Commit: in-place byte mutations (all bounds pre-validated) ===
	// Leader: animated → static-default + separated flags.
	pb.tdsb.Data[2] = 0x08
	pb.tdsb.Data[3] |= 0x02
	pb.tdb4.Data[0x05] |= 0x01
	pb.tdb4.Data[0x44] = 0x00
	pb.tdb4.Data[0x4f] |= 0x01
	pb.tdbs.Children[leaderKflIdx] = leaderCdat
	pb.cdat = leaderCdat
	pb.lhd3 = nil
	pb.ldat = nil
	pb.bytesPerKF = 0
	p.Keyframes = nil
	p.StaticValue = []float64{def[0], def[1], def[2]}

	// Followers Position_0/_1: static → animated, in place.
	convertFollowerToAnimated(pos0, kfl0, ctx)
	convertFollowerToAnimated(pos1, kfl1, ctx)

	// Splice the synthesized Position_2 chunks + scene node after Position_1.
	spliced := make([]*rifx.Chunk, 0, len(groupChildren)+2)
	spliced = append(spliced, groupChildren[:insertAt]...)
	spliced = append(spliced, newTdmn, newTdbs)
	spliced = append(spliced, groupChildren[insertAt:]...)
	grp.propertyGroupBack().chunk.Children = spliced
	insertChildAfter(grp, pos1, pos2)
	layer.Properties = append(layer.Properties, pos2)

	return nil
}

// temporalEaseLinear reports whether a side's temporal ease is the AE-default
// "linear path" ease (speed + influence both ≈ 0) across all components.
func temporalEaseLinear(es []TemporalEase) bool {
	for _, e := range es {
		if math.Abs(e.Speed) > 1e-9 || math.Abs(e.Influence) > 1e-9 {
			return false
		}
	}
	return true
}

// aeSepDimSpeedFactor is AE's per-axis speed conversion for separated
// dimensions: speed = (centralDifference of axis values) × 100, i.e.
// speed = Δ × (100/6). AE's stored factor is NOT the clean f64 100.0/6.0
// (0x4030aaaaaaaaaaab) — empirically it is 0x4030aaaaaaac192b
// (≈16.666666666999998, the f64 one ULP below the literal 16.666666667),
// carrying a ~2e-11 relative offset that comes from AE's internal tick-time
// quantization. Using it reproduces AE's stored speeds exactly for most Δ
// (200/−50/400/650) and within 1 ULP for the rest (e.g. Δ=300) — vs ~2e-7
// off for a clean Δ/6×100. time / value / influence are byte-identical to AE
// regardless; the ≤1-ULP speed residual is AE's per-keyframe tick rounding,
// not replicated here. RE'd from re_sepdim_anim{,2,3}_after.aep (AE 2020).
var aeSepDimSpeedFactor = math.Float64frombits(0x4030aaaaaaac192b)

// buildSeparatedAxisKfl builds a LIST(kfl)(lhd3 + ldat) for one axis of a
// separated Position, mapping the leader's 3D spatial keyframes to 1D temporal
// keyframes (bpk=48, header07=0x08). See separatePositionAnimated for the map.
//
// The per-axis temporal speed AE writes is NOT the leader's stored spatial
// tangent (which is timing-weighted / asymmetric for non-uniform spacing).
// It is the timing-independent central difference of the axis VALUES:
//
//	speed[i] = (v[min(n-1,i+1)] − v[max(0,i-1)]) / 6 × 100  (same on in+out side)
//
// while the per-side influence carries the timing:
//
//	in_influence[i]  = 0.01 / (t[i] − t[i-1])   (0 at the first keyframe)
//	out_influence[i] = 0.01 / (t[i+1] − t[i])   (0 at the last keyframe)
//
// RE'd byte-exact across three AE 2020 fixtures: uniform 1.0s, non-uniform
// 0.5/1.0/1.5s, uniform 0.5s (re_sepdim_anim{,2,3}_after.aep).
func buildSeparatedAxisKfl(leaderKfs []*Keyframe, axis int, tickRate float64) *rifx.Chunk {
	const bpk = 48
	n := len(leaderKfs)

	val := func(i int) float64 { return leaderKfs[i].Value.([]float64)[axis] }

	lhd3 := make([]byte, 52)
	lhd3[1], lhd3[2], lhd3[3] = 0xd0, 0x0b, 0xee
	binary.BigEndian.PutUint32(lhd3[0x08:0x0C], uint32(n))
	binary.BigEndian.PutUint32(lhd3[0x0C:0x10], 1)
	binary.BigEndian.PutUint32(lhd3[0x10:0x14], bpk)
	binary.BigEndian.PutUint32(lhd3[0x14:0x18], 4)
	binary.BigEndian.PutUint32(lhd3[0x18:0x1C], 1)
	binary.BigEndian.PutUint32(lhd3[0x1C:0x20], 4)

	ldat := make([]byte, n*bpk)
	for i, kf := range leaderKfs {
		blk := ldat[i*bpk : (i+1)*bpk]

		lo, hi := i-1, i+1
		if lo < 0 {
			lo = 0
		}
		if hi > n-1 {
			hi = n - 1
		}
		speed := (val(hi) - val(lo)) * aeSepDimSpeedFactor

		inInf, outInf := 0.0, 0.0
		if i > 0 {
			inInf = 0.01 / (kf.Time - leaderKfs[i-1].Time)
		}
		if i < n-1 {
			outInf = 0.01 / (leaderKfs[i+1].Time - kf.Time)
		}

		binary.BigEndian.PutUint32(blk[0x00:0x04], uint32(math.Round(kf.Time*tickRate)))
		blk[0x04] = byte(InterpBezier)
		blk[0x05] = byte(InterpBezier)
		blk[0x06] = 0x00
		blk[0x07] = 0x08
		binary.BigEndian.PutUint64(blk[0x08:0x10], math.Float64bits(val(i)))
		binary.BigEndian.PutUint64(blk[0x10:0x18], math.Float64bits(speed))
		binary.BigEndian.PutUint64(blk[0x18:0x20], math.Float64bits(inInf))
		binary.BigEndian.PutUint64(blk[0x20:0x28], math.Float64bits(speed))
		binary.BigEndian.PutUint64(blk[0x28:0x30], math.Float64bits(outInf))
	}

	kfl := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDkfl}
	kfl.Children = append(kfl.Children,
		&rifx.Chunk{ID: rifx.IDLhd3, Data: lhd3},
		&rifx.Chunk{ID: rifx.IDLdat, Data: ldat},
	)
	return kfl
}

// buildAnimatedLeaderStaticCdat builds the 72-byte cdat AE synthesizes when an
// animated 3D Position collapses to a separated static default: the default
// value (3 f64) followed by the first keyframe's in/out spatial tangents.
func buildAnimatedLeaderStaticCdat(def []float64, kf0 *Keyframe) *rifx.Chunk {
	d := make([]byte, 72)
	for i := 0; i < 3; i++ {
		binary.BigEndian.PutUint64(d[i*8:i*8+8], math.Float64bits(def[i]))
		binary.BigEndian.PutUint64(d[24+i*8:24+i*8+8], math.Float64bits(kf0.InSpatialTangent[i]))
		binary.BigEndian.PutUint64(d[48+i*8:48+i*8+8], math.Float64bits(kf0.OutSpatialTangent[i]))
	}
	return &rifx.Chunk{ID: rifx.IDCdat, Data: d}
}

// findTdb4Chunk returns the property metadata chunk under a tdbs, accepting
// either the modern lowercase "tdb4" or the legacy uppercase "Tdb4" ID.
func findTdb4Chunk(tdbs *rifx.Chunk) *rifx.Chunk {
	if t := tdbs.FindFirst(rifx.ChunkID{'t', 'd', 'b', '4'}); t != nil {
		return t
	}
	return tdbs.FindFirst(rifx.IDTdb4)
}

// convertFollowerTdbsToAnimated rewrites a static per-axis follower tdbs into
// its animated form in place: flips the tdb4 static→animated flags, clears the
// tdsb dimensions-separated bit, and swaps the cdat child for the kf stream.
// Used for the synthesized Position_2 clone (pre-commit, hence fallible).
func convertFollowerTdbsToAnimated(tdbs, kfl *rifx.Chunk) error {
	tdb4 := findTdb4Chunk(tdbs)
	if tdb4 == nil || len(tdb4.Data) <= 0x44 {
		return fmt.Errorf("follower tdbs missing/short tdb4")
	}
	tdb4.Data[0x05] &^= 0x01
	tdb4.Data[0x44] = 0x01
	if tdsb := tdbs.FindFirst(rifx.IDTdsb); tdsb != nil && len(tdsb.Data) >= 4 {
		tdsb.Data[3] &^= 0x02
	}
	for j, ch := range tdbs.Children {
		if ch.ID == rifx.IDCdat {
			tdbs.Children[j] = kfl
			return nil
		}
	}
	return fmt.Errorf("follower tdbs has no cdat to replace")
}

// convertFollowerToAnimated converts an existing static per-axis follower
// Property to animated in place (chunks + scene state), using its parsed
// back-refs. All bounds are pre-validated by the caller, so it is infallible.
func convertFollowerToAnimated(f *Property, kfl *rifx.Chunk, ctx *parseCtx) {
	fb := f.propertyBack()
	fb.tdb4.Data[0x05] &^= 0x01
	fb.tdb4.Data[0x44] = 0x01
	fb.tdsb.Data[3] &^= 0x02
	for j, ch := range fb.tdbs.Children {
		if ch.ID == rifx.IDCdat {
			fb.tdbs.Children[j] = kfl
			break
		}
	}
	fb.cdat = nil
	f.StaticValue = nil
	parseKeyframes(f, kfl.FindFirst(rifx.IDLhd3), kfl.FindFirst(rifx.IDLdat), ctx)
}

// mergePosition collapses separated per-axis followers back into the Position
// leader: the leader takes the [X,Y,Z] value and every Position_0/1/2 follower
// chunk is removed (AE's merged-after-separate form is leader-only).
func mergePosition(p *Property, grp *AEPropertyGroup) error {
	if !p.DimensionsSeparated() {
		return fmt.Errorf("SetDimensionsSeparated: Position already merged")
	}

	// Animated followers route to the stream-merge path.
	for _, mn := range []string{MatchNamePosition0, MatchNamePosition1, MatchNamePosition2} {
		if f := grp.Property(mn); f != nil && f.IsAnimated() {
			return mergePositionAnimated(p, grp)
		}
	}

	var followers []*Property
	axisVal := [3]float64{} // Z stays 0 when no Position_2 (2D)
	for axis, mn := range []string{MatchNamePosition0, MatchNamePosition1, MatchNamePosition2} {
		f := grp.Property(mn)
		if f == nil {
			continue
		}
		if fb := f.propertyBack(); fb == nil || fb.tdbs == nil {
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
	pb := p.propertyBack()
	pb.tdsb.Data[2] = 0x00
	pb.tdsb.Data[3] &^= 0x02
	binary.BigEndian.PutUint64(pb.cdat.Data[0:8], math.Float64bits(axisVal[0]))
	binary.BigEndian.PutUint64(pb.cdat.Data[8:16], math.Float64bits(axisVal[1]))
	binary.BigEndian.PutUint64(pb.cdat.Data[16:24], math.Float64bits(axisVal[2]))
	p.StaticValue = []float64{axisVal[0], axisVal[1], axisVal[2]}

	removeFollowerChunks(grp, followers)
	return nil
}

// mergePositionAnimated collapses ANIMATED per-axis followers back into the
// Position leader by rebuilding a single 3D spatial motion-path keyframe
// stream — the inverse of separatePositionAnimated:
//
//	leader.kf[i].Value                   = [pos0,pos1,pos2 value at kf i]
//	leader.kf[i].OutSpatialTangent[axis] =  follower[axis].kf[i].out_speed / 100
//	leader.kf[i].InSpatialTangent[axis]  = −follower[axis].kf[i].in_speed  / 100
//	leader temporal ease = 0; interp linear (path rebuilt linear).
//
// The leader spatial block's @0x08 / @0x10 (segment count / arc-length) are
// recompute-on-load cache fields — AE writes inconsistent values there and does
// not validate them (confirmed against AE's own merge output
// re_sepdim_anim_merge_after.aep: @0x08 = 0/1/0, @0x10 constant), so they are
// emitted as 0. Followers are removed; AE re-allocates zeroed placeholders on
// load, same as the static merge.
//
// First slice: 3D (Position_0/1/2 present), keyframe times aligned across axes.
//
// Atomicity: all fallible work (validation + stream construction + locating the
// leader cdat) runs before any in-place mutation.
func mergePositionAnimated(p *Property, grp *AEPropertyGroup) error {
	pb := p.propertyBack()
	if pb == nil || pb.tdbs == nil || pb.tdb4 == nil || len(pb.tdb4.Data) <= 0x4f {
		return fmt.Errorf("SetDimensionsSeparated: separated leader missing tdbs/tdb4 back-refs")
	}
	pos0 := grp.Property(MatchNamePosition0)
	pos1 := grp.Property(MatchNamePosition1)
	pos2 := grp.Property(MatchNamePosition2)
	if pos0 == nil || pos1 == nil || pos2 == nil {
		return fmt.Errorf("SetDimensionsSeparated: animated Position merge currently requires 3D (Position_0/1/2 present)")
	}
	followers := []*Property{pos0, pos1, pos2}
	n := len(pos0.Keyframes)
	if n == 0 {
		return fmt.Errorf("SetDimensionsSeparated: animated follower Position_0 has no keyframes")
	}
	tickRate := 0.0
	if pos0.Keyframes[0].back != nil {
		if kb, ok := pos0.Keyframes[0].back.(*keyframeBackrefs); ok {
			tickRate = kb.tickRate
		}
	}
	if tickRate <= 0 {
		return fmt.Errorf("SetDimensionsSeparated: animated follower tickRate unavailable")
	}
	for ai, f := range followers {
		if fb := f.propertyBack(); fb == nil || fb.tdbs == nil {
			return fmt.Errorf("SetDimensionsSeparated: follower %q missing tdbs back-ref", f.MatchName)
		}
		if len(f.Keyframes) != n {
			return fmt.Errorf("SetDimensionsSeparated: follower axis %d has %d kf, want %d (misaligned keyframes unsupported)", ai, len(f.Keyframes), n)
		}
		for i, kf := range f.Keyframes {
			if _, ok := kf.Value.(float64); !ok {
				return fmt.Errorf("SetDimensionsSeparated: follower axis %d kf%d value not scalar", ai, i)
			}
			if math.Abs(kf.Time-pos0.Keyframes[i].Time) > 1e-9 {
				return fmt.Errorf("SetDimensionsSeparated: follower axis %d kf%d time misaligned with axis 0 (first slice requires aligned keyframes)", ai, i)
			}
			if len(kf.InTemporalEase) < 1 || len(kf.OutTemporalEase) < 1 {
				return fmt.Errorf("SetDimensionsSeparated: follower axis %d kf%d missing temporal ease", ai, i)
			}
		}
	}

	leaderKfl := buildMergedLeaderKfl(followers, n, tickRate)

	leaderCdatIdx := -1
	for j, ch := range pb.tdbs.Children {
		if ch.ID == rifx.IDCdat {
			leaderCdatIdx = j
			break
		}
	}
	if leaderCdatIdx < 0 {
		return fmt.Errorf("SetDimensionsSeparated: separated leader has no cdat to replace")
	}

	// === Commit: static-default leader → animated; clear separated flags. ===
	pb.tdsb.Data[2] = 0x00
	pb.tdsb.Data[3] &^= 0x02
	pb.tdb4.Data[0x05] &^= 0x01
	pb.tdb4.Data[0x44] = 0x01
	pb.tdb4.Data[0x4f] &^= 0x01
	pb.tdbs.Children[leaderCdatIdx] = leaderKfl
	pb.cdat = nil
	p.StaticValue = nil
	var warns []string
	ctx := newParseCtx(tickRate, "", &warns)
	parseKeyframes(p, leaderKfl.FindFirst(rifx.IDLhd3), leaderKfl.FindFirst(rifx.IDLdat), ctx)

	removeFollowerChunks(grp, followers)
	return nil
}

// buildMergedLeaderKfl builds the leader's 3D spatial motion-path keyframe
// stream (bpk=128, header07=0x07) from the per-axis animated followers. See
// mergePositionAnimated for the tangent map; @0x08/@0x10 are emitted 0
// (AE-recomputed cache fields).
func buildMergedLeaderKfl(followers []*Property, n int, tickRate float64) *rifx.Chunk {
	const bpk = 128
	lhd3 := make([]byte, 52)
	lhd3[1], lhd3[2], lhd3[3] = 0xd0, 0x0b, 0xee
	binary.BigEndian.PutUint32(lhd3[0x08:0x0C], uint32(n))
	binary.BigEndian.PutUint32(lhd3[0x0C:0x10], 1)
	binary.BigEndian.PutUint32(lhd3[0x10:0x14], bpk)
	binary.BigEndian.PutUint32(lhd3[0x14:0x18], 4)
	binary.BigEndian.PutUint32(lhd3[0x18:0x1C], 1)
	binary.BigEndian.PutUint32(lhd3[0x1C:0x20], 4)

	ldat := make([]byte, n*bpk)
	for i := 0; i < n; i++ {
		blk := ldat[i*bpk : (i+1)*bpk]
		binary.BigEndian.PutUint32(blk[0x00:0x04], uint32(math.Round(followers[0].Keyframes[i].Time*tickRate)))
		blk[0x04] = byte(InterpLinear)
		blk[0x05] = byte(InterpLinear)
		blk[0x06] = 0x00
		blk[0x07] = 0x07
		for axis := 0; axis < 3; axis++ {
			kf := followers[axis].Keyframes[i]
			val, _ := kf.Value.(float64)
			binary.BigEndian.PutUint64(blk[0x38+axis*8:0x40+axis*8], math.Float64bits(val))
			binary.BigEndian.PutUint64(blk[0x50+axis*8:0x58+axis*8], math.Float64bits(-kf.InTemporalEase[0].Speed/100))
			binary.BigEndian.PutUint64(blk[0x68+axis*8:0x70+axis*8], math.Float64bits(kf.OutTemporalEase[0].Speed/100))
		}
	}
	kfl := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDkfl}
	kfl.Children = append(kfl.Children,
		&rifx.Chunk{ID: rifx.IDLhd3, Data: lhd3},
		&rifx.Chunk{ID: rifx.IDLdat, Data: ldat},
	)
	return kfl
}

// removeFollowerChunks drops every follower's tdmn+tdbs pair from the group
// LIST + the scene tree (leader-only merged form; AE re-allocates placeholders).
func removeFollowerChunks(grp *AEPropertyGroup, followers []*Property) {
	remove := make(map[*rifx.Chunk]bool, len(followers)*2)
	for _, f := range followers {
		fb := f.propertyBack()
		if fb == nil {
			continue
		}
		idx := indexOfChunk(grp.propertyGroupBack().chunk.Children, fb.tdbs)
		if idx < 0 {
			continue
		}
		remove[fb.tdbs] = true
		if idx >= 1 && grp.propertyGroupBack().chunk.Children[idx-1].ID == rifx.IDTdmn {
			remove[grp.propertyGroupBack().chunk.Children[idx-1]] = true
		}
	}
	keep := grp.propertyGroupBack().chunk.Children[:0:0]
	for _, ch := range grp.propertyGroupBack().chunk.Children {
		if !remove[ch] {
			keep = append(keep, ch)
		}
	}
	grp.propertyGroupBack().chunk.Children = keep

	removeFollowers := make(map[PropertyBase]bool, len(followers))
	for _, f := range followers {
		removeFollowers[f] = true
	}
	grp.Children = filterPropertyBase(grp.Children, removeFollowers)
	filterLayerProperties(grp, followers)
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
