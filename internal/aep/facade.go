package aep

// Facade re-exports of the parse + structural-mutation entry points whose
// implementations live in the serializer layer. The public call form
// (aep.Open / aep.DeleteLayer / aep.NewComposition …) is unchanged; each
// function delegates verbatim. Types in these signatures are aep aliases that
// resolve to the same scene types the serializer uses, so the delegations are
// type-identical across the package boundary.

import (
	"io"

	"github.com/example/aep-parser/internal/serializer"
)

// @summary    Parse an .aep file by path
// @param      path  filesystem path to the .aep file to read
// @returns    the parsed Project
// @domain     meta
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @alias      open,read,parse,读取,打开,加载 aep
func Open(path string) (*Project, error) { return serializer.Open(path) }

// @summary    Parse an .aep file from an io.ReadSeeker
// @param      r  reader positioned at the start of the .aep byte stream
// @returns    the parsed Project
// @domain     meta
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @alias      from reader,read,流读取,io.ReadSeeker
func FromReader(r io.ReadSeeker) (*Project, error) { return serializer.FromReader(r) }

// @summary    Serialize then re-parse a project to fully materialize built layers
// @description Serializes the project to memory (WriteAEP) and re-parses the
//   bytes (FromReader), returning a fresh Project. The receiver is left
//   untouched; callers switch to the returned project and re-resolve item /
//   layer handles (by name or ID — IDs are preserved by the round-trip).
//
//   Layers built by the structural New* APIs (NewShapeLayer / NewCameraLayer /
//   NewLightLayer) exist only as pre-lowered chunks — they have no parsed
//   property tree, so write paths that splice into a parsed layer (AddEffect's
//   parade auto-create, the Camera / Light option setters) refuse them. One
//   Reopen upgrades every built layer into a fully parsed layer, after which
//   all parsed-layer APIs work with full fidelity.
//
//   The round-trip costs one serialize + parse of the whole project and returns
//   a new object graph; any Layer / Composition pointers into the old project
//   remain valid for the old project only.
// @param      p  the project to serialize and re-parse
// @returns    a fresh fully-parsed Project (the receiver is left unchanged)
// @domain     meta
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   upgrades New*-built layers to parsed layers, unlocking parsed-only write paths (AddEffect / AddMask / Camera & Light setters)
// @alias      reopen,reparse,重新打开,升级层,parsed layer
func Reopen(p *Project) (*Project, error) { return serializer.Reopen(p) }

// @summary    Create a fresh empty Project from an embedded AE skeleton
// @description Returns an empty Project parsed from the embedded AE skeleton
//   matching the requested target. Zero args = TargetAE2020 (maximum
//   compatibility); pass at most one target. Subsequent NewComposition calls
//   populate it.
//
//   Never returns an error: the embedded templates are build-time trusted, so a
//   parser failure panics with a "build bug" message rather than surfacing to
//   the caller. Panics on multiple target args or an unknown AETarget value.
// @param      target  optional target AE version (default TargetAE2020)
// @returns    the new empty Project
// @domain     project
// @stability  stable
// @verify     ae-accept
// @gate       TestV2_1_AEShipGate_AE2020,TestV2_1_AEShipGate_AE2025
// @since      AE2020
// @boundary   zero args defaults to TargetAE2020; targets 2020 / 2022 / 2025 supported
// @incident   ae25-acceptance-gate
// @alias      project,工程,新建工程,空工程,create project
func NewProject(target ...AETarget) *Project { return serializer.NewProject(target...) }

// @summary    Add an empty composition to the project's root folder
// @description Appends a new empty composition to the project's root folder. The
//   name must be non-empty; width and height must be 1..30000; FrameRateHz must
//   be > 0 (in Hz, e.g. 29.97 — the whole+fraction encoding is handled
//   internally); duration must be > 0 (in seconds, converted to whole frames via
//   the frame rate).
//
//   Optional settings default to AE-typical values (background black, pixel
//   aspect 1.0, full resolution, shutter 180°, motion blur defaults); override
//   them with the Set* methods after the call. The composition ID is
//   auto-assigned from the project's monotonic item-ID counter.
//
//   Atomic: if chunk parsing fails or any parser warning appears, the chunk
//   tree, typed index, and warning list roll back to the pre-call state.
// @param      p            the project to add the composition to
// @param      name         composition name (non-empty)
// @param      width        composition width in pixels (1..30000)
// @param      height       composition height in pixels (1..30000)
// @param      FrameRateHz  frame rate in Hz (> 0; e.g. 29.97)
// @param      duration     composition duration in seconds (> 0)
// @returns    the created Composition
// @domain     comp
// @stability  stable
// @verify     ae-accept
// @gate       TestV2_1_AEShipGate_AE2020,TestV2_1_AEShipGate_AE2025
// @since      AE2020
// @boundary   optional settings default to AE-typical values; change the rest via Set* methods
// @incident   ae25-acceptance-gate
// @alias      composition,合成,新建合成,comp,create comp
func NewComposition(
	p *Project,
	name string,
	width, height uint16,
	FrameRateHz, duration float64,
) (*Composition, error) {
	return serializer.NewComposition(p, name, width, height, FrameRateHz, duration)
}

// @summary    Deep-clone a composition as a new sibling comp
// @description Deep-clones src (a composition in this project) as a new sibling
//   composition named name, appended to the project. The duplicate gets a fresh
//   copy of every layer (with new layer IDs), and intra-comp parent and
//   track-matte references are remapped to the duplicate's own layers. Layer
//   sources (footage / precomp items) are shared verbatim, not duplicated —
//   matching AE's CompItem.duplicate().
//
//   Refuses when src is nil, its project or item-list back-reference is missing,
//   src is not in this project, the name is empty, the source item is not found
//   in the root folder, or a layer record is too short to carry a parent ID.
//
//   Atomic: snapshots the root folder, composition list, next-item-ID counter,
//   and warning count; on any new parser warning during the re-parse, all of
//   them roll back including the ID-counter bump.
// @param      p     the project that owns src and will own the duplicate
// @param      src   the composition to clone (must belong to p)
// @param      name  name for the new duplicate composition (non-empty)
// @returns    the cloned Composition
// @domain     comp
// @stability  stable
// @verify     ae-accept
// @gate       TestCompDuplicate_AEShipGate_AE2020,TestCompDuplicate_AEShipGate_AE2025
// @since      AE2020
// @boundary   AE-verified: distinct comp item, deep-copied layer list, intra-comp parent refs remap to the duplicate's own layers, layer sources shared (not copied), comp settings (size / fps / duration) cloned
// @alias      duplicate composition,复制合成,克隆合成
func DuplicateComposition(p *Project, src *Composition, name string) (*Composition, error) {
	return serializer.DuplicateComposition(p, src, name)
}

// @summary    Add an empty shape layer to a composition
// @description Appends a new empty shape layer and returns the typed ShapeLayer
//   wrapper; the embedded Layer is also appended to the comp's layer list so
//   lookups by ID or name work immediately. The ID is auto-assigned from the
//   project's monotonic item-ID counter (never reused; layer IDs share the
//   item-ID namespace).
//
//   Atomic: if lowering fails or a downstream parse emits any warning, all state
//   mutated by the call rolls back to the pre-call snapshot before the error is
//   returned.
// @param      c     the composition to add the shape layer to
// @param      name  shape layer name (non-empty)
// @returns    the created ShapeLayer
// @domain     layer-create
// @stability  stable
// @verify     ae-accept
// @gate       TestV2_2_Ellipse_AEShipGate_AE2020,TestV2_2_Ellipse_AEShipGate_AE2025
// @since      AE2020
// @boundary   the ellipse variant is gated; rectangle path/stroke embed bytes are still deferred
// @incident   v2-2-aelayer-structure,multi-layer-silent-drop
// @alias      shape,形状,矢量图层
func NewShapeLayer(c *Composition, name string) (*ShapeLayer, error) {
	return serializer.NewShapeLayer(c, name)
}

// @summary    Add a Camera layer to a composition
// @description Appends a new Camera layer and returns it. A camera is source-less:
//   it is defined entirely by its layer record plus a Camera Options property
//   group. The layer is cloned from an embedded AE-native Camera layer (so every
//   AE-internal flag byte is faithful), with the layer ID, name, and time span
//   (0 → comp duration) patched for this comp. Position / point of interest /
//   options inherit the template's AE defaults; adjust them via the Camera*
//   setters after the project is re-parsed (see Reopen).
//
//   Atomic (snapshot + rollback on any parser warning). AE accepts the Go-built
//   camera, types it correctly, and preserves it on resave.
// @param      c     the composition to add the camera to
// @param      name  camera layer name (non-empty)
// @returns    the created camera Layer
// @domain     layer-create
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @incident   camera-light-layer-create-re
// @alias      camera,摄像机
func NewCameraLayer(c *Composition, name string) (*Layer, error) {
	return serializer.NewCameraLayer(c, name)
}

// @summary    Add a Light layer to a composition
// @description Appends a new Light layer and returns it. Like a camera, a light is
//   source-less (a layer record plus a Light Options group), cloned from an
//   embedded AE-native Light layer with ID, name, and time span patched. Light
//   kind / color / intensity inherit the template's AE defaults; adjust them via
//   the Light* setters after the project is re-parsed (see Reopen).
//
//   Atomic (snapshot + rollback on any parser warning). AE accepts the Go-built
//   light, types it correctly, and preserves it on resave.
// @param      c     the composition to add the light to
// @param      name  light layer name (non-empty)
// @returns    the created light Layer
// @domain     layer-create
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @incident   camera-light-layer-create-re
// @alias      light,灯光
func NewLightLayer(c *Composition, name string) (*Layer, error) {
	return serializer.NewLightLayer(c, name)
}

// @summary    Add a point-text layer to a composition
// @description Appends a new point-text layer and returns it. A text layer is
//   source-less: its content lives in the text-engine document inside the layer's
//   Text Properties group. The layer is cloned from an embedded AE-native text
//   layer (point text "A", a single styled run) with the layer ID, name, and time
//   span (0 → comp duration) patched for this comp.
//
//   The returned layer reads its text source immediately and supports SetText
//   without a Reopen (the text-source back-reference is wired at create time).
//   SetText accepts arbitrary-length replacement text for the template's
//   single-paragraph, single-run document (see Layer.SetText for the refuse set).
//
//   Atomic (snapshot + rollback on any parser warning). AE types the Go-built
//   layer as a text layer, reads back the text, and preserves it on resave.
// @param      c     the composition to add the text layer to
// @param      name  text layer name (non-empty)
// @returns    the created text Layer
// @domain     layer-create
// @stability  stable
// @verify     ae-accept
// @gate       TestNewTextLayer_AEShipGate_AE2020,TestNewTextLayer_AEShipGate_AE2025
// @since      AE2020
// @incident   text-btdk-length-variable-write-scoping
// @alias      text,文字,文本图层
func NewTextLayer(c *Composition, name string) (*Layer, error) {
	return serializer.NewTextLayer(c, name)
}

// @summary    Replace a parsed layer's Transform Group with a lowering of t
// @description Replaces a parsed layer's Transform Group — the from-scratch path
//   for animated (or non-default static) transform on layers that are not shape
//   layers (text / precomp / footage / solid / null).
//
//   Why a dedicated call: those layers come from embedded templates whose
//   Transform Group elides default channels (AE default-omission). A fresh
//   NewTextLayer has no "ADBE Position" / "ADBE Opacity" / "ADBE Anchor Point"
//   chunk at all — only the per-axis Position, Orientation, rotation, and
//   environment streams — so SetPosition (no materialized property) and the
//   keyframe APIs (no value chunk to convert) cannot reach them.
//   SetLayerTransform lowers the full canonical Transform Group and swaps it in,
//   materializing every channel.
//
//   Build t with NewLayerTransform (defaults: anchor 0,0; position 0,0; scale
//   100,100; rotation 0; opacity 100), then set static values or keyframes on its
//   streams: Position and AnchorPoint are pixels, Scale is percent, Rotation is
//   degrees, Opacity is percent (0–100). The layer must be parsed (round-trip via
//   Reopen for a freshly built layer).
// @param      layer  the parsed layer to retarget (round-trip via Reopen first)
// @param      t      the transform to lower into the layer's Transform Group
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestSetLayerTransform_AEShipGate_AE2020,TestSetLayerTransform_AEShipGate_AE2025
// @since      AE2020
// @boundary   layer must be parsed (Reopen); replaces the whole Transform Group, materializing default-elided channels; gate covers a text layer (anchor + 3-keyframe Position + 3-keyframe Opacity, opacity in percent); Scale/Rotation share the path but are not separately gated
// @incident   transform-group-default-omission
// @alias      layer transform,层变换,animate layer position,animate layer opacity,materialize transform
func SetLayerTransform(layer *Layer, t *LayerTransform) error {
	return serializer.SetLayerTransform(layer, t)
}

// @summary    Add a solid-color layer to a composition
// @description Appends a new solid-color layer and returns it. A solid is
//   footage-backed: the call also creates a backing solid footage item (cloned
//   from an embedded AE-native template so every AE-internal byte stays faithful)
//   and points the layer's source at it. width and height must be 1..30000 (AE's
//   solid ceiling); rgb components are 0..1 (stored as float32, so exact
//   round-trips need float32-representable values such as 0.25 / 0.5). The
//   layer's time span is re-homed to 0 → comp duration. The returned layer is
//   fully parsed — all parsed-layer setters (transform, AddEffect, …) work
//   immediately without a Reopen.
//
//   Atomic (the cross-project import snapshot + rollback covers both the footage
//   import and the layer splice). AE reads back color / dimensions / flags and
//   preserves them on resave.
// @param      c       the composition to add the solid to
// @param      name    solid layer name (non-empty)
// @param      width   solid width in pixels (1..30000)
// @param      height  solid height in pixels (1..30000)
// @param      rgb     solid color as RGB components, each in 0..1
// @returns    the created solid Layer
// @domain     layer-create
// @stability  stable
// @verify     ae-accept
// @gate       TestNewSolidNull_AEShipGate_AE2020,TestNewSolidNull_AEShipGate_AE2025
// @since      AE2020
// @incident   new-layer-types-scoping,nextitemid-must-include-layer-ids
// @alias      solid,纯色,固态层
func NewSolidLayer(c *Composition, name string, width, height int, rgb [3]float64) (*Layer, error) {
	return serializer.NewSolidLayer(c, name, width, height, rgb)
}

// @summary    Add a null-object layer to a composition
// @description Appends a new null-object layer and returns it. A null is a 100×100
//   solid-backed layer with the null flag set — AE's standard parenting helper.
//   The backing solid footage item is created alongside (see NewSolidLayer). The
//   layer's time span is re-homed to 0 → comp duration. The returned layer is
//   fully parsed.
//
//   Atomic (rides the gated solid-family creation path).
// @param      c     the composition to add the null to
// @param      name  null layer name (non-empty)
// @returns    the created null Layer
// @domain     layer-create
// @stability  stable
// @verify     ae-accept
// @gate       TestNewSolidNull_AEShipGate_AE2020,TestNewSolidNull_AEShipGate_AE2025
// @since      AE2020
// @incident   new-layer-types-scoping
// @alias      null,空对象,空层
func NewNullLayer(c *Composition, name string) (*Layer, error) {
	return serializer.NewNullLayer(c, name)
}

// @summary    Add an adjustment layer to a composition
// @description Appends a new adjustment layer and returns it. An adjustment layer
//   is a comp-sized white solid with the adjustment flag set: effects applied to
//   it affect every layer below it. The backing solid footage item is created
//   alongside, sized to the comp's current dimensions (see NewSolidLayer). The
//   layer's time span is re-homed to 0 → comp duration. The returned layer is
//   fully parsed.
//
//   Atomic (rides the gated solid-family creation path).
// @param      c     the composition to add the adjustment layer to
// @param      name  adjustment layer name (non-empty)
// @returns    the created adjustment Layer
// @domain     layer-create
// @stability  stable
// @verify     ae-accept
// @gate       TestNewSolidNull_AEShipGate_AE2020,TestNewSolidNull_AEShipGate_AE2025
// @since      AE2020
// @incident   new-layer-types-scoping
// @alias      adjustment,调整图层
func NewAdjustmentLayer(c *Composition, name string) (*Layer, error) {
	return serializer.NewAdjustmentLayer(c, name)
}

// @summary    Add a precomp (nested-comp) layer to a composition
// @description Adds a layer to parent whose source is the composition child (a
//   nested / pre-composed comp) and returns it. A precomp layer is an ordinary AV
//   layer whose source points at an existing comp item — the source comp already
//   lives in the project, so (unlike the solid family) no backing footage item is
//   created. An AE-native precomp layer is cloned and spliced in, its source
//   repointed at child, and its time span re-homed to 0 → parent duration. The
//   returned layer is fully parsed; its SourceComposition resolves to child.
//
//   Refuses when either comp is nil, the name is empty, parent == child, the two
//   comps are in different projects, or the nesting would create a circular
//   composition reference.
//
//   Atomic (snapshot + rollback on any parser warning).
// @param      parent  the composition that will contain the new layer
// @param      child   the composition to use as the layer's source
// @param      name    precomp layer name (non-empty)
// @returns    the created precomp Layer
// @domain     layer-create
// @stability  stable
// @verify     render-pixel
// @gate       TestMGPrecomp_AEShipGate_AE2020,TestMGPrecomp_AEShipGate_AE2025
// @since      AE2020
// @incident   precomp-layer-source-id-re
// @alias      precomp,预合成,嵌套合成
func NewPrecompLayer(parent, child *Composition, name string) (*Layer, error) {
	return serializer.NewPrecompLayer(parent, child, name)
}

// @summary    Remove a layer from a composition by index
// @description Removes the layer at the given 0-based index in the comp's layer
//   list. Returns an error on a refuse-case (index out of range, comp lacks its
//   item-list back-reference, target is the last layer, target is not an AV
//   layer, or back-reference corruption).
//
//   Reference cleanup mirrors AE's own delete behavior: any other layer whose
//   parent is the deleted layer has its parent reset to none, and any layer using
//   the deleted layer as a track-matte source has that reference cleared (when
//   the layer record is long enough — AE 2022 and earlier did not write that
//   field). The neighbor's track-matte intent flag is left untouched to match AE.
//   String-level references (expressions, render queue, essential graphics) are
//   out of scope — scrub them manually if needed.
//
//   Atomic (snapshot + rollback on any parser warning).
// @param      c      the composition to remove the layer from
// @param      index  0-based index of the layer to remove
// @domain     structural
// @stability  stable
// @verify     ae-accept
// @gate       TestStructuralOps_AEShipGate_AE2020,TestStructuralOps_AEShipGate_AE2025
// @since      AE2020
// @boundary   AE-gated on an AV solid carrier with layer-order DOM readback; non-AV layers and single-layer comps are refused
// @alias      delete layer,删图层,删除图层,remove layer
func DeleteLayer(c *Composition, index int) error { return serializer.DeleteLayer(c, index) }

// @summary    Clone a layer in place by index
// @description Clones the layer at the given 0-based index and inserts the clone
//   at that same position, pushing the source and everything below it down by one
//   (mirrors AE's layer.duplicate()). Returns the cloned layer, or an error on a
//   refuse-case.
//
//   Clone semantics: the clone gets a new monotonic layer ID; its chunk block is
//   a deep byte-clone of the source's, with only the ID overwritten — source,
//   parent, and track-matte references are copied verbatim (no footage
//   duplication, no reference rewrites). The name is caller-supplied (an explicit
//   name avoids silent duplicate-name confusion). Incoming references still
//   resolve to the source, not the clone.
//
//   Refuses on: empty name, index out of range, comp lacking its item-list
//   back-reference, a non-AV source (camera / light / audio behavior not yet
//   reverse-engineered), a source with an implicit track matte, or back-reference
//   corruption. An AE 23+ explicit track matte is allowed (the clone byte-copies
//   the matte reference verbatim).
//
//   Atomic (snapshot + rollback on any parser warning, including the ID bump).
// @param      c      the composition that owns the layer
// @param      index  0-based index of the layer to clone
// @param      name   name for the cloned layer (non-empty)
// @returns    the cloned Layer
// @domain     structural
// @stability  stable
// @verify     ae-accept
// @gate       TestStructuralOps_AEShipGate_AE2020,TestStructuralOps_AEShipGate_AE2025
// @since      AE2020
// @boundary   AE-gated on an AV solid (clone placed above the source, layer-order DOM readback); AE 23+ explicit matte allowed, implicit matte and non-AV refused
// @alias      duplicate layer,复制图层
func DuplicateLayer(c *Composition, index int, name string) (*Layer, error) {
	return serializer.DuplicateLayer(c, index, name)
}

// @summary    Deep-clone a layer into a composition at an index
// @description Deep-clones src into the comp's layer list at atIdx (0-based;
//   atIdx == len(layers) appends) and returns the inserted clone. src may live in
//   a sibling comp of the same project or in a different project (cross-project).
//
//   Same-project: the clone gets a new monotonic layer ID; its block is a deep
//   byte-clone of src with the ID set, the track matte reset (a cross-comp matte
//   source is invalid), and the parent reset (src's parent named a layer in src's
//   own comp). The source reference is verbatim (the shared footage / comp item);
//   the name matches AE's copyToComp (verbatim).
//
//   Cross-project: additionally imports src's reachable item closure (footage +
//   precomp, transitively) into the destination project at root level with fresh
//   item IDs, then remaps the clone's source and alternate-source through that
//   map. File-backed footage already present in the destination (matched by path)
//   is reused, not re-cloned; comps and solids / placeholders are always cloned.
//
//   Refuses on: nil src, missing destination back-reference, atIdx out of range,
//   a detached or same-comp src, a non-AV src, a direct precomp loop (same-project
//   only), or structural corruption. Cross-project also refuses a project with no
//   root folder or a dangling closure source.
//
//   Atomic (snapshot + rollback on any parser warning, including the ID bump).
// @param      c      the composition to insert into
// @param      src    the layer to clone (a sibling comp or another project)
// @param      atIdx  0-based insertion index (len(layers) appends)
// @returns    the inserted clone Layer
// @domain     structural
// @stability  stable
// @verify     ae-accept
// @gate       TestStructuralOps_AEShipGate_AE2020,TestStructuralOps_AEShipGate_AE2025
// @since      AE2020
// @boundary   same- and cross-project insert; same-project is auto-gated (layer-order DOM readback); cross-project has assert-gate coverage (6/6) but is not in the automated gate
// @alias      insert layer,插入图层,跨工程复制,copy to comp
func InsertLayer(c *Composition, src *Layer, atIdx int) (*Layer, error) {
	return serializer.InsertLayer(c, src, atIdx)
}

// @summary    Reorder a layer within a composition
// @description Moves the layer at from to position to in the comp's layer list
//   (both 0-based). The source layer's entire chunk block is spliced out and
//   re-inserted at the target slot; afterward every layer's index is refreshed to
//   match its new position. from == to is a no-op.
//
//   Refuses when from or to is out of range, the comp lacks its item-list
//   back-reference, or the source block is corrupt. Unlike DeleteLayer and
//   DuplicateLayer, MoveLayer ignores layer type and track matte — a pure reorder
//   works for AV / camera / light / audio / shape / text / matted layers alike.
//
//   Atomic (snapshot + rollback on any parser warning).
// @param      c     the composition whose layers to reorder
// @param      from  0-based current index of the layer to move
// @param      to    0-based target index
// @domain     structural
// @stability  stable
// @verify     ae-accept
// @gate       TestStructuralOps_AEShipGate_AE2020,TestStructuralOps_AEShipGate_AE2025
// @since      AE2020
// @boundary   pure reorder (type-agnostic); AE-gated on an AV solid with layer-order DOM readback
// @alias      move layer,图层排序,reorder layer,改层级
func MoveLayer(c *Composition, from, to int) error { return serializer.MoveLayer(c, from, to) }

// @summary    Move a layer to the top of the layer stack
// @description Moves the layer to position 0 (top of the stack in AE's display,
//   AE-index 1). A convenience wrapper over MoveLayer.
// @param      l  the layer to move to the top
// @domain     structural
// @stability  stable
// @verify     ae-accept
// @gate       TestStructuralOps_AEShipGate_AE2020,TestStructuralOps_AEShipGate_AE2025
// @since      AE2020
// @boundary   MoveLayer convenience wrapper (move to top); AE-gated via structural ops
// @alias      move to beginning,移到顶部,置顶
func MoveToBeginning(l *Layer) error { return serializer.MoveToBeginning(l) }

// @summary    Move a layer to the bottom of the layer stack
// @description Moves the layer to the last position (bottom of the stack in AE's
//   display). A convenience wrapper over MoveLayer.
// @param      l  the layer to move to the bottom
// @domain     structural
// @stability  stable
// @verify     ae-accept
// @gate       TestStructuralOps_AEShipGate_AE2020,TestStructuralOps_AEShipGate_AE2025
// @since      AE2020
// @boundary   MoveLayer convenience wrapper (move to bottom); AE-gated via structural ops
// @alias      move to end,移到底部,置底
func MoveToEnd(l *Layer) error { return serializer.MoveToEnd(l) }

// @summary    Move a layer to just after another layer
// @description Moves the layer to the slot immediately after other (the receiver
//   lands just below other in the stack). Returns an error if other belongs to a
//   different comp, other == l, or either layer lacks a comp back-reference. A
//   convenience wrapper over MoveLayer.
// @param      l      the layer to move
// @param      other  the layer to position l after
// @domain     structural
// @stability  stable
// @verify     ae-accept
// @gate       TestStructuralOps_AEShipGate_AE2020,TestStructuralOps_AEShipGate_AE2025
// @since      AE2020
// @boundary   MoveLayer convenience wrapper (move after other); AE-gated via structural ops
// @alias      move after,移到之后
func MoveAfter(l, other *Layer) error { return serializer.MoveAfter(l, other) }

// @summary    Move a layer to just before another layer
// @description Moves the layer to the slot immediately before other (the receiver
//   lands just above other in the stack). A convenience wrapper over MoveLayer.
// @param      l      the layer to move
// @param      other  the layer to position l before
// @domain     structural
// @stability  stable
// @verify     ae-accept
// @gate       TestStructuralOps_AEShipGate_AE2020,TestStructuralOps_AEShipGate_AE2025
// @since      AE2020
// @boundary   MoveLayer convenience wrapper (move before other); AE-gated via structural ops
// @alias      move before,移到之前
func MoveBefore(l, other *Layer) error { return serializer.MoveBefore(l, other) }

// @summary    Append a composition marker at a given time
// @description Appends a new composition marker at the given time (in seconds) and
//   returns it for further Set* calls. The new marker is a clean point marker: no
//   duration, no label color, empty text fields.
//
//   To avoid reverse-engineering the canonical defaults of the marker's opaque
//   metadata, the new marker clones an existing marker's block verbatim (opaque
//   preservation), then resets the time plus the known semantic fields (duration,
//   label) to zero. This means the comp must already have at least one marker to
//   serve as the clone template; AddMarker returns an error otherwise. Markers
//   are appended without re-sorting.
// @param      c        the composition to add the marker to
// @param      seconds  marker time in seconds
// @returns    the created Marker
// @domain     structural
// @stability  stable
// @verify     ae-accept
// @gate       TestMarker_AEShipGate_AE2020,TestMarker_AEShipGate_AE2025
// @since      AE2020
// @boundary   requires the comp to already have >= 1 marker (empty-comp seeding is deferred); tail-insert without sorting
// @alias      marker,标记,合成标记,comp marker
func AddMarker(c *Composition, seconds float64) (*Marker, error) {
	return serializer.AddMarker(c, seconds)
}

// @summary    Remove a marker from its owning comp or layer
// @description Deletes this marker from its owning composition or layer marker set:
//   it splices out the marker's keyframe block, decrements the count, removes the
//   marker's record, shifts the trailing markers' offsets down, and drops the
//   marker from the public list. The receiver is detached afterward — a second
//   RemoveMarker (or any Set*) errors.
//
//   Errors (project untouched) when the marker was built outside the parser, is
//   already detached, or its chunk references are inconsistent.
// @param      m  the marker to remove
// @domain     structural
// @stability  stable
// @verify     ae-accept
// @gate       TestMarker_AEShipGate_AE2020,TestMarker_AEShipGate_AE2025
// @since      AE2020
// @alias      remove marker,删标记
func RemoveMarker(m *Marker) error { return serializer.RemoveMarker(m) }

// @summary    Insert a keyframe into a property's stream
// @description Builds a new keyframe block and inserts it into the property's
//   keyframe stream (time-sorted; ties land after existing keys at the same
//   time), then updates the count header. Returns the new keyframe and its index
//   in the property's keyframe list.
//
//   Requires the property to already have at least one keyframe so the new block
//   can clone the existing layout. For properties without keyframes, use a static
//   value or the Animate* APIs — synthesizing the keyframe chunks from scratch is
//   not supported here. value follows the SetValue rules: a float64 for a 1D
//   property, or a []float64 (length == component count) for a multi-component
//   property. The new keyframe's interpolation is Linear/Linear with zeroed ease
//   and tangents; refine it via the returned keyframe's setters.
// @param      p      the property to insert into (must already have >= 1 keyframe)
// @param      time   keyframe time in seconds
// @param      value  keyframe value (float64 for 1D, []float64 for multi-component)
// @returns    the new Keyframe and its index in the property's keyframe list
// @domain     keyframe
// @stability  stable
// @verify     ae-accept
// @gate       TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025
// @since      AE2020
// @boundary   requires >= 1 existing keyframe to clone layout (from-scratch unsupported — use the Animate* APIs); AE-gated on keyframe-count readback; new keyframe defaults to Linear
// @alias      insert keyframe,插入关键帧,加关键帧
func InsertKeyframe(p *Property, time float64, value any) (*Keyframe, int, error) {
	return serializer.InsertKeyframe(p, time, value)
}

// @summary    Delete a keyframe from a property's stream by index
// @description Removes the keyframe at index i from the property's keyframe stream
//   and decrements the count header. Returns an error when i is out of range or
//   the property has no keyframe stream.
// @param      p  the property to delete from
// @param      i  0-based index of the keyframe to delete
// @domain     keyframe
// @stability  stable
// @verify     ae-accept
// @gate       TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025
// @since      AE2020
// @boundary   AE-gated on keyframe-count readback
// @alias      delete keyframe,删关键帧,移除关键帧
func DeleteKeyframe(p *Property, i int) error { return serializer.DeleteKeyframe(p, i) }

// @summary    Toggle Separate Dimensions on a Position property
// @description Toggles AE's "Separate Dimensions" on a Position leader, in both
//   directions. Static Position (2D and 3D) and animated Position on 3D layers
//   (near-linear leader path-ease) are double-version ship-gated. An animated
//   leader routes to the keyframe-stream migration paths; animated cases outside
//   that shipped subset — a 2D layer, or a leader carrying custom spatial-path
//   temporal ease — are refused rather than written.
//
//   Separating migrates the leader's value into per-axis Position followers
//   (X/Y, plus Z for 3D layers, which is synthesized) and resets the leader;
//   merging takes the value back into the leader and removes the per-axis
//   followers. The only fallible step (re-parsing a synthesized Z follower) runs
//   before any in-place mutation, so a failure leaves the project untouched.
// @param      p          the Position leader property to toggle
// @param      separated  true to separate dimensions, false to merge them
// @domain     structural
// @stability  stable
// @verify     ae-accept
// @gate       TestSeparateDims_AEShipGate_AE2020,TestSeparateDims_AEShipGate_AE2025,TestMergeDims_AEShipGate_AE2020,TestMergeDims_AEShipGate_AE2025
// @since      AE2020
// @boundary   static 2D/3D and near-linear animated 3D are gated; animated 2D and custom spatial ease are refused
// @incident   separate-dimensions-write-mechanics
// @alias      separate dimensions,分离维度,position 分离,X Y 分离
func SetDimensionsSeparated(p *Property, separated bool) error {
	return serializer.SetDimensionsSeparated(p, separated)
}

// @summary    Remove a group from its parent indexed group
// @description Deletes this group from its parent indexed group. The receiver must
//   be a direct child of an indexed group (Effect Parade / Mask Parade / Root
//   Vectors Group / Text Animators); RemovePropertyGroup returns an error
//   otherwise, mirroring AE's refuse.
//
//   Atomic (snapshot + rollback on any parser warning).
// @param      g  the property group to remove (a direct child of an indexed group)
// @domain     structural
// @stability  alpha
// @verify     ae-accept
// @gate       TestPropStructRemove_AEShipGate_AE2020,TestPropStructRemove_AEShipGate_AE2025
// @since      AE2020
// @boundary   Effect Parade and Text Animators are gated in both versions; Mask and Root Vectors use the same mechanism but are not separately gated
// @incident   property-indexed-group-structural-re
// @alias      remove property group,删属性组,删动画器,删 indexed group 子项
func RemovePropertyGroup(g *AEPropertyGroup) error { return serializer.RemovePropertyGroup(g) }

// @summary    Reorder a group within its parent indexed group
// @description Reorders this group to position index (0-based) among its parent
//   indexed group's children; index is range-checked against the current child
//   count. Mirrors AE's PropertyBase.moveTo (which is 1-based; this API is 0-based
//   per project convention).
// @param      g      the property group to reorder (a direct child of an indexed group)
// @param      index  0-based target position among the parent's children
// @domain     structural
// @stability  alpha
// @verify     ae-accept
// @gate       TestPropStructMove_AEShipGate_AE2020,TestPropStructMove_AEShipGate_AE2025
// @since      AE2020
// @boundary   same indexed-group gate coverage as RemovePropertyGroup
// @incident   property-indexed-group-structural-re
// @alias      move property group,属性组排序,reorder group
func MovePropertyGroup(g *AEPropertyGroup, index int) error {
	return serializer.MovePropertyGroup(g, index)
}

// @summary    Duplicate a group within its parent indexed group
// @description Inserts a copy of this group immediately after it among its parent
//   indexed group's children — mirroring AE's PropertyBase.duplicate() — and
//   returns the clone. The receiver must be a direct child of an indexed group
//   (Effect Parade / Mask Parade / Root Vectors Group / Text Animators);
//   DuplicatePropertyGroup returns an error otherwise.
//
//   The clone reuses the source's match-name and on-disk payload verbatim. AE's
//   own duplicate additionally persists a deduplicated display name (the
//   localized "<name> 2"); this call deliberately does NOT synthesize that suffix
//   — it needs AE's localization database we do not carry, and a clone with no
//   display-name override is byte-for-byte an "add the same effect twice"
//   project, which AE accepts and re-derives the runtime dedup name from on open.
//   The persisted suffix is cosmetic; the structural duplicate is faithful.
//
//   Atomic (snapshot + rollback on any parser warning, or if the re-parse fails
//   to reproduce exactly one clone).
// @param      g  the property group to duplicate (a direct child of an indexed group)
// @returns    the cloned AEPropertyGroup
// @domain     structural
// @stability  alpha
// @verify     ae-accept
// @gate       TestPropStructDuplicate_AEShipGate_AE2020,TestPropStructDuplicate_AEShipGate_AE2025
// @since      AE2020
// @boundary   the display-name suffix is not synthesized (AE recomputes it); same indexed-group gate coverage as RemovePropertyGroup
// @incident   property-indexed-group-structural-re
// @alias      duplicate property group,复制属性组,复制效果
func DuplicatePropertyGroup(g *AEPropertyGroup) (*AEPropertyGroup, error) {
	return serializer.DuplicatePropertyGroup(g)
}

// @summary    Append a built-in effect to a layer
// @description Appends an effect to the layer's Effect Parade and returns the
//   parsed Effect, so the caller can immediately tune its parameters via
//   SetEffectParam (or the property tree after a Reopen).
//
//   effectMatchName must be one of SupportedEffects(); the effect's full
//   parameter sub-tree is supplied from an embedded AE-native template, which is
//   why only reverse-engineered effects are addable. AE looks the effect up by
//   match-name at load, so the named plugin must be installed in the opening AE —
//   the seeded effects are built-ins present since before the AE 2020 read floor
//   and are version-portable (the AE-2020-extracted bytes are accepted by AE
//   2025).
//
//   A parsed layer with no effects has no Effect Parade group at all (AE only
//   persists the parade once at least one effect exists); AddEffect splices an
//   empty parade in the AE-native form, then adds the effect into it. Camera and
//   light layers are refused (AE does not allow effects on them), as are layers
//   built by the structural New* APIs that were never parsed — call Reopen first
//   and add effects to the re-parsed layer.
//
//   Atomic (snapshot + rollback on any parser warning).
// @param      layer            the parsed layer to add the effect to
// @param      effectMatchName  the effect match-name (one of SupportedEffects)
// @returns    the created Effect
// @domain     effect
// @stability  stable
// @verify     ae-accept
// @gate       TestAddEffect_AEShipGate_AE2020,TestAddEffect_AEShipGate_AE2025,TestAddEffectWave5_AEShipGate_AE2020,TestAddEffectWave5_AEShipGate_AE2025,TestAddEffectWave6_AEShipGate_AE2020,TestAddEffectWave6_AEShipGate_AE2025,TestAddEffectWave7_AEShipGate_AE2020,TestAddEffectWave7_AEShipGate_AE2025,TestAddEffectWave9_AEShipGate_AE2020,TestAddEffectWave9_AEShipGate_AE2025,TestAddEffectAudio_AEShipGate_AE2020,TestAddEffectAudio_AEShipGate_AE2025,TestAddEffectWave11_AEShipGate_AE2020,TestAddEffectWave11_AEShipGate_AE2025,TestAddEffectWave12_AEShipGate_AE2020,TestAddEffectWave12_AEShipGate_AE2025
// @since      AE2020
// @boundary   covers the embedded library of 216 built-in effects (ADBE + Cycore CC + keying / simulation / utility + 10 audio effects accepted only on audio layers, plus 11 layer-reference effects whose source is set via SetEffectLayerParam); effects that open a file/font dialog and deprecated names are not included; camera/light and un-Reopened fresh layers are refused; no per-effect typed helpers yet
// @incident   add-effect-splice-re
// @alias      effect,特效,加效果,blur,模糊,glow,cc,cycore,lumetri,keying,抠像,audio,音频,声音,reverb,delay,eq
func AddEffect(layer *Layer, effectMatchName string) (*Effect, error) {
	return serializer.AddEffect(layer, effectMatchName)
}

// @summary    List the effect match-names AddEffect can add
// @description Returns the sorted effect match-names AddEffect can add from an
//   embedded template.
// @returns    the sorted list of supported effect match-names
// @domain     meta
// @stability  stable
// @verify     none
// @since      AE2020
// @alias      effects,效果列表,supported effects
func SupportedEffects() []string { return serializer.SupportedEffects() }

// @summary    Apply an .ffx Animation Preset pseudo effect to a layer
// @description Splices the pseudo effect carried by an After Effects Animation
//   Preset (.ffx) into the layer's Effect Parade and returns the parsed Effect.
//   ffxBytes is the raw .ffx file content.
//
//   A pseudo effect (built with the Pseudo Effect Maker) is a user-defined effect
//   — a named group of standard controls (slider / color / checkbox / point /
//   angle …) that looks like a native effect. Unlike a native effect (looked up
//   by match-name in the opening AE), a pseudo effect's full control definition
//   travels inside the .ffx, so it is self-contained: AE renders it from the
//   saved .aep bytes without the preset ever being registered. This makes
//   ApplyPseudoEffect the offline, pure-Go equivalent of AE's applyPreset — no
//   running AE required.
//
//   The effect is spliced with its controls at their defined defaults — the
//   authored values stored in the .ffx are not yet preserved (AE rejects the raw
//   .ffx value entries spliced into a parade). All controls are present and
//   tunable in AE; programmatic tuning via Set* after a Reopen is a future step.
//
//   Refused (same as AddEffect): camera / light layers, and New*-built layers
//   never parsed (call Reopen first). Returns an error for a malformed .ffx (not
//   a RIFX FaFX form, missing payload, or no extractable match-name).
// @param      layer     the parsed layer to apply the pseudo effect to
// @param      ffxBytes  the raw .ffx Animation Preset file content
// @returns    the created Effect
// @domain     effect
// @stability  alpha
// @verify     ae-accept
// @gate       TestApplyPseudoEffect_AEShipGate_AE2020,TestApplyPseudoEffect_AEShipGate_AE2025
// @since      AE2020
// @boundary   applies controls at their defined defaults — the .ffx authored values are not preserved (AE rejects raw .ffx value entries) and Set* tuning is not yet wired; camera/light and un-Reopened fresh layers refused; reading an .ffx into the scene model is not implemented
// @incident   add-effect-splice-re
// @alias      pseudo effect,pseudoeffect,伪效果,自定义效果,ffx,animation preset,动画预设,applypreset,custom effect
func ApplyPseudoEffect(layer *Layer, ffxBytes []byte) (*Effect, error) {
	return serializer.ApplyPseudoEffect(layer, ffxBytes)
}

// @summary    Apply an .ffx pseudo effect with a custom display name
// @description ApplyPseudoEffect with a custom effect-instance display name (the
//   label in AE's Effect Controls / timeline). displayName may be any UTF-8
//   string — including CJK such as "伪效果" — because AE stores the name as a
//   length-prefixed byte record and the library counts bytes, not runes, so
//   multi-byte names round-trip exactly. An empty displayName keeps the .ffx's
//   own name. The match-name (AE's ASCII lookup key) is unaffected. All other
//   behavior matches ApplyPseudoEffect.
// @param      layer        the parsed layer to apply the pseudo effect to
// @param      ffxBytes     the raw .ffx Animation Preset file content
// @param      displayName  custom instance display name (any UTF-8; empty keeps the .ffx name)
// @returns    the created Effect
// @domain     effect
// @stability  alpha
// @verify     ae-accept
// @gate       TestApplyPseudoEffectNamed_CJK_AEShipGate_AE2020,TestApplyPseudoEffectNamed_CJK_AEShipGate_AE2025
// @since      AE2020
// @boundary   same as ApplyPseudoEffect, plus a custom instance display name supporting any UTF-8 (the display name lives in the value group, byte-length counted so multi-byte names round-trip exactly); the match-name stays ASCII
// @incident   add-effect-splice-re
// @alias      pseudo effect named,伪效果命名,中文效果名,自定义显示名,cjk effect name,utf8 effect name
func ApplyPseudoEffectNamed(layer *Layer, ffxBytes []byte, displayName string) (*Effect, error) {
	return serializer.ApplyPseudoEffectNamed(layer, ffxBytes, displayName)
}

// PseudoControlKind is the AE control type of a from-scratch pseudo-effect
// control (see BuildPseudoEffect).
type PseudoControlKind = serializer.PseudoControlKind

// Pseudo-effect control kinds for BuildPseudoEffect.
const (
	PseudoSlider     = serializer.PseudoSlider     // Slider (scalar)
	PseudoColor      = serializer.PseudoColor      // Color swatch
	PseudoCheckbox   = serializer.PseudoCheckbox   // Checkbox
	PseudoAngle      = serializer.PseudoAngle      // Angle dial
	PseudoPoint      = serializer.PseudoPoint      // 2D Point
	PseudoPoint3D    = serializer.PseudoPoint3D    // 3D Point
	PseudoDropdown   = serializer.PseudoDropdown   // Dropdown menu (Options)
	PseudoGroupStart = serializer.PseudoGroupStart // Group start (nest controls until PseudoGroupEnd)
	PseudoGroupEnd   = serializer.PseudoGroupEnd   // Group end
	PseudoLabel      = serializer.PseudoLabel      // Static text label (no value)
	PseudoLayer      = serializer.PseudoLayer      // Layer picker (LayerID, 0 = None)
)

// PseudoLabelCodepage selects the ANSI codepage a pseudo effect's control labels
// are encoded in (see WithLabelCodepage / BuildPseudoEffect). The zero value is
// PseudoLabelGBK.
type PseudoLabelCodepage = serializer.PseudoLabelCodepage

// Control-label target codepages for BuildPseudoEffect's WithLabelCodepage.
const (
	PseudoLabelGBK      = serializer.PseudoLabelGBK      // Simplified Chinese (GBK / cp936) — default; ASCII passes through
	PseudoLabelShiftJIS = serializer.PseudoLabelShiftJIS // Japanese (Shift-JIS / cp932)
)

// PseudoOption customizes BuildPseudoEffect. The only option today is
// WithLabelCodepage; the type leaves room for further effect-wide knobs.
type PseudoOption func(*pseudoConfig)

type pseudoConfig struct{ codepage PseudoLabelCodepage }

// @summary    Set the ANSI codepage for a pseudo effect's control labels
// @description Sets the ANSI codepage the effect's control labels are encoded in
//   (default PseudoLabelGBK). Pass PseudoLabelShiftJIS for a Japanese effect.
//   ASCII labels are unaffected. See BuildPseudoEffect for the rationale.
// @param      cp  the target label codepage
// @returns    a PseudoOption for BuildPseudoEffect
// @domain     effect
// @stability  alpha
// @verify     ae-accept
// @gate       TestBuildPseudoEffect_AEShipGate_AE2020,TestBuildPseudoEffect_AEShipGate_AE2025
// @since      AE2020
// @boundary   selects the target ANSI codepage for control-label names (GBK simplified-Chinese default / Shift-JIS Japanese); structurally AE-accepted (the codepage only changes name-field bytes), but correct display is byte-equivalence-verified against AE's native locale output and only renders correctly on a matching-locale Windows
// @incident   pseudo-control-label-ansi-codepage
// @alias      pseudo label codepage,控件标签码页,日语标签,japanese label,shift-jis,gbk,locale,with label codepage
func WithLabelCodepage(cp PseudoLabelCodepage) PseudoOption {
	return func(c *pseudoConfig) { c.codepage = cp }
}

// PseudoControl is one control of a from-scratch pseudo effect: a kind + the
// label shown in AE's Effect Controls, plus optional per-kind customization
// (Slider Min/Max/Default, Angle Default, Checkbox Checked, Color, Dropdown
// Options, Point/Point3D PointDefault, Layer LayerID). The optional fields' zero
// values reproduce AE's plain type defaults.
type PseudoControl = serializer.PseudoControl

// @summary    Build a pseudo effect from scratch in Go and apply it
// @description Builds a Pseudo Effect entirely in Go — no .ffx file and no running
//   AE — and splices it into the layer's Effect Parade. This is the authoring
//   direction (what the Pseudo Effect Maker does in AE's UI), brought offline:
//   name a set of controls and get a live, AE-accepted custom effect on the
//   layer.
//
//   uid is the per-effect unique id (the "<uID>" in match-name
//   "Pseudo/<uID>/<name>"); name is the match-name segment; displayName is the
//   effect label (any UTF-8, incl. CJK — empty falls back to name); controls are
//   the controls in order. Every control definition is synthesized field-by-field
//   from the reverse-engineered layout, so the supported kinds are exactly what
//   PseudoControlKind enumerates.
//
//   Per-control customization (verified by AE read-back): slider min/max +
//   default, angle default, checkbox checked, color RGBA, dropdown options +
//   selected index, point / 3D-point default, and layer-picker binding; zero
//   values give AE's plain type defaults. A point default is a fraction of the
//   host layer's coordinate space (e.g. {0.25, 0.125} on a 400×400 source-less
//   layer reads back as [100, 50]). Group and Label kinds are flat marker
//   controls — AE's pseudo-effect "groups" are a visual grouping in the Effect
//   Controls panel, not a nested property group.
//
//   Control labels are written into the name field, which AE decodes in the
//   viewing machine's system ANSI codepage (not UTF-8). ASCII labels are exact
//   everywhere; a CJK label is encoded in the codepage chosen by WithLabelCodepage
//   (default GBK simplified-Chinese; Shift-JIS for Japanese) — byte-identical to
//   AE's own output on that locale, so it displays correctly on a matching Windows
//   and mojibakes elsewhere (an AE architecture limit).
//
//   Refused (same as AddEffect): camera / light layers, and New*-built layers
//   never parsed (call Reopen first).
// @param      layer        the parsed layer to add the pseudo effect to
// @param      uid          per-effect unique id (the match-name "<uID>" segment)
// @param      name         the match-name segment
// @param      displayName  effect label (any UTF-8; empty falls back to name)
// @param      controls     the controls to build, in order
// @param      opts         optional effect-wide options (e.g. WithLabelCodepage)
// @returns    the created Effect
// @domain     effect
// @stability  alpha
// @verify     ae-accept
// @gate       TestBuildPseudoEffect_AEShipGate_AE2020,TestBuildPseudoEffect_AEShipGate_AE2025,TestBuildPseudoEffectRich_AEShipGate_AE2020,TestBuildPseudoEffectRich_AEShipGate_AE2025,TestBuildPseudoEffectValueEntry_AEShipGate_AE2020,TestBuildPseudoEffectValueEntry_AEShipGate_AE2025
// @since      AE2020
// @boundary   synthesizes a pseudo effect from scratch (no .ffx, no AE, no cloned template); all control kinds (slider/color/checkbox/angle/point/point3d/dropdown/group/label/layer) read back live in AE; group/label are flat marker controls (only the built-in Compositing Options is truly nested); CJK labels are byte-equivalence-verified (not ship-gated) and locale-dependent; camera/light and un-Reopened fresh layers refused
// @incident   add-effect-splice-re
// @alias      build pseudo effect,从零造伪效果,pseudo effect maker,authoring,造效果,自定义控件,slider color checkbox dropdown group label layer point,slider min max,自定义范围,下拉菜单,分组,标签,图层选择,点坐标,中文标签,cjk label,gbk,离线造伪效果
func BuildPseudoEffect(layer *Layer, uid, name, displayName string, controls []PseudoControl, opts ...PseudoOption) (*Effect, error) {
	cfg := pseudoConfig{codepage: PseudoLabelGBK}
	for _, o := range opts {
		o(&cfg)
	}
	return serializer.BuildPseudoEffect(layer, uid, name, displayName, controls, cfg.codepage)
}

// AddTextOpacityAnimator adds a per-character Opacity animator with a Range
// Selector to a text layer — the kinetic-typography primitive (fade / wipe text
// in or out one character at a time). opacity (0–100) is applied to the
// selected characters; rangeStart / rangeEnd / rangeOffset are the Range
// Selector bounds in percent. A static reveal frame, e.g. opacity 0 + start 0 +
// end 50, hides the first ~half of the characters; animate the reveal over time
// by keyframing the Range Offset with AnimateTextRangeOffset.
//
// Text animators live in the "ADBE Text Animators" indexed group nested inside
// the layer's "ADBE Text Properties" group (NOT in the btdk document). Fresh
// text layers (NewTextLayer) carry no Animators group, so the first animator
// splices the whole group into Text Properties; later animators append into it.
// The animator's parameter sub-tree (Selectors + Animator Properties) is
// supplied from an embedded AE-native template authored with every cdat slot
// materialized (AE elides defaults), which the call overwrites with the
// supplied values — the same embed-AE-bytes + (tdmn, payload) splice vein
// AddEffect uses.
//
// Refused: non-text layers, and text layers built by New* that were never
// parsed (call aep.Reopen first). Returns a stand-in group node referencing the
// spliced animator chunk.
//
// Alpha / structural — the animator chunk structure is RE'd + double-version
// render-gated, but the typed parameter accessors are not yet wired; tune
// further via Reopen. Free function (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextAnimator_AEShipGate_AE2020,TestTextAnimator_AEShipGate_AE2025 incident=text-animator-create-re boundary="typed param accessor 未接,经 Reopen 调;未 parse 的 fresh 层 refused" alias="text opacity animator,文字不透明度动画,kinetic typography,逐字,打字机"
func AddTextOpacityAnimator(layer *Layer, opacity, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextOpacityAnimator(layer, opacity, rangeStart, rangeEnd, rangeOffset)
}

// AddTextPositionAnimator adds a per-character Position 3D animator with a Range
// Selector to a text layer — the kinetic-typography primitive that slides /
// drops characters into place one at a time. x / y / z is the position offset
// (pixels) applied to the selected characters; rangeStart / rangeEnd /
// rangeOffset are the Range Selector bounds in percent. The canonical reveal:
// offset (0, -100, 0), Start=0/End=100, then sweep the Range Offset 0→100 over
// time with AnimateTextRangeOffset — the displacement lands the characters as
// the selection window slides off them.
//
// Same mechanics as AddTextOpacityAnimator (it shares the splice + Range
// Selector path); the difference is the embedded template drives a Position 3D
// leaf (a spatial 3-component cdat) instead of the scalar Opacity. Fresh text
// layers carry no Animators group, so the first animator splices the whole group
// into "ADBE Text Properties"; later animators append into it.
//
// Refused: non-text layers, and text layers built by New* that were never parsed
// (call aep.Reopen first). Returns a stand-in group node referencing the spliced
// animator chunk.
//
// Alpha / structural — the animator chunk structure is RE'd + double-version
// render-gated, but the typed parameter accessors are not yet wired; tune
// further via Reopen. Free function (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextPosAnimator_AEShipGate_AE2020,TestTextPosAnimator_AEShipGate_AE2025 incident=text-animator-create-re boundary="typed param accessor 未接,经 Reopen 调;未 parse 的 fresh 层 refused" alias="text position animator,文字位移动画,字符滑入,drop in"
func AddTextPositionAnimator(layer *Layer, x, y, z, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextPositionAnimator(layer, x, y, z, rangeStart, rangeEnd, rangeOffset)
}

// AddTextScaleAnimator adds a per-character Scale 3D animator with a Range
// Selector to a text layer — the kinetic-typography primitive that pops / grows
// characters into place one at a time. sx / sy / sz is the scale percent (100 =
// unchanged) applied to the selected characters; rangeStart / rangeEnd /
// rangeOffset are the Range Selector bounds in percent. The canonical reveal:
// scale (0, 0, 100) for a pop-in (or an oversize like 220 for a shrink-in),
// Start=0/End=100, then sweep the Range Offset 0→100 over time with
// AnimateTextRangeOffset — the scale resolves to 100% as the selection window
// slides off the characters.
//
// Same mechanics as AddTextPositionAnimator (shared splice + Range Selector
// path); the embedded template drives a Scale 3D leaf (a 3-component cdat)
// instead of Position. Fresh text layers carry no Animators group, so the first
// animator splices the whole group into "ADBE Text Properties"; later animators
// append into it.
//
// Refused: non-text layers, and text layers built by New* that were never parsed
// (call aep.Reopen first). Returns a stand-in group node referencing the spliced
// animator chunk.
//
// Alpha / structural — the animator chunk structure is RE'd + double-version
// render-gated, but the typed parameter accessors are not yet wired; tune
// further via Reopen. Free function (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextScaleAnimator_AEShipGate_AE2020,TestTextScaleAnimator_AEShipGate_AE2025 incident=text-animator-create-re boundary="typed param accessor 未接,经 Reopen 调;未 parse 的 fresh 层 refused" alias="text scale animator,文字缩放动画,字符弹入,pop in"
func AddTextScaleAnimator(layer *Layer, sx, sy, sz, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextScaleAnimator(layer, sx, sy, sz, rangeStart, rangeEnd, rangeOffset)
}

// AddTextRotationAnimator adds a per-character Rotation animator with a Range
// Selector to a text layer — the kinetic-typography primitive that spins
// characters into place one at a time. rotation is the angle in degrees applied
// to the selected characters (each rotates about its own anchor); rangeStart /
// rangeEnd / rangeOffset are the Range Selector bounds in percent. The canonical
// reveal: rotation 90, Start=0/End=100, then sweep the Range Offset 0→100 over
// time with AnimateTextRangeOffset — the rotation resolves to 0° as the
// selection window slides off the characters.
//
// Same mechanics as AddTextOpacityAnimator (shared splice + Range Selector
// path); Rotation is a 1D scalar like Opacity, so the embedded template drives a
// Rotation leaf and the value is the angle (degrees). Fresh text layers carry no
// Animators group, so the first animator splices the whole group into "ADBE Text
// Properties"; later animators append into it.
//
// Refused: non-text layers, and text layers built by New* that were never parsed
// (call aep.Reopen first). Returns a stand-in group node referencing the spliced
// animator chunk.
//
// Alpha / structural — the animator chunk structure is RE'd + double-version
// render-gated, but the typed parameter accessors are not yet wired; tune
// further via Reopen. Free function (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextRotAnimator_AEShipGate_AE2020,TestTextRotAnimator_AEShipGate_AE2025 incident=text-animator-create-re boundary="typed param accessor 未接,经 Reopen 调;未 parse 的 fresh 层 refused" alias="text rotation animator,文字旋转动画,字符旋转,spin in"
func AddTextRotationAnimator(layer *Layer, rotation, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextRotationAnimator(layer, rotation, rangeStart, rangeEnd, rangeOffset)
}

// AddTextColorAnimator adds a per-character Fill Color animator with a Range
// Selector to a text layer — the kinetic-typography primitive that tints
// characters one at a time (e.g. a colour wipe sweeping across the text). r / g
// / b / a is the target colour applied to the selected characters (each channel
// 0..1); rangeStart / rangeEnd / rangeOffset are the Range Selector bounds in
// percent. The canonical reveal: set a target colour, Start=0/End=100, then
// sweep the Range Offset 0→100 over time with AnimateTextRangeOffset — the
// colour applies to the selected characters and resolves to the base text colour
// as the selection window slides off them.
//
// Same mechanics as AddTextOpacityAnimator (shared splice + Range Selector
// path); the embedded template drives a Fill Color leaf (a 4-channel colour
// cdat) which AE stores as [A,R,G,B] × 255 f64 BE, the same on-disk encoding as
// shape Fill/Stroke. Fresh text layers carry no Animators group, so the first
// animator splices the whole group into "ADBE Text Properties"; later animators
// append into it.
//
// Refused: non-text layers, and text layers built by New* that were never parsed
// (call aep.Reopen first). Returns a stand-in group node referencing the spliced
// animator chunk.
//
// Alpha / structural — the animator chunk structure is RE'd + double-version
// render-gated, but the typed parameter accessors are not yet wired; tune
// further via Reopen. Free function (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextColorAnimator_AEShipGate_AE2020,TestTextColorAnimator_AEShipGate_AE2025 incident=text-animator-create-re boundary="typed param accessor 未接,经 Reopen 调;未 parse 的 fresh 层 refused" alias="text color animator,文字颜色动画,颜色擦除,color wipe"
func AddTextColorAnimator(layer *Layer, r, g, b, a, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextColorAnimator(layer, r, g, b, a, rangeStart, rangeEnd, rangeOffset)
}

// AddTextFillOpacityAnimator adds a per-character Fill Opacity animator with a
// Range Selector to a text layer — like AddTextOpacityAnimator, but it fades only
// the glyph fill (leaving any stroke intact). opacity (0–100) is applied to the
// selected characters; rangeStart / rangeEnd / rangeOffset are the Range Selector
// bounds in percent. Sweep the Range Offset over time with AnimateTextRangeOffset
// for a fill-only reveal.
//
// Same mechanics as AddTextOpacityAnimator (shared splice + Range Selector path);
// Fill Opacity is a 1D scalar with the same on-disk layout as Opacity.
//
// Refused: non-text layers, and text layers built by New* that were never parsed
// (call aep.Reopen first). Returns a stand-in group node referencing the spliced
// animator chunk.
//
// Alpha / structural — RE'd + double-version render-gated; typed parameter
// accessors are not yet wired (tune via Reopen). Free function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextNeighborAnimators_AEShipGate_AE2020,TestTextNeighborAnimators_AEShipGate_AE2025 incident=text-animator-create-re boundary="typed param accessor 未接,经 Reopen 调" alias="text fill opacity,填充不透明度动画"
func AddTextFillOpacityAnimator(layer *Layer, opacity, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextFillOpacityAnimator(layer, opacity, rangeStart, rangeEnd, rangeOffset)
}

// AddTextStrokeOpacityAnimator adds a per-character Stroke Opacity animator with
// a Range Selector to a text layer — it fades only the glyph stroke. opacity
// (0–100) is applied to the selected characters' stroke; the text must carry a
// stroke (apply-stroke + non-zero stroke width) for the effect to be visible.
// rangeStart / rangeEnd / rangeOffset are the Range Selector bounds in percent.
//
// Same mechanics as AddTextOpacityAnimator (shared splice + Range Selector path);
// Stroke Opacity is a 1D scalar with the same on-disk layout as Opacity.
//
// Refused: non-text layers, and text layers built by New* that were never parsed
// (call aep.Reopen first). Returns a stand-in group node referencing the spliced
// animator chunk.
//
// Alpha / structural — RE'd + double-version render-gated; typed parameter
// accessors are not yet wired (tune via Reopen). Free function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextNeighborAnimators_AEShipGate_AE2020,TestTextNeighborAnimators_AEShipGate_AE2025 incident=text-animator-create-re boundary="需文字带 stroke 才可见;typed accessor 未接" alias="text stroke opacity,描边不透明度动画"
func AddTextStrokeOpacityAnimator(layer *Layer, opacity, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextStrokeOpacityAnimator(layer, opacity, rangeStart, rangeEnd, rangeOffset)
}

// AddTextStrokeWidthAnimator adds a per-character Stroke Width animator with a
// Range Selector to a text layer — it grows / shrinks the glyph stroke. width
// (pixels) is applied to the selected characters' stroke; the text must carry a
// stroke (apply-stroke enabled) for the effect to be visible. rangeStart /
// rangeEnd / rangeOffset are the Range Selector bounds in percent.
//
// Same mechanics as AddTextOpacityAnimator (shared splice + Range Selector path);
// Stroke Width is a 1D scalar with the same on-disk layout as Opacity.
//
// Refused: non-text layers, and text layers built by New* that were never parsed
// (call aep.Reopen first). Returns a stand-in group node referencing the spliced
// animator chunk.
//
// Alpha / structural — RE'd + double-version render-gated; typed parameter
// accessors are not yet wired (tune via Reopen). Free function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextNeighborAnimators_AEShipGate_AE2020,TestTextNeighborAnimators_AEShipGate_AE2025 incident=text-animator-create-re boundary="需文字带 stroke 才可见;typed accessor 未接" alias="text stroke width,描边宽度动画"
func AddTextStrokeWidthAnimator(layer *Layer, width, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextStrokeWidthAnimator(layer, width, rangeStart, rangeEnd, rangeOffset)
}

// AddTextSkewAnimator adds a per-character Skew animator with a Range Selector to
// a text layer — the kinetic-typography primitive that shears characters into
// place. skew is the shear angle in degrees applied to the selected characters;
// rangeStart / rangeEnd / rangeOffset are the Range Selector bounds in percent.
// Sweep the Range Offset over time with AnimateTextRangeOffset for a shear-in.
//
// Same mechanics as AddTextOpacityAnimator (shared splice + Range Selector path);
// Skew is a 1D scalar with the same on-disk layout as Opacity.
//
// Refused: non-text layers, and text layers built by New* that were never parsed
// (call aep.Reopen first). Returns a stand-in group node referencing the spliced
// animator chunk.
//
// Alpha / structural — RE'd + double-version render-gated; typed parameter
// accessors are not yet wired (tune via Reopen). Free function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextNeighborAnimators_AEShipGate_AE2020,TestTextNeighborAnimators_AEShipGate_AE2025 incident=text-animator-create-re boundary="typed param accessor 未接,经 Reopen 调" alias="text skew,文字倾斜动画,shear"
func AddTextSkewAnimator(layer *Layer, skew, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextSkewAnimator(layer, skew, rangeStart, rangeEnd, rangeOffset)
}

// AddTextRotationXAnimator adds a per-character Rotation X animator with a Range
// Selector to a text layer — a 3D rotation about each character's horizontal
// axis (the characters tumble forward / back). rotation is the angle in degrees
// applied to the selected characters; rangeStart / rangeEnd / rangeOffset are the
// Range Selector bounds in percent.
//
// Same mechanics as AddTextOpacityAnimator (shared splice + Range Selector path);
// Rotation X is a 1D scalar (degrees). AE auto-adds an inert companion Z-rotation
// slot to the template, which stays at its default.
//
// Refused: non-text layers, and text layers built by New* that were never parsed
// (call aep.Reopen first). Returns a stand-in group node referencing the spliced
// animator chunk.
//
// Alpha / write-only — the value is written and survives AE (round-trip
// verified), but a per-character 3D rotation is VISUALLY INERT in a plain 2D text
// layer (it needs Per-character 3D enabled, a separate 3D capability not yet
// supported). Not render-gated; see incidents/text-animator-create-re.md. Free
// function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=roundtrip incident=text-animator-create-re boundary="2D 文字层视觉惰性(需 per-character 3D,未支持);值 write-only round-trip,未 render-gate" alias="text rotation x,3D 旋转 X,字符前后翻转"
func AddTextRotationXAnimator(layer *Layer, rotation, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextRotationXAnimator(layer, rotation, rangeStart, rangeEnd, rangeOffset)
}

// AddTextRotationYAnimator adds a per-character Rotation Y animator with a Range
// Selector to a text layer — a 3D rotation about each character's vertical axis
// (the characters swing left / right). rotation is the angle in degrees applied
// to the selected characters; rangeStart / rangeEnd / rangeOffset are the Range
// Selector bounds in percent.
//
// Same mechanics as AddTextOpacityAnimator (shared splice + Range Selector path);
// Rotation Y is a 1D scalar (degrees). AE auto-adds an inert companion Z-rotation
// slot to the template, which stays at its default.
//
// Alpha / write-only — the value is written and survives AE (round-trip
// verified), but a per-character 3D rotation is VISUALLY INERT in a plain 2D text
// layer (it needs Per-character 3D enabled, a separate 3D capability not yet
// supported). Not render-gated; see incidents/text-animator-create-re.md. Free
// function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=roundtrip incident=text-animator-create-re boundary="2D 文字层视觉惰性(需 per-character 3D,未支持);值 write-only round-trip,未 render-gate" alias="text rotation y,3D 旋转 Y,字符左右翻转"
func AddTextRotationYAnimator(layer *Layer, rotation, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextRotationYAnimator(layer, rotation, rangeStart, rangeEnd, rangeOffset)
}

// AddTextStrokeColorAnimator adds a per-character Stroke Color animator with a
// Range Selector to a text layer — it tints only the glyph stroke. r / g / b / a
// is the target colour (each channel 0..1) applied to the selected characters'
// stroke; the text must carry a stroke (apply-stroke + non-zero stroke width) for
// the effect to be visible. rangeStart / rangeEnd / rangeOffset are the Range
// Selector bounds in percent.
//
// Same mechanics as AddTextColorAnimator (shared splice + Range Selector path);
// Stroke Color is a 4-channel colour cdat stored as [A,R,G,B] × 255 f64 BE, the
// same on-disk encoding as shape Fill/Stroke and the text Fill Color leaf.
//
// Refused: non-text layers, and text layers built by New* that were never parsed
// (call aep.Reopen first). Returns a stand-in group node referencing the spliced
// animator chunk.
//
// Alpha / structural — RE'd + double-version render-gated; typed parameter
// accessors are not yet wired (tune via Reopen). Free function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextNeighborAnimators_AEShipGate_AE2020,TestTextNeighborAnimators_AEShipGate_AE2025 incident=text-animator-create-re boundary="需文字带 stroke 才可见;typed accessor 未接" alias="text stroke color,描边颜色动画"
func AddTextStrokeColorAnimator(layer *Layer, r, g, b, a, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextStrokeColorAnimator(layer, r, g, b, a, rangeStart, rangeEnd, rangeOffset)
}

// AddTextRangeSelector adds another Range Selector to the layer's FIRST text
// animator (a fresh animator carries one selector). Multiple selectors combine
// per each selector's Mode — the default is Add (union of the ranges); set a
// selector's Mode via SetTextRangeAdvanced for Subtract / Intersect / etc. start
// / end / offset are the new selector's bounds in percent. Sweep any selector's
// Offset over time with AnimateTextRangeOffset (which targets the first selector).
//
// Same indexed-group splice vein as AddTextOpacityAnimator (the "ADBE Text
// Selectors" group is INDEXED). Refused: non-text layers, text layers built by
// New* that were never parsed (call aep.Reopen first), and layers with no text
// animator. Returns a stand-in group node referencing the spliced selector.
//
// Alpha / structural — RE'd + double-version render-gated (two selectors union to
// extend the selection). Free function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextMultiSelector_AEShipGate_AE2020,TestTextMultiSelector_AEShipGate_AE2025 incident=text-animator-create-re boundary="作用于第一个 animator;Mode 经 SetTextRangeAdvanced;未 parse 的 fresh 层 refused" alias="text range selector,多选择器,range selector,叠加选择"
func AddTextRangeSelector(layer *Layer, start, end, offset float64) (*AEPropertyGroup, error) {
	return serializer.AddTextRangeSelector(layer, start, end, offset)
}

// AddTextWigglySelector adds a Wiggly Selector to the layer's FIRST text animator
// — a selector whose selection amount wobbles randomly (but deterministically per
// seed) over time, so the characters flicker / jitter in and out under the
// animator (the "wiggle" kinetic-typography primitive). The embedded selector
// uses AE's defaults (Temporal Freq 2/s, Max 100 / Min 0), so it animates on its
// own with no keyframes; combine it with a range selector via Mode, or use an
// empty range (End=0) so the wiggle drives selection alone.
//
// Same indexed-group splice vein as AddTextRangeSelector. Refused: non-text
// layers, text layers built by New* that were never parsed (call aep.Reopen
// first), and layers with no text animator. Returns a stand-in group node.
//
// Alpha / structural — RE'd + double-version render-gated (the rendered frames
// vary over time as the wiggle re-selects characters). Free function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextWigglySelector_AEShipGate_AE2020,TestTextWigglySelector_AEShipGate_AE2025 incident=text-animator-create-re boundary="作用于第一个 animator;用 AE 默认(Temporal Freq 2/s)自动摆动;未 parse 的 fresh 层 refused" alias="wiggly selector,摆动选择器,抖动,flicker,jitter"
func AddTextWigglySelector(layer *Layer) (*AEPropertyGroup, error) {
	return serializer.AddTextWigglySelector(layer)
}

// AddTextExpressibleSelector adds an Expressible Selector to the layer's FIRST
// text animator and drives its per-character selection with amountExpr, an
// ExtendScript expression returning the selection percentage (0..100). This is
// the expression-driven kinetic-typography primitive: the expression (which can
// read textIndex / textTotal / time / selectorValue) decides which glyphs the
// animator affects and by how much.
//
// Unlike the Range / Wiggly selectors the Expressible Amount is EXPRESSION-ONLY —
// it has no usable static value (a fresh Amount's .value throws in AE), so an
// empty expression yields an inert selector and amountExpr must be non-empty.
// Typical idioms: "textIndex <= 3 ? 100 : 0" (first 3 glyphs), "selectorValue"
// (all glyphs), or a time-driven sweep.
//
// Same indexed-group splice vein as AddTextWigglySelector, plus a SetExpression
// on the spliced Amount param (the expression Utf8 lands in canonical order —
// see incidents/expression-enable-byte-pair.md). Refused: non-text layers, text
// layers built by New* that were never parsed (call aep.Reopen first), layers
// with no text animator, and an empty amountExpr. Returns a stand-in group node.
//
// Alpha / structural — RE'd + double-version render-gated (the expression-driven
// selection visibly affects the glyphs AE renders). Free function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextExpressibleSelector_AEShipGate_AE2020,TestTextExpressibleSelector_AEShipGate_AE2025 incident=text-animator-create-re,expression-enable-byte-pair boundary="作用于第一个 animator;Amount 表达式驱动(必填,无静态值);未 parse 的 fresh 层 refused" alias="expressible selector,表达式选择器,expression selector,selectorValue,textIndex"
func AddTextExpressibleSelector(layer *Layer, amountExpr string) (*AEPropertyGroup, error) {
	return serializer.AddTextExpressibleSelector(layer, amountExpr)
}

// TextRangeAdvanced holds the "Advanced" sub-params of a text animator's Range
// Selector (Units / Based On / Mode / Amount / Shape / Smoothness / Ease High·Low
// / Randomize Order / Random Seed). See DefaultTextRangeAdvanced + the type doc
// for the field meanings and enum codings.
type TextRangeAdvanced = serializer.TextRangeAdvanced

// DefaultTextRangeAdvanced returns the Range Advanced params at their AE defaults
// (Units=Percentage, BasedOn=Characters, Mode=Add, Amount=100, Shape=Square,
// Smoothness=100, eases=0, no randomize). Tweak the fields you want, then pass
// the result to SetTextRangeAdvanced.
//
//aep:cap domain=meta tier=stable verify=none alias="range advanced defaults,默认高级范围"
func DefaultTextRangeAdvanced() TextRangeAdvanced { return serializer.DefaultTextRangeAdvanced() }

// SetTextRangeAdvanced sets the Range Advanced params on the layer's FIRST text
// animator's Range Selector — the selector-shaping controls behind a kinetic-
// typography reveal (how strongly the animator applies via Amount, the selection
// falloff Shape, the combination Mode for multi-selector setups, etc.). The
// Advanced group is elided on a fresh selector, so this materializes it from an
// embedded AE-native template, resets every slot to its AE default, then writes
// adv's values; it is idempotent (re-materializes on each call). Build adv with
// DefaultTextRangeAdvanced and tweak fields.
//
// Refused: non-text layers, text layers built by New* that were never parsed
// (call aep.Reopen first), and layers with no text animator (add one first).
//
// Alpha / structural — Amount is double-version render-gated; the other params
// round-trip (write + survive AE) but their visual effect is selector-internal /
// coupled (Mode needs multiple selectors, Smoothness only affects Shape=Square),
// so they are not individually render-gated. Free function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextRangeAdvancedAmount_AEShipGate_AE2020,TestTextRangeAdvancedAmount_AEShipGate_AE2025 incident=text-animator-create-re boundary="Amount render-gated;其余 param(Mode/Shape/Smoothness…)仅 round-trip(选择器内部/耦合)" alias="range advanced,高级范围,amount,shape,mode,ease high"
func SetTextRangeAdvanced(layer *Layer, adv TextRangeAdvanced) error {
	return serializer.SetTextRangeAdvanced(layer, adv)
}

// AnimateTextRangeOffset keyframes a text animator's Range Selector Offset,
// turning a static reveal into an animated sweep — the kinetic-typography
// payoff. Pair it with an Opacity-0 animator (Start=0/End=100): sweeping the
// Offset 0→100 over time reveals the characters one by one as the selection
// window (and the invisibility it carries) slides off the text. Operates on the
// layer's first animator; needs >= 2 keyframes. tickRate <= 0 uses the comp's.
//
// Builds a parsed property over the spliced Offset slot and delegates to the
// same 1D non-spatial static→animated conversion the effect-param / shape-scalar
// animate paths use (tdb4 flag flip + keyframe-stream synthesis).
//
// Alpha / structural — RE'd + double-version render-gated as the reveal sweep.
// Free function (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextAnimator_AEShipGate_AE2020,TestTextAnimator_AEShipGate_AE2025 incident=text-animator-create-re boundary="作用于第一个 animator;需 >=2 关键帧;tickRate<=0 用 comp 的" alias="animate range offset,范围偏移动画,逐字揭示,reveal sweep,打字机动画"
func AnimateTextRangeOffset(layer *Layer, tickRate float64, kfs []ScalarKeyframe) error {
	return serializer.AnimateTextRangeOffset(layer, tickRate, kfs)
}

// AnimateTextOpacity keyframes the per-character Opacity leaf of a text layer's
// first animator (added via AddTextOpacityAnimator) — animating the driven value
// itself rather than sweeping the Range Selector. Every selected character
// shares the opacity curve, so the text fades as one synchronized group (a pulse
// / blink), which the AnimateTextRangeOffset sweep cannot express. Builds a
// parsed property over the spliced Opacity slot and delegates to the same 1D
// non-spatial static→animated conversion the effect-param / shape-scalar animate
// paths use. Needs >= 2 keyframes; tickRate <= 0 uses the comp's.
//
// Refused: non-text layers, layers without a text animator carrying an Opacity
// leaf, and an Opacity leaf that is already animated.
//
// Alpha / structural — RE'd + double-version render-gated via the 1D scalar leaf
// path (shared with the Range Offset sweep). Free function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=roundtrip incident=text-animator-create-re boundary="1D scalar leaf 路径已 render-gated(经 AnimateTextRangeOffset/Rotation);本函数本身仅 round-trip" alias="animate text opacity,文字不透明度关键帧,同步闪烁,pulse"
func AnimateTextOpacity(layer *Layer, tickRate float64, kfs []ScalarKeyframe) error {
	return serializer.AnimateTextOpacity(layer, tickRate, kfs)
}

// AnimateTextRotation keyframes the per-character Rotation leaf of a text layer's
// first animator (added via AddTextRotationAnimator) — animating the driven angle
// itself rather than sweeping the Range Selector. Every selected character shares
// the rotation curve, so the text spins as one synchronized group (e.g. a
// continuous 0→360 spin), which the AnimateTextRangeOffset sweep cannot express.
// Builds a parsed property over the spliced Rotation slot and delegates to the
// same 1D non-spatial static→animated conversion the effect-param / shape-scalar
// animate paths use. Needs >= 2 keyframes; tickRate <= 0 uses the comp's.
//
// Refused: non-text layers, layers without a text animator carrying a Rotation
// leaf, and a Rotation leaf that is already animated.
//
// Alpha / structural — RE'd + double-version render-gated via the 1D scalar leaf
// path (shared with the Range Offset sweep). Free function (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextRotLeafAnimator_AEShipGate_AE2020,TestTextRotLeafAnimator_AEShipGate_AE2025 incident=text-animator-create-re boundary="需先 AddTextRotationAnimator;leaf 已动画则 refuse;>=2 关键帧" alias="animate text rotation,文字旋转关键帧,同步旋转,持续旋转"
func AnimateTextRotation(layer *Layer, tickRate float64, kfs []ScalarKeyframe) error {
	return serializer.AnimateTextRotation(layer, tickRate, kfs)
}

// AnimateTextPosition keyframes the per-character Position 3D leaf of a text
// layer's first animator (added via AddTextPositionAnimator) — animating the
// driven offset itself rather than sweeping the Range Selector. Each keyframe
// Value is the [x, y, z] offset in pixels; every selected character shares the
// motion curve, so the text glides as one synchronized group. Builds a parsed
// property over the spliced Position slot and delegates to AnimateVectorKeyframes
// (the spatial 3-component keyframe block — verified byte-matching the AE-saved
// text leaf). Needs >= 2 keyframes; tickRate <= 0 uses the comp's.
//
// Refused: non-text layers, layers without a text animator carrying a Position
// leaf, and a Position leaf that is already animated.
//
// Alpha / structural — RE'd + double-version render-gated. Free function
// (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextColorLeafAnimator_AEShipGate_AE2020,TestTextColorLeafAnimator_AEShipGate_AE2025 incident=text-animator-create-re boundary="需先 AddTextPositionAnimator;leaf 已动画则 refuse;>=2 关键帧" alias="animate text position,文字位移关键帧,同步滑动"
func AnimateTextPosition(layer *Layer, tickRate float64, kfs []VectorKeyframe) error {
	return serializer.AnimateTextPosition(layer, tickRate, kfs)
}

// AnimateTextScale keyframes the per-character Scale 3D leaf of a text layer's
// first animator (added via AddTextScaleAnimator) — animating the driven scale
// itself over time (e.g. a pulse / grow), which a Range-Offset sweep cannot
// express. Each keyframe Value is the [sx, sy, sz] scale percent (100 =
// unchanged); every selected character shares the curve. Scale 3D is a
// non-spatial 3-component leaf, so it uses a different animated keyframe block
// (value@0x08) than the spatial Position/Color leaves — RE'd byte-matching the
// AE-saved text Scale leaf. Needs >= 2 keyframes; tickRate <= 0 uses the comp's.
//
// Refused: non-text layers, layers without a text animator carrying a Scale leaf,
// and a Scale leaf that is already animated.
//
// Alpha / structural — RE'd + double-version render-gated. Free function
// (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextColorLeafAnimator_AEShipGate_AE2020,TestTextColorLeafAnimator_AEShipGate_AE2025 incident=text-animator-create-re boundary="需先 AddTextScaleAnimator;leaf 已动画则 refuse;>=2 关键帧" alias="animate text scale,文字缩放关键帧,同步缩放,pulse"
func AnimateTextScale(layer *Layer, tickRate float64, kfs []VectorKeyframe) error {
	return serializer.AnimateTextScale(layer, tickRate, kfs)
}

// AnimateTextColor keyframes the per-character Fill Color leaf of a text layer's
// first animator (added via AddTextColorAnimator) — animating the driven colour
// itself over time (e.g. a red→blue cycle), which a Range-Offset sweep cannot
// express. Each keyframe Value is an [r, g, b, a] colour with channels 0..1
// (converted internally to the on-disk [A,R,G,B]×255 encoding). Builds a parsed
// property over the spliced Fill Color slot and delegates to
// AnimateVectorKeyframes (the 4-channel keyframe block — verified byte-matching
// the AE-saved text leaf). Needs >= 2 keyframes; tickRate <= 0 uses the comp's.
//
// Refused: non-text layers, layers without a text animator carrying a Fill Color
// leaf, a Fill Color leaf that is already animated, and keyframe values that are
// not 4-channel.
//
// Alpha / structural — RE'd + double-version render-gated. Free function
// (CLAUDE.md #2).
//
//aep:cap domain=text tier=alpha verify=render-pixel gate=TestTextColorLeafAnimator_AEShipGate_AE2020,TestTextColorLeafAnimator_AEShipGate_AE2025 incident=text-animator-create-re boundary="需先 AddTextColorAnimator;leaf 已动画则 refuse;>=2 关键帧;值须 4 通道" alias="animate text color,文字颜色关键帧,颜色循环,color cycle"
func AnimateTextColor(layer *Layer, tickRate float64, kfs []VectorKeyframe) error {
	return serializer.AnimateTextColor(layer, tickRate, kfs)
}

// SetEffectParam sets an effect parameter's static value by full parameter
// match-name (e.g. "ADBE Gaussian Blur 2-0001"), returning the parameter's
// *Property. It is the typed-parameter entry for AddEffect-style workflows:
//
//	fx, _ := aep.AddEffect(layer, aep.EffectGaussianBlur)
//	_, err := aep.SetEffectParam(layer, fx, "ADBE Gaussian Blur 2-0001", 25.0)
//
// Why this exists: AE persists an effect parameter only while its value
// differs from the default — on a default instance the tunable params have no
// value stream at all (only "<effect>-0000" survives), so plain
// Property.SetStaticValue has nothing to target. When the parameter is
// already present, SetEffectParam is exactly SetStaticValue. When it is
// default-elided, the parameter's (tdmn, tdbs) value stream is first
// materialized from an embedded AE-native template (synthesis-lite) at its
// definition-order position, then the caller's value is written — matching
// what AE itself persists for a touched parameter. Any scalar / enum /
// boolean / angle / color / 2D-point / 3D-point / slider parameter of any
// effect materializes via the generic per-control-type template, patched
// (match-name, display name, scalar/slider min/max) from the host effect's
// own pard definition — parameter definitions are never elided, so the
// metadata is always in-file. Rarer control types (curve, layer, …) return
// an error when elided; params already present on the effect are settable
// regardless of control type.
//
// Values use the property's on-disk (StaticValue) encoding — the same units
// a parsed file exposes:
//   - scalar / slider / angle (degrees) / enum / boolean (0 or 1): float64,
//     1:1 with the AE UI value;
//   - color: []float64{A, R, G, B}, each channel 0–255;
//   - 2D / 3D point: []float64 fractions of the layer's coordinate space —
//     the SOURCE item's pixel size for footage/solid/precomp layers, the
//     COMPOSITION's for source-less layers (shape/text); the z component is
//     divided by the same space's HEIGHT (RE:
//     test_data/re_effect_param_types_units.aep).
//
// The materialized stream carries no tdpi host binding (only the
// always-present -0000 stream does), so no retarget is needed. Atomic
// mutation: snapshot value-group chunk children + flat Parameters + warnings;
// roll back on any parser warning or value-encode failure.
//
// Stable / structural — AE 2020 + AE 2025 ship-gate green (per-param and
// generic materialization, values read back on open and after AE's own
// resave); promoted from Alpha in the 2026-06-12 audit batch. Free function
// (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=effect tier=stable verify=ae-accept gate=TestSetEffectParam_AEShipGate_AE2020,TestSetEffectParam_AEShipGate_AE2025 incident=effect-param-elision-synthesis-lite boundary="scalar/enum/bool/angle/color/2D·3D point/slider 泛型物化;curve/layer-ref 类 elided 时 refuse" alias="effect param,效果参数,设参数,blurriness"
func SetEffectParam(layer *Layer, fx *Effect, paramMatchName string, value any) (*Property, error) {
	return serializer.SetEffectParam(layer, fx, paramMatchName, value)
}

// SupportedEffectParams returns the sorted parameter match-names with a
// dedicated per-param template. SetEffectParam is NOT limited to this list —
// scalar / enum / boolean / angle / color / 2D / 3D / slider params of any
// effect materialize via the generic per-control-type fallback, and
// already-present params are settable regardless.
//
//aep:cap domain=meta tier=stable verify=none alias="effect params,参数列表,supported params"
func SupportedEffectParams() []string { return serializer.SupportedEffectParams() }

// AnimateEffectParam keyframes a 1D-scalar effect parameter over time — N
// keyframes (>= 2), each a ScalarKeyframe{Time (seconds), Value, optional ease}.
// It materializes the parameter if it is default-elided (like SetEffectParam,
// from the host effect's pard definition), then converts its static value stream
// into an animated keyframe container from scratch — the case InsertKeyframe
// refuses (it requires a pre-existing keyframe to clone). Returns the animated
// *Property.
//
// On disk the parameter's static cdat is replaced by a LIST(list){lhd3, ldat}
// keyframe stream (non-spatial 1D layout, byte-matched to an AE-saved animated
// Gaussian-Blur-Blurriness fixture) and the tdb4 static→animated flags flip;
// WriteAEP recomputes the enclosing LIST sizes. Drives the classic MG rigs —
// an animated blur amount, or a Slider Control whose value an expression reads.
//
// fx must be on a parsed layer (round-trip through aep.Reopen after the
// structural New*/AddEffect APIs). Scalar (1D) params only — use
// AnimateEffectParamVec for color / 2D / 3D point params. Linear interp unless
// ScalarKeyframe.InEase/OutEase are set.
//
// Stable / structural — AE 2020 + AE 2025 ship-gate green. Free function
// (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=effect tier=stable verify=render-pixel gate=TestAnimEffect_AEShipGate_AE2020,TestAnimEffect_AEShipGate_AE2025 boundary="1D scalar only;color/point 用 AnimateEffectParamVec;fx 须在 parsed 层(Reopen)" alias="animate effect,效果关键帧,动画模糊,slider rig"
func AnimateEffectParam(layer *Layer, fx *Effect, paramMatchName string, kfs []ScalarKeyframe) (*Property, error) {
	return serializer.AnimateEffectParam(layer, fx, paramMatchName, kfs)
}

// AnimateEffectParamVec keyframes a multi-component effect parameter — the
// color / 2D-point / 3D-point counterpart of AnimateEffectParam. Each
// VectorKeyframe carries a Time (seconds), a []float64 Value whose length
// matches the parameter's component count, and optional ease. Values are in the
// parameter's on-disk units, identical to SetEffectParam: a color is
// [A,R,G,B] in 0-255; a 2D/3D point is a fraction of the layer's coordinate
// space (for a source-backed layer divide by the source's w/h, for a
// source-less layer by the comp's — and z by the same space's height).
//
// Like the scalar form it materializes the parameter if default-elided, then
// replaces its static cdat with a keyframe stream — but using the SPATIAL block
// layout AE writes for animated effect color/point params (value at 0x38, a
// per-type @0x08 marker: 2 for color, 3 for point), RE'd byte-for-byte from an
// AE-native fixture. The tdb4 static→animated flip is the same as the scalar /
// shape paths. Returns the animated *Property.
//
// fx must be on a parsed layer (round-trip through aep.Reopen). Components 2/3/4
// only (use AnimateEffectParam for 1D scalars). Linear interp unless ease is set.
//
// Stable / structural — AE 2020 + AE 2025 ship-gate green. Free function
// (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=effect tier=stable verify=render-pixel gate=TestAnimEffectVec_AEShipGate_AE2020,TestAnimEffectVec_AEShipGate_AE2025 boundary="2/3/4 分量(color/point);1D 用 AnimateEffectParam;fx 须在 parsed 层(Reopen)" alias="animate effect color,效果颜色关键帧,point 动画"
func AnimateEffectParamVec(layer *Layer, fx *Effect, paramMatchName string, kfs []VectorKeyframe) (*Property, error) {
	return serializer.AnimateEffectParamVec(layer, fx, paramMatchName, kfs)
}

// SetEffectLayerParam points a layer-reference effect parameter at target —
// e.g. Set Matte's "Take Matte From Layer" (paramMatchName
// "ADBE Set Matte3-0001"), which mattes the host layer with another layer's
// channel. AE stores the reference as target's layer ID in the parameter's tdpi
// chunk (the same binding the effect's always-present host stream uses, aimed
// elsewhere), so this is a length-preserving 4-byte rewrite. target must be a
// layer in the same composition.
//
// fx must be on a parsed layer (round-trip through aep.Reopen). The parameter
// must already be present in the effect (Set Matte's -0001 ships materialized in
// the AddEffect template); materializing a default-elided layer-ref param is a
// follow-up.
//
// Stable / structural — AE 2020 + AE 2025 ship-gate green. Free function
// (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=effect tier=stable verify=render-pixel gate=TestSetMatte_AEShipGate_AE2020,TestSetMatte_AEShipGate_AE2025,TestLayerRefDispMap_AEShipGate_AE2020,TestLayerRefDispMap_AEShipGate_AE2025,TestLayerRefCompoundBlur_AEShipGate_AE2020,TestLayerRefCompoundBlur_AEShipGate_AE2025,TestAddEffectWave11_AEShipGate_AE2020,TestAddEffectWave11_AEShipGate_AE2025,TestAddEffectWave12_AEShipGate_AE2020,TestAddEffectWave12_AEShipGate_AE2025 boundary="参数须已物化(随模板带出):Set Matte -0001 · Displacement Map -0001 · Compound Blur -0001 · CC Vector Blur -0005 均 render-pixel/accept 双版本 gated;CC Vector Blur=accept-only(render-pixel deferred);wave11 layer-ref(Warp Stabilizer -0046 · 3D Glasses -0001/-0002 · Timewarp -0029/-0031 · CC Particle World -0045) + wave12 layer-ref(Texturize -0001 · Color Link -0001 · Compound Arithmetic -0001 · Set Channels -0001/-0003/-0005/-0007 四源)=accept+readback+resave 双版本 gated(render-pixel deferred:无单帧像素证明面);其它 default-elided layer-ref 物化未做" alias="set matte,layer reference,蒙版层,displacement map,compound blur,vector blur,take matte from layer"
func SetEffectLayerParam(layer *Layer, fx *Effect, paramMatchName string, target *Layer) error {
	return serializer.SetEffectLayerParam(layer, fx, paramMatchName, target)
}

// SetMaterialOption sets a 3D layer's Material-Options property by AE match-name
// (e.g. "ADBE Casts Shadows", "ADBE Accepts Lights", "ADBE Diffuse Coefficient"),
// returning the property's *Property. It is the from-scratch entry for material
// properties — the synthesis-lite sibling of SetEffectParam.
//
// Why this exists: a from-scratch shape/solid layer made 3D (SetIs3D) emits an
// EMPTY Material Options group — AE materializes the full material tree (all
// defaults) in its DOM on open, but on disk the leaves are elided, so
// MaterialCastsShadows() / SetMaterialCastsShadows return "property not present".
// When the property is already present (a parsed/fixture layer, or a prior
// splice) SetMaterialOption is exactly SetStaticValue. When it is default-elided,
// the (tdmn, tdbs) leaf is first materialized from an embedded AE-native template
// (the 15-leaf material tree from re_material_options.aep) at its canonical
// position, then the caller's value is written — which is non-default by intent,
// exactly the state AE itself persists (a default-valued spliced leaf is dropped
// by AE on open).
//
// The headline use is making a from-scratch 3D layer CAST shadows (Casts Shadows
// defaults Off): SetMaterialOption(caster, "ADBE Casts Shadows", float64(aep.MaterialCastsOn)).
// The catcher needs nothing — an empty material group accepts shadows + lights by
// default. The layer must be round-tripped through aep.Reopen first (the material
// group chunk must exist to splice into).
//
// Values use the property's on-disk (StaticValue) encoding: scalar / enum /
// boolean (0 or 1) as float64; Shadow Color as []float64{A,R,G,B} 0–255. Atomic:
// snapshots the group chunk children + flat Properties + tree + warnings; rolls
// back on any parser warning or value-encode failure.
//
// Alpha / structural — AE 2020 + AE 2025 render ship-gate green for Casts Shadows
// (from-scratch 3D caster drops a visible shadow). Free function (CLAUDE.md #2
// structural-op call-form).
//
//aep:cap domain=layer-set tier=alpha verify=render-pixel gate=TestLayer3DShadow_AEShipGate_AE2020,TestLayer3DShadow_AEShipGate_AE2025 boundary="Casts Shadows render-gated;需 Reopen(material group 须存在);其他 material 属性 synthesis-lite 未逐个 gate" alias="material option,材质选项,casts shadows,投影,3D 材质"
func SetMaterialOption(layer *Layer, matchName string, value any) (*Property, error) {
	return serializer.SetMaterialOption(layer, matchName, value)
}

// AddEssentialProperty exposes one parameter of an effect on layer in the
// owning composition's Essential Graphics panel — mirrors AE's
// "addProperty to Essential Graphics" / Property.addToMotionGraphicsTemplate.
// Returns the new controller (Name / Type / UUID), also appended to
// Composition.EssentialGraphicsControllers.
//
// displayName is the controller's panel label; empty → the parameter's own
// display name. Supported parameter control types in this first slice:
// scalar / slider (EG slider controller, min/max from the pard definition),
// boolean (checkbox), and color (color controller — requires the parameter
// to have a materialized non-default value, since AE stores no color default
// in pard; set one first via SetEffectParam). Other types (point, dropdown,
// text) return an error for now.
//
// Mechanics (RE 2026-06-12, three coordinated chunk sites):
//   - The comp Item's three EG panel generations (CIFO/CIF2/CIF3 — AE keeps
//     them byte-identical) each gain a LIST:CCtl entry (localized name,
//     fresh v4 UUID, CTyp, type-keyed CVal/CDef[/Smin/Smax], and a CPrp
//     property ref: comp item ID + host layer ID + a JSON matchName path
//     whose element indexes are 0-based positions within the parent group,
//     -1 for fixed groups) and a CcCt count bump.
//   - The host layer's "ADBE Layer Overrides" parade triple (tdmn + OvG2 +
//     tdgp; auto-created in AE-native empty form when the layer lacks it)
//     gains an OvG2 CPrp uuid slot and an override value stream — a clone of
//     the parameter's materialized tdbs, or a template materialization
//     carrying the current value when the parameter is default-elided.
//
// Refused: layers built by the structural New* APIs that were never parsed
// (no property tree — call aep.Reopen first), comps without the EG panel
// shell, and parameters absent from the effect's pard definitions.
//
// Atomic mutation: every mutated site (3×CIF*, OvG2, override tdgp, scene
// controller list) is snapshotted; the panel is re-decoded after commit and
// any mismatch or parser warning rolls everything back.
//
// Stable / structural — AE 2020 + AE 2025 ship-gate green on an all-Go-built
// project (file accepted, panel read back via the scripting API, controller
// identity preserved across AE's own resave; the gate also covers
// SetMotionGraphicsTemplateName); promoted from Alpha in the 2026-06-12 audit
// batch. Free function (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=eg tier=stable verify=ae-accept gate=TestEGAdd_AEShipGate_AE2020,TestEGAdd_AEShipGate_AE2025 boundary="scalar/slider/checkbox/color 控件;point/dropdown/text/Transform deferred;未 Reopen 的 fresh 层 refused" alias="essential graphics,主图形,EG,模板控件,addToMotionGraphicsTemplate"
func AddEssentialProperty(layer *Layer, fx *Effect, paramMatchName, displayName string) (*EssentialGraphicsController, error) {
	return serializer.AddEssentialProperty(layer, fx, paramMatchName, displayName)
}

// RemoveEffect removes the effect at the given 0-based index from the layer's
// Effect Parade — the inverse of AddEffect. It is a thin, index-validated
// wrapper over RemovePropertyGroup (AE 2020 + AE 2025 ship-gate green for
// Effect-Parade child removal). Returns an error if the layer has no Effect
// Parade or index is out of range.
//
// Stable / structural — rides the AE 2020 + AE 2025 ship-gated Effect-Parade
// child removal. Free function (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=effect tier=stable verify=ae-accept gate=TestPropStructRemove_AEShipGate_AE2020,TestPropStructRemove_AEShipGate_AE2025 incident=property-indexed-group-structural-re boundary="rides RemovePropertyGroup 的 Effect-Parade gate;无独立 RemoveEffect AE gate" alias="remove effect,删效果"
func RemoveEffect(layer *Layer, index int) error { return serializer.RemoveEffect(layer, index) }

// AddMask appends a vector mask to the layer's "ADBE Mask Parade" and returns
// the parsed *Mask. The mask is created with the given display name (empty →
// "Mask N"), the given static Bezier path, and AE defaults everywhere else:
// mode Add, not inverted, zero feather, full opacity (Feather / Opacity /
// Expansion are default-elided on disk, exactly as AE persists an untouched
// mask).
//
// The path is parameterizable at creation time even though mutating an
// EXISTING mask's path is refused (structural write): the atom is built from
// scratch, reusing the ship-gated shape-path encoding for the "ADBE Mask
// Shape" value (mask paths share the byte layout of "ADBE Vector Shape").
// path.Vertices are in layer pixel coordinates; per-vertex tangents are
// relative to the anchor (AE Shape semantics). path.Closed selects a closed
// region vs an open polyline. (On disk AE stores mask coordinates as
// fractions of the SOURCE item's pixel space for footage/solid/precomp
// layers and as raw pixels for source-less layers (shape/text) — AddMask
// performs that conversion, so callers always pass pixels.)
//
// Mechanics: AE stores each mask as a (tdmn "ADBE Mask Atom", mkif, LIST:tdgp)
// chunk triple inside the parade — one chunk more than an effect's pair; the
// 48-byte mkif carries mode / inverted / locked / motion-blur / an internal
// per-layer index (monotonic, AE keeps gaps) / the label color. AddMask
// splices a fresh triple in just before the "ADBE Group End" sentinel. LIST
// sizes grow automatically (rifx recomputes bottom-up on write).
//
// Parade auto-create: a parsed layer with no masks has no Mask Parade group at
// all. AddMask splices an empty parade into the layer's property tree
// immediately before "ADBE Effect Parade" when present, else before "ADBE
// Transform Group" (AE's emitted group order — the Mask Parade precedes both).
//
// Refused layers: camera / light layers (AE does not allow masks on them), and
// layers built by the structural New* APIs that were never parsed — those have
// no property tree to splice into; call aep.Reopen first and add masks to the
// re-parsed layer.
//
// Atomic mutation: snapshot parade chunk + scene children + flat Masks slice
// (+ the pre-auto-create tree state); re-parse the spliced triple to obtain a
// back-ref-correct *Mask (its Set* setters work immediately); roll back on any
// parser warning.
//
// Stable / structural — AE 2020 + AE 2025 ship-gate green (4/4: AE-native
// fixture splice next to an existing Effect Parade, plus a 100% Go-built
// project; open + closed paths; AE reads names / modes / vertices back
// exactly and keeps the masks across its own resave); promoted from Alpha in
// the 2026-06-12 audit batch. Free function (CLAUDE.md #2 structural-op
// call-form).
//
//aep:cap domain=mask tier=stable verify=ae-accept gate=TestAddMask_AEShipGate_AE2020,TestAddMask_AEShipGate_AE2025 boundary="camera/light 层 + 未 Reopen 的 fresh 层 refused;Feather/Opacity/Expansion default-elided" alias="mask,蒙版,遮罩,vector mask,加蒙版"
func AddMask(layer *Layer, name string, path BezierPath) (*Mask, error) {
	return serializer.AddMask(layer, name, path)
}

// SetMaskPath rewrites an existing mask's outline in place with a new static
// path (layer-pixel coordinates, the same space AddMask accepts). Unlike the
// length-preserving Mask.Set* setters, the path is a variable-length subtree, so
// this rebuilds the "ADBE Mask Shape" om-s and swaps it in; WriteAEP recomputes
// the enclosing LIST sizes. The vertex count may differ from the original (e.g.
// reshape a 4-point rectangle into a 3-point triangle) — the mask-strictness
// lhd3/shph patching AddMask uses is reused so AE accepts non-4-vertex masks.
//
// mask must be one of layer.Masks obtained from a parsed project (it needs its
// atom-group chunk back-ref); call aep.Reopen first for masks built by the
// structural New*/AddMask APIs without an intervening parse.
//
// Stable / structural — AE 2020 + AE 2025 ship-gate green. Free function
// (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=mask tier=stable verify=ae-accept gate=TestMGMaskPath_AEShipGate_AE2020,TestMGMaskPath_AEShipGate_AE2025 boundary="mask 须来自 parsed 工程(Reopen);顶点数可与原不同" alias="mask path,蒙版路径,改蒙版形状,reshape mask"
func SetMaskPath(layer *Layer, mask *Mask, path BezierPath) error {
	return serializer.SetMaskPath(layer, mask, path)
}

// SetMaskPathKeyframes replaces an existing mask's outline with an ANIMATED
// path — N keyframes (>= 2), each a BezierPath snapshot at a time in seconds
// (the layer-pixel space AddMask / SetMaskPath accept), with optional temporal
// ease per side (zero = linear). Vertex counts may differ between keyframes
// (AE interpolates the outline; the mask-strictness lhd3/shph patching makes
// non-4-vertex frames safe).
//
// On disk this is byte-isomorphic to AE's own animated mask/shape path: the
// "ADBE Mask Shape" om-s carries a TIME-table tdbs (one 64-byte block per
// keyframe) plus one geometry shap per keyframe. WriteAEP recomputes the
// enclosing LIST sizes. mask must come from a parsed project (it needs its
// atom-group chunk back-ref); call aep.Reopen first for masks built by the
// structural New*/AddMask APIs without an intervening parse.
//
// Stable / structural — AE 2020 + AE 2025 ship-gate green. Free function
// (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=mask tier=stable verify=render-pixel gate=TestMGMaskPathKf_AEShipGate_AE2020,TestMGMaskPathKf_AEShipGate_AE2025 boundary=">=2 关键帧,gate 实测到 6kf(=2 lhd3 容量页,验 mask 严格性下分页正确,非仅 2kf 单页);逐帧顶点数可不同;mask 须来自 parsed 工程(Reopen)" alias="mask path keyframes,蒙版路径动画,animated mask,变形蒙版"
func SetMaskPathKeyframes(layer *Layer, mask *Mask, keys []MaskPathKey) error {
	return serializer.SetMaskPathKeyframes(layer, mask, keys)
}

// RemoveMask deletes mask m from layer's "ADBE Mask Parade" — the inverse of
// AddMask. m must be one of layer.Masks obtained from a parsed project; pass
// the same layer the mask belongs to (masks carry no owning-layer back-ref).
//
// Mechanics: each mask is a (tdmn "ADBE Mask Atom", mkif, LIST:tdgp) chunk
// triple — one chunk more than an effect's pair, which is why the generic
// indexed-group RemovePropertyGroup refuses a mask atom (its tdgp is preceded
// by the mkif, not the tdmn). RemoveMask is triple-aware: it anchors on the
// mask's own mkif, validates the framing "ADBE Mask Atom" tdmn and trailing
// atom tdgp, and splices all three out, then drops the mask from the scene
// property tree and the flat layer.Masks slice. LIST sizes shrink
// automatically (rifx recomputes bottom-up on write). The removed chunks ride
// out verbatim, so no opaque content is regenerated (CLAUDE.md #5).
//
// Refused (project untouched): a nil layer/mask, a mask not in layer.Masks
// (e.g. already removed), a mask built outside the parser (no mkif back-ref),
// or a layer with no Mask Parade. Removing the last mask leaves an empty
// parade group in place (AE tolerates it on reopen); collapsing the parade is
// a separate slice.
//
// Stable / structural — AE 2020 + AE 2025 ship-gate green (build three masks,
// remove the middle one, AE accepts the spliced-out triple next to a real
// Effect Parade and reads back both survivors with geometry intact and the
// effects untouched). Free function (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=mask tier=stable verify=ae-accept gate=TestRemoveMask_AEShipGate_AE2020,TestRemoveMask_AEShipGate_AE2025 boundary="删最后一个 mask 留空 parade(AE 容忍);mask 须来自 parsed 工程" alias="remove mask,删蒙版"
func RemoveMask(layer *Layer, m *Mask) error {
	return serializer.RemoveMask(layer, m)
}

// DuplicateMask inserts a copy of mask m immediately after it in layer's "ADBE
// Mask Parade" — mirroring AE's PropertyBase.duplicate() on a mask — and
// returns the clone. m must be one of layer.Masks from a parsed project; pass
// the layer it belongs to (masks carry no owning-layer back-ref).
//
// Mechanics: triple-aware, like RemoveMask. A mask is a (tdmn "ADBE Mask
// Atom", mkif, LIST:tdgp) triple, so the generic DuplicatePropertyGroup
// refuses it (its tdgp is preceded by the mkif, not the tdmn). DuplicateMask
// deep-clones all three chunks (opaque content rides along verbatim —
// CLAUDE.md #5), bumps only the clone's internal mask index (mkif @0x08) to
// max+1 so it stays unique, splices the clone in just after the source, and
// re-parses it into a *Mask whose Set* setters work immediately. The clone
// keeps the source's name, mode, color, inverted/locked flags and path.
//
// Refused (project untouched): a nil layer/mask, a mask not in layer.Masks, a
// mask built outside the parser (no mkif back-ref), or a layer with no Mask
// Parade.
//
// Stable / structural — AE 2020 + AE 2025 ship-gate green (add a mask,
// duplicate it, AE accepts the cloned triple with a distinct internal index
// and reads back both masks with geometry intact and the effects untouched).
// Free function (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=mask tier=stable verify=ae-accept gate=TestDuplicateMask_AEShipGate_AE2020,TestDuplicateMask_AEShipGate_AE2025 boundary="mask 须来自 parsed 工程(Reopen)" alias="duplicate mask,复制蒙版"
func DuplicateMask(layer *Layer, m *Mask) (*Mask, error) {
	return serializer.DuplicateMask(layer, m)
}

// MoveMask reorders mask m to position toIndex (0-based) among layer's masks,
// the other masks keeping their relative order — mirroring AE's
// PropertyBase.moveTo() on a mask. m must be one of layer.Masks from a parsed
// project; pass the layer it belongs to (masks carry no owning-layer back-ref).
// toIndex == m's current index is a no-op.
//
// Mechanics: triple-aware, like RemoveMask / DuplicateMask. Each mask is a
// (tdmn "ADBE Mask Atom", mkif, LIST:tdgp) triple; MoveMask locates every
// mask's triple by its mkif, re-emits the contiguous triple run in the target
// order (the same chunk pointers — opaque content rides along unchanged,
// CLAUDE.md #5), and applies the same permutation to the scene property tree
// and the flat layer.Masks slice. No chunk is created or destroyed, so no LIST
// size changes.
//
// Refused (project untouched): a nil layer/mask, a mask not in layer.Masks,
// toIndex out of range, a mask built outside the parser (no mkif back-ref), a
// layer with no Mask Parade, or a parade whose mask triples are not contiguous.
//
// Stable / structural — AE 2020 + AE 2025 ship-gate green (build three masks,
// move the last to the front, AE accepts the re-emitted triple run and reads
// the masks back in the new order with the effects untouched). Free function
// (CLAUDE.md #2 structural-op call-form).
//
//aep:cap domain=mask tier=stable verify=ae-accept gate=TestMoveMask_AEShipGate_AE2020,TestMoveMask_AEShipGate_AE2025 boundary="mask 须来自 parsed 工程(Reopen)" alias="move mask,蒙版排序,reorder mask"
func MoveMask(layer *Layer, m *Mask, toIndex int) error {
	return serializer.MoveMask(layer, m, toIndex)
}

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
	EffectDropShadow         = serializer.EffectDropShadow         // Drop Shadow
	EffectSharpen            = serializer.EffectSharpen            // Sharpen
	EffectMosaic             = serializer.EffectMosaic             // Mosaic
	EffectNoise              = serializer.EffectNoise              // Noise
	EffectTransform          = serializer.EffectTransform          // Transform
	EffectGradientRamp       = serializer.EffectGradientRamp       // Gradient Ramp
	EffectFractalNoise       = serializer.EffectFractalNoise       // Fractal Noise
	EffectMotionTile         = serializer.EffectMotionTile         // Motion Tile
	EffectDirectionalBlur    = serializer.EffectDirectionalBlur    // Directional Blur
	EffectLinearWipe         = serializer.EffectLinearWipe         // Linear Wipe
	EffectWaveWarp           = serializer.EffectWaveWarp           // Wave Warp
	EffectCurves             = serializer.EffectCurves             // Curves
	EffectSliderControl      = serializer.EffectSliderControl      // Slider Control
	EffectPointControl       = serializer.EffectPointControl       // Point Control
	EffectColorControl       = serializer.EffectColorControl       // Color Control
	EffectAngleControl       = serializer.EffectAngleControl       // Angle Control
	EffectCheckboxControl    = serializer.EffectCheckboxControl    // Checkbox Control
	EffectPoint3DControl     = serializer.EffectPoint3DControl     // 3D Point Control
	EffectSetMatte           = serializer.EffectSetMatte           // Set Matte (layer-reference effect)
	EffectTurbulentDisplace  = serializer.EffectTurbulentDisplace  // Turbulent Displace
	EffectRoughenEdges       = serializer.EffectRoughenEdges       // Roughen Edges
	EffectEcho               = serializer.EffectEcho               // Echo
	EffectRadialBlur         = serializer.EffectRadialBlur         // Radial Blur
	EffectFourColorGradient  = serializer.EffectFourColorGradient  // 4-Color Gradient
	EffectCheckerboard       = serializer.EffectCheckerboard       // Checkerboard
	EffectGrid               = serializer.EffectGrid               // Grid
	EffectStroke             = serializer.EffectStroke             // Stroke
	EffectCornerPin          = serializer.EffectCornerPin          // Corner Pin
	EffectVenetianBlinds     = serializer.EffectVenetianBlinds     // Venetian Blinds
	EffectTwirl              = serializer.EffectTwirl              // Twirl
	EffectPolarCoordinates   = serializer.EffectPolarCoordinates   // Polar Coordinates
	EffectSpherize           = serializer.EffectSpherize           // Spherize
	EffectMagnify            = serializer.EffectMagnify            // Magnify
	EffectRipple             = serializer.EffectRipple             // Ripple
	EffectOpticsCompensation = serializer.EffectOpticsCompensation // Optics Compensation
	EffectPosterize          = serializer.EffectPosterize          // Posterize
	EffectThreshold          = serializer.EffectThreshold          // Threshold
	EffectFindEdges          = serializer.EffectFindEdges          // Find Edges
	EffectColorEmboss        = serializer.EffectColorEmboss        // Color Emboss
	EffectEmboss             = serializer.EffectEmboss             // Emboss
	EffectStrobeLight        = serializer.EffectStrobeLight        // Strobe Light
	EffectBrushStrokes       = serializer.EffectBrushStrokes       // Brush Strokes
	EffectBevelAlpha         = serializer.EffectBevelAlpha         // Bevel Alpha
	EffectBevelEdges         = serializer.EffectBevelEdges         // Bevel Edges
	EffectPhotoFilter        = serializer.EffectPhotoFilter        // Photo Filter
	EffectVibrance           = serializer.EffectVibrance           // Vibrance
	EffectColorBalance       = serializer.EffectColorBalance       // Color Balance
	EffectColorBalanceHLS    = serializer.EffectColorBalanceHLS    // Color Balance (HLS)
	EffectBlackAndWhite      = serializer.EffectBlackAndWhite      // Black & White
	EffectGammaPedestalGain  = serializer.EffectGammaPedestalGain  // Gamma/Pedestal/Gain
	EffectChannelBlur        = serializer.EffectChannelBlur        // Channel Blur
	EffectBilateralBlur      = serializer.EffectBilateralBlur      // Bilateral Blur
	EffectSmartBlur          = serializer.EffectSmartBlur          // Smart Blur
	EffectUnsharpMask        = serializer.EffectUnsharpMask        // Unsharp Mask
	EffectShiftChannels      = serializer.EffectShiftChannels      // Shift Channels
	EffectSolidComposite     = serializer.EffectSolidComposite     // Solid Composite
	EffectMinimax            = serializer.EffectMinimax            // Minimax
	EffectArithmetic         = serializer.EffectArithmetic         // Arithmetic
	EffectCircle             = serializer.EffectCircle             // Circle
	EffectLensFlare          = serializer.EffectLensFlare          // Lens Flare
	EffectCellPattern        = serializer.EffectCellPattern        // Cell Pattern
	EffectAdvancedLightning  = serializer.EffectAdvancedLightning  // Advanced Lightning
	EffectBeam               = serializer.EffectBeam               // Beam
	EffectPaintBucket        = serializer.EffectPaintBucket        // Paint Bucket
	EffectPosterizeTime      = serializer.EffectPosterizeTime      // Posterize Time
	EffectSimpleChoker       = serializer.EffectSimpleChoker       // Simple Choker
	EffectMatteChoker        = serializer.EffectMatteChoker        // Matte Choker
	EffectBulge            = serializer.EffectBulge            // Bulge
	EffectOffset           = serializer.EffectOffset           // Offset
	EffectMirror           = serializer.EffectMirror           // Mirror
	EffectFractal          = serializer.EffectFractal          // Fractal
	EffectWriteOn          = serializer.EffectWriteOn          // Write-on
	EffectScribble         = serializer.EffectScribble         // Scribble
	EffectEyedropperFill   = serializer.EffectEyedropperFill   // Eyedropper Fill
	EffectAudioSpectrum    = serializer.EffectAudioSpectrum    // Audio Spectrum
	EffectAudioWaveform    = serializer.EffectAudioWaveform    // Audio Waveform
	EffectAutoLevels       = serializer.EffectAutoLevels       // Auto Levels
	EffectAutoColor        = serializer.EffectAutoColor        // Auto Color
	EffectAutoContrast     = serializer.EffectAutoContrast     // Auto Contrast
	EffectEqualize         = serializer.EffectEqualize         // Equalize
	EffectLeaveColor       = serializer.EffectLeaveColor       // Leave Color
	EffectChangeToColor    = serializer.EffectChangeToColor    // Change to Color
	EffectChangeColor      = serializer.EffectChangeColor      // Change Color
	EffectRadialShadow     = serializer.EffectRadialShadow     // Radial Shadow
	EffectRemoveColorMatte = serializer.EffectRemoveColorMatte // Remove Color Matting
	EffectDustAndScratches = serializer.EffectDustAndScratches // Dust & Scratches
	EffectNoiseAlpha       = serializer.EffectNoiseAlpha       // Noise Alpha
	EffectNoiseHLS         = serializer.EffectNoiseHLS         // Noise HLS
	EffectRadialWipe       = serializer.EffectRadialWipe       // Radial Wipe
	EffectBlockDissolve    = serializer.EffectBlockDissolve    // Block Dissolve
	EffectLumetri             = serializer.EffectLumetri             // Lumetri Color
	EffectLightning           = serializer.EffectLightning           // Lightning
	EffectCCRadialFastBlur    = serializer.EffectCCRadialFastBlur    // CC Radial Fast Blur
	EffectCCRadialBlur        = serializer.EffectCCRadialBlur        // CC Radial Blur
	EffectCCCrossBlur         = serializer.EffectCCCrossBlur         // CC Cross Blur
	EffectCCBendIt            = serializer.EffectCCBendIt            // CC Bend It
	EffectCCBender            = serializer.EffectCCBender            // CC Bender
	EffectCCBlobbylize        = serializer.EffectCCBlobbylize        // CC Blobbylize
	EffectCCFloMotion         = serializer.EffectCCFloMotion         // CC Flo Motion
	EffectCCGriddler          = serializer.EffectCCGriddler          // CC Griddler
	EffectCCLens              = serializer.EffectCCLens              // CC Lens
	EffectCCPageTurn          = serializer.EffectCCPageTurn          // CC Page Turn
	EffectCCPowerPin          = serializer.EffectCCPowerPin          // CC Power Pin
	EffectCCRipplePulse       = serializer.EffectCCRipplePulse       // CC Ripple Pulse
	EffectCCSlant             = serializer.EffectCCSlant             // CC Slant
	EffectCCSmear             = serializer.EffectCCSmear             // CC Smear
	EffectCCSplit             = serializer.EffectCCSplit             // CC Split
	EffectCCSplit2            = serializer.EffectCCSplit2            // CC Split 2
	EffectCCTiler             = serializer.EffectCCTiler             // CC Tiler
	EffectCCWarpoMatic        = serializer.EffectCCWarpoMatic        // CC WarpoMatic
	EffectCCLightBurst        = serializer.EffectCCLightBurst        // CC Light Burst 2.5
	EffectCCLightRays         = serializer.EffectCCLightRays         // CC Light Rays
	EffectCCLightSweep        = serializer.EffectCCLightSweep        // CC Light Sweep
	EffectCCThreads           = serializer.EffectCCThreads           // CC Threads
	EffectCCCylinder          = serializer.EffectCCCylinder          // CC Cylinder
	EffectCCSphere            = serializer.EffectCCSphere            // CC Sphere
	EffectCCSpotlight         = serializer.EffectCCSpotlight         // CC Spotlight
	EffectCCGlass             = serializer.EffectCCGlass             // CC Glass
	EffectCCHexTile           = serializer.EffectCCHexTile           // CC HexTile
	EffectCCKaleida           = serializer.EffectCCKaleida           // CC Kaleida
	EffectCCMrSmoothie        = serializer.EffectCCMrSmoothie        // CC Mr. Smoothie
	EffectCCPlastic           = serializer.EffectCCPlastic           // CC Plastic
	EffectCCRepeTile          = serializer.EffectCCRepeTile          // CC RepeTile
	EffectCCThreshold         = serializer.EffectCCThreshold         // CC Threshold
	EffectCCThresholdRGB      = serializer.EffectCCThresholdRGB      // CC Threshold RGB
	EffectCCPixelPolly        = serializer.EffectCCPixelPolly        // CC Pixel Polly
	EffectCCScatterize        = serializer.EffectCCScatterize        // CC Scatterize
	EffectCCStarBurst         = serializer.EffectCCStarBurst         // CC Star Burst
	EffectCCForceMotionBlur   = serializer.EffectCCForceMotionBlur   // CC Force Motion Blur
	EffectCCWideTime          = serializer.EffectCCWideTime          // CC Wide Time
	EffectCCColorOffset       = serializer.EffectCCColorOffset       // CC Color Offset
	EffectCCToner             = serializer.EffectCCToner             // CC Toner
	EffectCCBurnFilm          = serializer.EffectCCBurnFilm          // CC Burn Film
	EffectCCVignette          = serializer.EffectCCVignette          // CC Vignette
	EffectCCSimpleWireRemoval = serializer.EffectCCSimpleWireRemoval // CC Simple Wire Removal
	EffectDisplacementMap = serializer.EffectDisplacementMap // Displacement Map (layer-reference)
	EffectCompoundBlur    = serializer.EffectCompoundBlur    // Compound Blur (layer-reference)
	EffectCCVectorBlur    = serializer.EffectCCVectorBlur    // CC Vector Blur (layer-reference)

	EffectBasic3D               = serializer.EffectBasic3D
	EffectBroadcastColors       = serializer.EffectBroadcastColors
	EffectChannelCombiner       = serializer.EffectChannelCombiner
	EffectCineonConverter       = serializer.EffectCineonConverter
	EffectColorKey              = serializer.EffectColorKey
	EffectColorRange            = serializer.EffectColorRange
	EffectExtract               = serializer.EffectExtract
	EffectGeometryLegacy        = serializer.EffectGeometryLegacy
	EffectGradientWipe          = serializer.EffectGradientWipe
	EffectGrowBounds            = serializer.EffectGrowBounds
	EffectKeyCleaner            = serializer.EffectKeyCleaner
	EffectLayerControl          = serializer.EffectLayerControl
	EffectLumaKey               = serializer.EffectLumaKey
	EffectMedian                = serializer.EffectMedian
	EffectNoiseHLSAuto          = serializer.EffectNoiseHLSAuto
	EffectColorProfileConverter = serializer.EffectColorProfileConverter
	EffectTimeDisplacement      = serializer.EffectTimeDisplacement
	EffectTimecode              = serializer.EffectTimecode
	EffectCCBallAction          = serializer.EffectCCBallAction
	EffectCCBubbles             = serializer.EffectCCBubbles
	EffectCCComposite           = serializer.EffectCCComposite
	EffectCCDrizzle             = serializer.EffectCCDrizzle
	EffectCCEnvironment         = serializer.EffectCCEnvironment
	EffectCCGlassWipe           = serializer.EffectCCGlassWipe
	EffectCCGlueGun             = serializer.EffectCCGlueGun
	EffectCCGridWipe            = serializer.EffectCCGridWipe
	EffectCCHair                = serializer.EffectCCHair
	EffectCCImageWipe           = serializer.EffectCCImageWipe
	EffectCCJaws                = serializer.EffectCCJaws
	EffectCCLightWipe           = serializer.EffectCCLightWipe
	EffectCCMrMercury           = serializer.EffectCCMrMercury
	EffectCCParticleSystemsII   = serializer.EffectCCParticleSystemsII
	EffectCCRadialScaleWipe     = serializer.EffectCCRadialScaleWipe
	EffectCCRain                = serializer.EffectCCRain
	EffectCCScaleWipe           = serializer.EffectCCScaleWipe
	EffectCCSnow                = serializer.EffectCCSnow
	EffectCCTwister             = serializer.EffectCCTwister
	EffectCCBlockLoad           = serializer.EffectCCBlockLoad
	EffectCCColorNeutralizer    = serializer.EffectCCColorNeutralizer
	EffectCCKernel              = serializer.EffectCCKernel
	EffectCCLineSweep           = serializer.EffectCCLineSweep
	EffectCCRainfall            = serializer.EffectCCRainfall
	EffectCCSnowfall            = serializer.EffectCCSnowfall

	EffectAudioBackwards    = serializer.EffectAudioBackwards    // Backwards
	EffectAudioBassTreble   = serializer.EffectAudioBassTreble   // Bass & Treble
	EffectAudioDelay        = serializer.EffectAudioDelay        // Delay
	EffectAudioFlangeChorus = serializer.EffectAudioFlangeChorus // Flange & Chorus
	EffectAudioHighLowPass  = serializer.EffectAudioHighLowPass  // High-Low Pass
	EffectAudioModulator    = serializer.EffectAudioModulator    // Modulator
	EffectAudioParametricEQ = serializer.EffectAudioParametricEQ // Parametric EQ
	EffectAudioReverb       = serializer.EffectAudioReverb       // Reverb
	EffectAudioStereoMixer  = serializer.EffectAudioStereoMixer  // Stereo Mixer
	EffectAudioTone         = serializer.EffectAudioTone         // Tone

	// Wave 11 layer-reference effects.
	EffectWarpStabilizer  = serializer.EffectWarpStabilizer  // Warp Stabilizer
	Effect3DGlasses       = serializer.Effect3DGlasses       // 3D Glasses
	EffectTimewarp        = serializer.EffectTimewarp        // Timewarp
	EffectCCParticleWorld = serializer.EffectCCParticleWorld // CC Particle World

	// Wave 12 parked classic effects.
	EffectBezierWarp         = serializer.EffectBezierWarp         // Bezier Warp
	EffectMeshWarp           = serializer.EffectMeshWarp           // Mesh Warp
	EffectChannelMixer       = serializer.EffectChannelMixer       // Channel Mixer
	EffectReshape            = serializer.EffectReshape            // Reshape
	EffectVectorPaint        = serializer.EffectVectorPaint        // Vector Paint
	EffectTexturize          = serializer.EffectTexturize          // Texturize
	EffectColorLink          = serializer.EffectColorLink          // Color Link
	EffectCompoundArithmetic = serializer.EffectCompoundArithmetic // Compound Arithmetic
	EffectSetChannels        = serializer.EffectSetChannels        // Set Channels

	EffectDisplacementMapLayer = serializer.EffectDisplacementMapLayer // Displacement Map Layer param
	EffectCompoundBlurLayer    = serializer.EffectCompoundBlurLayer    // Compound Blur "Blur Layer" param
	EffectCCVectorBlurMap      = serializer.EffectCCVectorBlurMap      // CC Vector Blur "Vector Map" param
	// Wave 11 layer-ref params.
	EffectWarpStabilizerRefLayer = serializer.EffectWarpStabilizerRefLayer // Warp Stabilizer reference layer
	Effect3DGlassesLeftView      = serializer.Effect3DGlassesLeftView      // 3D Glasses left view
	Effect3DGlassesRightView     = serializer.Effect3DGlassesRightView     // 3D Glasses right view
	EffectTimewarpMatteLayer     = serializer.EffectTimewarpMatteLayer     // Timewarp matte layer
	EffectTimewarpSourceLayer    = serializer.EffectTimewarpSourceLayer    // Timewarp source layer
	EffectCCParticleWorldTexture = serializer.EffectCCParticleWorldTexture // CC Particle World texture layer
	// Wave 12 layer-ref params.
	EffectTexturizeLayer                 = serializer.EffectTexturizeLayer                 // Texturize texture layer
	EffectColorLinkSourceLayer           = serializer.EffectColorLinkSourceLayer           // Color Link source layer
	EffectCompoundArithmeticSecondSource = serializer.EffectCompoundArithmeticSecondSource // Compound Arithmetic 2nd source
	EffectSetChannelsSource1             = serializer.EffectSetChannelsSource1             // Set Channels source 1
	EffectSetChannelsSource2             = serializer.EffectSetChannelsSource2             // Set Channels source 2
	EffectSetChannelsSource3             = serializer.EffectSetChannelsSource3             // Set Channels source 3
	EffectSetChannelsSource4             = serializer.EffectSetChannelsSource4             // Set Channels source 4
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
//
//aep:cap domain=render-queue tier=alpha verify=roundtrip incident=render-queue-delete-mechanics boundary="需队列已有 >=1 item 作模板;输出模块沿用模板路径;未 AE-gate" alias="render queue,渲染队列,add item,RQ,导出"
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
//
//aep:cap domain=render-queue tier=alpha verify=roundtrip incident=render-queue-delete-mechanics boundary="仅单输出模块 item 的 Rout RE 覆盖;未 AE-gate" alias="render queue,渲染队列,remove item"
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
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestSetRenderer_AEShipGate_AE2020,TestSetRenderer_AEShipGate_AE2025 boundary="binary 或 ExtendScript 名;各 AE 版本暴露引擎不同" alias="renderer,渲染器,3D 引擎,advanced 3d,cinema 4d"
func SetRenderer(c *Composition, name string) error { return serializer.SetRenderer(c, name) }
