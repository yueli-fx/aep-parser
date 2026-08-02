package serializer

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

func TestPropertyTreePreservesUndecodedLeaf(t *testing.T) {
	tdbs := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDTdbs,
		Children: []*rifx.Chunk{{ID: rifx.ChunkID{'x', 'F', 'u', 't'}, Data: []byte{0xde, 0xad, 0xbe, 0xef}}},
	}
	tdgp := &rifx.Chunk{
		ID:       rifx.IDList,
		FormType: rifx.IDTdgp,
		Children: []*rifx.Chunk{makeTdmn("ADBE Future Property"), tdbs, makeTdmn("ADBE Group End")},
	}
	layr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDLayr, Children: []*rifx.Chunk{tdgp}}
	var before bytes.Buffer
	if err := layr.Write(&before); err != nil {
		t.Fatalf("write before parse: %v", err)
	}

	tree := buildAEPropertyGroupTree(layr)
	wirePropertyTreeLeaves(tree, nil)
	if tree.NumProperties() != 1 {
		t.Fatalf("tree children = %d, want undecoded leaf preserved", tree.NumProperties())
	}
	if observed, preserved := tree.ChildIntegrity(); observed != 1 || preserved != 1 {
		t.Fatalf("tree child integrity = %d/%d, want 1/1", observed, preserved)
	}
	opaque, ok := tree.ChildByIndex(0).(*AEOpaqueProperty)
	if !ok {
		t.Fatalf("undecoded leaf type = %T, want *AEOpaqueProperty", tree.ChildByIndex(0))
	}
	if opaque.PropertyMatchName() != "ADBE Future Property" || opaque.ValuePropertyType() != PVTUnknown {
		t.Fatalf("opaque leaf = %+v", opaque)
	}
	var after bytes.Buffer
	if err := layr.Write(&after); err != nil {
		t.Fatalf("write after parse: %v", err)
	}
	if !bytes.Equal(before.Bytes(), after.Bytes()) {
		t.Fatal("opaque property parsing changed its preserved RIFX bytes")
	}
}

func TestPropertyTreePreservesOrphanNamedEntry(t *testing.T) {
	tdgp := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDTdgp,
		Children: []*rifx.Chunk{makeTdmn("ADBE Orphan Property")},
	}
	layr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDLayr, Children: []*rifx.Chunk{tdgp}}
	tree := buildAEPropertyGroupTree(layr)
	if tree.NumProperties() != 1 {
		t.Fatalf("tree children = %d, want orphan entry preserved", tree.NumProperties())
	}
	opaque, ok := tree.ChildByIndex(0).(*AEOpaqueProperty)
	if !ok || opaque.DecodeEvidence().DecodeStatus != "failed" {
		t.Fatalf("orphan entry = %#v, want failed opaque property", tree.ChildByIndex(0))
	}
	if observed, preserved := tree.ChildIntegrity(); observed != 1 || preserved != 1 {
		t.Fatalf("tree child integrity = %d/%d, want 1/1", observed, preserved)
	}
}

func TestUnknownTdgpGroupStaysUnknown(t *testing.T) {
	inner := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDTdgp,
		Children: []*rifx.Chunk{makeTdmn("ADBE Group End")},
	}
	outer := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDTdgp,
		Children: []*rifx.Chunk{makeTdmn("ADBE Future Container"), inner, makeTdmn("ADBE Group End")},
	}
	layr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDLayr, Children: []*rifx.Chunk{outer}}
	tree := buildAEPropertyGroupTree(layr)
	group, ok := tree.ChildByIndex(0).(*AEPropertyGroup)
	if !ok {
		t.Fatalf("future container = %T, want *AEPropertyGroup", tree.ChildByIndex(0))
	}
	if got := group.PropertyType(); got != AEPropertyTypeUnknown {
		t.Fatalf("future container type = %v, want UNKNOWN", got)
	}
}

func TestWrappedGroupKeepsNestedIntegrityEvidence(t *testing.T) {
	value := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdbs}
	inner := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDTdgp,
		Children: []*rifx.Chunk{makeTdmn("ADBE Future Leaf"), value, makeTdmn("ADBE Group End")},
	}
	wrapper := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.ChunkID{'x', 'w', 'r', 'p'}, Children: []*rifx.Chunk{inner},
	}
	outer := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDTdgp,
		Children: []*rifx.Chunk{makeTdmn("ADBE Future Wrapper"), wrapper, makeTdmn("ADBE Group End")},
	}
	layr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDLayr, Children: []*rifx.Chunk{outer}}
	tree := buildAEPropertyGroupTree(layr)
	group, ok := tree.ChildByIndex(0).(*AEPropertyGroup)
	if !ok {
		t.Fatalf("wrapped entry = %T, want *AEPropertyGroup", tree.ChildByIndex(0))
	}
	if observed, preserved := group.ChildIntegrity(); observed != 1 || preserved != 1 {
		t.Fatalf("wrapped child integrity = %d/%d, want 1/1", observed, preserved)
	}
}

func TestMutationCapabilitiesRejectShortBackingChunks(t *testing.T) {
	tdb4 := make([]byte, 4)
	tdb4[3] = 2
	tdbs := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDTdbs,
		Children: []*rifx.Chunk{
			{ID: rifx.IDtdb4, Data: tdb4},
			{ID: rifx.IDCdat, Data: make([]byte, 8)},
		},
	}
	property := parseLeafProperty("ADBE Short TwoD", tdbs, &parseCtx{})
	if property == nil {
		t.Fatal("parseLeafProperty returned nil")
	}
	capabilities := property.MutationCapabilities()
	if !capabilities.Known || capabilities.StaticValue || capabilities.Expression {
		t.Fatalf("short static backing capabilities = %+v", capabilities)
	}
	if evidence := property.DecodeEvidence(); evidence.DecodeStatus != "partially-decoded" {
		t.Fatalf("short static decode evidence = %+v", evidence)
	}

	lhd3 := make([]byte, 20)
	binary.BigEndian.PutUint32(lhd3[0x08:0x0c], 1)
	binary.BigEndian.PutUint32(lhd3[0x10:0x14], 8)
	backing := &propertyBackrefs{
		lhd3: &rifx.Chunk{ID: rifx.IDLhd3, Data: lhd3},
		ldat: &rifx.Chunk{ID: rifx.IDLdat, Data: make([]byte, 8)}, bytesPerKF: 8,
	}
	capabilities = backing.MutationCapabilities(2, 1)
	if capabilities.KeyframeValues || capabilities.KeyframeStructure {
		t.Fatalf("short keyframe backing capabilities = %+v", capabilities)
	}
}

func TestShortOrientationCdatIsPartialAndNotWritable(t *testing.T) {
	tdb4 := make([]byte, 124)
	tdb4[3] = 1
	cdat := make([]byte, 8)
	binary.LittleEndian.PutUint64(cdat, math.Float64bits(12.5))
	tdbs := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDTdbs,
		Children: []*rifx.Chunk{
			{ID: rifx.IDtdb4, Data: tdb4},
			{ID: rifx.IDCdat, Data: cdat},
		},
	}
	otst := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDOtst, Children: []*rifx.Chunk{tdbs}}
	property := parseOrientationProperty("ADBE Orientation", otst, &parseCtx{})
	if property == nil {
		t.Fatal("parseOrientationProperty returned nil")
	}
	values, ok := property.StaticValue.([]float64)
	if !ok || len(values) != 1 || values[0] != 12.5 {
		t.Fatalf("partial orientation value = %#v", property.StaticValue)
	}
	if evidence := property.DecodeEvidence(); evidence.DecodeStatus != "partially-decoded" {
		t.Fatalf("orientation decode evidence = %+v", evidence)
	}
	if capabilities := property.MutationCapabilities(); capabilities.StaticValue {
		t.Fatalf("orientation capabilities = %+v", capabilities)
	}
}

func TestAnimatedOrientationDoesNotClaimGenericKeyframeWrites(t *testing.T) {
	tdb4 := make([]byte, 124)
	tdb4[3] = 1
	lhd3 := make([]byte, 20)
	binary.BigEndian.PutUint32(lhd3[0x08:0x0c], 1)
	binary.BigEndian.PutUint32(lhd3[0x10:0x14], 48)
	kfl := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDkfl,
		Children: []*rifx.Chunk{
			{ID: rifx.IDLhd3, Data: lhd3},
			{ID: rifx.IDLdat, Data: make([]byte, 48)},
		},
	}
	tdbs := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDTdbs,
		Children: []*rifx.Chunk{{ID: rifx.IDtdb4, Data: tdb4}, kfl},
	}
	otky := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDOtky,
		Children: []*rifx.Chunk{{ID: rifx.IDOtda, Data: make([]byte, 24)}},
	}
	otst := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDOtst, Children: []*rifx.Chunk{tdbs, otky}}
	property := parseOrientationProperty("ADBE Orientation", otst, &parseCtx{tickRate: 1})
	if property == nil || len(property.Keyframes) != 1 {
		t.Fatalf("animated orientation = %#v", property)
	}
	capabilities := property.MutationCapabilities()
	if capabilities.KeyframeValues || capabilities.KeyframeStructure {
		t.Fatalf("animated orientation capabilities = %+v", capabilities)
	}
}

func TestShortKeyframeEaseMarksPropertyPartiallyDecoded(t *testing.T) {
	property := &Property{MatchName: "ADBE Opacity", Components: 1}
	scene.SetPropertyBack(property, &propertyBackrefs{decodeStatus: "decoded"})
	lhd3 := make([]byte, 20)
	binary.BigEndian.PutUint32(lhd3[0x08:0x0c], 1)
	binary.BigEndian.PutUint32(lhd3[0x10:0x14], 16)
	parseKeyframes(property,
		&rifx.Chunk{ID: rifx.IDLhd3, Data: lhd3},
		&rifx.Chunk{ID: rifx.IDLdat, Data: make([]byte, 16)},
		&parseCtx{tickRate: 1},
	)
	evidence := property.DecodeEvidence()
	if evidence.DecodeStatus != "partially-decoded" || evidence.TemporalEaseStatus != "invalid-preserved" || len(property.Keyframes) != 1 {
		t.Fatalf("short ease property = %+v", property)
	}
}

func TestPropertyTreeUsesDecodedInstanceName(t *testing.T) {
	layout := valueLayout{dim: 1}
	tdbs := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDTdbs,
		Children: []*rifx.Chunk{
			makeTdsn("实例位置"),
			makeTdb4(layout, 30720),
			makeCdat(make([]byte, 8), layout),
		},
	}
	tdgp := &rifx.Chunk{
		ID: rifx.IDList, FormType: rifx.IDTdgp,
		Children: []*rifx.Chunk{makeTdmn("ADBE Position"), tdbs, makeTdmn("ADBE Group End")},
	}
	layr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDLayr, Children: []*rifx.Chunk{tdgp}}
	property := parseLeafProperty("ADBE Position", tdbs, &parseCtx{})
	if property == nil {
		t.Fatal("parseLeafProperty returned nil")
	}
	tree := buildAEPropertyGroupTree(layr)
	wirePropertyTreeLeaves(tree, []*Property{property})
	linked, ok := tree.ChildByIndex(0).(*Property)
	if !ok {
		t.Fatalf("tree leaf = %T, want *Property", tree.ChildByIndex(0))
	}
	if linked != property || linked.Name != "实例位置" || linked.NameSource != "decoded" || linked.MatchName != "ADBE Position" {
		t.Fatalf("decoded identity = %+v", linked)
	}
	capabilities := linked.MutationCapabilities()
	if !capabilities.Known || !capabilities.StaticValue || !capabilities.Expression || capabilities.KeyframeStructure {
		t.Fatalf("mutation capabilities = %+v", capabilities)
	}
}
