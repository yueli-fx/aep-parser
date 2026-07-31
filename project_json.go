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
const ProjectJSONSchemaVersion = 1

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
	SchemaVersion    int                      `json:"schema_version"`
	Project          *internal.JSONProject    `json:"project"`
	CompositionFacts []projectCompositionFact `json:"composition_facts,omitempty"`
	FootageFacts     []projectFootageFact     `json:"footage_facts,omitempty"`
	PropertyTrees    []projectPropertyTree    `json:"property_trees,omitempty"`
	Warnings         []projectWarning         `json:"warnings,omitempty"`
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
	Kind       string                `json:"kind"`
	MatchName  string                `json:"match_name,omitempty"`
	Name       string                `json:"name,omitempty"`
	Occurrence int                   `json:"occurrence,omitempty"`
	FlatIndex  *int                  `json:"flat_index,omitempty"`
	Children   []projectPropertyNode `json:"children,omitempty"`
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
	for _, footage := range project.Footage {
		if footage != nil {
			footage.Path = ""
		}
	}
	project.RenderQueue = nil

	envelope := projectJSONEnvelope{
		SchemaVersion:    ProjectJSONSchemaVersion,
		Project:          project,
		CompositionFacts: projectCompositionFacts(d.project),
		FootageFacts:     projectFootageFacts(d.project),
		PropertyTrees:    projectPropertyTrees(d.project),
		Warnings:         projectWarnings(d.project.ParseWarnings, d.project.Warnings),
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

func projectPropertyTrees(project *internal.Project) []projectPropertyTree {
	if project == nil {
		return nil
	}
	trees := make([]projectPropertyTree, 0)
	for _, composition := range project.Compositions {
		if composition == nil {
			continue
		}
		for _, layer := range composition.Layers {
			if layer == nil || layer.PropertyTree() == nil {
				continue
			}
			propertyIndexes := make(map[*internal.Property]int, len(layer.Properties))
			for index, property := range layer.Properties {
				propertyIndexes[property] = index
			}
			trees = append(trees, projectPropertyTree{
				CompositionID: composition.ID,
				LayerID:       layer.ID,
				LayerIndex:    layer.Index,
				Children:      projectPropertyChildren(layer.PropertyTree(), propertyIndexes),
			})
		}
	}
	return trees
}

func projectPropertyChildren(
	group *internal.AEPropertyGroup,
	propertyIndexes map[*internal.Property]int,
) []projectPropertyNode {
	if group == nil {
		return nil
	}
	occurrences := map[string]int{}
	children := make([]projectPropertyNode, 0, group.NumProperties())
	for index := 0; index < group.NumProperties(); index++ {
		child := group.ChildByIndex(index)
		if child == nil {
			continue
		}
		matchName := child.PropertyMatchName()
		occurrence := occurrences[matchName]
		occurrences[matchName] = occurrence + 1
		node := projectPropertyNode{
			MatchName:  matchName,
			Name:       child.PropertyName(),
			Occurrence: occurrence,
		}
		switch typed := child.(type) {
		case *internal.AEPropertyGroup:
			node.Kind = "group"
			node.Children = projectPropertyChildren(typed, propertyIndexes)
		case *internal.Property:
			node.Kind = "property"
			if flatIndex, exists := propertyIndexes[typed]; exists {
				value := flatIndex
				node.FlatIndex = &value
			}
		default:
			continue
		}
		children = append(children, node)
	}
	return children
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
