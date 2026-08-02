package serializer

import (
	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// Property-tree parse stage: builds the hierarchical AEPropertyGroup mirror
// of a Layr's tdgp tree and wires its leaves to the flat Layer.Properties.
// The AEPropertyGroup scene type itself (and its accessors) live in
// scene_property_group.go; everything chunk-coupled is here.

// buildAEPropertyGroupTree constructs the hierarchical AEPropertyGroup mirror
// of layr's tdgp tree. The returned root has empty MatchName / Name; its
// Children are the named top-level subgroups AE writes under the Layr.
//
// Effects, masks, and markers are still decoded into their typed slices by
// the existing parseProperties / parseMasks calls. This tree reuses those
// decoded property objects by binary chunk identity, while preserving opaque
// wrappers for node kinds that do not yet have a structured codec.
func buildAEPropertyGroupTree(layr *rifx.Chunk) *AEPropertyGroup {
	root := &AEPropertyGroup{SemanticType: AEPropertyTypeNamedGroup}
	scene.SetPropertyGroupBack(root, &propertyGroupBackrefs{})
	for _, ch := range layr.Children {
		if !ch.IsList() || ch.FormType != rifx.IDTdgp {
			continue
		}
		addNamedChildren(ch, root)
	}
	return root
}

// addNamedChildren walks one tdgp LIST and appends each named entry to
// `parent.Children` as a *AEPropertyGroup (when the payload is a nested
// tdgp) or as a *Property reference resolved against the parent layer's
// flat property list. Leaves are NOT re-parsed here — the existing
// parseLeafProperty work is shared by looking up the match-name in the
// caller-supplied resolver.
//
// We defer the leaf resolver wire-up to wirePropertyTreeLeaves, called
// after parseProperties has populated layer.Properties. This avoids a
// second parseLeafProperty pass (which would create duplicate *Property
// instances and double-count tdb4 refs).
func addNamedChildren(group *rifx.Chunk, parent *AEPropertyGroup) {
	before := len(parent.Children)
	observed := walkPropertyTreeEntries(group, func(name string, payload *rifx.Chunk) bool {
		if payload == nil {
			parent.Children = append(parent.Children, newOpaqueProperty(name, name, "fallback", AEPropertyTypeUnknown, PVTUnknown, "failed"))
			return true
		}
		displayName, nameSource := propertyTreeDisplayName(payload, name)
		if !payload.IsList() {
			valueType, status, known := knownOpaqueProperty(name, rifx.ChunkID{})
			if !known {
				valueType, status = PVTUnknown, "unsupported"
			}
			parent.Children = append(parent.Children, newOpaqueProperty(name, displayName, nameSource, AEPropertyTypeUnknown, valueType, status))
			return true
		}
		switch payload.FormType {
		case rifx.IDTdbs:
			// Leaf placeholder — resolved later by wirePropertyTreeLeaves.
			placeholder := &propertyTreeLeafRef{
				matchName: name, name: displayName, nameSource: nameSource, tdbs: payload,
			}
			parent.Children = append(parent.Children, placeholder)
		case rifx.IDTdgp:
			sub := &AEPropertyGroup{
				MatchName: name, Name: displayName, NameSource: nameSource, SemanticType: scene.ParsedPropertyGroupType(name),
			}
			scene.SetPropertyGroupParent(sub, parent)
			scene.SetPropertyGroupBack(sub, &propertyGroupBackrefs{chunk: payload})
			addNamedChildren(payload, sub)
			parent.Children = append(parent.Children, sub)
		case rifx.IDOtst, rifx.IDGCst:
			// These LISTs are binary wrappers around one AE leaf property, not
			// AE PropertyGroup nodes. Link their inner tdbs to the specialized
			// Property parsed by parseOrientationProperty/parseGradientStopsProperty.
			if tdbs := payload.FindFirstList(rifx.IDTdbs); tdbs != nil {
				parent.Children = append(parent.Children, &propertyTreeLeafRef{
					matchName: name, name: displayName, nameSource: nameSource, tdbs: tdbs,
				})
			} else {
				parent.Children = append(parent.Children, newOpaqueProperty(name, displayName, nameSource, AEPropertyTypeProperty, PVTUnknown, "unsupported"))
			}
		default:
			if valueType, status, ok := knownOpaqueProperty(name, payload.FormType); ok {
				parent.Children = append(parent.Children, newOpaqueProperty(name, displayName, nameSource, AEPropertyTypeProperty, valueType, status))
				break
			}

			// Effect/mask/shape instances use non-tdgp wrappers but are AE
			// groups. A child of an indexed group is a named group; wrappers
			// containing a nested tdgp are also groups. Descend so decoded
			// effect parameters can share identity with Effect.Parameters.
			sub := &AEPropertyGroup{
				MatchName: name, Name: displayName, NameSource: nameSource,
			}
			scene.SetPropertyGroupBack(sub, &propertyGroupBackrefs{chunk: payload})
			hasPropertyEntries := addWrappedPropertyChildren(payload, sub)
			if parent.PropertyType() == AEPropertyTypeIndexedGroup {
				sub.SemanticType = AEPropertyTypeNamedGroup
			}
			if parent.PropertyType() == AEPropertyTypeIndexedGroup || hasPropertyEntries {
				scene.SetPropertyGroupParent(sub, parent)
				parent.Children = append(parent.Children, sub)
			} else {
				parent.Children = append(parent.Children, newOpaqueProperty(name, displayName, nameSource, AEPropertyTypeUnknown, PVTUnknown, "unsupported"))
			}
		}
		return true
	})
	addPropertyGroupChildIntegrity(parent, observed, len(parent.Children)-before)
}

// walkPropertyTreeEntries is the property-tree counterpart of walkTdmnPairs.
// Mask atoms are the one observed exception to the usual adjacent pair:
// their tdmn is followed by mkif metadata and only then the child tdgp.
func walkPropertyTreeEntries(group *rifx.Chunk, fn func(name string, payload *rifx.Chunk) bool) int {
	observed := 0
	kids := group.Children
	for index := 0; index < len(kids); index++ {
		if kids[index].ID != rifx.IDTdmn {
			continue
		}
		name := trimNUL(kids[index].Data)
		if name == "ADBE Group End" {
			return observed
		}
		observed++
		payloadIndex := index + 1
		if payloadIndex >= len(kids) {
			if !fn(name, nil) {
				return observed
			}
			continue
		}
		if !kids[payloadIndex].IsList() && name == "ADBE Mask Atom" && payloadIndex+1 < len(kids) && kids[payloadIndex+1].IsList() && kids[payloadIndex+1].FormType == rifx.IDTdgp {
			payloadIndex++
		}
		payload := kids[payloadIndex]
		if !payload.IsList() {
			if !fn(name, payload) {
				return observed
			}
			continue
		}
		index = payloadIndex
		if !fn(name, payload) {
			return observed
		}
	}
	return observed
}

func propertyTreeDisplayName(payload *rifx.Chunk, fallback string) (string, string) {
	if name := vectorGroupName(payload); name != "" && name != aeDefaultGroupName {
		return name, "decoded"
	}
	return fallback, "fallback"
}

func knownOpaqueProperty(name string, formType rifx.ChunkID) (PropertyValueType, string, bool) {
	switch {
	case formType == rifx.IDMrst || name == "ADBE Marker":
		return PVTMarker, "partially-decoded", true
	case formType == rifx.IDBtds || name == "ADBE Text Document":
		return PVTTextDocument, "partially-decoded", true
	case formType == rifx.IDOmS || name == "ADBE Mask Shape":
		return PVTShape, "partially-decoded", true
	default:
		return PVTUnknown, "", false
	}
}

func newOpaqueProperty(name, displayName, nameSource string, semanticType AEPropertyType, valueType PropertyValueType, status string) *AEOpaqueProperty {
	property := &AEOpaqueProperty{
		MatchName: name, Name: displayName, NameSource: nameSource,
		SemanticType: semanticType, ValueType: valueType,
	}
	scene.SetOpaquePropertyDecodeEvidence(property, &opaquePropertyDecodeEvidence{decodeStatus: status})
	return property
}

func addWrappedPropertyChildren(wrapper *rifx.Chunk, parent *AEPropertyGroup) bool {
	before := len(parent.Children)
	var visit func(*rifx.Chunk)
	visit = func(container *rifx.Chunk) {
		for _, child := range container.Children {
			if child.ID == rifx.IDTdmn {
				addNamedChildren(container, parent)
				return
			}
		}
		for _, child := range container.Children {
			if child.IsList() {
				visit(child)
			}
		}
	}
	visit(wrapper)
	return len(parent.Children) > before
}

// propertyTreeLeafRef is a transient placeholder inserted by
// addNamedChildren and replaced by wirePropertyTreeLeaves with the
// actual *Property from layer.Properties (matched by tdbs chunk identity).
type propertyTreeLeafRef struct {
	matchName  string
	name       string
	nameSource string
	tdbs       *rifx.Chunk
}

func (r *propertyTreeLeafRef) PropertyMatchName() string { return r.matchName }
func (r *propertyTreeLeafRef) PropertyName() string      { return r.name }

// wirePropertyTreeLeaves replaces every propertyTreeLeafRef placeholder
// in the tree with the matching *Property from props. Match is by tdbs
// chunk pointer identity (each tdbs has exactly one *Property; the parser
// records it as Property.tdbs). Unresolved placeholders become
// AEOpaqueProperty leaves so their identity and position remain observable.
// Resolved leaves get their `parentTreeGroup` back-ref set so
// Property.ParentGroup() works.
func wirePropertyTreeLeaves(root *AEPropertyGroup, props []*Property) {
	byTdbs := make(map[*rifx.Chunk]*Property, len(props))
	for _, p := range props {
		if pb := propertyBack(p); pb != nil && pb.tdbs != nil {
			byTdbs[pb.tdbs] = p
		}
	}
	var walk func(g *AEPropertyGroup)
	walk = func(g *AEPropertyGroup) {
		filtered := g.Children[:0]
		for _, c := range g.Children {
			switch v := c.(type) {
			case *propertyTreeLeafRef:
				if p, ok := byTdbs[v.tdbs]; ok {
					if p.Name == "" || p.Name == p.MatchName {
						p.Name = v.name
					}
					p.NameSource = v.nameSource
					scene.SetPropertyParentTreeGroup(p, g)
					filtered = append(filtered, p)
				} else {
					filtered = append(filtered, newOpaqueProperty(
						v.matchName, v.name, v.nameSource, AEPropertyTypeProperty, PVTUnknown, "unsupported",
					))
				}
			case *AEPropertyGroup:
				walk(v)
				filtered = append(filtered, v)
			default:
				filtered = append(filtered, c)
			}
		}
		g.Children = filtered
	}
	walk(root)
}

func layerTreeProperties(layer *Layer) []*Property {
	if layer == nil {
		return nil
	}
	properties := append([]*Property(nil), layer.Properties...)
	for _, effect := range layer.Effects {
		if effect != nil {
			properties = append(properties, effect.Parameters...)
		}
	}
	for _, mask := range layer.Masks {
		if mask != nil {
			properties = append(properties, mask.Properties...)
		}
	}
	return properties
}
