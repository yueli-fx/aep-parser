package technique

import (
	"fmt"
	"sort"
	"strings"
)

func BuildExplanation(portrait *Portrait) (*Explanation, error) {
	if portrait == nil {
		return nil, fmt.Errorf("technique: nil portrait")
	}
	explanation := &Explanation{
		SchemaVersion:        SchemaVersion,
		SourcePath:           portrait.SourcePath,
		Portrait:             *portrait,
		Overview:             buildOverview(portrait),
		Archetypes:           buildArchetypes(portrait),
		Techniques:           buildTechniqueExplanations(portrait),
		TopSignalLayers:      topSignalLayers(portrait.SignalLayers, 3),
		ReproducibilityNotes: buildReproducibilityNotes(portrait),
		UnknownNotes:         buildUnknownNotes(portrait),
	}
	return explanation, nil
}

func buildOverview(portrait *Portrait) []string {
	fp := portrait.Fingerprint
	overview := []string{fmt.Sprintf(
		"Project has %d comps, %d layers, %d effects, %d text animators, %d shape operators, and %d dependency edges.",
		fp.CompCount,
		fp.LayerCount,
		fp.EffectCount,
		fp.TextAnimatorCount,
		fp.ShapeOperatorCount,
		fp.DependencyCount,
	)}
	if roles := formatTopCounts(fp.LayerRoleCounts, 4); roles != "" {
		overview = append(overview, "Dominant layer roles: "+roles+".")
	}
	if effects := formatTopCounts(portrait.Mechanisms.EffectMatchCounts, 4); effects != "" {
		overview = append(overview, "Most-used effects: "+effects+".")
	}
	if shapes := formatTopCounts(portrait.Mechanisms.ShapeFamilyCounts, 4); shapes != "" {
		overview = append(overview, "Most-used shape mechanisms: "+shapes+".")
	}
	return overview
}

func buildArchetypes(portrait *Portrait) []ProjectArchetype {
	var out []ProjectArchetype
	fp := portrait.Fingerprint
	roles := fp.LayerRoleCounts
	graph := portrait.Graph.RelationCounts
	repro := portrait.Mechanisms.ReproducibilityCounts
	hints := map[string]bool{}
	for _, hint := range portrait.TechniqueHints {
		hints[hint.ID] = true
	}

	if fp.ShapeOperatorCount >= 10 || roles["shape"] >= 3 || hints["shape_operator_stack"] {
		out = append(out, ProjectArchetype{
			ID:      "shape_system",
			Label:   "Shape system",
			Score:   fp.ShapeOperatorCount + roles["shape"]*2,
			Summary: "The project relies heavily on shape layers and vector operators.",
			Signals: []string{
				fmt.Sprintf("shape operators: %d", fp.ShapeOperatorCount),
				fmt.Sprintf("shape layers: %d", roles["shape"]),
			},
		})
	}
	if fp.EffectCount >= 3 || hints["effect_driven_layer"] {
		out = append(out, ProjectArchetype{
			ID:      "effect_stack",
			Label:   "Effect stack",
			Score:   fp.EffectCount*3 + len(portrait.Mechanisms.EffectMatchCounts),
			Summary: "The look is built from effect stacks and tuned effect parameters.",
			Signals: []string{
				fmt.Sprintf("effects: %d", fp.EffectCount),
				"top effects: " + fallbackText(formatTopCounts(portrait.Mechanisms.EffectMatchCounts, 3), "none"),
			},
		})
	}
	if fp.TextAnimatorCount > 0 || fp.TextLayerCount > 0 || hints["kinetic_text"] {
		out = append(out, ProjectArchetype{
			ID:      "text_animation",
			Label:   "Text animation",
			Score:   fp.TextAnimatorCount*4 + fp.TextLayerCount*2,
			Summary: "The project contains text layers or text animator mechanisms.",
			Signals: []string{
				fmt.Sprintf("text animators: %d", fp.TextAnimatorCount),
				"text animator kinds: " + fallbackText(formatTopCounts(portrait.Mechanisms.TextAnimatorKindCounts, 3), "none"),
			},
		})
	}
	if graph["source"] > 0 || roles["precomp"] > 0 || hints["precomp_assembly"] {
		out = append(out, ProjectArchetype{
			ID:      "precomp_system",
			Label:   "Precomp system",
			Score:   graph["source"]*3 + roles["precomp"]*2,
			Summary: "Composition nesting is part of the project structure.",
			Signals: []string{
				fmt.Sprintf("precomp layers: %d", roles["precomp"]),
				fmt.Sprintf("source edges: %d", graph["source"]),
			},
		})
	}
	if hints["controller_rig"] || roles["controller"] > 0 || graph["effect_param_layer"] > 0 {
		out = append(out, ProjectArchetype{
			ID:      "controller_rig",
			Label:   "Controller rig",
			Score:   roles["controller"]*5 + graph["effect_param_layer"]*2,
			Summary: "Controller layers or layer-reference parameters coordinate other layers.",
			Signals: []string{
				fmt.Sprintf("controller layers: %d", roles["controller"]),
				fmt.Sprintf("effect layer-reference edges: %d", graph["effect_param_layer"]),
			},
		})
	}
	if repro["third_party"] > 0 || hints["plugin_dependent"] {
		out = append(out, ProjectArchetype{
			ID:      "plugin_dependent",
			Label:   "Plugin dependent",
			Score:   repro["third_party"] * 5,
			Summary: "The project uses third-party effects that affect exact recreation.",
			Signals: []string{fmt.Sprintf("third-party effects: %d", repro["third_party"])},
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func buildTechniqueExplanations(portrait *Portrait) []TechniqueExplanation {
	out := make([]TechniqueExplanation, 0, len(portrait.TechniqueHints))
	for _, hint := range portrait.TechniqueHints {
		out = append(out, explainHint(hint, portrait))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ID != out[j].ID {
			return out[i].ID < out[j].ID
		}
		return out[i].Summary < out[j].Summary
	})
	return out
}

func explainHint(hint TechniqueHint, portrait *Portrait) TechniqueExplanation {
	note := TechniqueExplanation{
		ID:         hint.ID,
		Title:      techniqueTitle(hint.ID),
		Confidence: hint.Confidence,
		Evidence:   append([]string(nil), hint.Signals...),
	}
	sort.Strings(note.Evidence)
	switch hint.ID {
	case "controller_rig":
		note.Summary = "A controller/null or effect layer-reference rig coordinates values across other layers."
	case "effect_driven_layer":
		effects := formatTopCounts(portrait.Mechanisms.EffectMatchCounts, 4)
		if effects == "" {
			effects = "detected effects"
		}
		note.Summary = "Layer appearance is driven by tuned effects, keyframes, expressions, or layer-reference parameters; prominent effects: " + effects + "."
	case "kinetic_text":
		textKinds := formatTopCounts(portrait.Mechanisms.TextAnimatorKindCounts, 4)
		if textKinds == "" {
			textKinds = "text animator properties"
		}
		note.Summary = "Text layers use animator properties to drive typography motion; detected animator kinds: " + textKinds + "."
	case "matte_composite":
		note.Summary = "Track matte dependencies reveal cutout or layered compositing structure."
	case "plugin_dependent":
		thirdParty := portrait.Mechanisms.ReproducibilityCounts["third_party"]
		note.Summary = fmt.Sprintf("Third-party effects are present; exact recreation depends on plugin availability (%d third-party effect occurrences).", thirdParty)
	case "precomp_assembly":
		note.Summary = "Nested compositions are used to assemble reusable or staged scene parts."
	case "shape_operator_stack":
		shapes := formatTopCounts(portrait.Mechanisms.ShapeFamilyCounts, 5)
		if shapes == "" {
			shapes = "multiple shape operators"
		}
		note.Summary = "Shape layers combine multiple shape operators; prominent shape families: " + shapes + "."
	default:
		note.Summary = "Detected technique signal: " + strings.ReplaceAll(hint.ID, "_", " ") + "."
	}
	return note
}

func techniqueTitle(id string) string {
	switch id {
	case "controller_rig":
		return "Controller rig"
	case "effect_driven_layer":
		return "Effect-driven layer"
	case "kinetic_text":
		return "Kinetic text"
	case "matte_composite":
		return "Matte composite"
	case "plugin_dependent":
		return "Plugin-dependent effect"
	case "precomp_assembly":
		return "Precomp assembly"
	case "shape_operator_stack":
		return "Shape operator stack"
	default:
		return titleFromID(id)
	}
}

func titleFromID(id string) string {
	words := strings.Fields(strings.ReplaceAll(id, "_", " "))
	for i, word := range words {
		if word == "" {
			continue
		}
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}

func topSignalLayers(layers []SignalLayer, max int) []SignalLayer {
	if max <= 0 || len(layers) == 0 {
		return nil
	}
	out := append([]SignalLayer(nil), layers...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		if out[i].CompName != out[j].CompName {
			return out[i].CompName < out[j].CompName
		}
		return out[i].LayerName < out[j].LayerName
	})
	if len(out) > max {
		out = out[:max]
	}
	return out
}

func buildReproducibilityNotes(portrait *Portrait) []string {
	counts := portrait.Mechanisms.ReproducibilityCounts
	notes := make([]string, 0, 4)
	if counts["native"] > 0 {
		notes = append(notes, fmt.Sprintf("native effects: %d", counts["native"]))
	}
	if counts["cycore"] > 0 {
		notes = append(notes, fmt.Sprintf("cycore bundled effects: %d", counts["cycore"]))
	}
	if counts["third_party"] > 0 {
		notes = append(notes, fmt.Sprintf("third-party effects: %d", counts["third_party"]))
	}
	if counts["unknown"] > 0 {
		notes = append(notes, fmt.Sprintf("unknown effect reproducibility: %d", counts["unknown"]))
	}
	if len(notes) == 0 {
		notes = append(notes, "No effect reproducibility risks detected.")
	}
	return notes
}

func buildUnknownNotes(portrait *Portrait) []string {
	if portrait.Unknowns.Count == 0 {
		return []string{"No unknown parsed fields were reported."}
	}
	return []string{fmt.Sprintf("%d unknown parsed items need inspection before claiming exact recreation.", portrait.Unknowns.Count)}
}

type countRow struct {
	name  string
	count int
}

func formatTopCounts(counts map[string]int, max int) string {
	rows := topCounts(counts, max)
	parts := make([]string, 0, len(rows))
	for _, row := range rows {
		parts = append(parts, fmt.Sprintf("%s (%d)", row.name, row.count))
	}
	return strings.Join(parts, ", ")
}

func fallbackText(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func topCounts(counts map[string]int, max int) []countRow {
	if max <= 0 || len(counts) == 0 {
		return nil
	}
	rows := make([]countRow, 0, len(counts))
	for name, count := range counts {
		if name == "" || count <= 0 {
			continue
		}
		rows = append(rows, countRow{name: name, count: count})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].count != rows[j].count {
			return rows[i].count > rows[j].count
		}
		return rows[i].name < rows[j].name
	})
	if len(rows) > max {
		rows = rows[:max]
	}
	return rows
}
