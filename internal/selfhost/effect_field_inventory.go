package selfhost

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

type EffectFieldInventoryOptions struct {
	Root    string
	OutPath string
	Limit   int
}

type EffectFieldInventory struct {
	SchemaVersion int                           `json:"schema_version"`
	Root          string                        `json:"root,omitempty"`
	OutputPath    string                        `json:"output_path,omitempty"`
	Summary       EffectFieldInventorySummary   `json:"summary"`
	Effects       []EffectFieldInventoryEffect  `json:"effects"`
	Failures      []EffectFieldInventoryFailure `json:"failures,omitempty"`
}

type EffectFieldInventorySummary struct {
	ProjectCount      int        `json:"project_count"`
	FailedProjects    int        `json:"failed_projects"`
	EffectKinds       int        `json:"effect_kinds"`
	EffectOccurrences int        `json:"effect_occurrences"`
	ParamKinds        int        `json:"param_kinds"`
	ParamOccurrences  int        `json:"param_occurrences"`
	Classes           []CountRow `json:"classes,omitempty"`
	Capabilities      []CountRow `json:"capabilities,omitempty"`
}

type EffectFieldInventoryEffect struct {
	MatchName            string                       `json:"match_name"`
	Class                string                       `json:"class"`
	DisplayNames         []CountRow                   `json:"display_names,omitempty"`
	DependencyClasses    []CountRow                   `json:"dependency_classes,omitempty"`
	Occurrences          int                          `json:"occurrences"`
	ProjectSamples       []string                     `json:"project_samples,omitempty"`
	ParamKinds           int                          `json:"param_kinds"`
	ParamOccurrences     int                          `json:"param_occurrences"`
	InferredCapabilities []string                     `json:"inferred_capabilities,omitempty"`
	Params               []EffectFieldInventoryParam  `json:"params,omitempty"`
	UnknownParams        []EffectFieldInventoryReason `json:"unknown_params,omitempty"`
}

type EffectFieldInventoryParam struct {
	MatchName              string     `json:"match_name"`
	Names                  []CountRow `json:"names,omitempty"`
	ValueTypes             []CountRow `json:"value_types,omitempty"`
	Occurrences            int        `json:"occurrences"`
	StaticValueOccurrences int        `json:"static_value_occurrences,omitempty"`
	DefaultOccurrences     int        `json:"default_occurrences,omitempty"`
	ChangedOccurrences     int        `json:"changed_occurrences,omitempty"`
	KeyframedOccurrences   int        `json:"keyframed_occurrences,omitempty"`
	ExpressionOccurrences  int        `json:"expression_occurrences,omitempty"`
	LayerRefOccurrences    int        `json:"layer_ref_occurrences,omitempty"`
	GradientOccurrences    int        `json:"gradient_occurrences,omitempty"`
	ExampleStaticValues    []any      `json:"example_static_values,omitempty"`
}

type EffectFieldInventoryReason struct {
	Path   string `json:"path,omitempty"`
	Reason string `json:"reason"`
}

type EffectFieldInventoryFailure struct {
	InputPath string `json:"input_path"`
	Error     string `json:"error"`
}

func RunEffectFieldInventory(opts EffectFieldInventoryOptions) (EffectFieldInventory, error) {
	if opts.Root == "" {
		opts.Root = filepath.Join("data", "samples")
	}
	if opts.OutPath == "" {
		opts.OutPath = filepath.Join("tmp", "effect_field_inventory", "inventory.json")
	}
	inputs, err := DiscoverSampleShellInputs(opts.Root, opts.Limit)
	if err != nil {
		return EffectFieldInventory{}, err
	}
	profiles := make([]*profile.Profile, 0, len(inputs))
	failures := make([]EffectFieldInventoryFailure, 0)
	for _, input := range inputs {
		prof, err := openProfile(input)
		if err != nil {
			failures = append(failures, EffectFieldInventoryFailure{InputPath: input, Error: err.Error()})
			continue
		}
		profiles = append(profiles, prof)
	}
	inventory := BuildEffectFieldInventory(profiles)
	inventory.Root = opts.Root
	inventory.OutputPath = opts.OutPath
	inventory.Failures = failures
	inventory.Summary.FailedProjects = len(failures)
	if err := os.MkdirAll(filepath.Dir(opts.OutPath), 0o755); err != nil {
		return EffectFieldInventory{}, err
	}
	if err := writeIndentedJSON(opts.OutPath, inventory); err != nil {
		return EffectFieldInventory{}, err
	}
	return inventory, nil
}

func BuildEffectFieldInventory(profiles []*profile.Profile) EffectFieldInventory {
	builder := newEffectFieldInventoryBuilder()
	for _, prof := range profiles {
		builder.addProfile(prof)
	}
	return builder.build()
}

type effectFieldInventoryBuilder struct {
	projectCount int
	effects      map[string]*effectFieldAccumulator
}

type effectFieldAccumulator struct {
	matchName         string
	class             string
	displayNames      map[string]int
	dependencyClasses map[string]int
	occurrences       int
	projectSamples    map[string]struct{}
	params            map[string]*effectFieldParamAccumulator
	unknownParams     []EffectFieldInventoryReason
	capabilities      map[string]struct{}
}

type effectFieldParamAccumulator struct {
	matchName              string
	names                  map[string]int
	valueTypes             map[string]int
	occurrences            int
	staticValueOccurrences int
	defaultOccurrences     int
	changedOccurrences     int
	keyframedOccurrences   int
	expressionOccurrences  int
	layerRefOccurrences    int
	gradientOccurrences    int
	exampleStaticValues    []any
}

func newEffectFieldInventoryBuilder() *effectFieldInventoryBuilder {
	return &effectFieldInventoryBuilder{effects: map[string]*effectFieldAccumulator{}}
}

func (b *effectFieldInventoryBuilder) addProfile(prof *profile.Profile) {
	if prof == nil {
		return
	}
	b.projectCount++
	projectPath := prof.Meta.Path
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			for _, effect := range layer.Effects {
				b.addEffect(projectPath, effect)
			}
		}
	}
}

func (b *effectFieldInventoryBuilder) addEffect(projectPath string, effect profile.Effect) {
	matchName := effect.MatchName
	if matchName == "" {
		matchName = "(unknown)"
	}
	acc := b.effects[matchName]
	if acc == nil {
		acc = &effectFieldAccumulator{
			matchName:         matchName,
			class:             classifyEffectFieldInventoryEffect(matchName),
			displayNames:      map[string]int{},
			dependencyClasses: map[string]int{},
			projectSamples:    map[string]struct{}{},
			params:            map[string]*effectFieldParamAccumulator{},
			capabilities:      map[string]struct{}{},
		}
		b.effects[matchName] = acc
	}
	acc.occurrences++
	addNonEmptyCount(acc.displayNames, effect.DisplayName)
	addNonEmptyCount(acc.dependencyClasses, effect.DependencyClass)
	if projectPath != "" && len(acc.projectSamples) < 5 {
		acc.projectSamples[projectPath] = struct{}{}
	}
	for _, capability := range inferEffectFieldCapabilities(matchName, effect.DisplayName, "") {
		acc.capabilities[capability] = struct{}{}
	}
	for _, param := range effect.Params {
		acc.addParam(param)
	}
	for _, unknown := range effect.UnknownParams {
		acc.unknownParams = append(acc.unknownParams, EffectFieldInventoryReason{Path: unknown.Path, Reason: unknown.Reason})
	}
}

func (acc *effectFieldAccumulator) addParam(param profile.Property) {
	matchName := param.MatchName
	if matchName == "" {
		matchName = param.Name
	}
	if matchName == "" {
		matchName = "(unknown)"
	}
	p := acc.params[matchName]
	if p == nil {
		p = &effectFieldParamAccumulator{
			matchName:  matchName,
			names:      map[string]int{},
			valueTypes: map[string]int{},
		}
		acc.params[matchName] = p
	}
	p.occurrences++
	addNonEmptyCount(p.names, param.Name)
	addNonEmptyCount(p.valueTypes, effectFieldValueType(param))
	if param.StaticValue != nil {
		p.staticValueOccurrences++
		if len(p.exampleStaticValues) < 5 {
			p.exampleStaticValues = append(p.exampleStaticValues, effectFieldSampleValue(param.StaticValue))
		}
	}
	if param.Default != nil {
		p.defaultOccurrences++
	}
	if param.Changed {
		p.changedOccurrences++
	}
	if len(param.Keyframes) > 0 {
		p.keyframedOccurrences++
		acc.capabilities["keyframes"] = struct{}{}
	}
	if param.Expression != "" || param.ExpressionEnabled != nil {
		p.expressionOccurrences++
		acc.capabilities["expression"] = struct{}{}
	}
	if param.LayerRef != nil {
		p.layerRefOccurrences++
		acc.capabilities["layer_ref"] = struct{}{}
	}
	if param.Gradient != nil {
		p.gradientOccurrences++
		acc.capabilities["gradient"] = struct{}{}
	}
	for _, capability := range inferEffectFieldCapabilities(param.MatchName, param.Name, param.Expression, effectFieldValueType(param)) {
		acc.capabilities[capability] = struct{}{}
	}
}

func (b *effectFieldInventoryBuilder) build() EffectFieldInventory {
	inventory := EffectFieldInventory{
		SchemaVersion: 1,
		Summary: EffectFieldInventorySummary{
			ProjectCount: b.projectCount,
			EffectKinds:  len(b.effects),
		},
		Effects: []EffectFieldInventoryEffect{},
	}
	classCounts := map[string]int{}
	capabilityCounts := map[string]int{}
	for _, acc := range b.effects {
		effect := acc.build()
		inventory.Effects = append(inventory.Effects, effect)
		inventory.Summary.EffectOccurrences += effect.Occurrences
		inventory.Summary.ParamKinds += effect.ParamKinds
		inventory.Summary.ParamOccurrences += effect.ParamOccurrences
		classCounts[effect.Class] += effect.Occurrences
		for _, capability := range effect.InferredCapabilities {
			capabilityCounts[capability] += effect.Occurrences
		}
	}
	sort.Slice(inventory.Effects, func(i, j int) bool {
		if inventory.Effects[i].Occurrences != inventory.Effects[j].Occurrences {
			return inventory.Effects[i].Occurrences > inventory.Effects[j].Occurrences
		}
		return inventory.Effects[i].MatchName < inventory.Effects[j].MatchName
	})
	inventory.Summary.Classes = countMapRows(classCounts)
	inventory.Summary.Capabilities = countMapRows(capabilityCounts)
	return inventory
}

func (acc *effectFieldAccumulator) build() EffectFieldInventoryEffect {
	effect := EffectFieldInventoryEffect{
		MatchName:            acc.matchName,
		Class:                acc.class,
		DisplayNames:         countMapRows(acc.displayNames),
		DependencyClasses:    countMapRows(acc.dependencyClasses),
		Occurrences:          acc.occurrences,
		ProjectSamples:       sortedStringSet(acc.projectSamples),
		Params:               []EffectFieldInventoryParam{},
		UnknownParams:        append([]EffectFieldInventoryReason(nil), acc.unknownParams...),
		InferredCapabilities: sortedStringSet(acc.capabilities),
	}
	for _, param := range acc.params {
		p := param.build()
		effect.Params = append(effect.Params, p)
		effect.ParamOccurrences += p.Occurrences
	}
	sort.Slice(effect.Params, func(i, j int) bool {
		if effect.Params[i].Occurrences != effect.Params[j].Occurrences {
			return effect.Params[i].Occurrences > effect.Params[j].Occurrences
		}
		return effect.Params[i].MatchName < effect.Params[j].MatchName
	})
	effect.ParamKinds = len(effect.Params)
	return effect
}

func (acc *effectFieldParamAccumulator) build() EffectFieldInventoryParam {
	return EffectFieldInventoryParam{
		MatchName:              acc.matchName,
		Names:                  countMapRows(acc.names),
		ValueTypes:             countMapRows(acc.valueTypes),
		Occurrences:            acc.occurrences,
		StaticValueOccurrences: acc.staticValueOccurrences,
		DefaultOccurrences:     acc.defaultOccurrences,
		ChangedOccurrences:     acc.changedOccurrences,
		KeyframedOccurrences:   acc.keyframedOccurrences,
		ExpressionOccurrences:  acc.expressionOccurrences,
		LayerRefOccurrences:    acc.layerRefOccurrences,
		GradientOccurrences:    acc.gradientOccurrences,
		ExampleStaticValues:    append([]any(nil), acc.exampleStaticValues...),
	}
}

func classifyEffectFieldInventoryEffect(matchName string) string {
	switch ClassifySampleShellUnsupportedEffect(matchName) {
	case "supported_native_pending":
		return "native_supported"
	case "native_template_gap":
		return "native_template_gap"
	case "pseudo":
		return "pseudo"
	case "third_party":
		return "third_party"
	default:
		return "unknown_or_alias"
	}
}

func effectFieldValueType(param profile.Property) string {
	switch {
	case param.LayerRef != nil:
		return "layer_ref"
	case param.Gradient != nil:
		return "gradient"
	case param.StaticValue == nil:
		return "none"
	default:
		return effectFieldScalarType(param.StaticValue)
	}
}

func effectFieldScalarType(value any) string {
	switch v := value.(type) {
	case bool:
		return "bool"
	case string:
		return "string"
	case float64, float32, int, int64, int32, uint, uint32, uint64:
		return "number"
	case []float64:
		return effectFieldVectorType(len(v))
	case []int:
		return effectFieldVectorType(len(v))
	case []any:
		return effectFieldVectorType(len(v))
	case [2]float64:
		return "vector2"
	case [3]float64:
		return "vector3"
	case [4]float64:
		return "vector4"
	default:
		return fmt.Sprintf("%T", value)
	}
}

func effectFieldVectorType(n int) string {
	switch n {
	case 2:
		return "vector2"
	case 3:
		return "vector3"
	case 4:
		return "vector4"
	default:
		return fmt.Sprintf("vector%d", n)
	}
}

func effectFieldSampleValue(value any) any {
	switch v := value.(type) {
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil
		}
		return v
	case float32:
		n := float64(v)
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return nil
		}
		return n
	case []float64:
		out := make([]float64, 0, len(v))
		for _, n := range v {
			if math.IsNaN(n) || math.IsInf(n, 0) {
				out = append(out, 0)
				continue
			}
			out = append(out, n)
		}
		return out
	default:
		return value
	}
}

func inferEffectFieldCapabilities(parts ...string) []string {
	text := strings.ToLower(strings.Join(parts, " "))
	capabilities := map[string]struct{}{}
	keywordCapabilities := []struct {
		keyword    string
		capability string
	}{
		{"color", "color"},
		{"colour", "color"},
		{"opacity", "opacity"},
		{"position", "position"},
		{"point", "position"},
		{"center", "position"},
		{"anchor", "position"},
		{"scale", "scale"},
		{"rotation", "rotation"},
		{"angle", "rotation"},
		{"blur", "blur"},
		{"radius", "radius"},
		{"amount", "amount"},
		{"intensity", "amount"},
		{"slider", "slider"},
		{"checkbox", "boolean"},
		{"layer", "layer_ref"},
		{"time", "time"},
		{"seed", "random_seed"},
		{"random", "random_seed"},
		{"gradient", "gradient"},
		{"blend", "blend"},
		{"vector2", "position"},
		{"vector3", "position"},
		{"vector4", "color"},
		{"number", "scalar"},
		{"bool", "boolean"},
	}
	for _, item := range keywordCapabilities {
		if strings.Contains(text, item.keyword) {
			capabilities[item.capability] = struct{}{}
		}
	}
	return sortedStringSet(capabilities)
}

func addNonEmptyCount(counts map[string]int, value string) {
	if value != "" {
		counts[value]++
	}
}

func sortedStringSet(set map[string]struct{}) []string {
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}
