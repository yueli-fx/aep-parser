package aep

import "github.com/example/aep-parser/internal/rifx"

// Property-tree parse stage: builds the hierarchical AEPropertyGroup mirror
// of a Layr's tdgp tree and wires its leaves to the flat Layer.Properties.
// The AEPropertyGroup scene type itself (and its accessors) live in
// scene_property_group.go; everything chunk-coupled is here.

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
	root := &AEPropertyGroup{back: &propertyGroupBackrefs{}}
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
			sub := &AEPropertyGroup{MatchName: name, Name: name, parent: parent, back: &propertyGroupBackrefs{chunk: payload}}
			addNamedChildren(payload, sub)
			parent.Children = append(parent.Children, sub)
		default:
			// otst, parT, mrst, etc. — wrap as an opaque group so chained
			// lookup still finds the name, but don't descend further.
			sub := &AEPropertyGroup{MatchName: name, Name: name, parent: parent, back: &propertyGroupBackrefs{chunk: payload}}
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
// tree. Resolved leaves get their `parentTreeGroup` back-ref set so
// Property.ParentGroup() works.
func wirePropertyTreeLeaves(root *AEPropertyGroup, props []*Property) {
	byTdbs := make(map[*rifx.Chunk]*Property, len(props))
	for _, p := range props {
		if pb := p.propertyBack(); pb != nil && pb.tdbs != nil {
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
					p.parentTreeGroup = g
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
