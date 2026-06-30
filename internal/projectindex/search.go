package projectindex

import (
	"strconv"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// HitKind classifies a search hit by the project object it matched.
type HitKind string

const (
	HitProjectItem HitKind = "project_item"
	HitComposition HitKind = "composition"
	HitLayer       HitKind = "layer"
	HitEffect      HitKind = "effect"
	HitProperty    HitKind = "property"
	HitExpression  HitKind = "expression"
	HitTextStyle   HitKind = "text_style"
)

// Hit is a stable search location plus the value that matched.
type Hit struct {
	Kind     HitKind     `json:"kind"`
	Match    Match       `json:"match"`
	Location Location    `json:"location"`
	Pointers HitPointers `json:"-"`
}

// Match identifies the field and value that satisfied a search query.
type Match struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// Location is the machine-readable identity of a hit within one project.
type Location struct {
	ItemID   uint32 `json:"item_id"`
	ItemKind string `json:"item_kind"`
	ItemName string `json:"item_name"`

	CompID   uint32 `json:"comp_id"`
	CompName string `json:"comp_name"`

	LayerID    uint32 `json:"layer_id"`
	LayerIndex int    `json:"layer_index"`
	LayerName  string `json:"layer_name"`

	EffectMatchName  string `json:"effect_match_name"`
	EffectName       string `json:"effect_name"`
	EffectOccurrence int    `json:"effect_occurrence"`

	PropertyMatchName string `json:"property_match_name"`
	PropertyName      string `json:"property_name"`
	PropertyPath      string `json:"property_path"`
}

// HitPointers are single-project conveniences. They are not durable corpus
// identity; use Location for serialized results.
type HitPointers struct {
	Item     aep.AVItem
	Comp     *aep.Composition
	Layer    *aep.Layer
	Effect   *aep.Effect
	Property *aep.Property
}

// SearchLayersBySourceID returns layer hits whose SourceID equals sourceID.
func (idx *Index) SearchLayersBySourceID(sourceID uint32) []Hit {
	if idx == nil || idx.project == nil || sourceID == 0 {
		return nil
	}
	item := idx.AVItemByID(sourceID)
	itemKind, itemName := itemLocation(item)
	hits := []Hit{}
	for _, comp := range idx.project.Compositions {
		if comp == nil {
			continue
		}
		for _, layer := range comp.Layers {
			if layer == nil || layer.SourceID != sourceID {
				continue
			}
			hits = append(hits, Hit{
				Kind:  HitLayer,
				Match: Match{Field: "source_id", Value: strconv.FormatUint(uint64(sourceID), 10)},
				Location: Location{
					ItemID:     sourceID,
					ItemKind:   itemKind,
					ItemName:   itemName,
					CompID:     comp.ID,
					CompName:   comp.Name,
					LayerID:    layer.ID,
					LayerIndex: layer.Index,
					LayerName:  layer.Name,
				},
				Pointers: HitPointers{
					Item:  item,
					Comp:  comp,
					Layer: layer,
				},
			})
		}
	}
	return hits
}

// SearchEffectsByMatchName returns effect hits whose MatchName equals
// matchName. Results follow project composition/layer/effect order.
func (idx *Index) SearchEffectsByMatchName(matchName string) []Hit {
	if idx == nil || idx.project == nil || matchName == "" {
		return nil
	}
	hits := []Hit{}
	for _, comp := range idx.project.Compositions {
		if comp == nil {
			continue
		}
		for _, layer := range comp.Layers {
			if layer == nil {
				continue
			}
			occurrences := map[string]int{}
			for _, effect := range layer.Effects {
				if effect == nil || effect.MatchName == "" {
					continue
				}
				occurrences[effect.MatchName]++
				if effect.MatchName != matchName {
					continue
				}
				hits = append(hits, Hit{
					Kind:  HitEffect,
					Match: Match{Field: "effect.match_name", Value: matchName},
					Location: Location{
						CompID:           comp.ID,
						CompName:         comp.Name,
						LayerID:          layer.ID,
						LayerIndex:       layer.Index,
						LayerName:        layer.Name,
						EffectMatchName:  effect.MatchName,
						EffectName:       effect.Name,
						EffectOccurrence: occurrences[effect.MatchName],
					},
					Pointers: HitPointers{
						Comp:   comp,
						Layer:  layer,
						Effect: effect,
					},
				})
			}
		}
	}
	return hits
}

func itemLocation(item aep.AVItem) (string, string) {
	switch v := item.(type) {
	case *aep.Composition:
		return "composition", v.Name
	case *aep.Footage:
		return "footage", v.Name
	default:
		return "", ""
	}
}
