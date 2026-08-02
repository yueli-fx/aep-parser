package aep_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/yueli-fx/aep-parser"
)

func TestProjectJSONIsDetachedVersionedAndPathSafe(t *testing.T) {
	path := filepath.Join("test_data", "fixtures", "v2_2_transform_kf_re.aep")
	document, err := aep.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	payload, err := document.ProjectJSON()
	if err != nil {
		t.Fatalf("ProjectJSON: %v", err)
	}
	var snapshot struct {
		SchemaVersion     int `json:"schema_version"`
		PropertyIntegrity struct {
			ObservedTreeEntryCount  int `json:"observed_tree_entry_count"`
			PreservedTreeEntryCount int `json:"preserved_tree_entry_count"`
			TreeNodeCount           int `json:"tree_node_count"`
			GroupCount              int `json:"group_count"`
			PropertyCount           int `json:"property_count"`
			LinkedPropertyCount     int `json:"linked_property_count"`
			UnlinkedPropertyCount   int `json:"unlinked_property_count"`
			DuplicateLinkCount      int `json:"duplicate_property_link_count"`
			PreservedUnknownCount   int `json:"preserved_unknown_count"`
			DroppedUnknownCount     int `json:"dropped_unknown_count"`
		} `json:"property_integrity"`
		PropertyDiagnostics []struct {
			Code        string `json:"code"`
			PropertyRef string `json:"property_ref"`
			MatchName   string `json:"match_name"`
			Message     string `json:"message"`
		} `json:"property_diagnostics"`
		Project struct {
			Compositions []struct {
				Layers []struct {
					Properties []struct {
						PropertyRef string `json:"property_ref"`
					} `json:"properties"`
					Effects []struct {
						Parameters []struct {
							PropertyRef string `json:"property_ref"`
						} `json:"parameters"`
					} `json:"effects"`
					Masks []struct {
						Properties []struct {
							PropertyRef string `json:"property_ref"`
						} `json:"properties"`
					} `json:"masks"`
				} `json:"layers"`
			} `json:"compositions"`
			Footage []struct {
				Path string `json:"path"`
			} `json:"footage"`
			RenderQueue json.RawMessage `json:"render_queue"`
		} `json:"project"`
		PropertyTrees []struct {
			Children []projectJSONNode `json:"children"`
		} `json:"property_trees"`
		PropertyRecords []projectJSONRecord `json:"property_records"`
	}
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		t.Fatalf("decode ProjectJSON: %v", err)
	}
	if snapshot.SchemaVersion != aep.ProjectJSONSchemaVersion {
		t.Fatalf("schema version = %d", snapshot.SchemaVersion)
	}
	if len(snapshot.Project.Compositions) != 1 ||
		len(snapshot.Project.Compositions[0].Layers) != 1 ||
		len(snapshot.Project.Compositions[0].Layers[0].Properties) == 0 ||
		len(snapshot.PropertyTrees) != 1 ||
		len(snapshot.PropertyTrees[0].Children) == 0 {
		t.Fatalf("unexpected detached project summary: %+v", snapshot)
	}
	if snapshot.SchemaVersion != 2 {
		t.Fatalf("schema version = %d, want v2", snapshot.SchemaVersion)
	}
	if snapshot.PropertyIntegrity.TreeNodeCount == 0 ||
		snapshot.PropertyIntegrity.ObservedTreeEntryCount != snapshot.PropertyIntegrity.TreeNodeCount ||
		snapshot.PropertyIntegrity.PreservedTreeEntryCount != snapshot.PropertyIntegrity.TreeNodeCount ||
		snapshot.PropertyIntegrity.PropertyCount == 0 ||
		snapshot.PropertyIntegrity.LinkedPropertyCount == 0 ||
		snapshot.PropertyIntegrity.UnlinkedPropertyCount != 0 ||
		snapshot.PropertyIntegrity.DuplicateLinkCount != 0 ||
		snapshot.PropertyIntegrity.DroppedUnknownCount != 0 {
		t.Fatalf("unexpected property integrity: %+v diagnostics=%+v", snapshot.PropertyIntegrity, snapshot.PropertyDiagnostics)
	}
	flatRefs := map[string]bool{}
	for _, record := range snapshot.PropertyRecords {
		if record.PropertyRef == "" || record.PropertyType == "" || record.PropertyValueType == "" || record.ElidedStatus == "" || record.NameSource == "" || record.DecodeStatus == "" || !record.RawPreserved || record.WriteCapability == "" || record.TemporalEaseStatus == "" || record.OriginEvidence == nil {
			t.Fatalf("property record missing v2 facts: %+v", record)
		}
		if record.PropertyType == "PROPERTY" && record.KindDecoded() && (record.CanVaryOverTime == nil || record.IsSpatial == nil || record.Dimensions < 1) {
			t.Fatalf("decoded property record missing animation facts: %+v", record)
		}
		for _, keyframe := range record.Keyframes {
			if keyframe.TemporalEaseUnit != "ratio" || keyframe.TemporalEaseStatus == "" {
				t.Fatalf("record keyframe missing ease contract: %+v", keyframe)
			}
		}
		flatRefs[record.PropertyRef] = true
	}
	for _, composition := range snapshot.Project.Compositions {
		for _, layer := range composition.Layers {
			for _, property := range layer.Properties {
				if property.PropertyRef == "" {
					t.Fatal("flat property missing property_ref")
				}
				if !flatRefs[property.PropertyRef] {
					t.Fatalf("legacy flat property ref %q absent from property_records", property.PropertyRef)
				}
			}
			for _, effect := range layer.Effects {
				for _, property := range effect.Parameters {
					if property.PropertyRef == "" {
						t.Fatal("effect parameter missing property_ref")
					}
					if !flatRefs[property.PropertyRef] {
						t.Fatalf("effect property ref %q absent from property_records", property.PropertyRef)
					}
				}
			}
			for _, mask := range layer.Masks {
				for _, property := range mask.Properties {
					if property.PropertyRef == "" {
						t.Fatal("mask property missing property_ref")
					}
					if !flatRefs[property.PropertyRef] {
						t.Fatalf("mask property ref %q absent from property_records", property.PropertyRef)
					}
				}
			}
		}
	}
	var leafCount int
	for _, tree := range snapshot.PropertyTrees {
		walkProjectJSONNodes(t, tree.Children, flatRefs, &leafCount)
	}
	if leafCount == 0 {
		t.Fatal("ProjectJSON v2 tree contained no property leaves")
	}
	for _, footage := range snapshot.Project.Footage {
		if footage.Path != "" {
			t.Fatalf("footage path leaked: %q", footage.Path)
		}
	}
	if len(snapshot.Project.RenderQueue) != 0 &&
		string(snapshot.Project.RenderQueue) != "null" {
		t.Fatalf("render queue leaked: %s", snapshot.Project.RenderQueue)
	}
	if strings.Contains(string(payload), filepath.Clean(path)) {
		t.Fatal("input path leaked into ProjectJSON")
	}

	var nilDocument *aep.Document
	if _, err := nilDocument.ProjectJSON(); err == nil {
		t.Fatal("ProjectJSON succeeded on nil document")
	}
}

func TestProjectJSONV2LinksEffectAndMaskParameters(t *testing.T) {
	document, err := aep.Open(filepath.Join("test_data", "fixtures", "re_batch.aep"))
	if err != nil {
		t.Skipf("re_batch.aep not present: %v", err)
	}
	payload, err := document.ProjectJSON()
	if err != nil {
		t.Fatalf("ProjectJSON: %v", err)
	}
	var snapshot struct {
		PropertyIntegrity struct {
			Linked   int `json:"linked_property_count"`
			Unlinked int `json:"unlinked_property_count"`
			Dropped  int `json:"dropped_unknown_count"`
		} `json:"property_integrity"`
		Diagnostics []struct {
			Code        string `json:"code"`
			MatchName   string `json:"match_name"`
			PropertyRef string `json:"property_ref"`
		} `json:"property_diagnostics"`
		Project struct {
			Compositions []struct {
				Layers []struct {
					Effects []struct {
						Parameters []struct {
							PropertyRef string `json:"property_ref"`
						} `json:"parameters"`
					} `json:"effects"`
					Masks []struct {
						Properties []struct {
							PropertyRef string `json:"property_ref"`
						} `json:"properties"`
					} `json:"masks"`
				} `json:"layers"`
			} `json:"compositions"`
		} `json:"project"`
	}
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.PropertyIntegrity.Linked == 0 || snapshot.PropertyIntegrity.Unlinked != 0 || snapshot.PropertyIntegrity.Dropped != 0 {
		t.Fatalf("effect project integrity = %+v diagnostics=%+v", snapshot.PropertyIntegrity, snapshot.Diagnostics)
	}
	found := false
	foundMaskProperty := false
	for _, composition := range snapshot.Project.Compositions {
		for _, layer := range composition.Layers {
			for _, effect := range layer.Effects {
				for _, parameter := range effect.Parameters {
					found = true
					if parameter.PropertyRef == "" {
						t.Fatal("effect parameter missing property_ref")
					}
				}
			}
			for _, mask := range layer.Masks {
				for _, property := range mask.Properties {
					foundMaskProperty = true
					if property.PropertyRef == "" {
						t.Fatal("mask property missing property_ref")
					}
				}
			}
		}
	}
	if !found {
		t.Fatal("fixture exposed no effect parameters")
	}
	if !foundMaskProperty {
		t.Fatal("fixture exposed no mask properties")
	}
}

type projectJSONNode struct {
	Kind           string                     `json:"kind"`
	PropertyType   string                     `json:"property_type"`
	PropertyRef    string                     `json:"property_ref"`
	ElidedStatus   string                     `json:"elided_status"`
	NameSource     string                     `json:"name_source"`
	CanonicalPath  string                     `json:"canonical_path"`
	OriginEvidence *projectJSONOriginEvidence `json:"origin_evidence"`
	Children       []projectJSONNode          `json:"children"`
}

type projectJSONOriginEvidence struct {
	EvidenceKind       string `json:"evidence_kind"`
	SourceOrdinal      int    `json:"source_ordinal"`
	CanonicalPath      string `json:"canonical_path"`
	PreservationStatus string `json:"preservation_status"`
}

type projectJSONRecord struct {
	PropertyType       string                     `json:"property_type"`
	PropertyValueType  string                     `json:"property_value_type"`
	PropertyRef        string                     `json:"property_ref"`
	ElidedStatus       string                     `json:"elided_status"`
	NameSource         string                     `json:"name_source"`
	DecodeStatus       string                     `json:"decode_status"`
	TemporalEaseStatus string                     `json:"temporal_ease_status"`
	RawPreserved       bool                       `json:"raw_preserved"`
	WriteCapability    string                     `json:"write_capability"`
	CanVaryOverTime    *bool                      `json:"can_vary_over_time"`
	IsSpatial          *bool                      `json:"is_spatial"`
	Dimensions         int                        `json:"dimensions"`
	OriginEvidence     *projectJSONOriginEvidence `json:"origin_evidence"`
	Keyframes          []struct {
		TemporalEaseUnit   string `json:"temporal_ease_unit"`
		TemporalEaseStatus string `json:"temporal_ease_status"`
	} `json:"keyframes"`
}

func (r projectJSONRecord) KindDecoded() bool { return r.DecodeStatus == "decoded" }

func walkProjectJSONNodes(t *testing.T, nodes []projectJSONNode, flatRefs map[string]bool, leafCount *int) {
	t.Helper()
	for _, node := range nodes {
		if node.PropertyType == "" || node.ElidedStatus == "" || node.NameSource == "" || node.CanonicalPath == "" || node.OriginEvidence == nil || node.OriginEvidence.PreservationStatus != "preserved" {
			t.Errorf("tree node missing v2 identity facts: %+v", node)
		}
		switch node.PropertyType {
		case "PROPERTY":
			*leafCount++
			if node.PropertyRef == "" {
				t.Errorf("property leaf missing property_ref: %+v", node)
			}
			if !flatRefs[node.PropertyRef] {
				t.Errorf("tree leaf ref %q absent from property_records", node.PropertyRef)
			}
		case "NAMED_GROUP", "INDEXED_GROUP", "UNKNOWN":
			if node.PropertyRef != "" && !flatRefs[node.PropertyRef] {
				t.Errorf("opaque tree ref %q absent from property_records", node.PropertyRef)
			}
		default:
			t.Errorf("unexpected property_type %q", node.PropertyType)
		}
		walkProjectJSONNodes(t, node.Children, flatRefs, leafCount)
	}
}

func TestProjectJSONV2ExternalFixtureIntegrity(t *testing.T) {
	path := os.Getenv("AEP_PROJECT_JSON_FIXTURE")
	if path == "" {
		t.Skip("set AEP_PROJECT_JSON_FIXTURE to run the large-project integrity gate")
	}
	document, err := aep.Open(path)
	if err != nil {
		t.Fatalf("Open(%q): %v", path, err)
	}
	payload, err := document.ProjectJSON()
	if err != nil {
		t.Fatalf("ProjectJSON(%q): %v", path, err)
	}
	var snapshot struct {
		SchemaVersion int `json:"schema_version"`
		Integrity     struct {
			Observed       int `json:"observed_tree_entry_count"`
			Preserved      int `json:"preserved_tree_entry_count"`
			TreeNodes      int `json:"tree_node_count"`
			FlatProperties int `json:"flat_property_count"`
			Linked         int `json:"linked_property_count"`
			Unlinked       int `json:"unlinked_property_count"`
			Duplicate      int `json:"duplicate_property_link_count"`
			Opaque         int `json:"preserved_opaque_count"`
			Unknown        int `json:"preserved_unknown_count"`
			DroppedUnknown int `json:"dropped_unknown_count"`
		} `json:"property_integrity"`
		Diagnostics []json.RawMessage `json:"property_diagnostics"`
	}
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		t.Fatalf("decode ProjectJSON: %v", err)
	}
	if snapshot.SchemaVersion != aep.ProjectJSONSchemaVersion || snapshot.Integrity.TreeNodes == 0 || snapshot.Integrity.FlatProperties == 0 || snapshot.Integrity.Observed != snapshot.Integrity.TreeNodes || snapshot.Integrity.Preserved != snapshot.Integrity.TreeNodes {
		t.Fatalf("invalid v2 integrity envelope: %+v", snapshot.Integrity)
	}
	if snapshot.Integrity.Linked+snapshot.Integrity.Unlinked != snapshot.Integrity.FlatProperties {
		t.Fatalf("flat property accounting mismatch: %+v", snapshot.Integrity)
	}
	if snapshot.Integrity.DroppedUnknown != 0 || snapshot.Integrity.Duplicate != 0 {
		t.Fatalf("property-tree preservation gate failed: %+v", snapshot.Integrity)
	}
	if len(snapshot.Diagnostics) < snapshot.Integrity.Unlinked {
		t.Fatalf("unlinked properties lack diagnostics: integrity=%+v diagnostics=%d", snapshot.Integrity, len(snapshot.Diagnostics))
	}
	t.Logf("ProjectJSON v2 %s: bytes=%d integrity=%+v diagnostics=%d", filepath.Base(path), len(payload), snapshot.Integrity, len(snapshot.Diagnostics))
}

func TestClassifyErrorReturnsStableCategories(t *testing.T) {
	path := filepath.Join("test_data", "fixtures", "v2_smoke_ae2020.aep")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	limits := aep.DefaultLimits()
	limits.MaxInputBytes = uint64(len(data) - 1)
	_, parseErr := aep.ParseWithLimits(bytes.NewReader(data), limits)
	info, ok := aep.ClassifyError(parseErr)
	if !ok || info.Code != "resource-limit" ||
		info.Resource != "input bytes" ||
		info.Actual != uint64(len(data)) {
		t.Fatalf("limit classification = %+v, %v", info, ok)
	}

	_, parseErr = aep.Parse(bytes.NewReader(nil))
	info, ok = aep.ClassifyError(parseErr)
	if !ok || info.Code != "invalid-format" {
		t.Fatalf("format classification = %+v, %v (%v)", info, ok, parseErr)
	}
	if _, ok := aep.ClassifyError(nil); ok {
		t.Fatal("nil error classified")
	}
}
