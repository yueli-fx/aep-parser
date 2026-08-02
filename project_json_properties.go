package aep

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	internal "github.com/yueli-fx/aep-parser/internal/aep"
)

type projectPropertyLocation struct {
	compositionID uint32
	layerID       uint32
}

type projectPropertyRegistry struct {
	refs          map[*internal.Property]string
	locations     map[string]projectPropertyLocation
	recordIndexes map[string]int
	records       []projectPropertyRecord
	next          int
}

func newProjectPropertyRegistry(project *internal.Project, jsonProject *internal.JSONProject) *projectPropertyRegistry {
	r := &projectPropertyRegistry{
		refs:          map[*internal.Property]string{},
		locations:     map[string]projectPropertyLocation{},
		recordIndexes: map[string]int{},
		next:          1,
	}
	if project == nil || jsonProject == nil {
		return r
	}
	for compositionIndex, composition := range project.Compositions {
		if composition == nil || compositionIndex >= len(jsonProject.Compositions) {
			continue
		}
		jsonComposition := jsonProject.Compositions[compositionIndex]
		for layerIndex, layer := range composition.Layers {
			if layer == nil || layerIndex >= len(jsonComposition.Layers) {
				continue
			}
			jsonLayer := jsonComposition.Layers[layerIndex]
			location := projectPropertyLocation{compositionID: composition.ID, layerID: layer.ID}
			for index, property := range layer.Properties {
				if index < len(jsonLayer.Properties) {
					r.register(property, jsonLayer.Properties[index], location)
				}
			}
			for effectIndex, effect := range layer.Effects {
				if effect == nil || effectIndex >= len(jsonLayer.Effects) {
					continue
				}
				jsonEffect := jsonLayer.Effects[effectIndex]
				for parameterIndex, property := range effect.Parameters {
					if parameterIndex < len(jsonEffect.Parameters) {
						r.register(property, jsonEffect.Parameters[parameterIndex], location)
					}
				}
			}
			for maskIndex, mask := range layer.Masks {
				if mask == nil || maskIndex >= len(jsonLayer.Masks) {
					continue
				}
				jsonMask := jsonLayer.Masks[maskIndex]
				for propertyIndex, property := range mask.Properties {
					if propertyIndex < len(jsonMask.Properties) {
						r.register(property, jsonMask.Properties[propertyIndex], location)
					}
				}
			}
		}
	}
	return r
}

func (r *projectPropertyRegistry) register(property *internal.Property, jsonProperty *internal.JSONProperty, location projectPropertyLocation) string {
	if property == nil {
		return ""
	}
	ref, exists := r.refs[property]
	if !exists {
		ref = r.nextRef()
		r.refs[property] = ref
		r.locations[ref] = location
		r.recordIndexes[ref] = len(r.records)
		r.records = append(r.records, projectPropertyRecord{
			PropertyType: "PROPERTY", MatchName: property.MatchName, Name: property.Name,
			NameSource: propertyBaseNameSource(property), ElidedStatus: "unknown",
			CompositionID: location.compositionID, LayerID: location.layerID,
			projectPropertyFacts: propertyFacts(property, ref),
		})
	}
	if jsonProperty != nil {
		jsonProperty.PropertyRef = ref
	}
	return ref
}

func (r *projectPropertyRegistry) ref(property *internal.Property) string {
	if ref, ok := r.refs[property]; ok {
		return ref
	}
	return r.register(property, nil, projectPropertyLocation{})
}

func (r *projectPropertyRegistry) registerOpaque(property *internal.AEOpaqueProperty, location projectPropertyLocation) string {
	ref := r.nextRef()
	facts := opaquePropertyFacts(property, ref)
	r.locations[ref] = location
	r.recordIndexes[ref] = len(r.records)
	r.records = append(r.records, projectPropertyRecord{
		PropertyType: property.PropertyType().String(), MatchName: property.MatchName, Name: property.Name,
		NameSource: propertyBaseNameSource(property), ElidedStatus: "unknown",
		CompositionID: location.compositionID, LayerID: location.layerID,
		projectPropertyFacts: facts,
	})
	return ref
}

func (r *projectPropertyRegistry) setOrigin(ref string, evidence *projectPropertyOriginEvidence) {
	if r == nil || ref == "" || evidence == nil {
		return
	}
	index, ok := r.recordIndexes[ref]
	if !ok || index < 0 || index >= len(r.records) {
		return
	}
	r.records[index].OriginEvidence = evidence
}

func (r *projectPropertyRegistry) nextRef() string {
	ref := fmt.Sprintf("property-%d", r.next)
	r.next++
	return ref
}

type projectPropertyTreeBuild struct {
	registry    *projectPropertyRegistry
	linked      map[string]int
	location    projectPropertyLocation
	integrity   projectPropertyIntegrity
	diagnostics []projectPropertyDiagnostic
}

func newProjectPropertyTreeBuild(registry *projectPropertyRegistry) *projectPropertyTreeBuild {
	return &projectPropertyTreeBuild{registry: registry, linked: map[string]int{}}
}

func (b *projectPropertyTreeBuild) children(group *internal.AEPropertyGroup, propertyIndexes map[*internal.Property]int, parentPath string) []projectPropertyNode {
	if group == nil {
		return nil
	}
	observed, preserved := group.ChildIntegrity()
	b.integrity.ObservedTreeEntryCount += observed
	b.integrity.PreservedTreeEntryCount += preserved
	if observed > preserved {
		dropped := observed - preserved
		b.integrity.DroppedUnknownCount += dropped
		b.diagnostics = append(b.diagnostics, projectPropertyDiagnostic{
			Code: "dropped-tree-entry", CanonicalPath: nonEmpty(parentPath, "/"),
			Message: fmt.Sprintf("%d named source entries were not retained in this property group", dropped),
		})
	}
	occurrences := map[string]int{}
	sourceOrderPreserved := observed > 0 && observed == preserved && preserved == group.NumProperties()
	children := make([]projectPropertyNode, 0, group.NumProperties())
	for index := 0; index < group.NumProperties(); index++ {
		child := group.ChildByIndex(index)
		if child == nil {
			continue
		}
		matchName := child.PropertyMatchName()
		occurrence := occurrences[matchName]
		occurrences[matchName] = occurrence + 1
		path := canonicalPropertyPath(parentPath, index, matchName, occurrence)
		node := projectPropertyNode{
			MatchName: matchName, Name: child.PropertyName(), NameSource: propertyBaseNameSource(child),
			Occurrence: occurrence, SiblingIndex: index, CanonicalPath: path, ElidedStatus: "unknown",
		}
		if sourceOrderPreserved {
			node.OriginEvidence = &projectPropertyOriginEvidence{
				EvidenceKind: "named-entry-order", SourceOrdinal: index, CanonicalPath: path, PreservationStatus: "preserved",
			}
		}
		switch typed := child.(type) {
		case *internal.AEPropertyGroup:
			node.Kind = "group"
			node.PropertyType = typed.PropertyType().String()
			node.Children = b.children(typed, propertyIndexes, path)
			b.integrity.GroupCount++
			if node.PropertyType == "UNKNOWN" {
				b.integrity.PreservedUnknownCount++
			}
		case *internal.Property:
			node.Kind = "property"
			node.PropertyType = "PROPERTY"
			ref := b.registry.ref(typed)
			node.PropertyRef = ref
			b.registry.setOrigin(ref, node.OriginEvidence)
			if flatIndex, exists := propertyIndexes[typed]; exists {
				value := flatIndex
				node.FlatIndex = &value
			}
			b.linked[ref]++
			if b.linked[ref] > 1 {
				b.integrity.DuplicateLinkCount++
				location := b.registry.locations[ref]
				b.diagnostics = append(b.diagnostics, projectPropertyDiagnostic{
					Code: "duplicate-property-link", PropertyRef: ref, MatchName: typed.MatchName,
					CompositionID: location.compositionID, LayerID: location.layerID, CanonicalPath: path,
					Message: "decoded property is referenced more than once by the semantic property tree",
				})
			}
			b.integrity.PropertyCount++
			if typed.ValuePropertyType() == internal.PVTUnknown {
				b.integrity.PreservedUnknownCount++
			}
		case *internal.AEOpaqueProperty:
			node.Kind = "opaque"
			node.PropertyType = typed.PropertyType().String()
			ref := b.registry.registerOpaque(typed, b.location)
			node.PropertyRef = ref
			b.registry.setOrigin(ref, node.OriginEvidence)
			b.linked[ref]++
			if node.PropertyType == "PROPERTY" {
				b.integrity.PropertyCount++
			}
			b.integrity.PreservedOpaqueCount++
			if node.PropertyType == "UNKNOWN" || typed.ValuePropertyType() == internal.PVTUnknown {
				b.integrity.PreservedUnknownCount++
			}
		default:
			b.integrity.DroppedUnknownCount++
			b.diagnostics = append(b.diagnostics, projectPropertyDiagnostic{
				Code: "unsupported-tree-node", MatchName: matchName, CanonicalPath: path,
				Message: "semantic property tree contained an unsupported node implementation",
			})
			continue
		}
		b.integrity.TreeNodeCount++
		children = append(children, node)
	}
	return children
}

func (b *projectPropertyTreeBuild) finish() {
	b.integrity.FlatPropertyCount = len(b.registry.records)
	b.integrity.LinkedPropertyCount = len(b.linked)
	for _, record := range b.registry.records {
		ref := record.PropertyRef
		if b.linked[ref] > 0 {
			continue
		}
		location := b.registry.locations[ref]
		b.integrity.UnlinkedPropertyCount++
		b.diagnostics = append(b.diagnostics, projectPropertyDiagnostic{
			Code: "unlinked-property", PropertyRef: ref, MatchName: record.MatchName,
			CompositionID: location.compositionID, LayerID: location.layerID,
			Message: "decoded property is not referenced by the semantic property tree",
		})
	}
}

func propertyFacts(property *internal.Property, ref string) *projectPropertyFacts {
	valueType := property.ValuePropertyType()
	canVary, spatial := property.CanVaryOverTime(), property.IsSpatial()
	leader, separated := property.IsSeparationLeader(), property.DimensionsSeparated()
	follower := property.IsSeparationFollower()
	facts := &projectPropertyFacts{
		PropertyRef: ref, PropertyValueType: propertyValueTypeName(valueType),
		Dimensions: semanticDimensions(valueType, property.Components), StorageDimensions: property.Components,
		CanVaryOverTime: &canVary, IsSpatial: &spatial, IsSeparationLeader: &leader,
		DimensionsSeparated: &separated, IsSeparationFollower: &follower,
		DecodeStatus: propertyDecodeStatus(property, valueType), RawPreserved: true,
		StaticValue: property.StaticValue, Gradient: property.Gradient, LayerRefID: property.LayerRefID,
		Expression: property.Expression,
	}
	if property.Expression != "" {
		enabled := property.ExpressionEnabled
		facts.ExpressionEnabled = &enabled
	}
	if dimension := property.SeparationDimension(); dimension >= 0 {
		facts.SeparationDimension = &dimension
	}
	facts.WriteCapability, facts.WriteCapabilities = propertyWriteCapabilities(property, valueType)
	for _, keyframe := range property.Keyframes {
		facts.Keyframes = append(facts.Keyframes, projectKeyframeFrom(property, keyframe))
	}
	facts.TemporalEaseStatus = propertyTemporalEaseStatus(property, facts.Keyframes)
	return facts
}

func opaquePropertyFacts(property *internal.AEOpaqueProperty, ref string) *projectPropertyFacts {
	status := property.DecodeEvidence().DecodeStatus
	if status == "" {
		status = "unsupported"
	}
	return &projectPropertyFacts{
		PropertyRef: ref, PropertyValueType: propertyValueTypeName(property.ValuePropertyType()),
		DecodeStatus: status, RawPreserved: true, WriteCapability: "preserve-only", TemporalEaseStatus: "unknown",
		WriteCapabilities: projectWriteCapabilities{
			StaticValue: "preserve-only", KeyframeValues: "preserve-only", KeyframeStructure: "unsupported",
			Expression: "unknown", Separation: "unsupported",
		},
	}
}

func propertyWriteCapabilities(property *internal.Property, valueType internal.PropertyValueType) (string, projectWriteCapabilities) {
	capabilities := projectWriteCapabilities{
		StaticValue: "not-applicable", KeyframeValues: "not-applicable", KeyframeStructure: "not-applicable",
		Expression: "unknown", Separation: "not-applicable",
	}
	writer := property.MutationCapabilities()
	if !writer.Known {
		if property.StaticValue != nil {
			capabilities.StaticValue = "unknown"
		}
		if len(property.Keyframes) > 0 {
			capabilities.KeyframeValues = "unknown"
			capabilities.KeyframeStructure = "unknown"
		}
		if property.IsSeparationLeader() || property.IsSeparationFollower() {
			capabilities.Separation = "unknown"
		}
		return "unknown", capabilities
	}
	capabilities.Expression = capabilityStatus(writer.Expression)
	if property.IsSeparationLeader() || property.IsSeparationFollower() {
		// The current capability provider intentionally does not claim this
		// until all leader/follower/layout preconditions are checked.
		capabilities.Separation = "unknown"
	}
	switch valueType {
	case internal.PVTUnknown, internal.PVTCustomValue, internal.PVTMarker, internal.PVTShape, internal.PVTTextDocument:
		capabilities.StaticValue = preserveCapability(property.StaticValue != nil)
		capabilities.KeyframeValues = preserveCapability(len(property.Keyframes) > 0)
		capabilities.KeyframeStructure = "unsupported"
		capabilities.Expression = preserveCapability(property.Expression != "")
		capabilities.Separation = "unsupported"
		return "preserve-only", capabilities
	default:
		if len(property.Keyframes) > 0 {
			capabilities.KeyframeValues = capabilityStatus(writer.KeyframeValues)
			capabilities.KeyframeStructure = capabilityStatus(writer.KeyframeStructure)
		} else if property.StaticValue != nil {
			capabilities.StaticValue = capabilityStatus(writer.StaticValue)
		}
		if capabilities.StaticValue == "supported" || capabilities.KeyframeValues == "supported" {
			return "supported", capabilities
		}
		return "preserve-only", capabilities
	}
}

func capabilityStatus(supported bool) string {
	if supported {
		return "supported"
	}
	return "unsupported"
}

func projectKeyframeFrom(property *internal.Property, keyframe *internal.Keyframe) projectKeyframe {
	if keyframe == nil {
		return projectKeyframe{TemporalEaseUnit: "ratio", InTemporalEaseStatus: "failed", OutTemporalEaseStatus: "failed", TemporalEaseStatus: "failed"}
	}
	result := projectKeyframe{
		Time: keyframe.Time, Value: keyframe.Value, InInterp: keyframe.InInterp.String(), OutInterp: keyframe.OutInterp.String(),
		InSpatialTangent: keyframe.InSpatialTangent, OutSpatialTangent: keyframe.OutSpatialTangent,
		TemporalEaseUnit: "ratio",
	}
	for _, ease := range keyframe.InTemporalEase {
		result.InTemporalEase = append(result.InTemporalEase, projectTemporalEaseFrom(ease))
	}
	for _, ease := range keyframe.OutTemporalEase {
		result.OutTemporalEase = append(result.OutTemporalEase, projectTemporalEaseFrom(ease))
	}
	expected := temporalEaseDimensions(property)
	result.InTemporalEaseStatus = temporalEaseStatus(keyframe.InInterp, keyframe.InTemporalEase, expected)
	result.OutTemporalEaseStatus = temporalEaseStatus(keyframe.OutInterp, keyframe.OutTemporalEase, expected)
	result.TemporalEaseStatus = combinedTemporalEaseStatus(result.InTemporalEaseStatus, result.OutTemporalEaseStatus)
	if property != nil && property.DecodeEvidence().TemporalEaseStatus == "invalid-preserved" {
		result.InTemporalEaseStatus = "invalid-preserved"
		result.OutTemporalEaseStatus = "invalid-preserved"
		result.TemporalEaseStatus = "invalid-preserved"
	}
	return result
}

func projectTemporalEaseFrom(ease internal.TemporalEase) projectTemporalEase {
	result := projectTemporalEase{}
	result.Speed, result.SpeedRaw = finiteJSONNumber(ease.Speed)
	result.Influence, result.InfluenceRaw = finiteJSONNumber(ease.Influence)
	return result
}

func finiteJSONNumber(value float64) (*float64, string) {
	switch {
	case math.IsNaN(value):
		return nil, "NaN"
	case math.IsInf(value, 1):
		return nil, "+Infinity"
	case math.IsInf(value, -1):
		return nil, "-Infinity"
	default:
		return &value, ""
	}
}

func temporalEaseDimensions(property *internal.Property) int {
	switch property.ValuePropertyType() {
	case internal.PVTTwoD:
		return 2
	case internal.PVTThreeD:
		return 3
	default:
		return 1
	}
}

func temporalEaseStatus(interp internal.InterpType, eases []internal.TemporalEase, expected int) string {
	if len(eases) == 0 {
		return "not-present"
	}
	if len(eases) != expected {
		return "invalid-preserved"
	}
	for _, ease := range eases {
		if math.IsNaN(ease.Speed) || math.IsInf(ease.Speed, 0) || math.IsNaN(ease.Influence) || math.IsInf(ease.Influence, 0) {
			return "invalid-preserved"
		}
		if ease.Influence < 0 || ease.Influence > 1 || (interp == internal.InterpBezier && ease.Influence == 0) {
			return "invalid-preserved"
		}
	}
	if interp != internal.InterpBezier {
		return "not-applicable-preserved"
	}
	return "valid"
}

func combinedTemporalEaseStatus(in, out string) string {
	if in == "invalid-preserved" || out == "invalid-preserved" || in == "failed" || out == "failed" {
		return "invalid-preserved"
	}
	if in == "valid" || out == "valid" {
		return "valid"
	}
	if in == "not-present" && out == "not-present" {
		return "not-present"
	}
	return "not-applicable-preserved"
}

func propertyDecodeStatus(property *internal.Property, valueType internal.PropertyValueType) string {
	if property != nil {
		switch status := property.DecodeEvidence().DecodeStatus; status {
		case "failed", "unsupported", "partially-decoded":
			return status
		case "decoded":
			if valueType != internal.PVTUnknown {
				return "decoded"
			}
		}
	}
	if valueType == internal.PVTUnknown {
		return "partially-decoded"
	}
	return "decoded"
}

func propertyTemporalEaseStatus(property *internal.Property, keyframes []projectKeyframe) string {
	if property != nil {
		if status := property.DecodeEvidence().TemporalEaseStatus; status != "" {
			return status
		}
	}
	status := "not-present"
	for _, keyframe := range keyframes {
		switch keyframe.TemporalEaseStatus {
		case "invalid-preserved":
			return "invalid-preserved"
		case "valid":
			status = "valid"
		case "not-applicable-preserved":
			if status == "not-present" {
				status = "not-applicable-preserved"
			}
		}
	}
	return status
}

func propertyValueTypeName(valueType internal.PropertyValueType) string {
	switch valueType {
	case internal.PVTNoValue:
		return "NO_VALUE"
	case internal.PVTOneD:
		return "OneD"
	case internal.PVTTwoD:
		return "TwoD"
	case internal.PVTTwoDSpatial:
		return "TwoD_SPATIAL"
	case internal.PVTThreeD:
		return "ThreeD"
	case internal.PVTThreeDSpatial:
		return "ThreeD_SPATIAL"
	case internal.PVTColor:
		return "COLOR"
	case internal.PVTCustomValue:
		return "CUSTOM_VALUE"
	case internal.PVTMarker:
		return "MARKER"
	case internal.PVTLayerIndex:
		return "LAYER_INDEX"
	case internal.PVTMaskIndex:
		return "MASK_INDEX"
	case internal.PVTShape:
		return "SHAPE"
	case internal.PVTTextDocument:
		return "TEXT_DOCUMENT"
	default:
		return "UNKNOWN"
	}
}

func semanticDimensions(valueType internal.PropertyValueType, storage int) int {
	switch valueType {
	case internal.PVTOneD, internal.PVTLayerIndex, internal.PVTMaskIndex:
		return 1
	case internal.PVTTwoD, internal.PVTTwoDSpatial:
		return 2
	case internal.PVTThreeD, internal.PVTThreeDSpatial:
		return 3
	case internal.PVTColor:
		return 4
	default:
		return storage
	}
}

func projectPropertyNameSource(name, matchName string) string {
	switch {
	case name == "":
		return "missing"
	case name == matchName:
		return "fallback"
	default:
		return "decoded"
	}
}

func propertyBaseNameSource(property internal.PropertyBase) string {
	switch typed := property.(type) {
	case *internal.Property:
		if typed.NameSource != "" {
			return typed.NameSource
		}
	case *internal.AEPropertyGroup:
		if typed.NameSource != "" {
			return typed.NameSource
		}
	case *internal.AEOpaqueProperty:
		if typed.NameSource != "" {
			return typed.NameSource
		}
	}
	return projectPropertyNameSource(property.PropertyName(), property.PropertyMatchName())
}

func canonicalPropertyPath(parent string, siblingIndex int, matchName string, occurrence int) string {
	segment := strconv.Itoa(siblingIndex) + ":" + strings.ReplaceAll(strings.ReplaceAll(matchName, "~", "~0"), "/", "~1") + "#" + strconv.Itoa(occurrence)
	return parent + "/" + segment
}

func nonEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func preserveCapability(present bool) string {
	if present {
		return "preserve-only"
	}
	return "not-applicable"
}
