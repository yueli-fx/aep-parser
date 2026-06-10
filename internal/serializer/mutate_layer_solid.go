// Public-API entry for footage-backed utility layer creation (Solid / Null /
// Adjustment).
//
// Unlike Camera/Light (source-less, embed a single Layr), these layers need a
// backing solid footage Item (`opti "Soli"` + Pin/sspc). Rather than
// synthesizing that Item from scratch, we embed a whole AE-2020-native template
// project (templates/solidnull_2020.aep, one comp with a solid + null +
// adjustment layer) and reuse the cross-Project InsertLayer machinery: it
// imports the layer's footage closure with fresh item IDs and remaps the
// clone's SourceID, exactly what AE-faithful solid creation needs. Per-instance
// fields (layer name / time span / solid color / dimensions / footage name) are
// patched afterward — all length-preserving (the opti name + color live inside
// the fixed 282-byte chunk) or managed splices (Utf8 layer name).
package serializer

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/scene"
)

//go:embed templates/solidnull_2020.aep
var solidLibraryBytes []byte

// opti "Soli" layout (RE: test_data/re_solidnull.aep, AE 2020; 282-byte chunk):
// tag "Soli" @0x00, ARGB 4×float32 BE @0x0A (alpha always 1.0), display name
// @0x1A (NUL-terminated UTF-8 inside the fixed tail buffer). sspc offsets match
// parseFootage's real-AE layout (u16 BE width/height @0x20/@0x24).
const (
	optiSoliColorA = 0x0A
	optiSoliColorR = 0x0E
	optiSoliColorG = 0x12
	optiSoliColorB = 0x16
	optiSoliName   = 0x1A

	sspcWidthOff  = 0x20
	sspcHeightOff = 0x24

	solidMaxDim = 30000 // AE's solid dimension ceiling
)

// Template layer names inside templates/solidnull_2020.aep, comp "TplComp".
const (
	tplSolidLayer      = "TplSolid"
	tplNullLayer       = "TplNull"
	tplAdjustmentLayer = "TplAdj"
)

// NewSolidLayer adds a new solid-backed layer to the composition.
// (Full contract lives on the aep.NewSolidLayer facade — docgen source.)
func NewSolidLayer(c *Composition, name string, width, height int, rgb [3]float64) (*Layer, error) {
	if width < 1 || width > solidMaxDim || height < 1 || height > solidMaxDim {
		return nil, fmt.Errorf("NewSolidLayer: dimensions %dx%d out of range 1..%d", width, height, solidMaxDim)
	}
	for i, v := range rgb {
		if math.IsNaN(v) || v < 0 || v > 1 {
			return nil, fmt.Errorf("NewSolidLayer: color[%d]=%v out of range 0..1", i, v)
		}
	}
	return newSolidBackedLayer(c, "NewSolidLayer", name, tplSolidLayer, width, height, &rgb)
}

// NewNullLayer adds a new null (100×100 invisible parenting helper) layer.
// (Full contract lives on the aep.NewNullLayer facade — docgen source.)
func NewNullLayer(c *Composition, name string) (*Layer, error) {
	return newSolidBackedLayer(c, "NewNullLayer", name, tplNullLayer, 0, 0, nil)
}

// NewAdjustmentLayer adds a new comp-sized adjustment layer.
// (Full contract lives on the aep.NewAdjustmentLayer facade — docgen source.)
func NewAdjustmentLayer(c *Composition, name string) (*Layer, error) {
	if c == nil {
		return nil, fmt.Errorf("NewAdjustmentLayer: composition cannot be nil")
	}
	return newSolidBackedLayer(c, "NewAdjustmentLayer", name, tplAdjustmentLayer, int(c.Width), int(c.Height), nil)
}

// newSolidBackedLayer clones tplLayerName out of the embedded template project
// via cross-Project InsertLayer (footage closure import + SourceID remap), then
// patches per-instance fields. The template footage structure is validated
// BEFORE the insert so the post-insert patches — which operate on a
// byte-identical clone — cannot fail and leave partial state behind.
func newSolidBackedLayer(c *Composition, op, name, tplLayerName string, width, height int, rgb *[3]float64) (*Layer, error) {
	if c == nil {
		return nil, fmt.Errorf("%s: composition cannot be nil", op)
	}
	if name == "" {
		return nil, fmt.Errorf("%s: layer name cannot be empty", op)
	}
	if len(name) > 255 {
		return nil, fmt.Errorf("%s: name exceeds the 255-byte opti buffer (%d bytes)", op, len(name))
	}

	// Parse a fresh template instance per call (same pattern as NewProject) so
	// nothing can alias the embedded library across calls.
	lib, err := FromReader(bytes.NewReader(solidLibraryBytes))
	if err != nil {
		return nil, fmt.Errorf("%s: corrupt embedded solid template (build bug): %w", op, err)
	}
	var src *Layer
	for _, tc := range lib.Compositions {
		if tc.Name != "TplComp" {
			continue
		}
		for _, l := range tc.Layers {
			if l.Name == tplLayerName {
				src = l
				break
			}
		}
	}
	if src == nil {
		return nil, fmt.Errorf("%s: template layer %q missing from embedded library (build bug)", op, tplLayerName)
	}
	if fErr := validateSolidTemplateFootage(lib, src.SourceID); fErr != nil {
		return nil, fmt.Errorf("%s: %w", op, fErr)
	}

	// Splice via the cross-Project import machinery directly (not the public
	// InsertLayer wrapper, whose refuse-matrix only admits AV src — a template
	// null parses as LayerTypeNull). The wrapper's src-side prerequisites are
	// re-established here against our own pre-validated template.
	if compositionBack(c) == nil || compositionBack(c).itemList == nil {
		return nil, fmt.Errorf("%s: dest comp %q has no itemList back-ref (built outside parser?)", op, c.Name)
	}
	if scene.CompositionProj(c) == nil {
		return nil, fmt.Errorf("%s: dest comp %q has no project back-ref", op, c.Name)
	}
	srcBack := layerBack(src)
	srcCb := compositionBack(scene.LayerComp(src))
	if srcBack == nil || srcBack.layrList == nil || srcCb == nil || srcCb.itemList == nil {
		return nil, fmt.Errorf("%s: template layer %q missing chunk back-refs (build bug)", op, tplLayerName)
	}
	srcLayrIdx := findLayrIndexInItemList(srcCb.itemList, srcBack.layrList)
	if srcLayrIdx < 0 {
		return nil, fmt.Errorf("%s: template layer %q Layr not found in template itemList (build bug)", op, tplLayerName)
	}
	clone, err := insertLayerCrossProject(c, src, len(c.Layers), srcLayrIdx, srcCb.itemList.Children)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Pad/trim the cloned ldta to the dest target's capability size (160
	// AE2020/22, 164 AE2025) — the template is AE-2020-form (160B) and a
	// mixed-form layer inside a 2025-form project risks an AE reject. Solid
	// fields all sit below 0xA0, so size only varies the trailing zero-pad
	// (same invariant as newTemplatedLayer for camera/light).
	if lb := layerBack(clone); lb != nil && lb.ldta != nil {
		wantSize := codec.LdtaSize2025
		if caps := Capabilities(scene.ProjectTarget(scene.CompositionProj(c))); caps.LdtaSize > 0 {
			wantSize = caps.LdtaSize
		}
		if len(lb.ldta.Data) < wantSize {
			lb.ldta.Data = append(lb.ldta.Data, make([]byte, wantSize-len(lb.ldta.Data))...)
		} else if len(lb.ldta.Data) > wantSize {
			lb.ldta.Data = lb.ldta.Data[:wantSize]
		}
	}

	// Per-instance patches. All chunk shapes were pre-validated on the
	// template, so errors here indicate an internal bug, not user input.
	if err := clone.SetName(name); err != nil {
		return nil, fmt.Errorf("%s: rename clone: %w", op, err)
	}
	// Re-home the time span into THIS comp: start=0, in=0, out=comp duration
	// (the template carries its source comp's 10s span; an out-point past the
	// dest comp's duration risks an AE clamp/reject — see camera/light RE).
	if err := clone.SetStartTime(0); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := clone.SetInPoint(0); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := clone.SetOutPoint(c.Duration); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	proj := scene.CompositionProj(c)
	f := footageByID(proj, clone.SourceID)
	if f == nil {
		return nil, fmt.Errorf("%s: imported solid footage id=%d not found (internal)", op, clone.SourceID)
	}
	fb := footageBack(f)
	d := fb.optiChunk.Data
	for i := optiSoliName; i < len(d); i++ {
		d[i] = 0
	}
	copy(d[optiSoliName:], name)
	f.Name = name
	if rgb != nil {
		putF32 := func(off int, v float64) {
			binary.BigEndian.PutUint32(d[off:off+4], math.Float32bits(float32(v)))
		}
		putF32(optiSoliColorA, 1)
		putF32(optiSoliColorR, rgb[0])
		putF32(optiSoliColorG, rgb[1])
		putF32(optiSoliColorB, rgb[2])
		f.SolidColor = [3]float64{rgb[0], rgb[1], rgb[2]}
	}
	if width > 0 && height > 0 {
		sd := fb.sspcChunk.Data
		binary.BigEndian.PutUint16(sd[sspcWidthOff:sspcWidthOff+2], uint16(width))
		binary.BigEndian.PutUint16(sd[sspcHeightOff:sspcHeightOff+2], uint16(height))
		f.Width, f.Height = uint16(width), uint16(height)
	}

	// cdta @0x18: bump fresh-comp marker (600) to TickRate — AE's "comp has
	// user content" gate (see NewShapeLayer).
	if cb := compositionBack(c); cb != nil && cb.cdta != nil && len(cb.cdta.Data) >= codec.CdtaSecondaryDivisor18+4 {
		if tr := uint32(c.TickRate); tr > 0 {
			binary.BigEndian.PutUint32(cb.cdta.Data[codec.CdtaSecondaryDivisor18:codec.CdtaSecondaryDivisor18+4], tr)
		}
	}

	return clone, nil
}

// validateSolidTemplateFootage checks that the template layer's backing solid
// footage has the chunk shapes every post-insert patch relies on. Failures are
// build bugs (corrupt/mismatched embedded template), surfaced before any dest
// mutation happens.
func validateSolidTemplateFootage(lib *Project, sourceID uint32) error {
	f := footageByID(lib, sourceID)
	if f == nil || !f.IsSolid {
		return fmt.Errorf("template footage id=%d missing or not a solid (build bug)", sourceID)
	}
	fb := footageBack(f)
	if fb == nil || fb.optiChunk == nil || len(fb.optiChunk.Data) <= optiSoliName {
		return fmt.Errorf("template solid footage id=%d opti chunk missing/short (build bug)", sourceID)
	}
	if fb.sspcChunk == nil || len(fb.sspcChunk.Data) < sspcHeightOff+2 {
		return fmt.Errorf("template solid footage id=%d sspc chunk missing/short (build bug)", sourceID)
	}
	return nil
}

// footageByID returns the project's footage item with the given item ID, or nil.
func footageByID(p *Project, id uint32) *Footage {
	if p == nil {
		return nil
	}
	for _, f := range p.Footage {
		if f.ID == id {
			return f
		}
	}
	return nil
}
