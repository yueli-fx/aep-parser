package aep

import (
	"encoding/json"
	"math"
	"testing"

	internal "github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

func TestTemporalEaseStatusPreservesInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		interp internal.InterpType
		eases  []internal.TemporalEase
		want   string
	}{
		{name: "valid-bezier", interp: internal.InterpBezier, eases: []internal.TemporalEase{{Influence: 0.333}}, want: "valid"},
		{name: "linear-zero", interp: internal.InterpLinear, eases: []internal.TemporalEase{{Influence: 0}}, want: "not-applicable-preserved"},
		{name: "out-of-range", interp: internal.InterpBezier, eases: []internal.TemporalEase{{Influence: 1.5}}, want: "invalid-preserved"},
		{name: "nan", interp: internal.InterpBezier, eases: []internal.TemporalEase{{Influence: math.NaN()}}, want: "invalid-preserved"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := temporalEaseStatus(test.interp, test.eases, 1); got != test.want {
				t.Fatalf("status = %q, want %q", got, test.want)
			}
		})
	}
}

func TestPropertyIntegrityReportsLayerWithoutTree(t *testing.T) {
	property := &internal.Property{MatchName: "ADBE Position", Name: "ADBE Position", NameSource: "fallback", Components: 2, StaticValue: []float64{1, 2}}
	layer := &internal.Layer{ID: 20, Properties: []*internal.Property{property}}
	project := &internal.Project{Compositions: []*internal.Composition{{ID: 10, Layers: []*internal.Layer{layer}}}}
	registry := newProjectPropertyRegistry(project, project.ToJSON())
	trees, integrity, diagnostics := projectPropertyTrees(project, registry)
	if len(trees) != 1 || len(trees[0].Children) != 0 {
		t.Fatalf("missing-tree layer snapshot = %+v", trees)
	}
	if integrity.UnlinkedPropertyCount != 1 || len(diagnostics) != 2 {
		t.Fatalf("integrity=%+v diagnostics=%+v", integrity, diagnostics)
	}
	seenMissing, seenUnlinked := false, false
	for _, diagnostic := range diagnostics {
		seenMissing = seenMissing || diagnostic.Code == "missing-property-tree"
		seenUnlinked = seenUnlinked || diagnostic.Code == "unlinked-property"
	}
	if !seenMissing || !seenUnlinked {
		t.Fatalf("missing diagnostics: %+v", diagnostics)
	}
}

func TestOpaqueTreeLeafHasFlatPropertyRecord(t *testing.T) {
	opaque := &internal.AEOpaqueProperty{
		MatchName: "ADBE Future Property", Name: "ADBE Future Property", NameSource: "fallback",
		SemanticType: internal.AEPropertyTypeProperty, ValueType: internal.PVTUnknown,
	}
	root := &internal.AEPropertyGroup{SemanticType: internal.AEPropertyTypeNamedGroup, Children: []internal.PropertyBase{opaque}}
	layer := &internal.Layer{ID: 20}
	scene.SetLayerPropertyTree(layer, root)
	project := &internal.Project{Compositions: []*internal.Composition{{ID: 10, Layers: []*internal.Layer{layer}}}}
	registry := newProjectPropertyRegistry(project, project.ToJSON())
	_, integrity, diagnostics := projectPropertyTrees(project, registry)
	if len(registry.records) != 1 || registry.records[0].PropertyRef == "" {
		t.Fatalf("opaque records = %+v", registry.records)
	}
	if integrity.FlatPropertyCount != 1 || integrity.LinkedPropertyCount != 1 || integrity.UnlinkedPropertyCount != 0 || len(diagnostics) != 0 {
		t.Fatalf("integrity=%+v diagnostics=%+v", integrity, diagnostics)
	}
}

func TestTemporalEaseNonFiniteValuesRemainJSONRepresentable(t *testing.T) {
	ease := projectTemporalEaseFrom(internal.TemporalEase{Speed: math.Inf(1), Influence: math.NaN()})
	payload, err := json.Marshal(ease)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if got := string(payload); got != `{"speed":null,"influence":null,"speed_raw":"+Infinity","influence_raw":"NaN"}` {
		t.Fatalf("encoded ease = %s", got)
	}
	canonical, err := json.Marshal(internal.JSONTemporalEase{Speed: math.Inf(1), Influence: math.NaN()})
	if err != nil {
		t.Fatalf("json.Marshal canonical ease: %v", err)
	}
	if string(canonical) != string(payload) {
		t.Fatalf("canonical ease = %s, supplemental ease = %s", canonical, payload)
	}
}

func TestPropertyValueTypeNamesCoverAEKinds(t *testing.T) {
	tests := map[internal.PropertyValueType]string{
		internal.PVTNoValue: "NO_VALUE", internal.PVTOneD: "OneD", internal.PVTTwoD: "TwoD",
		internal.PVTTwoDSpatial: "TwoD_SPATIAL", internal.PVTThreeD: "ThreeD",
		internal.PVTThreeDSpatial: "ThreeD_SPATIAL", internal.PVTColor: "COLOR",
		internal.PVTCustomValue: "CUSTOM_VALUE", internal.PVTMarker: "MARKER",
		internal.PVTLayerIndex: "LAYER_INDEX", internal.PVTMaskIndex: "MASK_INDEX",
		internal.PVTShape: "SHAPE", internal.PVTTextDocument: "TEXT_DOCUMENT", internal.PVTUnknown: "UNKNOWN",
	}
	for valueType, want := range tests {
		if got := propertyValueTypeName(valueType); got != want {
			t.Errorf("propertyValueTypeName(%v) = %q, want %q", valueType, got, want)
		}
	}
}

func TestDeclaredEffectControlTypesDriveIndexValueTypes(t *testing.T) {
	tests := []struct {
		control internal.PropertyControlType
		want    internal.PropertyValueType
	}{
		{control: internal.PCTLLayer, want: internal.PVTLayerIndex},
		{control: internal.PCTLMask, want: internal.PVTMaskIndex},
		{control: internal.PCTLCurve, want: internal.PVTCustomValue},
	}
	for _, test := range tests {
		property := &internal.Property{DeclaredControlType: test.control, HasDeclaredControlType: true, Components: 1}
		if got := property.ValuePropertyType(); got != test.want {
			t.Errorf("control %v value type = %v, want %v", test.control, got, test.want)
		}
		if got := property.ControlType(); got != test.control {
			t.Errorf("control type = %v, want %v", got, test.control)
		}
	}
}

func TestWriteCapabilitiesDoNotGuessWithoutWriterBacking(t *testing.T) {
	property := &internal.Property{MatchName: "ADBE Position", Components: 2, StaticValue: []float64{1, 2}}
	overall, operations := propertyWriteCapabilities(property, internal.PVTTwoD)
	if overall != "unknown" || operations.StaticValue != "unknown" || operations.Expression != "unknown" {
		t.Fatalf("write capabilities = %q %+v", overall, operations)
	}
}

func TestPropertyFactsFallbackWithoutDecodeEvidence(t *testing.T) {
	property := &internal.Property{
		MatchName: "ADBE Opacity", Components: 1, StaticValue: float64(100),
	}
	facts := propertyFacts(property, "property-1")
	if facts.DecodeStatus != "partially-decoded" || facts.TemporalEaseStatus != "not-present" {
		t.Fatalf("property facts = %+v", facts)
	}
}

func TestKeyframeFactsInheritInvalidPreservedEaseEvidence(t *testing.T) {
	property := &internal.Property{MatchName: "ADBE Opacity", Components: 1}
	scene.SetPropertyBack(property, partialDecodePropertyWriter{})
	keyframe := projectKeyframeFrom(property, &internal.Keyframe{})
	if keyframe.InTemporalEaseStatus != "invalid-preserved" || keyframe.OutTemporalEaseStatus != "invalid-preserved" || keyframe.TemporalEaseStatus != "invalid-preserved" {
		t.Fatalf("keyframe ease status = %+v", keyframe)
	}
	facts := propertyFacts(property, "property-1")
	if facts.DecodeStatus != "partially-decoded" || facts.TemporalEaseStatus != "invalid-preserved" {
		t.Fatalf("property facts = %+v", facts)
	}
}

type partialDecodePropertyWriter struct{}

func (partialDecodePropertyWriter) SetStaticValue(any) error        { return nil }
func (partialDecodePropertyWriter) SetExpressionEnabled(bool) error { return nil }
func (partialDecodePropertyWriter) SetExpression(string) error      { return nil }
func (partialDecodePropertyWriter) SetLockedRatio(bool) error       { return nil }
func (partialDecodePropertyWriter) Tdb4Byte(int) (byte, bool)       { return 0, false }
func (partialDecodePropertyWriter) TdsbByte(int) (byte, bool)       { return 0, false }
func (partialDecodePropertyWriter) MinValueBytes() []byte           { return nil }
func (partialDecodePropertyWriter) MaxValueBytes() []byte           { return nil }
func (partialDecodePropertyWriter) DecodeEvidence() scene.PropertyDecodeEvidence {
	return scene.PropertyDecodeEvidence{DecodeStatus: "partially-decoded", TemporalEaseStatus: "invalid-preserved"}
}
