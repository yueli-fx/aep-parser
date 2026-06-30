package projectindex

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// Corpus is a durable fact set extracted from one or more projects.
type Corpus struct {
	Projects []CorpusProject `json:"projects,omitempty"`
	Facts    []CorpusFact    `json:"facts,omitempty"`
}

// CorpusProject summarizes one ingested project without retaining its graph.
type CorpusProject struct {
	ID          string `json:"id"`
	Path        string `json:"path,omitempty"`
	Fingerprint string `json:"fingerprint"`
	CompCount   int    `json:"comp_count"`
	LayerCount  int    `json:"layer_count"`
	EffectCount int    `json:"effect_count"`
}

// CorpusFactKind classifies a durable project fact.
type CorpusFactKind string

const (
	FactLayerSource CorpusFactKind = "layer_source"
	FactEffectUsage CorpusFactKind = "effect_usage"
	FactExpression  CorpusFactKind = "expression"
	FactTextStyle   CorpusFactKind = "text_style"
	FactShapeUsage  CorpusFactKind = "shape_usage"
)

// CorpusFact is a serializable search/learning fact tied to a project ID.
type CorpusFact struct {
	ProjectID string         `json:"project_id"`
	Kind      CorpusFactKind `json:"kind"`
	Location  Location       `json:"location"`
	Match     Match          `json:"match"`
	Summary   string         `json:"summary,omitempty"`
}

// CorpusBuilder ingests projects and emits durable facts without keeping
// parsed Project or Index pointers alive.
type CorpusBuilder interface {
	AddProject(path string, project *aep.Project) error
	Build() (*Corpus, error)
}

// Builder is the default in-memory CorpusBuilder implementation.
type Builder struct {
	projects []CorpusProject
	facts    []CorpusFact
}

var _ CorpusBuilder = (*Builder)(nil)

// NewCorpusBuilder returns an empty corpus builder.
func NewCorpusBuilder() *Builder {
	return &Builder{}
}

// AddProject extracts first-slice corpus facts from project. Nil projects are
// accepted and produce a project record with zero counts and no facts.
func (b *Builder) AddProject(path string, project *aep.Project) error {
	idx := Build(project)
	meta := summarizeProject(path, project)
	b.projects = append(b.projects, meta)
	if project == nil {
		return nil
	}
	for _, comp := range project.Compositions {
		if comp == nil {
			continue
		}
		for _, layer := range comp.Layers {
			if layer == nil {
				continue
			}
			if layer.SourceID != 0 {
				b.facts = append(b.facts, layerSourceFact(meta.ID, idx, comp, layer))
			}
			occurrences := map[string]int{}
			for _, effect := range layer.Effects {
				if effect == nil || effect.MatchName == "" {
					continue
				}
				occurrences[effect.MatchName]++
				b.facts = append(b.facts, effectUsageFact(meta.ID, comp, layer, effect, occurrences[effect.MatchName]))
			}
		}
	}
	idx.walkProperties(func(comp *aep.Composition, layer *aep.Layer, effect *aep.Effect, effectOccurrence int, property *aep.Property, propertyOccurrence int, propertyPath string) {
		if property.Expression == "" {
			return
		}
		hit := propertyHit(HitExpression, Match{Field: "property.expression", Value: property.Expression}, comp, layer, effect, effectOccurrence, property, propertyPath)
		b.facts = append(b.facts, expressionFact(meta.ID, hit))
	})
	idx.walkTextStyles(func(comp *aep.Composition, layer *aep.Layer, runIndex int, font string) {
		b.facts = append(b.facts, textStyleFact(meta.ID, textStyleHit(comp, layer, runIndex, font)))
	})
	return nil
}

// Build returns a copy of the accumulated corpus.
func (b *Builder) Build() (*Corpus, error) {
	if b == nil {
		return &Corpus{}, nil
	}
	return &Corpus{
		Projects: append([]CorpusProject(nil), b.projects...),
		Facts:    append([]CorpusFact(nil), b.facts...),
	}, nil
}

func summarizeProject(path string, project *aep.Project) CorpusProject {
	compCount, layerCount, effectCount := projectCounts(project)
	fingerprint := fingerprintProject(path, project)
	return CorpusProject{
		ID:          fingerprint[:12],
		Path:        path,
		Fingerprint: fingerprint,
		CompCount:   compCount,
		LayerCount:  layerCount,
		EffectCount: effectCount,
	}
}

func projectCounts(project *aep.Project) (int, int, int) {
	if project == nil {
		return 0, 0, 0
	}
	compCount := 0
	layerCount := 0
	effectCount := 0
	for _, comp := range project.Compositions {
		if comp == nil {
			continue
		}
		compCount++
		for _, layer := range comp.Layers {
			if layer == nil {
				continue
			}
			layerCount++
			for _, effect := range layer.Effects {
				if effect != nil && effect.MatchName != "" {
					effectCount++
				}
			}
		}
	}
	return compCount, layerCount, effectCount
}

func layerSourceFact(projectID string, idx *Index, comp *aep.Composition, layer *aep.Layer) CorpusFact {
	item := idx.AVItemByID(layer.SourceID)
	itemKind, itemName := itemLocation(item)
	sourceLabel := itemKind
	if sourceLabel == "" {
		sourceLabel = "unknown"
	}
	return CorpusFact{
		ProjectID: projectID,
		Kind:      FactLayerSource,
		Location: Location{
			ItemID:     layer.SourceID,
			ItemKind:   itemKind,
			ItemName:   itemName,
			CompID:     comp.ID,
			CompName:   comp.Name,
			LayerID:    layer.ID,
			LayerIndex: layer.Index,
			LayerName:  layer.Name,
		},
		Match:   Match{Field: "source_id", Value: strconv.FormatUint(uint64(layer.SourceID), 10)},
		Summary: fmt.Sprintf("layer %q references %s source %d", layer.Name, sourceLabel, layer.SourceID),
	}
}

func effectUsageFact(projectID string, comp *aep.Composition, layer *aep.Layer, effect *aep.Effect, occurrence int) CorpusFact {
	return CorpusFact{
		ProjectID: projectID,
		Kind:      FactEffectUsage,
		Location: Location{
			CompID:           comp.ID,
			CompName:         comp.Name,
			LayerID:          layer.ID,
			LayerIndex:       layer.Index,
			LayerName:        layer.Name,
			EffectMatchName:  effect.MatchName,
			EffectName:       effect.Name,
			EffectOccurrence: occurrence,
		},
		Match:   Match{Field: "effect.match_name", Value: effect.MatchName},
		Summary: fmt.Sprintf("layer %q uses effect %q", layer.Name, effect.MatchName),
	}
}

func expressionFact(projectID string, hit Hit) CorpusFact {
	return CorpusFact{
		ProjectID: projectID,
		Kind:      FactExpression,
		Location:  hit.Location,
		Match:     hit.Match,
		Summary:   fmt.Sprintf("layer %q uses expression on %q", hit.Location.LayerName, hit.Location.PropertyMatchName),
	}
}

func textStyleFact(projectID string, hit Hit) CorpusFact {
	return CorpusFact{
		ProjectID: projectID,
		Kind:      FactTextStyle,
		Location:  hit.Location,
		Match:     hit.Match,
		Summary:   fmt.Sprintf("layer %q uses font %q", hit.Location.LayerName, hit.Match.Value),
	}
}

func fingerprintProject(path string, project *aep.Project) string {
	hash := sha256.New()
	fmt.Fprintf(hash, "path=%s\n", path)
	if project == nil {
		fmt.Fprintln(hash, "nil")
		return hex.EncodeToString(hash.Sum(nil))
	}
	for _, comp := range project.Compositions {
		if comp == nil {
			fmt.Fprintln(hash, "comp=<nil>")
			continue
		}
		fmt.Fprintf(hash, "comp=%d:%s\n", comp.ID, comp.Name)
		for _, layer := range comp.Layers {
			if layer == nil {
				fmt.Fprintln(hash, "layer=<nil>")
				continue
			}
			fmt.Fprintf(hash, "layer=%d:%d:%s:%d\n", layer.ID, layer.Index, layer.Name, layer.SourceID)
			for _, effect := range layer.Effects {
				if effect == nil {
					fmt.Fprintln(hash, "effect=<nil>")
					continue
				}
				fmt.Fprintf(hash, "effect=%s:%s\n", effect.MatchName, effect.Name)
			}
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}
