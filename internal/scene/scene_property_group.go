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
	Children  []PropertyBase

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
// the parser / lives under an Effect or Mask subtree (those roots carry no
// layer back-ref). Exported for the serializer stage (internal/aep).
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
// built outside the parser, lives inside an Effect/Mask (not the
// layer-level tdgp), or hasn't been wired through wirePropertyTreeLeaves.
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

// IsIndexedGroup reports whether this group is one of AE's INDEXED_GROUP
// containers, i.e. whether its direct children support structural Remove /
// MoveTo / Duplicate. Named groups (Transform, Material Options, …) and leaf
// properties return false.
func (g *AEPropertyGroup) IsIndexedGroup() bool {
	return g != nil && indexedGroupMatchNames[g.MatchName]
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
