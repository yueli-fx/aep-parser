package scene

// PropertyBase is the common interface satisfied by both *Property (leaf)
// and *AEPropertyGroup (interior node). Used so AEPropertyGroup.Children can
// hold mixed leaves and subgroups in their original on-disk order — which
// mirrors AE's property index: every node has a match-name
// and a (usually empty) display name.
type PropertyBase interface {
	PropertyMatchName() string
	PropertyName() string
}

// AEPropertyType mirrors the semantic PropertyBase.propertyType values AE
// exposes. Unknown is a parser extension used when the binary wrapper is
// preserved but its AE runtime role has not been established.
type AEPropertyType uint8

const (
	AEPropertyTypeUnknown AEPropertyType = iota
	AEPropertyTypeProperty
	AEPropertyTypeNamedGroup
	AEPropertyTypeIndexedGroup
)

func (t AEPropertyType) String() string {
	switch t {
	case AEPropertyTypeProperty:
		return "PROPERTY"
	case AEPropertyTypeNamedGroup:
		return "NAMED_GROUP"
	case AEPropertyTypeIndexedGroup:
		return "INDEXED_GROUP"
	default:
		return "UNKNOWN"
	}
}

// AEOpaqueProperty preserves an on-disk named node that the parser cannot
// materialize as Property yet. Keeping it in the semantic tree makes loss
// explicit while the original RIFX subtree remains available for round-trip.
type AEOpaqueProperty struct {
	MatchName    string
	Name         string
	NameSource   string
	SemanticType AEPropertyType
	ValueType    PropertyValueType

	decodeEvidence propertyDecodeEvidenceProvider
}

func (p *AEOpaqueProperty) PropertyMatchName() string { return p.MatchName }
func (p *AEOpaqueProperty) PropertyName() string      { return p.Name }
func (p *AEOpaqueProperty) PropertyType() AEPropertyType {
	if p.SemanticType == AEPropertyTypeUnknown {
		return AEPropertyTypeUnknown
	}
	return p.SemanticType
}
func (p *AEOpaqueProperty) ValuePropertyType() PropertyValueType {
	return p.ValueType
}
func (p *AEOpaqueProperty) DecodeEvidence() PropertyDecodeEvidence {
	if p == nil || p.decodeEvidence == nil {
		return PropertyDecodeEvidence{}
	}
	return p.decodeEvidence.DecodeEvidence()
}

// SetOpaquePropertyDecodeEvidence wires serializer-owned parse evidence onto
// an opaque semantic node without exposing binary chunks to scene.
func SetOpaquePropertyDecodeEvidence(p *AEOpaqueProperty, provider interface {
	DecodeEvidence() PropertyDecodeEvidence
}) {
	if p != nil {
		p.decodeEvidence = provider
	}
}

// PropertyMatchName returns the property's AE match-name (e.g.
// "ADBE Position"). Implements PropertyBase.
func (p *Property) PropertyMatchName() string { return p.MatchName }

// PropertyName returns the property's display name. Implements PropertyBase.
func (p *Property) PropertyName() string { return p.Name }

// AEPropertyGroup is the hierarchical wrapper for AE's tdgp groups
// ("ADBE Transform Group", "ADBE Mask Parade", "ADBE Effect Parade", etc.).
// It mirrors AE's PropertyGroup interface — children are ordered and
// can be looked up by match-name or walked positionally.
//
// The layer-level root group has empty MatchName / Name; its Children are
// the top-level subgroups AE writes under the Layr tdgp.
//
// The on-disk hierarchy is preserved during parse alongside the flat
// `Layer.Properties / Effects / Markers / Masks` slices — existing flat-
// API consumers are unaffected.
type AEPropertyGroup struct {
	MatchName string
	Name      string
	// NameSource distinguishes an on-disk instance name from a parser fallback.
	NameSource string
	Children   []PropertyBase
	// SemanticType is assigned by the parser from the property-group grammar.
	// Unknown is retained when a wrapper does not establish AE's group role.
	SemanticType AEPropertyType

	parent *AEPropertyGroup
	back   PropertyGroupWriter // underlying tdgp LIST shard, interface-typed (concrete via propertyGroupBack); see back_property_group.go

	// layer back-ref, set only on the synthetic root by parseLayer. Lets a
	// leaf reach its owning Layer (walk parent → root → layer) for structural
	// mutations like SetDimensionsSeparated that must append a new follower
	// Property to Layer.Properties. nil for groups built outside the parser.
	layer *Layer
}

// OwnerLayer walks from a parsed leaf up to the synthetic property-tree root
// and returns the owning Layer, or nil when the property was built outside
// the parser or is not attached to a layer property tree. Exported for the
// serializer stage (internal/aep).
func (p *Property) OwnerLayer() *Layer {
	g := p.parentTreeGroup
	for g != nil {
		if g.layer != nil {
			return g.layer
		}
		g = g.parent
	}
	return nil
}

// PropertyMatchName implements PropertyBase.
func (g *AEPropertyGroup) PropertyMatchName() string { return g.MatchName }

// PropertyName implements PropertyBase.
func (g *AEPropertyGroup) PropertyName() string { return g.Name }

// ParentGroup returns the containing group, or nil for the root.
func (g *AEPropertyGroup) ParentGroup() *AEPropertyGroup { return g.parent }

// NumProperties returns the number of direct children (mix of leaves and
// subgroups). Mirrors AE's PropertyGroup.numProperties.
func (g *AEPropertyGroup) NumProperties() int { return len(g.Children) }

// Property returns the direct-child *Property whose match-name equals
// matchName, or nil if no direct child by that name exists (or the child
// exists but is a AEPropertyGroup, not a leaf). Does NOT recurse — call
// Group(...).Property(...) for nested access.
func (g *AEPropertyGroup) Property(matchName string) *Property {
	for _, c := range g.Children {
		if c.PropertyMatchName() != matchName {
			continue
		}
		if p, ok := c.(*Property); ok {
			return p
		}
	}
	return nil
}

// Group returns the direct-child *AEPropertyGroup whose match-name equals
// matchName, or nil if no such subgroup exists at this level. Does NOT
// recurse.
func (g *AEPropertyGroup) Group(matchName string) *AEPropertyGroup {
	for _, c := range g.Children {
		if c.PropertyMatchName() != matchName {
			continue
		}
		if sub, ok := c.(*AEPropertyGroup); ok {
			return sub
		}
	}
	return nil
}

// ChildByIndex returns the child at the given index, or nil if out of
// bounds. Mirrors AE's `property(index)` / 1-based AE addressing
// translated to 0-based Go.
func (g *AEPropertyGroup) ChildByIndex(i int) PropertyBase {
	if i < 0 || i >= len(g.Children) {
		return nil
	}
	return g.Children[i]
}

// PropertyByPath walks the group hierarchy following each match-name in
// turn. The terminal match-name must resolve to a *Property (leaf);
// intermediate names must resolve to *AEPropertyGroup. Returns nil if any
// step fails.
//
// Example: layer.PropertyTree().PropertyByPath("ADBE Transform Group", "ADBE Position")
func (g *AEPropertyGroup) PropertyByPath(matchNames ...string) *Property {
	if len(matchNames) == 0 {
		return nil
	}
	cur := g
	for i, name := range matchNames {
		if i == len(matchNames)-1 {
			return cur.Property(name)
		}
		sub := cur.Group(name)
		if sub == nil {
			return nil
		}
		cur = sub
	}
	return nil
}

// ParentGroup returns the AEPropertyGroup that contains this property in
// the layer's hierarchical property tree, or nil when the property was
// built outside the parser or hasn't been wired through
// wirePropertyTreeLeaves.
func (p *Property) ParentGroup() *AEPropertyGroup { return p.parentTreeGroup }

// indexedGroupMatchNames are AE's INDEXED_GROUP container match-names whose
// direct children support structural Remove / MoveTo / Duplicate.
var indexedGroupMatchNames = map[string]bool{
	"ADBE Effect Parade":      true,
	"ADBE Mask Parade":        true,
	"ADBE Effect Mask Parade": true,
	"ADBE Text Animators":     true,
	"ADBE Root Vectors Group": true,
}

// namedGroupMatchNames contains tdgp-backed groups whose fixed-child,
// NAMED_GROUP semantics have been established. Unknown tdgp wrappers stay
// UNKNOWN instead of being guessed from the absence of indexed semantics.
var namedGroupMatchNames = map[string]bool{
	"ADBE Transform Group":        true,
	"ADBE Audio Group":            true,
	"ADBE Text Properties":        true,
	"ADBE Camera Options Group":   true,
	"ADBE Light Options Group":    true,
	"ADBE Material Options Group": true,
	"ADBE Extrsn Options Group":   true,
	"ADBE Layer Sets":             true,
	"ADBE Source Options Group":   true,
	"ADBE Blend Options Group":    true,
	"ADBE Adv Blend Group":        true,
	"ADBE Vector Transform Group": true,
	"ADBE Vector Materials Group": true,
}

// ParsedPropertyGroupType classifies a tdgp-backed AE group using the
// parser's established semantic catalog. Wrapper grammar establishes that the
// node is a group; the catalog distinguishes AE's indexed containers from
// fixed named groups.
func ParsedPropertyGroupType(matchName string) AEPropertyType {
	if indexedGroupMatchNames[matchName] {
		return AEPropertyTypeIndexedGroup
	}
	if namedGroupMatchNames[matchName] {
		return AEPropertyTypeNamedGroup
	}
	return AEPropertyTypeUnknown
}

// IsIndexedGroup reports whether this group is one of AE's INDEXED_GROUP
// containers, i.e. whether its direct children support structural Remove /
// MoveTo / Duplicate. Named groups (Transform, Material Options, …) and leaf
// properties return false.
func (g *AEPropertyGroup) IsIndexedGroup() bool {
	if g == nil {
		return false
	}
	if g.SemanticType != AEPropertyTypeUnknown {
		return g.SemanticType == AEPropertyTypeIndexedGroup
	}
	// Mutation-created groups have no parsed semantic metadata. Retain the
	// established structural API behaviour for those in-memory objects.
	return indexedGroupMatchNames[g.MatchName]
}

// PropertyType reports AE's semantic group type, or UNKNOWN for a preserved
// binary wrapper that has not been classified as an AE property group.
func (g *AEPropertyGroup) PropertyType() AEPropertyType {
	if g == nil {
		return AEPropertyTypeUnknown
	}
	if g.SemanticType != AEPropertyTypeUnknown {
		return g.SemanticType
	}
	return AEPropertyTypeUnknown
}

type propertyGroupIntegrityProvider interface {
	ChildIntegrity() (observed, preserved int)
}

// ChildIntegrity returns parser preservation evidence for the group's direct
// children when its serializer-owned backing provides it. In-memory groups
// without parser backing return zero values.
func (g *AEPropertyGroup) ChildIntegrity() (observed, preserved int) {
	if g == nil || g.back == nil {
		return 0, 0
	}
	provider, ok := g.back.(propertyGroupIntegrityProvider)
	if !ok {
		return 0, 0
	}
	return provider.ChildIntegrity()
}

// PropertyIndex returns the position of child within this group's
// Children slice (0-based), or -1 when child is not a direct child of g.
// Compared by pointer identity.
func (g *AEPropertyGroup) PropertyIndex(child PropertyBase) int {
	for i, c := range g.Children {
		// Compare by interface equality, which collapses to pointer
		// equality for *Property and *AEPropertyGroup (both pointer types).
		if c == child {
			return i
		}
	}
	return -1
}

// Depth returns the number of hops from this group to the synthetic
// root: 0 for the root itself, 1 for top-level subgroups
// (Transform Group, etc.), 2 for nested subgroups, and so on.
func (g *AEPropertyGroup) Depth() int {
	d := 0
	for cur := g.parent; cur != nil; cur = cur.parent {
		d++
	}
	return d
}
