package aep

// Facade re-exports of the parse + structural-mutation entry points whose
// implementations moved to internal/serializer (M8 P3.1 stage 2). The public
// call form (aep.Open / aep.DeleteLayer / aep.NewComposition …) is unchanged;
// each function delegates verbatim. Types in these signatures are aep aliases
// that resolve to the same scene types serializer uses, so the delegations are
// type-identical across the package boundary.

import (
	"io"

	"github.com/example/aep-parser/internal/serializer"
)

// Open parses an .aep file by path and returns the Project.
func Open(path string) (*Project, error) { return serializer.Open(path) }

// FromReader parses an .aep file from an io.ReadSeeker.
func FromReader(r io.ReadSeeker) (*Project, error) { return serializer.FromReader(r) }

// Reopen serializes the project to memory (WriteAEP) and re-parses the bytes
// (FromReader), returning the fresh *Project. The receiver is left untouched;
// callers switch to the returned project and re-resolve item / layer handles
// (e.g. by name or ID — IDs are preserved by the round-trip).
//
// Why: layers built by the structural New* APIs (NewShapeLayer / NewCameraLayer
// / NewLightLayer) exist only as pre-lowered chunks — they have no parsed
// property tree, so write paths that splice into a parsed Layr (AddEffect's
// parade auto-create, the Camera* / Light* option setters) refuse them. One
// Reopen upgrades every built layer into a fully parsed layer, after which all
// parsed-layer APIs work with full fidelity.
//
// The round-trip costs one serialize + parse of the whole project and returns a
// new object graph; any *Layer / *Composition pointers into the old project
// remain valid for the old project only.
func Reopen(p *Project) (*Project, error) { return serializer.Reopen(p) }

// NewProject returns a fresh empty Project parsed from the embedded
// AE skeleton matching the requested target.
//
// Optional target arg: zero args = TargetAE2020 (max compatibility). Pass
// at most one target. Subsequent NewComposition calls populate it.
//
// Never returns an error: the embedded templates are build-time trusted;
// parser bugs panic with a "build bug" message (not user-facing).
// Panics on: multiple target args, or unknown AETarget value (forward-incompat).
func NewProject(target ...AETarget) *Project { return serializer.NewProject(target...) }

// NewComposition adds an empty composition to the project's root folder.
//
// Required:
//
//	name        — non-empty string
//	width/height — > 0 (uint16; AE max 30000)
//	FrameRateHz   — > 0 (Hz; 29.97 etc.; whole+frac/65536 encoding handled internally)
//	duration    — > 0 (seconds; converted to whole frames via fps internally)
//
// Optional fields default to AE-typical (BGColor=0/PAR=1.0/ResFac=1,1/Shutter=180,0/MotionBlur=128,16).
// Override via existing Set* methods after the call.
//
// Composition.ID is auto-assigned (Project.nextItemID++, monotonic).
// New comp appends to the project's root folder.
//
// Atomic mutation: if chunk parse fails or warnings appear, rollback
// chunk-tree + typed index + warnings to pre-call state.
//
// Warnings-as-failure: builder must produce zero parser warnings —
// if any appear, that's a builder bug; rollback + return internal error.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Project.NewComposition method form.
func NewComposition(
	p *Project,
	name string,
	width, height uint16,
	FrameRateHz, duration float64,
) (*Composition, error) {
	return serializer.NewComposition(p, name, width, height, FrameRateHz, duration)
}

// DuplicateComposition deep-clones src (a comp in this Project) as a new
// sibling comp named name, appended to p.Compositions. The dup contains a
// fresh copy of every layer (new layer IDs), with intra-comp parent +
// track-matte refs remapped to the dup's own layers; layer SOURCES
// (footage / precomp items) are shared verbatim, not duplicated — matching
// AE ScriptingAPI's CompItem.duplicate(). Returns the new *Composition.
//
// Clone semantics (same-Project comp only):
//
//   - new comp item ID = allocItemID(p)             (idta @0x10)
//   - per layer: new layer ID = allocItemID(p)      (ldta @0x00)
//   - intra-comp ParentID @0x84 / TrackMatteLayerID @0xA0 remapped via
//     srcLayerID→dupLayerID map (matte guarded by len(ldta) >= 0xA4)
//   - SourceID @0x28 verbatim (shared Footage/Comp items)
//   - comp name = caller-supplied (length-variable Utf8 rewrite)
//
// Refuse-cases (R1..R7): nil src, project backref missing, src itemList
// backref missing, src not in this Project, empty name, src Item not
// found in rootFold, layer ldta too short for ParentID write.
//
// Atomic mutation: snapshot rootFold.Children + p.Compositions +
// scene.ProjectNextItemID(p) + len(p.Warnings); on any new parser warning during the
// re-parse, roll all back including the nextItemID bump.
//
// Stable — passed AE 2020 + AE 2025 ship-gate: AE accepts the
// Go-emitted file and the dup's intra-comp parent ref resolves to the dup's
// own layer (remap confirmed by AE), with sources shared with the original.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Project.DuplicateComposition method form.
func DuplicateComposition(p *Project, src *Composition, name string) (*Composition, error) {
	return serializer.DuplicateComposition(p, src, name)
}

// NewShapeLayer adds a new empty ShapeLayer to the composition.
//
// Required:
//
//	name — non-empty string (matches NewComposition validation contract)
//
// Returns the typed *ShapeLayer wrapper; the embedded *Layer is also
// appended to comp.Layers so V1 lookup paths (Composition.LayerByID /
// LayerByName) work immediately. ID is auto-assigned via the project's
// monotonic item-ID counter (never reused; layer IDs share the item-ID
// namespace per V1 parser convention).
//
// Atomic mutation: if lowering fails, or downstream parse emits any warning,
// all state mutated by this call is rolled back to the pre-call snapshot
// before the error is returned.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Composition.NewShapeLayer method form.
func NewShapeLayer(c *Composition, name string) (*ShapeLayer, error) {
	return serializer.NewShapeLayer(c, name)
}

// NewCameraLayer adds a new Camera layer to the composition and returns it.
//
// A camera is source-less: it is defined entirely by its ldta + Camera Options
// property group. The new layer is cloned from an embedded AE-native Camera Layr
// (so every AE-internal flag byte is faithful), with the layer ID, name, and
// time span (0 → comp duration) patched for this comp. Camera position / point
// of interest / options inherit the template's AE defaults; adjust afterward via
// the Camera* setters once the project is re-parsed.
//
// Atomic mutation (snapshot + warnings-as-failure rollback). Alpha / structural.
// Free function (CLAUDE.md #2 structural-op call-form).
func NewCameraLayer(c *Composition, name string) (*Layer, error) {
	return serializer.NewCameraLayer(c, name)
}

// NewLightLayer adds a new Light layer to the composition and returns it.
//
// Like NewCameraLayer, a light is source-less (ldta + Light Options group),
// cloned from an embedded AE-native Light Layr with ID / name / time span
// patched. Light kind / color / intensity inherit the template's AE defaults;
// adjust afterward via the Light* setters once the project is re-parsed.
//
// Atomic mutation (snapshot + warnings-as-failure rollback). Alpha / structural.
// Free function (CLAUDE.md #2 structural-op call-form).
func NewLightLayer(c *Composition, name string) (*Layer, error) {
	return serializer.NewLightLayer(c, name)
}

// DeleteLayer removes the layer at the given 0-based index in c.Layers.
// Returns nil on success, or an error if a refuse-case triggers (index
// out of range / comp lacks itemList back-ref / target is the last
// layer / target is not an AV layer / backref corruption).
//
// Reference cleanup — per AE's own delete behavior (RE'd via the
// re_delete_layer_*.aep fixtures):
//
//   - any other layer's Layer.ParentID == deleted.ID → reset to 0
//     (ldta @0x84..0x87)
//   - any other layer's Layer.TrackMatteLayerID == deleted.ID → reset
//     to 0 (ldta @0xA0..0xA3, when ldta is long enough — AE ≤22 didn't
//     write this field)
//   - Layer.TrackMatte byte (ldta @0x6B) on those neighbors is LEFT
//     UNTOUCHED to match AE: the matte intent flag persists even after
//     the matte source is gone (AE re-resolves via implicit "layer
//     above" at render time, which now returns nothing — matches AE)
//   - Project.nextItemID counter: untouched (IDs never reused)
//
// String-level references to the deleted layer's ID (expressions,
// render queue, essential graphics) are out of scope — callers must
// scrub these manually if needed.
//
// Atomic mutation: snapshot pre-call state of itemList.Children, c.Layers,
// neighbor refs / ldta bytes, and Project.Warnings; on any new parser
// warning surfaced during the call, roll all of them back and return the
// warnings as an error.
//
// Stable: AE 2020 + AE 2025 ship-gate green (8/8 PASS across baseline /
// middle / parent / matte modes). Future RE can lift the non-AV refuse
// and the single-layer-comp refuse — both are conservative defaults
// because AE's behavior for those scenarios hasn't been verified.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Composition.DeleteLayer method form.
func DeleteLayer(c *Composition, index int) error { return serializer.DeleteLayer(c, index) }

// DuplicateLayer clones the layer at the given 0-based index in c.Layers
// and inserts the clone at that same position, pushing source and
// everything below down by one (mirrors AE ScriptingAPI's
// layer.duplicate()). Returns the cloned *Layer on success, or an error
// if a refuse-case triggers.
//
// Clone semantics (RE'd via 4 AE-saved fixtures + byte-diff):
//
//   - new layer ID = allocItemID(proj) (head counter +1, monotonic)
//   - clone's 16-chunk block (Layr + Ewst + 14 follower leaves in
//     AE-saved files; 2 chunks in Go-built layers) is a deep byte-clone
//     of source's block, with ldta @0x00..0x03 overwritten with the new
//     ID. All other body bytes (SourceID @0x28, ParentID @0x84,
//     TrackMatte @0x6B) are verbatim from source.
//   - Layer.SourceID/ParentID/TrackMatteLayerID/TrackMatte struct fields
//     on the clone = source values (no footage duplication; no
//     reference rewrites).
//   - Name = caller-supplied (AE keeps source's name verbatim; we
//     require an explicit name to avoid silent duplicate-name confusion).
//   - Children's outgoing ParentID is NOT updated — clone is a fresh
//     sibling shadow; source remains the canonical parent for any
//     incoming refs (F6).
//
// Refuse-cases (conservative):
//
//   - name empty
//   - index out of range
//   - comp lacks parsed itemList back-ref
//   - source is not an AV layer (camera/light/audio behavior not RE'd)
//   - source has implicit TrackMatte (TrackMatte != None &&
//     TrackMatteLayerID == 0). F2 quirk: AE relocates clone above the
//     positional matte source to preserve original's matte; not yet
//     supported. AE 23+ explicit matte (TrackMatteLayerID != 0) is
//     ALLOWED (Stable — clone byte-copies @0xA0 + @0x6B verbatim;
//     passed AE 2025 ship-gate).
//   - backref corruption (Layr formType / Ewst sibling mismatch)
//
// Atomic mutation: snapshot pre-call state of itemList.Children,
// c.Layers, scene.ProjectNextItemID(proj), and proj.Warnings; on any parser warning
// surfaced during the re-parse, roll all of them back (including the
// nextItemID bump) and return the warnings as an error.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Composition.DuplicateLayer method form.
func DuplicateLayer(c *Composition, index int, name string) (*Layer, error) {
	return serializer.DuplicateLayer(c, index, name)
}

// InsertLayer deep-clones src into c.Layers at atIdx (0-based; atIdx ==
// len(c.Layers) appends). Returns the inserted clone *Layer on success. src may
// live in a sibling comp of the same Project, or in a different Project
// (cross-Project).
//
// Same-Project clone semantics (scene.CompositionProj(scene.LayerComp(src)) == scene.CompositionProj(c)):
//
//   - new layer ID = allocItemID(scene.CompositionProj(c)) (head counter +1, monotonic)
//   - clone block = deep byte-clone of src's [Layr, Ewst, leaf-followers)
//     range, with per-byte ldta mutations:
//     @0x00..0x03 ← newID
//     @0x6B       ← TrackMatteNone (cross-comp matte source is invalid)
//     @0x84..0x87 ← 0 (ParentID; src's ParentID named a layer in scene.LayerComp(src))
//     @0xA0..0xA3 ← 0 (explicit matte ID, guarded by len(ldta) >= 0xA4)
//   - clone.SourceID = src.SourceID (verbatim — the shared Footage/Comp item).
//   - clone.Name = src.Name (verbatim — matches AE's layer.copyToComp).
//
// Cross-Project semantics (scene.CompositionProj(scene.LayerComp(src)) != scene.CompositionProj(c)):
// additionally imports src's reachable ITEM CLOSURE (footage + precomp,
// transitively) into c's Project at root level with fresh dest item IDs, then
// remaps the inserted clone's SourceID @0x28 + AlternateSourceID through the
// srcItemID→destItemID map. File-backed footage already present in dest (matched
// by Path) is reused, not re-cloned; comps and solids/placeholders are always
// cloned. ParentID / track matte are still reset (cross-comp). Folders are not
// recreated.
//
// Refuse-cases: nil src, dest backref missing, atIdx out of range, src
// detached, same-comp redirect, non-AV, direct pre-comp loop (same-Project
// only), src backref missing, structural corruption. Cross-Project adds:
// dest/src Project has no root Fold; dangling closure source.
//
// Atomic mutation: snapshot dest itemList.Children + c.Layers +
// scene.ProjectNextItemID(scene.CompositionProj(c)) + len(scene.CompositionProj(c).Warnings) (cross-Project also snapshots
// rootFold.Children + Compositions + Footage); on any new parser warning
// during re-parse, roll all back including the nextItemID bump.
//
// Stable (both paths) — same-Project passed AE 2020 + AE 2025 ship-gate (3 modes
// [basic/footage/precomp] × 2 = 6/6 PASS); cross-Project passed the assert-based
// AE 2020 + AE 2025 gate (3 modes [footage/precomp/dedup] × 2 = 6/6 PASS): AE
// accepts the Go-emitted file, the inserted clone's source resolves (imported /
// dedup'd), and footage is not duplicated on path match.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Composition.InsertLayer method form.
func InsertLayer(c *Composition, src *Layer, atIdx int) (*Layer, error) {
	return serializer.InsertLayer(c, src, atIdx)
}

// MoveLayer reorders the layer at `from` to position `to` in c.Layers
// (both 0-based). The source layer's entire chunk block — Layr + Ewst
// + leaf followers (adaptive scan to next LIST/EOF, same machinery as
// DeleteLayer / DuplicateLayer) — is spliced out and re-inserted at the
// target slot. After the call, c.Layers[to] == the moved layer, and
// every layer's Layer.Index field is refreshed to match its new slice
// position.
//
// Refuse-cases (conservative):
//
//   - `from` or `to` out of range (note: `to == len(c.Layers)-1` IS in
//     range and means "move to last slot")
//   - comp lacks parsed itemList back-ref
//   - source layer lacks Layr back-ref / corrupted block (Layr formType
//     / Ewst sibling mismatch)
//
// `from == to` is a no-op (returns nil, no state change).
//
// Unlike DeleteLayer / DuplicateLayer, MoveLayer does NOT care about
// layer Type or TrackMatte — pure reorder works for AV / Camera / Light
// / Audio / Shape / Text / matted layers alike.
//
// Atomic mutation: snapshot pre-call itemList.Children + c.Layers +
// each layer's Index + Warnings count; on any new parser warning during
// the call, roll all of them back. No re-parse and no new chunks
// created, so the warnings path is defensive.
//
// Stable: no Alpha gate — AE behavior is known (layer order = order of
// Layr LISTs in itemList.Children, same model that DeleteLayer and
// DuplicateLayer already exercise and ship-gate across AE 2020 + AE
// 2025). The reorder path is ship-gate validated for AE acceptance.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Composition.MoveLayer method form.
func MoveLayer(c *Composition, from, to int) error { return serializer.MoveLayer(c, from, to) }

// MoveToBeginning moves the receiver to position 0 (top of layer stack
// in AE's display, AE-index 1).
//
// Free function (not a method) — see MoveLayer. BREAKING vs the former
// Layer.MoveToBeginning method form; the aep facade re-exports it post-split.
func MoveToBeginning(l *Layer) error { return serializer.MoveToBeginning(l) }

// MoveToEnd moves the receiver to the last position in c.Layers
// (bottom of layer stack in AE's display, AE-index c.numLayers).
//
// Free function (not a method) — see MoveLayer. BREAKING vs the former
// Layer.MoveToEnd method form; the aep facade re-exports it post-split.
func MoveToEnd(l *Layer) error { return serializer.MoveToEnd(l) }

// MoveAfter moves the receiver to the slot immediately after `other`
// (i.e., other.Index < receiver.Index post-call, both viewed in
// c.Layers slice order — receiver lands just below other in the stack).
// Returns an error if other belongs to a different comp, other == l,
// or either layer is missing a comp back-ref.
//
// Free function (not a method) — see MoveLayer. BREAKING vs the former
// Layer.MoveAfter method form; the aep facade re-exports it post-split.
func MoveAfter(l, other *Layer) error { return serializer.MoveAfter(l, other) }

// MoveBefore moves the receiver to the slot immediately before `other`
// (receiver lands just above other in the stack).
//
// Free function (not a method) — see MoveLayer. BREAKING vs the former
// Layer.MoveBefore method form; the aep facade re-exports it post-split.
func MoveBefore(l, other *Layer) error { return serializer.MoveBefore(l, other) }

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
	return serializer.AddMarker(c, seconds)
}

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
func RemoveMarker(m *Marker) error { return serializer.RemoveMarker(m) }

// InsertKeyframe builds a new bpk-byte keyframe block and inserts it
// into the property's ldat stream, then updates the lhd3 count header.
// Returns the new Keyframe and its index in Property.Keyframes
// (insertion is time-sorted; ties land after existing keys at the
// same time).
//
// Requires the property to already have ≥1 keyframe so the new block
// can clone the existing layout (header byte @0x07, bpk, etc.). For
// properties without keyframes, use SetStaticValue or build keyframes
// in AE first — synthesizing the lhd3/ldat chunks from scratch isn't
// supported yet.
//
// `value` follows the same rules as Keyframe.SetValue:
//   - 1D property: pass float64
//   - multi-component: pass []float64 (length == Property.Components)
//
// The new keyframe's interpolation is Linear/Linear; ease + tangents
// are zeroed. Call SetInInterp / SetInTemporalEase / SetInSpatialTangent
// on the returned Keyframe to refine.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Property.InsertKeyframe method form.
func InsertKeyframe(p *Property, time float64, value any) (*Keyframe, int, error) {
	return serializer.InsertKeyframe(p, time, value)
}

// DeleteKeyframe removes the keyframe at index i from the property's
// ldat stream and decrements the lhd3 count header. Returns an error
// when i is out of range or the property has no keyframe stream.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former Property.DeleteKeyframe method form.
func DeleteKeyframe(p *Property, i int) error { return serializer.DeleteKeyframe(p, i) }

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
	return serializer.SetDimensionsSeparated(p, separated)
}

// RemovePropertyGroup deletes this group from its parent INDEXED_GROUP. The receiver must
// be a direct child of an indexed group (Effect Parade / Mask Parade / Root
// Vectors Group / Text Animators); RemovePropertyGroup returns an error otherwise, mirroring
// AE's ScriptingAPI refuse.
//
// Atomic: snapshots the parent chunk LIST, scene children, the mirrored flat
// slice, and Project.Warnings; on any new parser warning everything rolls back
// and the warnings are returned as an error.
//
// Alpha — see file header for ship-gate status. Free function (not a method) so
// the impl can live in internal/serializer after the M8 split (CLAUDE.md #2
// structural-op call-form carve-out); the aep facade re-exports it. Renamed +
// BREAKING vs the former AEPropertyGroup.Remove method form.
func RemovePropertyGroup(g *AEPropertyGroup) error { return serializer.RemovePropertyGroup(g) }

// MovePropertyGroup reorders this group to position index (0-based) among its parent
// INDEXED_GROUP's children. index is clamped-checked against the current child
// count. Mirrors AE's PropertyBase.moveTo (which is 1-based; the Go API is
// 0-based per project convention).
//
// Alpha — see file header for ship-gate status. Free function (not a method) so
// the impl can live in internal/serializer after the M8 split (CLAUDE.md #2
// structural-op call-form carve-out); the aep facade re-exports it. Renamed +
// BREAKING vs the former AEPropertyGroup.MoveTo method form.
func MovePropertyGroup(g *AEPropertyGroup, index int) error {
	return serializer.MovePropertyGroup(g, index)
}

// DuplicatePropertyGroup inserts a copy of this group immediately after it among its parent
// INDEXED_GROUP's children — mirroring AE's PropertyBase.duplicate() structural
// effect — and returns the clone. The receiver must be a direct child of an
// indexed group (Effect Parade / Mask Parade / Root Vectors Group / Text
// Animators); DuplicatePropertyGroup returns an error otherwise, mirroring AE's refuse.
//
// The clone reuses the source's match-name and on-disk payload verbatim. AE's
// own .duplicate() additionally persists a deduplicated display name (the
// localized "<name> 2") into a length-variable tdsn on the clone's inner tdgp
// (RE'd 2026-06-03, see incidents/property-indexed-group-structural-re.md
// slice 2: the source carries NO tdsn, the clone gains one reading "高斯模糊 2").
// We deliberately do NOT synthesize that suffix: the base is AE's *localized*
// effect name, which needs the AE schema/localization DB we don't carry (the
// same blocker as Property.ValueText), and a clone with no tdsn is byte-for-byte
// an "add the same effect twice" project — which AE accepts and re-derives the
// runtime dedup name from on open. The persisted suffix is cosmetic; AE
// recomputes it. The structural duplicate is faithful.
//
// Chunk mechanics: pure (tdmn, payload) pair insert immediately after the
// source pair, no count/index chunk (RE: parade 9→11 children, nothing else
// touched).
//
// Atomic: snapshots the parent chunk LIST, scene children, the mirrored flat
// slice, and Project.Warnings; on any new parser warning — or a flat-mirror
// re-parse that fails to reproduce exactly one clone — everything rolls back
// and an error is returned.
//
// Alpha — see file header for ship-gate status. Free function (not a method) so
// the impl can live in internal/serializer after the M8 split (CLAUDE.md #2
// structural-op call-form carve-out); the aep facade re-exports it. Renamed +
// BREAKING vs the former AEPropertyGroup.Duplicate method form.
func DuplicatePropertyGroup(g *AEPropertyGroup) (*AEPropertyGroup, error) {
	return serializer.DuplicatePropertyGroup(g)
}

// AddEffect appends an effect to the layer's "ADBE Effect Parade" and returns
// the parsed *Effect, so the caller can immediately tune its parameters via
// Effect.Parameters (Property.SetStaticValue works on effect params — e.g. set
// "ADBE Gaussian Blur 2-0001" to change Blurriness).
//
// effectMatchName must be one of SupportedEffects(); the effect's full
// parameter sub-tree (sspc payload) is supplied from an embedded AE-native
// template, which is why only RE'd effects are addable. AE looks the effect up
// by match-name at load, so the named plugin must be installed in the opening
// AE — the seeded effects are built-ins present since before the AE 2020 read
// floor and are version-portable (the AE-2020-extracted bytes are accepted by
// AE 2025).
//
// Mechanics: the parade stores effects as (tdmn, LIST:sspc) pairs terminated by
// an "ADBE Group End" tdmn sentinel; AddEffect splices a fresh pair in just
// before that sentinel — the same (tdmn, payload) splice DuplicatePropertyGroup
// is ship-gate-green with, sourced from a template instead of a sibling. LIST
// sizes grow automatically (rifx recomputes bottom-up on write).
//
// Parade auto-create: a parsed layer with no effects has no Effect Parade group
// at all (AE only persists the parade once ≥1 effect exists). AddEffect splices
// an empty parade — tdsb + default-name tdsn + Group End, the AE-native form —
// into the layer's property tree immediately before "ADBE Transform Group"
// (AE's emitted group order), then adds the effect into it.
//
// Refused layers: camera / light layers (AE does not allow effects on them),
// and layers built by the structural New* APIs that were never parsed — those
// have no property tree to splice into; call aep.Reopen first and add effects
// to the re-parsed layer.
//
// Atomic mutation: snapshot parade chunk + scene children + flat Effects slice
// (+ the pre-auto-create tree state); re-parse the spliced pair to obtain a
// back-ref-correct *Effect; roll back on any parser warning.
//
// Alpha / structural. Free function (not a method) so the impl can live in
// internal/serializer (CLAUDE.md #2 structural-op call-form carve-out).
func AddEffect(layer *Layer, effectMatchName string) (*Effect, error) {
	return serializer.AddEffect(layer, effectMatchName)
}

// SupportedEffects returns the sorted effect match-names AddEffect can add from
// an embedded template.
func SupportedEffects() []string { return serializer.SupportedEffects() }

// RemoveEffect removes the effect at the given 0-based index from the layer's
// Effect Parade — the inverse of AddEffect. It is a thin, index-validated
// wrapper over RemovePropertyGroup (AE 2020 + AE 2025 ship-gate green for
// Effect-Parade child removal). Returns an error if the layer has no Effect
// Parade or index is out of range.
//
// Alpha / structural. Free function (CLAUDE.md #2 structural-op call-form).
func RemoveEffect(layer *Layer, index int) error { return serializer.RemoveEffect(layer, index) }

// Effect match-name constants for AddEffect's built-in library. Use these
// instead of hardcoding AE's internal match-name strings. The trailing comment
// on each is the display name shown in AE's Effects panel.
const (
	EffectGaussianBlur       = serializer.EffectGaussianBlur       // Gaussian Blur
	EffectFill               = serializer.EffectFill               // Fill
	EffectTint               = serializer.EffectTint               // Tint
	EffectBrightnessContrast = serializer.EffectBrightnessContrast // Brightness & Contrast
	EffectTritone            = serializer.EffectTritone            // Tritone
	EffectLevels             = serializer.EffectLevels             // Levels
	EffectLevelsIndividual   = serializer.EffectLevelsIndividual   // Levels (Individual Controls)
	EffectHueSaturation      = serializer.EffectHueSaturation      // Hue/Saturation
	EffectBoxBlur            = serializer.EffectBoxBlur            // Fast Box Blur
	EffectGlow               = serializer.EffectGlow               // Glow
	EffectInvert             = serializer.EffectInvert             // Invert
	EffectExposure           = serializer.EffectExposure           // Exposure
)

// AddItem appends a render queue item for comp, mirroring ExtendScript
// RenderQueue.items.add(comp). Alpha / structural.
//
// Strategy (clone + remap, like InsertLayer): the queue's last item is the
// template — its 2246B settings block, [LIST:list + 'LOm '] group, and Rout
// per-item block are deep-cloned, then the clone's comp_id (settings @0x08) is
// repointed at comp. The settings ldat / Rout / lhd3 count grow in lock-step,
// mirroring AE's own items.add() delta (REd from a 1-item→2-item diff). The
// cloned output module keeps the template's path/template (AE accepts it; a
// fresh add would name it after comp — deferred).
//
// Requires at least one existing item to clone from (an empty queue has no
// template). The clone is taken from the template's scene-owned settingsBlock
// copy (single source of truth). The grown settings ldat reallocates, so every
// item's back.settingsSlice alias is re-pointed afterward, and WriteAEP syncs
// the copies back. See incidents/render-queue-delete-mechanics.md.
//
// Free function (not a method) — see RemoveItem. BREAKING vs rq.AddItem(comp).
func AddItem(rq *RenderQueue, comp *Composition) (*RenderQueueItem, error) {
	return serializer.AddItem(rq, comp)
}

// RemoveItem deletes the render queue item at index (0-based), mirroring
// ExtendScript RenderQueueItem.remove(). Alpha / structural.
//
// Free function (not a method) so the impl can live in internal/serializer
// after the M8 split (CLAUDE.md #2 structural-op call-form carve-out); the aep
// facade re-exports it. BREAKING vs the former rq.RemoveItem(i) method form.
//
// Byte mechanics REd from AE 2020 (test_data/re_rq_delete.jsx, 2-item→1-item
// diff): removing item i drops, in lock-step,
//
//   - the item's [RCom?] + LIST:list + LIST:'LOm ' from the LItm container,
//   - the item's 2246-byte block from the LRdr-level settings ldat, and
//     decrements the settings lhd3 count (@0x08 and @0x0C),
//   - the item's per-item block from the Rout flags chunk (4-byte header +
//     uniform per-item stride), decrementing the header proportionally.
//
// The scene-side settingsBlock buffers are independent copies (single source of
// truth); surviving items' back.settingsSlice aliases are re-pointed to their
// new offsets after the splice, and WriteAEP syncs the copies back.
//
// Alpha: structural delete is not yet AE-ship-gated. Only items with one output
// module are covered by the Rout RE (uniform per-item stride); see
// incidents/render-queue-delete-mechanics.md.
func RemoveItem(rq *RenderQueue, index int) error { return serializer.RemoveItem(rq, index) }

// SetRenderer switches the composition's 3D rendering engine. The name may be
// either a binary prin match_name ("ADBE Escher" / "ADBE Calder" /
// "ADBE Ernst" / "ADBE Picasso") or an ExtendScript module name
// ("ADBE Advanced 3d" → "ADBE Escher"); it is normalized to the binary name.
// The binary match_name + display name are rewritten in the prin chunk
// (length-preserving) and the prda chunk is replaced with the engine's default
// options (structural). Returns an error for an unknown renderer, a comp built
// outside the parser (no prin/prda back-ref), a comp whose prin is not the
// expected 104 bytes, or if the mutation surfaces a parser warning (rolled back).
//
// Which engines a given AE version actually exposes differs (AE 2020:
// Escher/Ernst + a Standard variant; AE 2025: Calder/Ernst + Picasso; AE 2025
// auto-promotes legacy Escher/Picasso to Advanced 3D on load). The binary
// match_name is the stable engine identity — see
// sketches/2026-06-01-renderer-write-re-findings.md.
//
// Ship-gated: AE 2025 (4/4) + AE 2020 (Ernst + Escher) green.
//
// Free function (not a method) so the rollback path can reach the concrete
// comp back-ref (prin/prda chunks) after the M8 split; the aep facade
// re-exports it. BREAKING vs the former Composition.SetRenderer method form.
func SetRenderer(c *Composition, name string) error { return serializer.SetRenderer(c, name) }
