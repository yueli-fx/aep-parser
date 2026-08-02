package aep

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	internal "github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

// ProjectJSONSchemaVersion identifies the detached read-only project
// snapshot returned by Document.ProjectJSON.
const ProjectJSONSchemaVersion = 2

// ErrorInfo is the stable classification of a parser error. Message text and
// internal error types are deliberately not part of this public contract.
type ErrorInfo struct {
	Code     string `json:"code"`
	Resource string `json:"resource,omitempty"`
	Limit    uint64 `json:"limit,omitempty"`
	Actual   uint64 `json:"actual,omitempty"`
	Offset   int64  `json:"offset,omitempty"`
}

type projectJSONEnvelope struct {
	SchemaVersion       int                         `json:"schema_version"`
	Project             *internal.JSONProject       `json:"project"`
	CompositionFacts    []projectCompositionFact    `json:"composition_facts,omitempty"`
	FootageFacts        []projectFootageFact        `json:"footage_facts,omitempty"`
	PropertyTrees       []projectPropertyTree       `json:"property_trees,omitempty"`
	PropertyRecords     []projectPropertyRecord     `json:"property_records,omitempty"`
	PropertyIntegrity   projectPropertyIntegrity    `json:"property_integrity"`
	PropertyDiagnostics []projectPropertyDiagnostic `json:"property_diagnostics,omitempty"`
	Warnings            []projectWarning            `json:"warnings,omitempty"`
}

type projectCompositionFact struct {
	ID               uint32  `json:"id"`
	PixelAspect      float64 `json:"pixel_aspect"`
	DisplayStartTime float64 `json:"display_start_time"`
}

type projectFootageFact struct {
	ID            uint32     `json:"id"`
	SolidColor    [3]float64 `json:"solid_color,omitempty"`
	IsPlaceholder bool       `json:"is_placeholder,omitempty"`
}

type projectPropertyTree struct {
	CompositionID uint32                `json:"composition_id"`
	LayerID       uint32                `json:"layer_id,omitempty"`
	LayerIndex    int                   `json:"layer_index"`
	Children      []projectPropertyNode `json:"children,omitempty"`
}

type projectPropertyNode struct {
	Kind           string                         `json:"kind"`
	PropertyType   string                         `json:"property_type"`
	Elided         *bool                          `json:"elided"`
	ElidedStatus   string                         `json:"elided_status"`
	MatchName      string                         `json:"match_name,omitempty"`
	Name           string                         `json:"name,omitempty"`
	NameSource     string                         `json:"name_source"`
	Occurrence     int                            `json:"occurrence"`
	SiblingIndex   int                            `json:"sibling_index"`
	CanonicalPath  string                         `json:"canonical_path"`
	OriginEvidence *projectPropertyOriginEvidence `json:"origin_evidence,omitempty"`
	PropertyRef    string                         `json:"property_ref,omitempty"`
	FlatIndex      *int                           `json:"flat_index,omitempty"`
	Children       []projectPropertyNode          `json:"children,omitempty"`
}

type projectPropertyFacts struct {
	PropertyRef          string                   `json:"property_ref"`
	PropertyValueType    string                   `json:"property_value_type"`
	Dimensions           int                      `json:"dimensions,omitempty"`
	StorageDimensions    int                      `json:"storage_dimensions,omitempty"`
	CanVaryOverTime      *bool                    `json:"can_vary_over_time"`
	IsSpatial            *bool                    `json:"is_spatial"`
	IsSeparationLeader   *bool                    `json:"is_separation_leader"`
	DimensionsSeparated  *bool                    `json:"dimensions_separated"`
	IsSeparationFollower *bool                    `json:"is_separation_follower"`
	SeparationDimension  *int                     `json:"separation_dimension"`
	DecodeStatus         string                   `json:"decode_status"`
	RawPreserved         bool                     `json:"raw_preserved"`
	WriteCapability      string                   `json:"write_capability"`
	WriteCapabilities    projectWriteCapabilities `json:"write_capabilities"`
	StaticValue          any                      `json:"static_value,omitempty"`
	Gradient             any                      `json:"gradient,omitempty"`
	LayerRefID           uint32                   `json:"layer_ref_id,omitempty"`
	Expression           string                   `json:"expression,omitempty"`
	ExpressionEnabled    *bool                    `json:"expression_enabled,omitempty"`
	TemporalEaseStatus   string                   `json:"temporal_ease_status"`
	Keyframes            []projectKeyframe        `json:"keyframes,omitempty"`
}

type projectPropertyRecord struct {
	PropertyType   string                         `json:"property_type"`
	MatchName      string                         `json:"match_name,omitempty"`
	Name           string                         `json:"name,omitempty"`
	NameSource     string                         `json:"name_source"`
	Elided         *bool                          `json:"elided"`
	ElidedStatus   string                         `json:"elided_status"`
	CompositionID  uint32                         `json:"composition_id,omitempty"`
	LayerID        uint32                         `json:"layer_id,omitempty"`
	OriginEvidence *projectPropertyOriginEvidence `json:"origin_evidence,omitempty"`
	*projectPropertyFacts
}

type projectPropertyOriginEvidence struct {
	EvidenceKind       string `json:"evidence_kind"`
	SourceOrdinal      int    `json:"source_ordinal"`
	CanonicalPath      string `json:"canonical_path"`
	PreservationStatus string `json:"preservation_status"`
}

type projectWriteCapabilities struct {
	StaticValue       string `json:"static_value"`
	KeyframeValues    string `json:"keyframe_values"`
	KeyframeStructure string `json:"keyframe_structure"`
	Expression        string `json:"expression"`
	Separation        string `json:"separation"`
}

type projectKeyframe struct {
	Time                  float64               `json:"time_seconds"`
	Value                 any                   `json:"value"`
	InInterp              string                `json:"in_interp,omitempty"`
	OutInterp             string                `json:"out_interp,omitempty"`
	InSpatialTangent      []float64             `json:"in_spatial_tangent,omitempty"`
	OutSpatialTangent     []float64             `json:"out_spatial_tangent,omitempty"`
	InTemporalEase        []projectTemporalEase `json:"in_temporal_ease,omitempty"`
	OutTemporalEase       []projectTemporalEase `json:"out_temporal_ease,omitempty"`
	TemporalEaseUnit      string                `json:"temporal_ease_unit"`
	InTemporalEaseStatus  string                `json:"in_temporal_ease_status"`
	OutTemporalEaseStatus string                `json:"out_temporal_ease_status"`
	TemporalEaseStatus    string                `json:"temporal_ease_status"`
}

type projectTemporalEase struct {
	Speed        *float64 `json:"speed"`
	Influence    *float64 `json:"influence"`
	SpeedRaw     string   `json:"speed_raw,omitempty"`
	InfluenceRaw string   `json:"influence_raw,omitempty"`
}

type projectPropertyIntegrity struct {
	ObservedTreeEntryCount  int `json:"observed_tree_entry_count"`
	PreservedTreeEntryCount int `json:"preserved_tree_entry_count"`
	TreeNodeCount           int `json:"tree_node_count"`
	GroupCount              int `json:"group_count"`
	PropertyCount           int `json:"property_count"`
	FlatPropertyCount       int `json:"flat_property_count"`
	LinkedPropertyCount     int `json:"linked_property_count"`
	UnlinkedPropertyCount   int `json:"unlinked_property_count"`
	DuplicateLinkCount      int `json:"duplicate_property_link_count"`
	PreservedOpaqueCount    int `json:"preserved_opaque_count"`
	PreservedUnknownCount   int `json:"preserved_unknown_count"`
	DroppedUnknownCount     int `json:"dropped_unknown_count"`
}

type projectPropertyDiagnostic struct {
	Code          string `json:"code"`
	PropertyRef   string `json:"property_ref,omitempty"`
	MatchName     string `json:"match_name,omitempty"`
	CompositionID uint32 `json:"composition_id,omitempty"`
	LayerID       uint32 `json:"layer_id,omitempty"`
	CanonicalPath string `json:"canonical_path,omitempty"`
	Message       string `json:"message"`
}

type projectWarning struct {
	Code    string `json:"code"`
	Offset  int64  `json:"offset,omitempty"`
	ChunkID string `json:"chunk_id,omitempty"`
	Message string `json:"message,omitempty"`
}

// ProjectJSON returns a detached, versioned read-only project snapshot.
// Uploaded source paths and render-output paths are omitted.
func (d *Document) ProjectJSON() ([]byte, error) {
	if d == nil || d.project == nil {
		return nil, fmt.Errorf("aep: nil document")
	}
	project := d.project.ToJSON()
	propertyRegistry := newProjectPropertyRegistry(d.project, project)
	for _, footage := range project.Footage {
		if footage != nil {
			footage.Path = ""
		}
	}
	project.RenderQueue = nil

	propertyTrees, propertyIntegrity, propertyDiagnostics := projectPropertyTrees(d.project, propertyRegistry)
	envelope := projectJSONEnvelope{
		SchemaVersion:       ProjectJSONSchemaVersion,
		Project:             project,
		CompositionFacts:    projectCompositionFacts(d.project),
		FootageFacts:        projectFootageFacts(d.project),
		PropertyTrees:       propertyTrees,
		PropertyRecords:     propertyRegistry.records,
		PropertyIntegrity:   propertyIntegrity,
		PropertyDiagnostics: propertyDiagnostics,
		Warnings:            projectWarnings(d.project.ParseWarnings, d.project.Warnings),
	}
	return json.Marshal(envelope)
}

func projectCompositionFacts(project *internal.Project) []projectCompositionFact {
	result := make([]projectCompositionFact, 0, len(project.Compositions))
	for _, composition := range project.Compositions {
		if composition == nil {
			continue
		}
		result = append(result, projectCompositionFact{
			ID: composition.ID, PixelAspect: composition.PixelAspect,
			DisplayStartTime: composition.DisplayStartTime,
		})
	}
	return result
}

func projectFootageFacts(project *internal.Project) []projectFootageFact {
	result := make([]projectFootageFact, 0, len(project.Footage))
	for _, footage := range project.Footage {
		if footage == nil {
			continue
		}
		result = append(result, projectFootageFact{
			ID: footage.ID, SolidColor: footage.SolidColor,
			IsPlaceholder: footage.IsPlaceholder,
		})
	}
	return result
}

// ClassifyError exposes stable parser error categories without exporting
// internal RIFX or serializer error types.
func ClassifyError(err error) (ErrorInfo, bool) {
	if err == nil {
		return ErrorInfo{}, false
	}
	var limitErr *rifx.LimitError
	if errors.As(err, &limitErr) {
		return ErrorInfo{
			Code:     "resource-limit",
			Resource: limitErr.Resource,
			Limit:    limitErr.Limit,
			Actual:   limitErr.Actual,
		}, true
	}
	var formatErr *rifx.FormatError
	if errors.As(err, &formatErr) {
		return ErrorInfo{
			Code:   "invalid-format",
			Offset: formatErr.Offset,
		}, true
	}
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		return ErrorInfo{Code: "invalid-format"}, true
	}
	return ErrorInfo{}, false
}

func projectPropertyTrees(project *internal.Project, registry *projectPropertyRegistry) ([]projectPropertyTree, projectPropertyIntegrity, []projectPropertyDiagnostic) {
	if project == nil {
		return nil, projectPropertyIntegrity{}, nil
	}
	build := newProjectPropertyTreeBuild(registry)
	trees := make([]projectPropertyTree, 0)
	for _, composition := range project.Compositions {
		if composition == nil {
			continue
		}
		for _, layer := range composition.Layers {
			if layer == nil {
				continue
			}
			tree := projectPropertyTree{
				CompositionID: composition.ID,
				LayerID:       layer.ID,
				LayerIndex:    layer.Index,
			}
			build.location = projectPropertyLocation{compositionID: composition.ID, layerID: layer.ID}
			if layer.PropertyTree() == nil {
				if projectLayerPropertyCount(layer) > 0 {
					trees = append(trees, tree)
					build.diagnostics = append(build.diagnostics, projectPropertyDiagnostic{
						Code: "missing-property-tree", CompositionID: composition.ID, LayerID: layer.ID,
						Message: "layer has decoded properties but no semantic property tree",
					})
				}
				continue
			}
			propertyIndexes := make(map[*internal.Property]int, len(layer.Properties))
			for index, property := range layer.Properties {
				propertyIndexes[property] = index
			}
			tree.Children = build.children(layer.PropertyTree(), propertyIndexes, "")
			trees = append(trees, tree)
		}
	}
	build.finish()
	return trees, build.integrity, build.diagnostics
}

func projectLayerPropertyCount(layer *internal.Layer) int {
	if layer == nil {
		return 0
	}
	count := len(layer.Properties)
	for _, effect := range layer.Effects {
		if effect != nil {
			count += len(effect.Parameters)
		}
	}
	for _, mask := range layer.Masks {
		if mask != nil {
			count += len(mask.Properties)
		}
	}
	return count
}

func projectWarnings(
	structured []internal.ParseWarning,
	legacy []string,
) []projectWarning {
	warnings := make([]projectWarning, 0, len(structured)+len(legacy))
	for _, warning := range structured {
		warnings = append(warnings, projectWarning{
			Code:    "parse-warning",
			Offset:  warning.Offset,
			ChunkID: warning.Chunk,
			Message: warning.Message,
		})
	}
	if len(structured) == 0 {
		for _, warning := range legacy {
			warnings = append(warnings, projectWarning{
				Code:    "parse-warning",
				Message: warning,
			})
		}
	}
	return warnings
}
