// Material-Options materialization (synthesis-lite) for from-scratch 3D layers.
//
// A from-scratch shape/solid layer emits an EMPTY "ADBE Material Options Group"
// (lower_layer.go emptyPropGroupFlags) — AE materializes the full material tree
// in its DOM from the empty group on open (all defaults), but on disk the leaves
// are elided, so MaterialCastsShadows()/SetMaterial* find no property. To make a
// from-scratch 3D layer CAST shadows (Casts Shadows defaults Off) we must splice
// the leaf back, exactly as SetEffectParam does for default-elided effect params
// and spliceCameraIrisLeaves does for camera DoF controls.
//
// SetMaterialOption clones the requested (tdmn, LIST:tdbs) pair from an embedded
// AE-native template (the 15-leaf material tree extracted from
// re_material_options.aep), splices it into the group in AE's canonical order,
// re-parses it, and writes the caller's value — which is non-default by intent,
// exactly the state AE itself persists (a default-valued spliced leaf is dropped
// by AE on open). RE: incidents/layer-3d-enable-bit-materializes.md.
package serializer

import (
	"bytes"
	_ "embed"
	"fmt"
	"sync"

	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

//go:embed templates/options/material_options_leaves.bin
var materialOptionsLeavesBytes []byte

// materialLeafOrder is AE's canonical child order in the Material Options group
// (= extraction order in material_options_leaves.bin). Splices preserve it;
// property order within an AE group is significant (out-of-order leaves get
// dropped on open).
var materialLeafOrder = map[string]int{
	"ADBE Casts Shadows":            0,
	"ADBE Light Transmission":       1,
	"ADBE Accepts Lights":           2,
	"ADBE Shadow Color":             3,
	"ADBE Appears in Reflections":   4,
	"ADBE Ambient Coefficient":      5,
	"ADBE Diffuse Coefficient":      6,
	"ADBE Specular Coefficient":     7,
	"ADBE Metal Coefficient":        8,
	"ADBE Reflection Coefficient":   9,
	"ADBE Glossiness Coefficient":   10,
	"ADBE Fresnel Coefficient":      11,
	"ADBE Transparency Coefficient": 12,
	"ADBE Transp Rolloff":           13,
	"ADBE Index of Refraction":      14,
}

var (
	materialLeavesWrapper *rifx.Chunk
	materialLeavesOnce    sync.Once
	materialLeavesErr     error
)

// cloneMaterialLeaf returns fresh (tdmn, tdbs) chunks for matchName from the
// embedded template, or an error if matchName is not a known material leaf.
func cloneMaterialLeaf(matchName string) (tdmn, tdbs *rifx.Chunk, err error) {
	materialLeavesOnce.Do(func() {
		materialLeavesWrapper, materialLeavesErr = rifx.ReadChunk(bytes.NewReader(materialOptionsLeavesBytes))
	})
	if materialLeavesErr != nil {
		return nil, nil, fmt.Errorf("parse material leaves template: %w", materialLeavesErr)
	}
	kids := materialLeavesWrapper.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == matchName &&
			kids[i+1].IsList() && kids[i+1].FormType == rifx.IDTdbs {
			return deepCloneChunk(kids[i]), deepCloneChunk(kids[i+1]), nil
		}
	}
	return nil, nil, fmt.Errorf("material leaf %q not in template", matchName)
}

// SetMaterialOption sets a Material-Options property's static value on a 3D
// layer, materializing the leaf first when it is default-elided (the from-scratch
// case). When the property is already present (a parsed/fixture layer) it just
// sets the value. matchName is the AE match-name, e.g. "ADBE Casts Shadows".
// (Full contract lives on the aep.SetMaterialOption facade — docgen source.)
func SetMaterialOption(layer *Layer, matchName string, value any) (*Property, error) {
	if layer == nil {
		return nil, fmt.Errorf("SetMaterialOption: layer is nil")
	}
	// Already present (parsed/fixture layer or a prior splice) — just set it.
	if p := layer.PropertyByMatchName(matchName); p != nil {
		if err := p.SetStaticValue(value); err != nil {
			return nil, err
		}
		return p, nil
	}
	myOrd, ok := materialLeafOrder[matchName]
	if !ok {
		return nil, fmt.Errorf("SetMaterialOption: %q is not a Material Options property", matchName)
	}

	tree := scene.LayerPropertyTree(layer)
	if tree == nil {
		return nil, fmt.Errorf("SetMaterialOption: layer %q has no property tree; round-trip through aep.Reopen first", layer.Name)
	}
	matGroup := tree.Group("ADBE Material Options Group")
	if matGroup == nil {
		return nil, fmt.Errorf("SetMaterialOption: layer %q has no Material Options group (not a 3D-capable layer?)", layer.Name)
	}
	pgb := propertyGroupBack(matGroup)
	if pgb == nil || pgb.chunk == nil {
		return nil, fmt.Errorf("SetMaterialOption: Material Options group has no chunk back-ref")
	}
	groupChunk := pgb.chunk

	comp := scene.LayerComp(layer)
	if comp == nil || scene.CompositionProj(comp) == nil {
		return nil, fmt.Errorf("SetMaterialOption: layer %q has no composition/project back-ref", layer.Name)
	}

	tdmnCh, tdbsCh, err := cloneMaterialLeaf(matchName)
	if err != nil {
		return nil, err
	}

	// Insertion point in canonical order: before the first existing material
	// leaf with a higher ordinal, else before the Group End sentinel.
	kids := groupChunk.Children
	insertIdx := len(kids)
	for i := 0; i < len(kids); i++ {
		if kids[i].ID != rifx.IDTdmn {
			continue
		}
		name := trimChunkNUL(kids[i].Data)
		if name == "ADBE Group End" {
			insertIdx = i
			break
		}
		if ord, ok := materialLeafOrder[name]; ok && ord > myOrd {
			insertIdx = i
			break
		}
	}

	// Snapshot for rollback.
	oldGroupChildren := append([]*rifx.Chunk(nil), kids...)
	oldProps := append([]*Property(nil), layer.Properties...)
	oldTreeChildren := append([]scene.PropertyBase(nil), matGroup.Children...)
	oldWarningsLen := warningsLen(layer)
	rollback := func() {
		groupChunk.Children = oldGroupChildren
		layer.Properties = oldProps
		matGroup.Children = oldTreeChildren
		rollbackWarnings(layer, oldWarningsLen)
	}

	spliced := make([]*rifx.Chunk, 0, len(kids)+2)
	spliced = append(spliced, kids[:insertIdx]...)
	spliced = append(spliced, tdmnCh, tdbsCh)
	spliced = append(spliced, kids[insertIdx:]...)
	groupChunk.Children = spliced

	ctx := newParseCtxFPS(comp.TickRate, comp.FrameRate, comp.Name, &scene.CompositionProj(comp).Warnings)
	prop := parseLeafProperty(matchName, tdbsCh, ctx)
	if prop == nil {
		rollback()
		return nil, fmt.Errorf("SetMaterialOption: spliced leaf %q re-parse produced no property", matchName)
	}
	if err := prop.SetStaticValue(value); err != nil {
		rollback()
		return nil, err
	}

	layer.Properties = append(layer.Properties, prop)
	matGroup.Children = append(matGroup.Children, prop)
	scene.SetPropertyParentTreeGroup(prop, matGroup)

	if newWarn := newWarningsSince(layer, oldWarningsLen); len(newWarn) > 0 {
		rollback()
		return nil, fmt.Errorf("SetMaterialOption: produced %d parser warning(s), rolled back: %v", len(newWarn), newWarn)
	}
	return prop, nil
}
