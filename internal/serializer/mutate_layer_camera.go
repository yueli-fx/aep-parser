// Public-API entry for source-less layer creation (Camera / Light).
//
// Unlike Solid/Null/Adjustment (which need a backing footage Item), a Camera or
// Light layer is defined entirely by its ldta + a layer-specific property group
// (Camera Options / Light Options) — the same 4-child Layr shape as a shape
// layer (ldta, Utf8 name, outer LIST(tdgp), Gide). We embed a whole AE-native
// Camera/Light Layr (extracted from re_cameralight.aep) and splice it in,
// patching only the layer ID, name, and time fields — the InsertLayer
// clone-a-Layr approach, so every AE-internal flag byte stays AE-faithful
// (sidesteps the from-scratch silent-drop pitfalls shape layers needed RE for).
package serializer

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

//go:embed templates/layers/layer_camera_body.bin
var layerCameraBodyBytes []byte

//go:embed templates/layers/layer_light_body.bin
var layerLightBodyBytes []byte

// light_color_leaf.bin is a LIST(tdgp) wrapper around the
// (tdmn "ADBE Light Color", LIST:tdbs) pair, extracted from an AE-2020-saved
// light whose color was authored non-default (AE elides the default white, so
// the from-scratch light template carries no Color slot). newTemplatedLayer
// splices this into a fresh light's Light Options group and resets the cdat to
// white — lighting up SetLightColor / LightColor() from scratch. Same
// synthesis-insert blueprint as SetEffectParam's materialized param streams.
//
//go:embed templates/options/light_color_leaf.bin
var lightColorLeafBytes []byte

// camera_iris_leaves.bin is a LIST(tdgp) wrapper around the 8 Iris*/Highlight*
// (tdmn, LIST:tdbs) pairs in AE's canonical order, extracted from an AE-2020
// camera with DoF on + those options authored non-default (AE elides these
// DoF-bokeh controls at default, so the from-scratch camera template carries no
// slots — even AE's own parsed cameras lack them). newTemplatedLayer splices
// them after Blur Level + resets each cdat to AE's default. Same synthesis-
// insert blueprint as the Light Color leaf.
//
//go:embed templates/options/camera_iris_leaves.bin
var cameraIrisLeavesBytes []byte

// NewCameraLayer adds a new Camera layer to the composition.
// (Full contract lives on the aep.NewCameraLayer facade — docgen source.)
func NewCameraLayer(c *Composition, name string) (*Layer, error) {
	return newTemplatedLayer(c, name, layerCameraBodyBytes, LayerTypeCamera)
}

// NewLightLayer adds a new Light layer to the composition.
// (Full contract lives on the aep.NewLightLayer facade — docgen source.)
func NewLightLayer(c *Composition, name string) (*Layer, error) {
	return newTemplatedLayer(c, name, layerLightBodyBytes, LayerTypeLight)
}

// spliceLightColorLeaf inserts the AE-native (tdmn "ADBE Light Color",
// LIST:tdbs) pair at the front of the cloned light's Light Options group
// (Color is canonically the group's first child, ahead of Intensity). The
// enclosing LIST sizes reflow on Write (length-variable, like a name edit), and
// AE property groups are Group-End-sentinel-delimited with no child-count
// header to bump — same as SetEffectParam's value-stream splice.
func spliceLightColorLeaf(layrChunk *rifx.Chunk) error {
	var outer *rifx.Chunk
	for _, ch := range layrChunk.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			outer = ch
			break
		}
	}
	if outer == nil {
		return fmt.Errorf("light template: no outer tdgp group")
	}
	opts := findGroupBody(outer, "ADBE Light Options Group")
	if opts == nil {
		return fmt.Errorf("light template: no Light Options group")
	}
	// Already present (re-extracted template carries it) — nothing to do.
	for _, ch := range opts.Children {
		if ch.ID == rifx.IDTdmn && trimChunkNUL(ch.Data) == "ADBE Light Color" {
			return nil
		}
	}
	wrapper, err := rifx.ReadChunk(bytes.NewReader(lightColorLeafBytes))
	if err != nil {
		return fmt.Errorf("parse light color leaf template: %w", err)
	}
	if len(wrapper.Children) != 2 {
		return fmt.Errorf("light color leaf template: want 2 children (tdmn, tdbs), got %d", len(wrapper.Children))
	}
	// AE's canonical Light Options order is Intensity[1], Color[2], … — insert
	// the pair right after Intensity's payload. Property order within a group is
	// significant: a Color spliced ahead of Intensity is dropped by AE on open.
	insertAt := 0
	kids := opts.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == "ADBE Light Intensity" {
			insertAt = i + 2 // after the (tdmn, payload) pair
			break
		}
	}
	spliced := make([]*rifx.Chunk, 0, len(kids)+2)
	spliced = append(spliced, kids[:insertAt]...)
	spliced = append(spliced, wrapper.Children...)
	spliced = append(spliced, kids[insertAt:]...)
	opts.Children = spliced
	return nil
}

// spliceCameraIrisLeaves inserts the 8 AE-native Iris*/Highlight* leaf pairs
// after "ADBE Camera Blur Level" in the cloned camera's Camera Options group
// (AE's canonical order: Zoom, DoF, Focus, Aperture, Blur Level, then the 8
// iris controls). Same group-splice mechanics as spliceLightColorLeaf.
func spliceCameraIrisLeaves(layrChunk *rifx.Chunk) error {
	var outer *rifx.Chunk
	for _, ch := range layrChunk.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			outer = ch
			break
		}
	}
	if outer == nil {
		return fmt.Errorf("camera template: no outer tdgp group")
	}
	opts := findGroupBody(outer, "ADBE Camera Options Group")
	if opts == nil {
		return fmt.Errorf("camera template: no Camera Options group")
	}
	// Idempotent: if the iris leaves are already present (re-extracted template),
	// do nothing.
	for _, ch := range opts.Children {
		if ch.ID == rifx.IDTdmn && trimChunkNUL(ch.Data) == "ADBE Iris Shape" {
			return nil
		}
	}
	wrapper, err := rifx.ReadChunk(bytes.NewReader(cameraIrisLeavesBytes))
	if err != nil {
		return fmt.Errorf("parse camera iris leaves template: %w", err)
	}
	if len(wrapper.Children) != 16 {
		return fmt.Errorf("camera iris leaves template: want 16 children (8 tdmn,tdbs pairs), got %d", len(wrapper.Children))
	}
	insertAt := len(opts.Children) // fallback: before Group End handled below
	kids := opts.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == "ADBE Camera Blur Level" {
			insertAt = i + 2 // after the (tdmn, payload) pair
			break
		}
	}
	spliced := make([]*rifx.Chunk, 0, len(kids)+16)
	spliced = append(spliced, kids[:insertAt]...)
	spliced = append(spliced, wrapper.Children...)
	spliced = append(spliced, kids[insertAt:]...)
	opts.Children = spliced
	return nil
}

// newTemplatedLayer clones an embedded AE-native Layr template, patches its
// per-instance fields (ID / name / time span), and splices it into the comp's
// Item LIST with the same atomic machinery as NewShapeLayer (Ewst sibling +
// fvdv… followers, ID-collision avoidance, head-counter + cdta bump,
// warnings-as-failure rollback).
func newTemplatedLayer(c *Composition, name string, templateBytes []byte, typ LayerType) (*Layer, error) {
	if name == "" {
		return nil, fmt.Errorf("layer name cannot be empty")
	}
	cProj := scene.CompositionProj(c)
	if cProj == nil {
		return nil, fmt.Errorf("internal: comp has no project back-ref")
	}
	cb := compositionBack(c)
	if cb == nil || cb.itemList == nil {
		return nil, fmt.Errorf("internal: comp has no itemList chunk")
	}

	// Clone the template Layr (ReadChunk yields fresh chunks each call).
	layrChunk, err := rifx.ReadChunk(bytes.NewReader(templateBytes))
	if err != nil {
		return nil, fmt.Errorf("parse layer template: %w", err)
	}
	ldta := layrChunk.FindFirst(rifx.IDLdta)
	if ldta == nil || len(ldta.Data) < 0x88 {
		return nil, fmt.Errorf("layer template ldta missing/short")
	}

	// Per-composition layer-ID namespace (must avoid template service-layer IDs).
	layerID := maxLayerIDInItemList(cb.itemList) + 1
	d := ldta.Data
	binary.BigEndian.PutUint32(d[codec.LdtaLayerID:codec.LdtaLayerID+4], layerID)

	// Re-home the layer's time span into THIS comp (the template carries its
	// source comp's ticks; an out-point past the new comp's duration risks an
	// AE clamp/reject). start=0, in=0, out=comp duration, divisor=TickRate.
	tickRate := uint32(c.TickRate)
	if tickRate == 0 {
		tickRate = 30720
	}
	outTicks := uint32(c.Duration * float64(tickRate))
	binary.BigEndian.PutUint32(d[codec.LdtaStartTimeDivd:codec.LdtaStartTimeDivd+4], 0)
	binary.BigEndian.PutUint32(d[codec.LdtaStartTimeDivs:codec.LdtaStartTimeDivs+4], tickRate)
	binary.BigEndian.PutUint32(d[codec.LdtaInPointDivd:codec.LdtaInPointDivd+4], 0)
	binary.BigEndian.PutUint32(d[codec.LdtaInPointDivs:codec.LdtaInPointDivs+4], tickRate)
	binary.BigEndian.PutUint32(d[codec.LdtaOutPointDivd:codec.LdtaOutPointDivd+4], outTicks)
	binary.BigEndian.PutUint32(d[codec.LdtaOutPointDivs:codec.LdtaOutPointDivs+4], tickRate)

	// A fresh layer has no parent (the template may carry a stale ParentID).
	if len(d) >= 0x88 {
		binary.BigEndian.PutUint32(d[0x84:0x88], 0)
	}

	// Pad/trim ldta to the target's capability size (160 AE2020/22, 164 AE2025).
	// Camera/Light fields all fit in the first 0xA0 bytes, so size only varies
	// the trailing zero-pad (same invariant as buildLdtaBytes for shapes).
	wantSize := codec.LdtaSize2025
	if caps := Capabilities(scene.ProjectTarget(cProj)); caps.LdtaSize > 0 {
		wantSize = caps.LdtaSize
	}
	if len(ldta.Data) < wantSize {
		ldta.Data = append(ldta.Data, make([]byte, wantSize-len(ldta.Data))...)
	} else if len(ldta.Data) > wantSize {
		ldta.Data = ldta.Data[:wantSize]
	}

	// Replace the Utf8 name child (length-variable).
	var nameChunk *rifx.Chunk
	for _, ch := range layrChunk.Children {
		if ch.ID == rifx.IDUtf8 {
			ch.Data = []byte(name)
			nameChunk = ch
			break
		}
	}

	// A fresh light's template has no `ADBE Light Color` slot (AE elides the
	// default white). Splice the AE-native leaf into its Light Options group so
	// LightColor()/SetLightColor work from scratch; it is reset to white below
	// (after the property tree is wired) so an untouched fresh light stays white.
	if typ == LayerTypeLight {
		if err := spliceLightColorLeaf(layrChunk); err != nil {
			return nil, err
		}
	}
	// A fresh camera's template carries only the 5 non-elided options; the 8
	// Iris*/Highlight* DoF-bokeh controls are AE-elided. Splice their AE-native
	// leaves after Blur Level so SetIris*/SetIrisHighlight* work from scratch;
	// each is reset to AE's default below (after the property tree is wired).
	if typ == LayerTypeCamera {
		if err := spliceCameraIrisLeaves(layrChunk); err != nil {
			return nil, err
		}
	}

	// Runtime Layer wrapper. Capture ldta + nameChunk so ldta-based setters
	// (SetSource for precomp, SetParent, flag bits) work on the fresh layer.
	base := &Layer{Type: typ, Name: name, ID: layerID}
	scene.SetLayerComp(base, c)
	scene.SetLayerBack(base, &layerBackrefs{layrList: layrChunk, ldta: ldta, nameChunk: nameChunk})
	if layerID >= scene.ProjectNextItemID(cProj) {
		scene.SetProjectNextItemID(cProj, layerID+1)
	}

	// Parse the cloned template's property tree + wire its leaf backrefs so the
	// Camera/Light Options accessors + setters (SetCameraZoom / SetLightIntensity
	// / …) work on this fresh layer exactly as on a parsed one. The template is
	// AE-native and parses clean; parse warnings go to a LOCAL sink so they never
	// trip the splice's warnings-as-failure rollback. Read-only on the bytes —
	// an untouched fresh layer still serializes byte-identically (the create
	// ship-gate stays green); a setter then overwrites only its own cdat in place.
	localWarn := []string{}
	pctx := newParseCtx(float64(tickRate), name, &localWarn)
	base.Properties, base.Effects, base.Markers = parseProperties(layrChunk, pctx)
	ptree := buildAEPropertyGroupTree(layrChunk)
	scene.SetLayerPropertyTree(base, ptree)
	scene.SetPropertyGroupLayer(ptree, base)
	wirePropertyTreeLeaves(ptree, base.Properties)

	// Reset the spliced Light Color leaf to AE's default white (the embedded
	// template carries the extraction fixture's authored colour). AE stores
	// light colour as raw [A,R,G,B] in 0..255, so white = all 255. Best-effort:
	// a missing slot just leaves LightColor() nil, as before.
	if typ == LayerTypeLight {
		if lc := base.LightColor(); lc != nil {
			_ = base.SetLightColor([]float64{255, 255, 255, 255})
		}
	}

	// Reset the spliced camera Iris*/Highlight* leaves to AE's defaults (the
	// embedded template carries the extraction fixture's authored values).
	// Best-effort: a missing slot just leaves the accessor nil, as before.
	if typ == LayerTypeCamera {
		for _, r := range []struct {
			p   *Property
			def float64
		}{
			{base.IrisShape(), 1},
			{base.IrisRotation(), 0},
			{base.IrisRoundness(), 0},
			{base.IrisAspectRatio(), 1},
			{base.IrisDiffractionFringe(), 0},
			{base.IrisHighlightGain(), 0},
			{base.IrisHighlightThreshold(), 1},
			{base.IrisHighlightSaturation(), 0},
		} {
			if r.p != nil {
				_ = r.p.SetStaticValue(r.def)
			}
		}
	}

	// cdta @0x18: bump fresh-comp marker (600) to TickRate — AE's "comp has user
	// content" gate (see NewShapeLayer).
	if cb.cdta != nil && len(cb.cdta.Data) >= codec.CdtaSecondaryDivisor18+4 && tickRate > 0 {
		binary.BigEndian.PutUint32(cb.cdta.Data[codec.CdtaSecondaryDivisor18:codec.CdtaSecondaryDivisor18+4], tickRate)
	}

	// Atomic splice: snapshot, insert the layer unit before the first template
	// service layer, roll back on any parser warning.
	oldItemChildren := append([]*rifx.Chunk(nil), cb.itemList.Children...)
	oldLayersLen := len(c.Layers)
	oldWarningsLen := len(cProj.Warnings)

	unit := append([]*rifx.Chunk{
		layrChunk,
		{ID: rifx.IDList, FormType: rifx.IDEwst},
	}, lowerLayerSiblings()...)
	insertIdx := insertLayrPosition(cb.itemList.Children)
	n := len(unit)
	cb.itemList.Children = append(cb.itemList.Children, make([]*rifx.Chunk, n)...)
	copy(cb.itemList.Children[insertIdx+n:], cb.itemList.Children[insertIdx:len(cb.itemList.Children)-n])
	copy(cb.itemList.Children[insertIdx:insertIdx+n], unit)
	c.Layers = append(c.Layers, base)

	if len(cProj.Warnings) != oldWarningsLen {
		cb.itemList.Children = oldItemChildren
		c.Layers = c.Layers[:oldLayersLen]
		newWarnings := append([]string(nil), cProj.Warnings[oldWarningsLen:]...)
		cProj.Warnings = cProj.Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("internal: NewLayer(%v) produced %d parser warning(s): %v", typ, len(newWarnings), newWarnings)
	}

	return base, nil
}
