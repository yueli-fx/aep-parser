// Public-API entry for exposing an effect parameter in a comp's Essential
// Graphics panel (.mogrt controller).
//
// AE persists an EG controller across three coordinated sites (RE 2026-06-12):
//
//  1. The comp Item's CIF* panel lists — AE writes three byte-identical
//     generations (CIFO / CIF2 / CIF3); each gains a LIST:CCtl describing the
//     controller (name, uuid, CTyp, type-keyed value chunks, CPrp property
//     ref with a JSON matchName path) and a CcCt count bump.
//  2. The host layer's "ADBE Layer Overrides" parade entry — a non-standard
//     triple (tdmn + LIST:OvG2 + LIST:tdgp). OvG2 gains a CPrp uuid slot
//     (CprC count bump); the sibling tdgp gains an override value stream
//     (tdmn = leaf param match-name + LIST:tdbs), in CPrp order.
//  3. Nothing on the source property itself — AE resolves the link through
//     the CPrp JSON path (element index = 0-based position within the parent
//     group; 4294967295 = -1 for fixed groups like the Effect Parade).
//
// Atomic: every chunk-side mutation site is snapshotted before commit; the
// re-decoded controller list and any parser warning trigger a full rollback.
package serializer

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// newEGUUID returns a random lowercase RFC-4122 v4 UUID — the format AE uses
// for EG controller identities (36-char, hyphenated).
func newEGUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("AddEssentialProperty: uuid generation: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func leafChunk(id rifx.ChunkID, data []byte) *rifx.Chunk {
	return &rifx.Chunk{ID: id, Data: data}
}

func u32BE(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

func u32LE(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

func f64BE(v float64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, math.Float64bits(v))
	return b
}

// egLocalizedList builds a LIST(CpS2/CapS) string holder. CsCt is
// little-endian on disk (observed 01000000 across every AE fixture, vs the
// big-endian CcCt / CprC-in-CCtl counts — replicated verbatim).
func egLocalizedList(formType rifx.ChunkID, value, locale string) *rifx.Chunk {
	kids := []*rifx.Chunk{leafChunk(rifx.IDCsCt, u32LE(1))}
	if formType == rifx.IDCapS {
		kids = append(kids, leafChunk(rifx.IDCapL, u32BE(0)))
	}
	kids = append(kids, leafChunk(rifx.IDUtf8, []byte(value)))
	if formType == rifx.IDCpS2 {
		kids = append(kids, leafChunk(rifx.IDUtf8, []byte(locale)))
	}
	return &rifx.Chunk{ID: rifx.IDList, FormType: formType, Children: kids}
}

// egPanelLocale reads the locale tag from a CIF* generation's template-name
// CpS2 (value Utf8 followed by locale Utf8), falling back to "en_US".
func egPanelLocale(gen *rifx.Chunk) string {
	cps2 := gen.FindFirstList(rifx.IDCpS2)
	if cps2 == nil {
		return "en_US"
	}
	seen := 0
	for _, ch := range cps2.Children {
		if ch.ID == rifx.IDUtf8 {
			seen++
			if seen == 2 {
				return string(ch.Data)
			}
		}
	}
	return "en_US"
}

// egPathJSON renders the CPrp property path for an effect parameter. Element
// "index" = 0-based position of the node within its parent group; the Effect
// Parade itself is a fixed layer-root group and carries -1 (serialized
// unsigned, 4294967295) — both verbatim from AE-native fixtures.
func egPathJSON(effectMatchName string, effectIdx int, paramMatchName string, paramIdx int) string {
	return fmt.Sprintf(
		`{"0":{"index":4294967295,"matchName":"ADBE Effect Parade"},"1":{"index":%d,"matchName":%q},"2":{"index":%d,"matchName":%q}}`,
		effectIdx, effectMatchName, paramIdx, paramMatchName)
}

// findLayerOverridesTriple locates the layer's "ADBE Layer Overrides" parade
// entry — the (tdmn, LIST:OvG2, LIST:tdgp) triple — inside the Layr property
// tdgp. Returns nils when the layer has none.
func findLayerOverridesTriple(layr *rifx.Chunk) (ovg2, overrides *rifx.Chunk) {
	var propsTdgp *rifx.Chunk
	for _, ch := range layr.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			propsTdgp = ch
			break
		}
	}
	if propsTdgp == nil {
		return nil, nil
	}
	kids := propsTdgp.Children
	for i := 0; i+2 < len(kids); i++ {
		if kids[i].ID != rifx.IDTdmn || trimNUL(kids[i].Data) != "ADBE Layer Overrides" {
			continue
		}
		if kids[i+1].IsList() && kids[i+1].FormType == rifx.IDOvG2 &&
			kids[i+2].IsList() && kids[i+2].FormType == rifx.IDTdgp {
			return kids[i+1], kids[i+2]
		}
	}
	return nil, nil
}

// ensureLayerOverrides returns the layer's OvG2 + override tdgp, splicing an
// empty "ADBE Layer Overrides" triple (CprC=0 OvG2 + bare tdgp) into the Layr
// property tree when absent — anchored before "ADBE Layer Sets" / "ADBE Source
// Options Group", matching AE's emitted group order. The returned undo
// restores the pre-splice children (no-op when the triple already existed).
func ensureLayerOverrides(layer *Layer) (ovg2, overrides *rifx.Chunk, undo func(), err error) {
	lb := layerBack(layer)
	if lb == nil || lb.layrList == nil {
		return nil, nil, nil, fmt.Errorf("AddEssentialProperty: layer %q has no Layr chunk back-ref; round-trip the project through aep.Reopen first", layer.Name)
	}
	if ovg2, overrides = findLayerOverridesTriple(lb.layrList); ovg2 != nil {
		return ovg2, overrides, func() {}, nil
	}

	var propsTdgp *rifx.Chunk
	for _, ch := range lb.layrList.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			propsTdgp = ch
			break
		}
	}
	if propsTdgp == nil {
		return nil, nil, nil, fmt.Errorf("AddEssentialProperty: layer %q has no property-group LIST in its Layr", layer.Name)
	}
	anchor := -1
	for _, anchorName := range []string{"ADBE Layer Sets", "ADBE Source Options Group"} {
		for i, ch := range propsTdgp.Children {
			if ch.ID == rifx.IDTdmn && trimNUL(ch.Data) == anchorName {
				anchor = i
				break
			}
		}
		if anchor >= 0 {
			break
		}
	}
	if anchor < 0 {
		return nil, nil, nil, fmt.Errorf("AddEssentialProperty: layer %q has no anchor group to position ADBE Layer Overrides", layer.Name)
	}

	ovg2 = &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDOvG2, Children: []*rifx.Chunk{
		leafChunk(rifx.IDCprC, u32LE(0)),
	}}
	overrides = &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
		makeTdsb(),
		makeTdsn(aeDefaultGroupName),
		makeTdmn("ADBE Group End"),
	}}

	oldChildren := append([]*rifx.Chunk(nil), propsTdgp.Children...)
	spliced := make([]*rifx.Chunk, 0, len(propsTdgp.Children)+3)
	spliced = append(spliced, propsTdgp.Children[:anchor]...)
	spliced = append(spliced, makeTdmn("ADBE Layer Overrides"), ovg2, overrides)
	spliced = append(spliced, propsTdgp.Children[anchor:]...)
	propsTdgp.Children = spliced

	return ovg2, overrides, func() { propsTdgp.Children = oldChildren }, nil
}

// egParamOrdinal returns the 0-based position of paramMatchName among the
// effect's parT (tdmn, pard) definition pairs — the value AE writes as the
// leaf "index" in the CPrp path JSON (the header "-0000" pair counts as 0).
// Returns -1 when the param is not defined.
func egParamOrdinal(sspc *rifx.Chunk, paramMatchName string) int {
	var parT *rifx.Chunk
	for _, ch := range sspc.Children {
		if ch.IsList() && ch.FormType == rifx.IDparT {
			parT = ch
			break
		}
	}
	if parT == nil {
		return -1
	}
	ordinal := 0
	kids := parT.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID != rifx.IDTdmn || kids[i+1].ID != rifx.IDpard {
			continue
		}
		if trimNUL(kids[i].Data) == paramMatchName {
			return ordinal
		}
		ordinal++
		i++
	}
	return -1
}

// findMaterializedParam returns the tdbs stream for paramMatchName in the
// effect sspc's value tdgp, or nil when the param is default-elided.
func findMaterializedParam(sspc *rifx.Chunk, paramMatchName string) *rifx.Chunk {
	var valueGroup *rifx.Chunk
	for _, ch := range sspc.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			valueGroup = ch
			break
		}
	}
	if valueGroup == nil {
		return nil
	}
	kids := valueGroup.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID != rifx.IDTdmn || trimNUL(kids[i].Data) != paramMatchName {
			continue
		}
		if kids[i+1].IsList() && kids[i+1].FormType == rifx.IDTdbs {
			return kids[i+1]
		}
	}
	return nil
}

// egControllerValueChunks builds the CTyp-keyed CCtl value chunks plus the EG
// controller type for one effect parameter. Color requires the param's
// materialized cdat (AE stores no color default in pard); colorCdat is nil
// for the other types.
func egControllerValueChunks(def *pardParamDef, colorCdat []byte) (scene.EGControllerType, []*rifx.Chunk, error) {
	switch def.controlType {
	case PCTLScalar, PCTLSlider:
		cur, _ := def.lastValue.(float64)
		minV, _ := def.minValue.(float64)
		maxV, ok := def.maxValue.(float64)
		if !ok {
			maxV = 100
		}
		return scene.EGSlider, []*rifx.Chunk{
			leafChunk(rifx.IDCVal, f64BE(cur)),
			leafChunk(rifx.IDCDef, f64BE(cur)),
			leafChunk(rifx.IDSmin, f64BE(minV)),
			leafChunk(rifx.IDSmax, f64BE(maxV)),
		}, nil

	case PCTLBoolean:
		cur, _ := def.lastValue.(int)
		dflt, _ := def.defaultVal.(int)
		return scene.EGCheckbox, []*rifx.Chunk{
			leafChunk(rifx.IDCVal, []byte{byte(cur & 1)}),
			leafChunk(rifx.IDCDef, []byte{byte(dflt & 1)}),
		}, nil

	case PCTLColor:
		// cdat carries 4 BE f64 = ARGB in 0..255; CVal/CDef are 4 BE f32 =
		// RGBA in 0..1 (eg_color_controller.aep ground truth).
		if len(colorCdat) < 32 {
			return 0, nil, fmt.Errorf("AddEssentialProperty: color parameter has no materialized value stream — set a value first (SetEffectParam), then expose it")
		}
		argb := [4]float64{}
		for i := range argb {
			argb[i] = math.Float64frombits(binary.BigEndian.Uint64(colorCdat[i*8 : i*8+8]))
		}
		val := make([]byte, 16)
		rgba := [4]float64{argb[1], argb[2], argb[3], argb[0]}
		for i, v := range rgba {
			binary.BigEndian.PutUint32(val[i*4:i*4+4], math.Float32bits(float32(v/255.0)))
		}
		return scene.EGColor, []*rifx.Chunk{
			leafChunk(rifx.IDCVal, append([]byte(nil), val...)),
			leafChunk(rifx.IDCDef, append([]byte(nil), val...)),
		}, nil

	case PCTLTwoD:
		vals, ok := def.lastValue.([]float64)
		if !ok || len(vals) < 2 {
			return 0, nil, fmt.Errorf("AddEssentialProperty: 2D point parameter has no 2D default value")
		}
		val := make([]byte, 16)
		binary.BigEndian.PutUint64(val[0:8], math.Float64bits(vals[0]))
		binary.BigEndian.PutUint64(val[8:16], math.Float64bits(vals[1]))
		return scene.EGPoint, []*rifx.Chunk{
			leafChunk(rifx.IDCVal, append([]byte(nil), val...)),
			leafChunk(rifx.IDCDef, append([]byte(nil), val...)),
		}, nil
	}
	return 0, nil, fmt.Errorf("AddEssentialProperty: control type %d is not supported yet (supported: scalar/slider, boolean, color, 2D point)", def.controlType)
}

// buildEGCCtl assembles one LIST:CCtl controller entry.
func buildEGCCtl(displayName, locale, uuid string, ctyp scene.EGControllerType, valueChunks []*rifx.Chunk, compID, layerID uint32, pathJSON string) *rifx.Chunk {
	kids := []*rifx.Chunk{
		egLocalizedList(rifx.IDCpS2, displayName, locale),
		egLocalizedList(rifx.IDCapS, displayName, ""),
		leafChunk(rifx.IDUtf8, []byte(uuid)),
		leafChunk(rifx.IDCTyp, u32BE(uint32(ctyp))),
	}
	kids = append(kids, valueChunks...)
	kids = append(kids,
		leafChunk(rifx.IDCprC, u32BE(1)),
		&rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDCPrp, Children: []*rifx.Chunk{
			leafChunk(rifx.IDCCId, u32BE(compID)),
			leafChunk(rifx.IDCLId, u32BE(layerID)),
			leafChunk(rifx.IDUtf8, []byte(pathJSON)),
		}},
	)
	return &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDCctl, Children: kids}
}

// AddEssentialProperty exposes one parameter of an effect on layer in the
// owning comp's Essential Graphics panel and returns the new controller.
// (Full contract lives on the aep.AddEssentialProperty facade — docgen source.)
func AddEssentialProperty(layer *Layer, fx *Effect, paramMatchName, displayName string) (*EssentialGraphicsController, error) {
	if layer == nil || fx == nil {
		return nil, fmt.Errorf("AddEssentialProperty: layer/effect is nil")
	}
	comp := scene.LayerComp(layer)
	if comp == nil || scene.CompositionProj(comp) == nil {
		return nil, fmt.Errorf("AddEssentialProperty: layer %q has no composition/project back-ref", layer.Name)
	}
	cb := compositionBack(comp)
	if cb == nil || cb.itemList == nil {
		return nil, fmt.Errorf("AddEssentialProperty: comp %q has no Item LIST reference (built outside parser?)", comp.Name)
	}
	gens := egGenerations(cb.itemList)
	if len(gens) == 0 {
		return nil, fmt.Errorf("AddEssentialProperty: comp %q has no Essential Graphics panel shell (CIFO/CIF2/CIF3)", comp.Name)
	}

	// Resolve the effect's parade position (= path JSON element-1 index).
	effectIdx := -1
	for i, e := range layer.Effects {
		if e == fx {
			effectIdx = i
			break
		}
	}
	if effectIdx < 0 {
		return nil, fmt.Errorf("AddEssentialProperty: effect %q is not on layer %q", fx.MatchName, layer.Name)
	}
	parade := layer.EffectsParade()
	if parade == nil || effectIdx >= parade.NumProperties() {
		return nil, fmt.Errorf("AddEssentialProperty: layer %q has no parsed Effect Parade entry for effect %d; round-trip the project through aep.Reopen first", layer.Name, effectIdx)
	}
	g, ok := parade.ChildByIndex(effectIdx).(*AEPropertyGroup)
	if !ok {
		return nil, fmt.Errorf("AddEssentialProperty: parade child %d is not a property group", effectIdx)
	}
	pgb := propertyGroupBack(g)
	if pgb == nil || pgb.chunk == nil {
		return nil, fmt.Errorf("AddEssentialProperty: effect %q has no chunk back-ref", fx.MatchName)
	}

	defs := parsePardParams(pgb.chunk)
	def := defs[paramMatchName]
	if def == nil {
		return nil, fmt.Errorf("AddEssentialProperty: effect %q has no parameter %q in its pard definitions", fx.MatchName, paramMatchName)
	}
	paramIdx := egParamOrdinal(pgb.chunk, paramMatchName)
	if paramIdx < 0 {
		return nil, fmt.Errorf("AddEssentialProperty: parameter %q not found in effect %q parT", paramMatchName, fx.MatchName)
	}
	if displayName == "" {
		displayName = def.name
	}

	srcTdbs := findMaterializedParam(pgb.chunk, paramMatchName)
	var colorCdat []byte
	if srcTdbs != nil {
		if cdat := srcTdbs.FindFirst(rifx.IDCdat); cdat != nil {
			colorCdat = cdat.Data
		}
	}
	ctyp, valueChunks, err := egControllerValueChunks(def, colorCdat)
	if err != nil {
		return nil, err
	}

	uuid, err := newEGUUID()
	if err != nil {
		return nil, err
	}
	pathJSON := egPathJSON(fx.MatchName, effectIdx, paramMatchName, paramIdx)

	// Override value stream: clone the materialized param stream when present
	// (its cdat already holds the current value); otherwise materialize from
	// the per-param / generic control-type template. tdsn carries the display
	// name only when it differs from the param's own name — AE writes the
	// "-_0_/-" placeholder otherwise.
	var overrideTdbs *rifx.Chunk
	if srcTdbs != nil {
		overrideTdbs = deepCloneChunk(srcTdbs)
	} else {
		_, overrideTdbs, err = cloneEffectParamTemplate(paramMatchName)
		if err != nil {
			_, overrideTdbs, err = cloneGenericParamTemplate(paramMatchName, def)
			if err != nil {
				return nil, err
			}
		}
		ctx := newParseCtxFPS(comp.TickRate, comp.FrameRate, comp.Name, &scene.CompositionProj(comp).Warnings)
		prop := parseLeafProperty(paramMatchName, overrideTdbs, ctx)
		if prop == nil {
			return nil, fmt.Errorf("AddEssentialProperty: override stream for %q failed to parse", paramMatchName)
		}
		var cur any
		switch v := def.lastValue.(type) {
		case float64:
			cur = v
		case int:
			cur = float64(v)
		case []float64:
			cur = v
		}
		if cur != nil {
			if err := prop.SetStaticValue(cur); err != nil {
				return nil, fmt.Errorf("AddEssentialProperty: writing current value into override stream: %w", err)
			}
		}
	}
	overrideName := aeDefaultGroupName
	if displayName != def.name {
		overrideName = displayName
	}
	for i, ch := range overrideTdbs.Children {
		if ch.ID == rifx.IDTdsn {
			overrideTdbs.Children[i] = makeTdsn(overrideName)
		}
	}

	ovg2, overrides, undoOverridesCreate, err := ensureLayerOverrides(layer)
	if err != nil {
		return nil, err
	}
	cprc := ovg2.FindFirst(rifx.IDCprC)
	if cprc == nil || len(cprc.Data) < 4 {
		undoOverridesCreate()
		return nil, fmt.Errorf("AddEssentialProperty: layer %q OvG2 has no CprC count", layer.Name)
	}

	// Snapshot every mutation site for atomic rollback.
	type genSnap struct {
		children []*rifx.Chunk
		ccCt     *rifx.Chunk
		ccCtData []byte
	}
	snaps := make([]genSnap, 0, len(gens))
	for _, gen := range gens {
		ccCt := gen.FindFirst(rifx.IDCcCt)
		if ccCt == nil || len(ccCt.Data) < 4 {
			undoOverridesCreate()
			return nil, fmt.Errorf("AddEssentialProperty: comp %q EG %s has no CcCt count", comp.Name, gen.FormType)
		}
		snaps = append(snaps, genSnap{
			children: append([]*rifx.Chunk(nil), gen.Children...),
			ccCt:     ccCt,
			ccCtData: append([]byte(nil), ccCt.Data...),
		})
	}
	oldOvg2Children := append([]*rifx.Chunk(nil), ovg2.Children...)
	oldCprcData := append([]byte(nil), cprc.Data...)
	oldOverridesChildren := append([]*rifx.Chunk(nil), overrides.Children...)
	oldControllers := append([]*EssentialGraphicsController(nil), comp.EssentialGraphicsControllers...)
	oldWarningsLen := warningsLen(layer)

	rollback := func() {
		for i, gen := range gens {
			gen.Children = snaps[i].children
			snaps[i].ccCt.Data = snaps[i].ccCtData
		}
		ovg2.Children = oldOvg2Children
		cprc.Data = oldCprcData
		overrides.Children = oldOverridesChildren
		comp.EssentialGraphicsControllers = oldControllers
		rollbackWarnings(layer, oldWarningsLen)
		undoOverridesCreate()
	}

	// 1. Comp Item: append one CCtl per generation (each gets its own clone —
	//    the three generations must stay independent trees) + bump CcCt (BE).
	for i, gen := range gens {
		cctl := buildEGCCtl(displayName, egPanelLocale(gen), uuid, ctyp, valueChunks, comp.ID, layer.ID, pathJSON)
		gen.Children = append(gen.Children, deepCloneChunk(cctl))
		binary.BigEndian.PutUint32(snaps[i].ccCt.Data[0:4], binary.BigEndian.Uint32(snaps[i].ccCt.Data[0:4])+1)
	}

	// 2. Layer OvG2: bump CprC (LE) + append the uuid slot.
	binary.LittleEndian.PutUint32(cprc.Data[0:4], binary.LittleEndian.Uint32(cprc.Data[0:4])+1)
	ovg2.Children = append(ovg2.Children, &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDCPrp, Children: []*rifx.Chunk{
		leafChunk(rifx.IDUtf8, []byte(uuid)),
	}})

	// 3. Override stream: insert (tdmn, tdbs) before the Group End sentinel.
	insertIdx := len(overrides.Children)
	for i, ch := range overrides.Children {
		if ch.ID == rifx.IDTdmn && trimNUL(ch.Data) == "ADBE Group End" {
			insertIdx = i
			break
		}
	}
	spliced := make([]*rifx.Chunk, 0, len(overrides.Children)+2)
	spliced = append(spliced, overrides.Children[:insertIdx]...)
	spliced = append(spliced, makeTdmn(paramMatchName), overrideTdbs)
	spliced = append(spliced, overrides.Children[insertIdx:]...)
	overrides.Children = spliced

	// Closed loop: re-decode the panel from the mutated chunks; the new
	// controller must round out the list with our uuid.
	_, ctrls := parseEssentialGraphics(cb.itemList)
	if len(ctrls) != len(oldControllers)+1 || ctrls[len(ctrls)-1].UUID != uuid {
		rollback()
		return nil, fmt.Errorf("AddEssentialProperty: panel re-decode mismatch (got %d controllers, want %d)", len(ctrls), len(oldControllers)+1)
	}
	if newWarn := newWarningsSince(layer, oldWarningsLen); len(newWarn) > 0 {
		rollback()
		return nil, fmt.Errorf("AddEssentialProperty: produced %d parser warning(s), rolled back: %v", len(newWarn), newWarn)
	}

	comp.EssentialGraphicsControllers = append(comp.EssentialGraphicsControllers, ctrls[len(ctrls)-1])
	return ctrls[len(ctrls)-1], nil
}
