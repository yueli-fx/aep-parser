package selfhost

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

type PseudoControllerRebuildOptions struct {
	InventoryPath     string
	UnderstandingPath string
	OutDir            string
	MaxFamilies       int
}

type PseudoControllerRebuildProof struct {
	SchemaVersion       int                             `json:"schema_version"`
	SourceInventory     string                          `json:"source_inventory"`
	SourceUnderstanding string                          `json:"source_understanding"`
	OutputPath          string                          `json:"output_path,omitempty"`
	Summary             PseudoControllerRebuildSummary  `json:"summary"`
	Families            []PseudoControllerRebuildFamily `json:"families"`
	Boundaries          []string                        `json:"boundaries"`
}

type PseudoControllerRebuildSummary struct {
	PseudoFamilies          int `json:"pseudo_families"`
	SelectedFamilies        int `json:"selected_families"`
	GeneratedFamilies       int `json:"generated_families"`
	GeneratedControls       int `json:"generated_controls"`
	UnsupportedFamilies     int `json:"unsupported_families"`
	UnsupportedControlCount int `json:"unsupported_control_count"`
	SystemParamCount        int `json:"system_param_count"`
}

type PseudoControllerRebuildFamily struct {
	MatchName           string                               `json:"match_name"`
	UID                 string                               `json:"uid"`
	Occurrences         int                                  `json:"occurrences"`
	ParamKinds          int                                  `json:"param_kinds"`
	ParamOccurrences    int                                  `json:"param_occurrences"`
	ProjectSamples      []string                             `json:"project_samples,omitempty"`
	Reproducibility     string                               `json:"reproducibility,omitempty"`
	GenerationPolicy    string                               `json:"generation_policy,omitempty"`
	StudyPriority       int                                  `json:"study_priority,omitempty"`
	Controls            []PseudoControllerControlPlan        `json:"controls"`
	UnsupportedControls []PseudoControllerUnsupportedControl `json:"unsupported_controls,omitempty"`
	SystemParamCount    int                                  `json:"system_param_count,omitempty"`
	Status              string                               `json:"status"`
	GeneratedAEP        string                               `json:"generated_aep,omitempty"`
	Boundary            string                               `json:"boundary,omitempty"`
}

type PseudoControllerControlPlan struct {
	ParamMatchName string   `json:"param_match_name"`
	Label          string   `json:"label"`
	ControlKind    string   `json:"control_kind"`
	ValueType      string   `json:"value_type"`
	Occurrences    int      `json:"occurrences"`
	DefaultValue   any      `json:"default_value,omitempty"`
	BehaviorNotes  []string `json:"behavior_notes,omitempty"`
}

type PseudoControllerUnsupportedControl struct {
	ParamMatchName string   `json:"param_match_name"`
	ValueTypes     []string `json:"value_types,omitempty"`
	Reason         string   `json:"reason"`
}

func RunPseudoControllerRebuildProof(opts PseudoControllerRebuildOptions) (PseudoControllerRebuildProof, error) {
	if opts.InventoryPath == "" {
		opts.InventoryPath = filepath.Join("tmp", "effect_field_inventory", "inventory.json")
	}
	if opts.UnderstandingPath == "" {
		opts.UnderstandingPath = filepath.Join("tmp", "effect_field_understanding", "understanding.json")
	}
	if opts.OutDir == "" {
		opts.OutDir = filepath.Join("tmp", "pseudo_controller_rebuild")
	}
	inventoryData, err := os.ReadFile(opts.InventoryPath)
	if err != nil {
		return PseudoControllerRebuildProof{}, err
	}
	var inventory EffectFieldInventory
	if err := json.Unmarshal(inventoryData, &inventory); err != nil {
		return PseudoControllerRebuildProof{}, fmt.Errorf("parse pseudo controller inventory %s: %w", opts.InventoryPath, err)
	}
	understandingData, err := os.ReadFile(opts.UnderstandingPath)
	if err != nil {
		return PseudoControllerRebuildProof{}, err
	}
	var understanding EffectFieldUnderstanding
	if err := json.Unmarshal(understandingData, &understanding); err != nil {
		return PseudoControllerRebuildProof{}, fmt.Errorf("parse pseudo controller understanding %s: %w", opts.UnderstandingPath, err)
	}
	proof, err := BuildPseudoControllerRebuildProof(inventory, understanding, opts)
	if err != nil {
		return PseudoControllerRebuildProof{}, err
	}
	proof.SourceInventory = opts.InventoryPath
	proof.SourceUnderstanding = opts.UnderstandingPath
	proof.OutputPath = filepath.Join(opts.OutDir, "proof.json")
	if err := os.MkdirAll(filepath.Join(opts.OutDir, "generated"), 0o755); err != nil {
		return PseudoControllerRebuildProof{}, err
	}
	for i := range proof.Families {
		if len(proof.Families[i].Controls) == 0 || len(proof.Families[i].UnsupportedControls) != 0 {
			continue
		}
		outPath := filepath.Join(opts.OutDir, "generated", safePseudoFamilySlug(proof.Families[i].UID)+".aep")
		if err := writePseudoControllerProofAEP(outPath, proof.Families[i]); err != nil {
			return PseudoControllerRebuildProof{}, err
		}
		proof.Families[i].GeneratedAEP = outPath
		proof.Families[i].Status = "generated"
		proof.Summary.GeneratedFamilies++
		proof.Summary.GeneratedControls += len(proof.Families[i].Controls)
	}
	if err := writeIndentedJSON(proof.OutputPath, proof); err != nil {
		return PseudoControllerRebuildProof{}, err
	}
	return proof, nil
}

func BuildPseudoControllerRebuildProof(inventory EffectFieldInventory, understanding EffectFieldUnderstanding, opts PseudoControllerRebuildOptions) (PseudoControllerRebuildProof, error) {
	if opts.MaxFamilies <= 0 {
		opts.MaxFamilies = 10
	}
	understandingByMatch := map[string]EffectFieldUnderstandingEffect{}
	for _, effect := range understanding.Effects {
		understandingByMatch[effect.MatchName] = effect
	}
	pseudoEffects := make([]EffectFieldInventoryEffect, 0)
	for _, effect := range inventory.Effects {
		if effect.Class == "pseudo" {
			pseudoEffects = append(pseudoEffects, effect)
		}
	}
	sort.Slice(pseudoEffects, func(i, j int) bool {
		if pseudoEffects[i].Occurrences != pseudoEffects[j].Occurrences {
			return pseudoEffects[i].Occurrences > pseudoEffects[j].Occurrences
		}
		if pseudoEffects[i].ParamKinds != pseudoEffects[j].ParamKinds {
			return pseudoEffects[i].ParamKinds > pseudoEffects[j].ParamKinds
		}
		return pseudoEffects[i].MatchName < pseudoEffects[j].MatchName
	})
	proof := PseudoControllerRebuildProof{
		SchemaVersion: 1,
		Summary: PseudoControllerRebuildSummary{
			PseudoFamilies: len(pseudoEffects),
		},
		Boundaries: []string{
			"Pseudo fields are parseable controller evidence, not a full expression or rig-behavior reconstruction.",
			"Generated controls use stable labels derived from parameter match names when source display labels are unavailable.",
			"Keyframes and expressions are recorded as behavior notes for a later wiring package.",
		},
	}
	limit := opts.MaxFamilies
	if limit > len(pseudoEffects) {
		limit = len(pseudoEffects)
	}
	for i := 0; i < limit; i++ {
		family := buildPseudoControllerRebuildFamily(pseudoEffects[i], understandingByMatch[pseudoEffects[i].MatchName])
		proof.Families = append(proof.Families, family)
		proof.Summary.SystemParamCount += family.SystemParamCount
		proof.Summary.UnsupportedControlCount += len(family.UnsupportedControls)
		if len(family.UnsupportedControls) > 0 {
			proof.Summary.UnsupportedFamilies++
		}
	}
	proof.Summary.SelectedFamilies = len(proof.Families)
	return proof, nil
}

func buildPseudoControllerRebuildFamily(effect EffectFieldInventoryEffect, understanding EffectFieldUnderstandingEffect) PseudoControllerRebuildFamily {
	family := PseudoControllerRebuildFamily{
		MatchName:        effect.MatchName,
		UID:              pseudoFamilyUID(effect.MatchName),
		Occurrences:      effect.Occurrences,
		ParamKinds:       effect.ParamKinds,
		ParamOccurrences: effect.ParamOccurrences,
		ProjectSamples:   append([]string(nil), effect.ProjectSamples...),
		Reproducibility:  understanding.Reproducibility,
		GenerationPolicy: understanding.GenerationPolicy,
		StudyPriority:    understanding.StudyPriority,
		Status:           "planned",
		Boundary:         "Control structure is rebuildable; expressions, keyframes, and downstream layer wiring remain separate behavior work.",
	}
	params := append([]EffectFieldInventoryParam(nil), effect.Params...)
	sort.Slice(params, func(i, j int) bool { return params[i].MatchName < params[j].MatchName })
	for _, param := range params {
		if isPseudoSystemParam(effect.MatchName, param.MatchName) {
			family.SystemParamCount++
			continue
		}
		control, unsupported := mapPseudoControllerParam(param)
		if unsupported.Reason != "" {
			family.UnsupportedControls = append(family.UnsupportedControls, unsupported)
			continue
		}
		family.Controls = append(family.Controls, control)
	}
	if len(family.Controls) == 0 && len(family.UnsupportedControls) > 0 {
		family.Status = "blocked"
	}
	return family
}

func mapPseudoControllerParam(param EffectFieldInventoryParam) (PseudoControllerControlPlan, PseudoControllerUnsupportedControl) {
	valueType := primaryCountName(param.ValueTypes)
	control := PseudoControllerControlPlan{
		ParamMatchName: param.MatchName,
		Label:          pseudoParamLabel(param.MatchName),
		ValueType:      valueType,
		Occurrences:    param.Occurrences,
		DefaultValue:   firstExampleValue(param.ExampleStaticValues),
	}
	switch valueType {
	case "layer_ref":
		control.ControlKind = "layer"
	case "bool":
		control.ControlKind = "checkbox"
	case "vector2":
		control.ControlKind = "point"
	case "vector3":
		control.ControlKind = "point3d"
	case "vector4":
		control.ControlKind = "color"
	case "number":
		control.ControlKind = "slider"
	case "none":
		if param.KeyframedOccurrences > 0 {
			control.ControlKind = "slider"
			control.BehaviorNotes = append(control.BehaviorNotes, "keyframed")
		}
	default:
		return PseudoControllerControlPlan{}, PseudoControllerUnsupportedControl{
			ParamMatchName: param.MatchName,
			ValueTypes:     countRowNames(param.ValueTypes),
			Reason:         "no BuildPseudoEffect control mapping for value type " + valueType,
		}
	}
	if control.ControlKind == "" {
		return PseudoControllerControlPlan{}, PseudoControllerUnsupportedControl{
			ParamMatchName: param.MatchName,
			ValueTypes:     countRowNames(param.ValueTypes),
			Reason:         "parameter has no static value and no keyframe evidence",
		}
	}
	if param.KeyframedOccurrences > 0 && !stringSliceContains(control.BehaviorNotes, "keyframed") {
		control.BehaviorNotes = append(control.BehaviorNotes, "keyframed")
	}
	if param.ExpressionOccurrences > 0 {
		control.BehaviorNotes = append(control.BehaviorNotes, "expression")
	}
	return control, PseudoControllerUnsupportedControl{}
}

func writePseudoControllerProofAEP(path string, family PseudoControllerRebuildFamily) error {
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Pseudo Controller Proof", 640, 360, 30, 5)
	if err != nil {
		return err
	}
	if _, err := aep.NewShapeLayer(comp, "Controller Host"); err != nil {
		return err
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return err
	}
	var host *aep.Layer
	for _, comp := range reopened.Compositions {
		for _, layer := range comp.Layers {
			if layer.Name == "Controller Host" {
				host = layer
				break
			}
		}
	}
	if host == nil {
		return fmt.Errorf("generated pseudo proof host layer missing")
	}
	controls := make([]aep.PseudoControl, 0, len(family.Controls))
	for _, plan := range family.Controls {
		controls = append(controls, pseudoControlFromPlan(plan))
	}
	if _, err := aep.BuildPseudoEffect(host, family.UID, "Rebuild", family.UID+" Rebuild", controls); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := reopened.WriteAEP(file); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	parsed, err := aep.Open(path)
	if err != nil {
		return err
	}
	if len(parsed.Warnings) > 0 {
		return fmt.Errorf("generated pseudo proof reparsed with warnings: %v", parsed.Warnings)
	}
	return nil
}

func pseudoControlFromPlan(plan PseudoControllerControlPlan) aep.PseudoControl {
	control := aep.PseudoControl{Name: plan.Label}
	switch plan.ControlKind {
	case "layer":
		control.Kind = aep.PseudoLayer
	case "checkbox":
		control.Kind = aep.PseudoCheckbox
		control.Checked = boolAny(plan.DefaultValue)
	case "point":
		control.Kind = aep.PseudoPoint
		control.PointDefault = vectorDefault(plan.DefaultValue, 2)
	case "point3d":
		control.Kind = aep.PseudoPoint3D
		control.PointDefault = vectorDefault(plan.DefaultValue, 3)
	case "color":
		control.Kind = aep.PseudoColor
		control.Color = colorDefault(plan.DefaultValue)
	default:
		control.Kind = aep.PseudoSlider
		value := floatAny(plan.DefaultValue)
		control.Min = 0
		control.Max = 100
		if value < 0 {
			control.Min = value
		}
		if value > 100 {
			control.Max = value
		}
		control.Default = value
	}
	return control
}

func isPseudoSystemParam(effectMatchName, paramMatchName string) bool {
	return paramMatchName == effectMatchName+"-0000"
}

func pseudoFamilyUID(matchName string) string {
	return strings.TrimPrefix(matchName, "Pseudo/")
}

func safePseudoFamilySlug(uid string) string {
	slug := strings.ToLower(uid)
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "pseudo-controller"
	}
	return slug
}

func pseudoParamLabel(matchName string) string {
	idx := matchName
	if dash := strings.LastIndex(matchName, "-"); dash >= 0 && dash+1 < len(matchName) {
		idx = matchName[dash+1:]
	}
	if n, err := strconv.Atoi(idx); err == nil {
		return fmt.Sprintf("Control %d", n)
	}
	return idx
}

func primaryCountName(rows []CountRow) string {
	if len(rows) == 0 {
		return ""
	}
	best := rows[0]
	for _, row := range rows[1:] {
		if row.Count > best.Count || row.Count == best.Count && row.Name < best.Name {
			best = row
		}
	}
	return best.Name
}

func countRowNames(rows []CountRow) []string {
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Name)
	}
	sort.Strings(names)
	return names
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func firstExampleValue(values []any) any {
	if len(values) == 0 {
		return nil
	}
	return values[0]
}

func floatAny(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		f, _ := v.Float64()
		return f
	default:
		return 0
	}
}

func boolAny(value any) bool {
	v, _ := value.(bool)
	return v
}

func vectorDefault(value any, dims int) []float64 {
	out := make([]float64, dims)
	switch v := value.(type) {
	case []float64:
		for i := 0; i < len(v) && i < dims; i++ {
			out[i] = v[i]
		}
	case []any:
		for i := 0; i < len(v) && i < dims; i++ {
			out[i] = floatAny(v[i])
		}
	}
	return out
}

func colorDefault(value any) []float64 {
	color := vectorDefault(value, 4)
	if color[3] == 0 {
		color[3] = 1
	}
	return color
}
