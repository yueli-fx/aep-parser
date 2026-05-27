package aep

import (
	"github.com/example/aep-parser/internal/rifx"
)

// PropertyBase is the common interface satisfied by both *Property (leaf)
// and *AEPropertyGroup (interior node). Used so AEPropertyGroup.Children can
// hold mixed leaves and subgroups in their original on-disk order — which
// mirrors AE's property index. py-aep parity: every node has a match-name
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
// It mirrors py-aep's AEPropertyGroup interface — children are ordered and
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
	chunk  *rifx.Chunk // the underlying tdgp LIST; nil for the synthetic root
}

// PropertyMatchName implements PropertyBase.
func (g *AEPropertyGroup) PropertyMatchName() string { return g.MatchName }

// PropertyName implements PropertyBase.
func (g *AEPropertyGroup) PropertyName() string { return g.Name }

// ParentGroup returns the containing group, or nil for the root.
func (g *AEPropertyGroup) ParentGroup() *AEPropertyGroup { return g.parent }

// NumProperties returns the number of direct children (mix of leaves and
// subgroups). Mirrors py-aep's AEPropertyGroup.num_properties.
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
// bounds. Mirrors py-aep's `property_by_index` / 1-based AE addressing
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

// buildAEPropertyGroupTree constructs the hierarchical AEPropertyGroup mirror
// of layr's tdgp tree. The returned root has empty MatchName / Name; its
// Children are the named top-level subgroups AE writes under the Layr.
//
// Effects, masks, and markers are still diverted to their typed slices by
// the existing parseProperties / parseMasks calls — this tree captures
// the layout structure without duplicating their decode work. Effects and
// marker subtrees DO appear as AEPropertyGroup nodes here (with their
// match-name) so chained lookup like
// `layer.PropertyTree().Group("ADBE Effect Parade")` finds them, but
// individual effect parameters are not re-parsed (the subgroup's
// Children are empty for those — use Layer.Effects for typed access).
func buildAEPropertyGroupTree(layr *rifx.Chunk) *AEPropertyGroup {
	root := &AEPropertyGroup{}
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
	walkTdmnPairs(group, func(name string, payload *rifx.Chunk) bool {
		switch payload.FormType {
		case rifx.IDTdbs:
			// Leaf placeholder — resolved later by wirePropertyTreeLeaves.
			placeholder := &propertyTreeLeafRef{matchName: name, tdbs: payload}
			parent.Children = append(parent.Children, placeholder)
		case rifx.IDTdgp:
			sub := &AEPropertyGroup{MatchName: name, Name: name, parent: parent, chunk: payload}
			addNamedChildren(payload, sub)
			parent.Children = append(parent.Children, sub)
		default:
			// otst, parT, mrst, etc. — wrap as an opaque group so chained
			// lookup still finds the name, but don't descend further.
			sub := &AEPropertyGroup{MatchName: name, Name: name, parent: parent, chunk: payload}
			parent.Children = append(parent.Children, sub)
		}
		return true
	})
}

// propertyTreeLeafRef is a transient placeholder inserted by
// addNamedChildren and replaced by wirePropertyTreeLeaves with the
// actual *Property from layer.Properties (matched by tdbs chunk identity).
type propertyTreeLeafRef struct {
	matchName string
	tdbs      *rifx.Chunk
}

func (r *propertyTreeLeafRef) PropertyMatchName() string { return r.matchName }
func (r *propertyTreeLeafRef) PropertyName() string      { return r.matchName }

// wirePropertyTreeLeaves replaces every propertyTreeLeafRef placeholder
// in the tree with the matching *Property from props. Match is by tdbs
// chunk pointer identity (each tdbs has exactly one *Property; the parser
// records it as Property.tdbs). Unresolved placeholders (e.g. leaves
// parseLeafProperty rejected as "nothing useful") are dropped from the
// tree to keep PropertyByPath / Group / Property accessors honest.
func wirePropertyTreeLeaves(root *AEPropertyGroup, props []*Property) {
	byTdbs := make(map[*rifx.Chunk]*Property, len(props))
	for _, p := range props {
		if p.tdbs != nil {
			byTdbs[p.tdbs] = p
		}
	}
	var walk func(g *AEPropertyGroup)
	walk = func(g *AEPropertyGroup) {
		filtered := g.Children[:0]
		for _, c := range g.Children {
			switch v := c.(type) {
			case *propertyTreeLeafRef:
				if p, ok := byTdbs[v.tdbs]; ok {
					filtered = append(filtered, p)
				}
				// else: parseLeafProperty rejected this tdbs — drop placeholder.
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
