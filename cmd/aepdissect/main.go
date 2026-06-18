// cmd/aepdissect — structured teardown of any After Effects .aep, for the
// technique-internalization pipeline (spec 2026-06-18-fx-technique-internalization,
// step PARSE; output shape per spec 2026-06-18-technique-ontology). Given a
// reference template, print
//   - effect-usage histogram with native / Cycore-bundled / third-party class
//   - third-party plugin dependency list (what the output needs installed to render)
//   - precomp nesting map
//   - per-comp layer detail: blend mode (named), source, mask/effect counts,
//     and each effect's SET or ANIMATED params, TRANSLATED via the effect
//     dictionary (matchName -> human name) and flagged against the AE default
//     so the params the author actually CHANGED stand out (= recipe signal;
//     AE elides params equal to default, so a stored param ~= an author decision).
//
// Usage:
//   go run ./cmd/aepdissect [-dict <effects.json>] <file.aep>
//
// Dictionary defaults to data/effects-dict/effects_en_US_25.1x68.json (run from
// repo root). Missing dict -> falls back to raw matchName output. Build the dict
// with scripts/dump_effects_dict.ps1 (see checklists/effects-dict.md).
//
// This is the mechanical step; turning the dump into role/technique knowledge is
// human/AI judgment (see docs/fx-techniques.md + the technique-ontology schema).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"

	aep "github.com/example/aep-parser/internal/aep"
)

var blendName = map[int]string{
	0: "NormalCamera", 2: "Normal", 3: "Dissolve", 4: "Add", 5: "Multiply",
	6: "Screen", 7: "Overlay", 8: "SoftLight", 9: "HardLight", 10: "Darken",
	11: "Lighten", 12: "ClassicDiff", 13: "Hue", 14: "Saturation", 15: "Color",
	16: "Luminosity", 17: "StencilAlpha", 18: "StencilLuma", 19: "SilhouetteAlpha",
	20: "SilhouetteLuma", 21: "LuminescentPremul", 22: "AlphaAdd", 23: "ClassicColorDodge",
	24: "ClassicColorBurn", 25: "Exclusion", 26: "Difference", 27: "ColorDodge",
	28: "ColorBurn", 29: "LinearDodge", 30: "LinearBurn", 31: "LinearLight",
	32: "VividLight", 33: "PinLight", 34: "HardMix", 35: "LighterColor",
	36: "DarkerColor", 37: "Subtract", 38: "Divide",
}

func blend(m int) string {
	if n, ok := blendName[m]; ok {
		return n
	}
	return fmt.Sprintf("blend#%d", m)
}

// classify an effect matchName by render dependency.
func classify(mn string) string {
	switch {
	case strings.HasPrefix(mn, "ADBE "):
		return "native"
	case strings.HasPrefix(mn, "CC "):
		return "Cycore(bundled)"
	default:
		return "⚠THIRD-PARTY"
	}
}

// --- effect dictionary (data/effects-dict/effects_<lang>_<ver>.json) ---

type dictParam struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Default any    `json:"default"`
}
type dictEffect struct {
	Name   string               `json:"name"`
	Params map[string]dictParam `json:"params"`
}
type effectsDict struct {
	AEVersion string                `json:"aeVersion"`
	Effects   map[string]dictEffect `json:"effects"`
}

func loadDict(path string) *effectsDict {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: no effect dict (%v) — falling back to raw matchName output\n", err)
		return nil
	}
	var d effectsDict
	if err := json.Unmarshal(b, &d); err != nil {
		fmt.Fprintf(os.Stderr, "warn: dict parse failed (%v) — raw output\n", err)
		return nil
	}
	return &d
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

func toFloatSlice(v any) ([]float64, bool) {
	switch s := v.(type) {
	case []float64:
		return s, true
	case []any:
		out := make([]float64, 0, len(s))
		for _, e := range s {
			f, ok := toFloat(e)
			if !ok {
				return nil, false
			}
			out = append(out, f)
		}
		return out, true
	}
	return nil, false
}

// equalsDefault reports whether stored value v matches the dict default def.
// ok=false means "not comparable" (no default / unhandled type) — treat as changed.
func equalsDefault(v, def any) (equal, ok bool) {
	if def == nil {
		return false, false
	}
	switch d := def.(type) {
	case float64:
		if f, ok2 := toFloat(v); ok2 {
			return math.Abs(f-d) < 1e-6, true
		}
	case []any, []float64:
		ds, _ := toFloatSlice(def)
		vs, ok2 := toFloatSlice(v)
		_ = d
		if !ok2 || len(vs) != len(ds) {
			return false, len(vs) > 0
		}
		for i := range ds {
			if math.Abs(vs[i]-ds[i]) > 1e-6 {
				return false, true
			}
		}
		return true, true
	case string:
		return fmt.Sprint(v) == d, true
	case bool:
		if b, ok2 := v.(bool); ok2 {
			return b == d, true
		}
	}
	return false, false
}

func fmtNum(v any) string {
	if f, ok := toFloat(v); ok {
		return fmt.Sprintf("%g", f)
	}
	return fmt.Sprintf("%v", v)
}

// --- timeline / motion (keyframe reading) ---

var xformName = map[string]string{
	"ADBE Position": "Position", "ADBE Anchor Point": "Anchor Point",
	"ADBE Scale": "Scale", "ADBE Rotate Z": "Rotation", "ADBE Opacity": "Opacity",
	"ADBE Rotate X": "X Rotation", "ADBE Rotate Y": "Y Rotation",
	"ADBE Orientation": "Orientation", "ADBE Position_0": "X Position",
	"ADBE Position_1": "Y Position", "ADBE Position_2": "Z Position",
	"ADBE Time Remapping": "Time Remap",
}

func propName(mn string) string {
	if n, ok := xformName[mn]; ok {
		return n
	}
	return mn
}

func kfSpan(kfs []*aep.Keyframe) (lo, hi float64) {
	if len(kfs) == 0 {
		return 0, 0
	}
	lo, hi = kfs[0].Time, kfs[0].Time
	for _, k := range kfs {
		if k.Time < lo {
			lo = k.Time
		}
		if k.Time > hi {
			hi = k.Time
		}
	}
	return lo, hi
}

// motionLabel classifies a property's keyframe interpolation into a motion
// technique tag (the "运动语言" — readable straight from the time table, no render).
func motionLabel(kfs []*aep.Keyframe) string {
	var hold, lin, bez int
	eased := false
	for _, k := range kfs {
		for _, it := range []aep.InterpType{k.InInterp, k.OutInterp} {
			switch it {
			case aep.InterpHold:
				hold++
			case aep.InterpLinear:
				lin++
			case aep.InterpBezier:
				bez++
			}
		}
		for _, te := range k.InTemporalEase {
			if te.Influence > 0.2 {
				eased = true
			}
		}
		for _, te := range k.OutTemporalEase {
			if te.Influence > 0.2 {
				eased = true
			}
		}
	}
	switch {
	case hold > 0 && bez == 0 && lin == 0:
		return "hold"
	case bez > 0 && eased:
		return "ease"
	case bez > 0:
		return "bezier"
	case lin > 0:
		return "linear"
	default:
		return "?"
	}
}

// walkAnimated collects leaf properties carrying keyframes from a property tree
// (Transform group etc; effect params live elsewhere and are handled separately).
func walkAnimated(g *aep.AEPropertyGroup, out *[]*aep.Property) {
	if g == nil {
		return
	}
	for _, c := range g.Children {
		switch n := c.(type) {
		case *aep.Property:
			if len(n.Keyframes) > 0 {
				*out = append(*out, n)
			}
		case *aep.AEPropertyGroup:
			walkAnimated(n, out)
		}
	}
}

// --- dependency edges ---

var exprRefRe = regexp.MustCompile(`(?:effect|comp|layer|footage)\(\s*"([^"]+)"`)

// exprRefs extracts the quoted names an expression references (effect("X"),
// thisComp.layer("Y"), comp("Z")...) — the property→property dependency edges.
func exprRefs(s string) []string {
	if s == "" {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, m := range exprRefRe.FindAllStringSubmatch(s, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

// --- structured profile (project_profile per technique-ontology spec §3) ---

type jParam struct {
	MatchName string `json:"matchName"`
	Name      string `json:"name,omitempty"`
	Value     any    `json:"value,omitempty"`
	Default   any    `json:"default,omitempty"`
	Changed   bool   `json:"changed"`
	Expr      string `json:"expr,omitempty"`
}
type jEffect struct {
	MatchName string   `json:"matchName"`
	Name      string   `json:"name,omitempty"`
	Class     string   `json:"class"`
	Tuned     []string `json:"tuned,omitempty"`
	Params    []jParam `json:"params,omitempty"`
}
type jAnim struct {
	Property string  `json:"property"`
	KF       int     `json:"kf"`
	Motion   string  `json:"motion"`
	Start    float64 `json:"start"`
	End      float64 `json:"end"`
}
type jLayer struct {
	Index    int       `json:"index"`
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Blend    string    `json:"blend"`
	Visible  bool      `json:"visible"`
	Source   string    `json:"source,omitempty"`
	InPoint  float64   `json:"inPoint"`
	OutPoint float64   `json:"outPoint"`
	Parent   int       `json:"parent,omitempty"` // parent layer Index, 0 = none
	Matte    int       `json:"matte,omitempty"`  // matte source layer Index
	Effects  []jEffect `json:"effects,omitempty"`
	Animated []jAnim   `json:"animated,omitempty"`
	ExprRefs []string  `json:"exprRefs,omitempty"`
}
type jComp struct {
	Name     string   `json:"name"`
	ID       uint32   `json:"id"`
	Width    int      `json:"width"`
	Height   int      `json:"height"`
	FPS      float64  `json:"fps"`
	Duration float64  `json:"duration"`
	Layers   []jLayer `json:"layers"`
}
type jProfile struct {
	Project     string         `json:"project"`
	EffectUsage map[string]int `json:"effectUsage"`
	ThirdParty  []string       `json:"thirdParty,omitempty"`
	Comps       []jComp        `json:"comps"`
}

func buildProfile(path string, p *aep.Project, dict *effectsDict, compName map[uint32]string) jProfile {
	prof := jProfile{Project: path, EffectUsage: map[string]int{}}
	tpSeen := map[string]bool{}
	for _, c := range p.Compositions {
		jc := jComp{Name: c.Name, ID: c.ID, Width: int(c.Width), Height: int(c.Height), FPS: c.FrameRate, Duration: c.Duration}
		idIndex := map[uint32]int{}
		for _, l := range c.Layers {
			idIndex[l.ID] = l.Index
		}
		for _, l := range c.Layers {
			jl := jLayer{
				Index: l.Index, Name: l.Name, Type: fmt.Sprintf("%v", l.Type),
				Blend: blend(int(l.BlendingMode)), Visible: l.Visible,
				InPoint: l.InPoint(), OutPoint: l.OutPoint(),
			}
			if l.SourceID != 0 {
				if nm, ok := compName[l.SourceID]; ok {
					jl.Source = "precomp:" + nm
				} else {
					jl.Source = fmt.Sprintf("footage:%d", l.SourceID)
				}
			}
			if l.ParentID != 0 {
				jl.Parent = idIndex[l.ParentID]
			}
			if l.TrackMatteLayerID != 0 {
				jl.Matte = idIndex[l.TrackMatteLayerID]
			}
			var ap []*aep.Property
			walkAnimated(l.PropertyTree(), &ap)
			for _, pr := range ap {
				t0, t1 := kfSpan(pr.Keyframes)
				jl.Animated = append(jl.Animated, jAnim{
					Property: propName(pr.MatchName), KF: len(pr.Keyframes),
					Motion: motionLabel(pr.Keyframes), Start: t0, End: t1,
				})
			}
			refSeen := map[string]bool{}
			for _, fx := range l.Effects {
				prof.EffectUsage[fx.MatchName]++
				cls := classify(fx.MatchName)
				if cls == "⚠THIRD-PARTY" && !tpSeen[fx.MatchName] {
					tpSeen[fx.MatchName] = true
					prof.ThirdParty = append(prof.ThirdParty, fx.MatchName)
				}
				je := jEffect{MatchName: fx.MatchName, Class: cls}
				de, haveEffect := dictEffect{}, false
				if dict != nil {
					de, haveEffect = dict.Effects[fx.MatchName]
				}
				if haveEffect {
					je.Name = de.Name
				}
				for _, pr := range fx.Parameters {
					anim := len(pr.Keyframes) > 0
					set := pr.StaticValue != nil
					if !anim && !set {
						continue
					}
					jp := jParam{MatchName: pr.MatchName, Value: pr.StaticValue, Expr: pr.Expression}
					if haveEffect {
						if dp, ok := de.Params[pr.MatchName]; ok {
							jp.Name = dp.Name
							jp.Default = dp.Default
							if set && dp.Default != nil {
								if eq, cmp := equalsDefault(pr.StaticValue, dp.Default); cmp && !eq {
									jp.Changed = true
								}
							}
						}
					}
					label := jp.Name
					if label == "" {
						label = pr.MatchName
					}
					if jp.Changed || anim {
						je.Tuned = append(je.Tuned, label)
					}
					je.Params = append(je.Params, jp)
					for _, r := range exprRefs(pr.Expression) {
						if !refSeen[r] {
							refSeen[r] = true
							jl.ExprRefs = append(jl.ExprRefs, r)
						}
					}
				}
				jl.Effects = append(jl.Effects, je)
			}
			jc.Layers = append(jc.Layers, jl)
		}
		prof.Comps = append(prof.Comps, jc)
	}
	return prof
}

func main() {
	dictPath := flag.String("dict", "data/effects-dict/effects_en_US_25.1x68.json",
		"effect dictionary json (matchName -> name + default); empty to disable")
	jsonOut := flag.Bool("json", false, "emit a structured project profile as JSON instead of the text report")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: aepdissect [-dict <effects.json>] <file.aep>")
		os.Exit(2)
	}
	path := args[0]

	var dict *effectsDict
	if *dictPath != "" {
		dict = loadDict(*dictPath)
	}

	p, err := aep.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		os.Exit(1)
	}

	compName := map[uint32]string{}
	for _, c := range p.Compositions {
		compName[c.ID] = c.Name
	}

	if *jsonOut {
		prof := buildProfile(path, p, dict, compName)
		b, err := json.MarshalIndent(prof, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "marshal:", err)
			os.Exit(1)
		}
		fmt.Println(string(b))
		return
	}

	fmt.Printf("=== PROJECT: %s ===\n", path)
	fmt.Printf("compositions: %d\n", len(p.Compositions))
	if dict != nil {
		fmt.Printf("dict: %s (AE %s, %d effects)\n", *dictPath, dict.AEVersion, len(dict.Effects))
	}

	// --- effect usage histogram + classification ---
	usage := map[string]int{}
	for _, c := range p.Compositions {
		for _, l := range c.Layers {
			for _, fx := range l.Effects {
				usage[fx.MatchName]++
			}
		}
	}
	type ue struct {
		name  string
		count int
	}
	var hist []ue
	for n, c := range usage {
		hist = append(hist, ue{n, c})
	}
	sort.Slice(hist, func(i, j int) bool {
		if hist[i].count != hist[j].count {
			return hist[i].count > hist[j].count
		}
		return hist[i].name < hist[j].name
	})
	fmt.Printf("\n=== EFFECT USAGE (%d distinct) ===\n", len(hist))
	var thirdParty []string
	for _, e := range hist {
		cls := classify(e.name)
		human := ""
		if dict != nil {
			if de, ok := dict.Effects[e.name]; ok && de.Name != "" {
				human = "  \"" + de.Name + "\""
			}
		}
		fmt.Printf("  %4d  %-28s [%s]%s\n", e.count, e.name, cls, human)
		if cls == "⚠THIRD-PARTY" {
			thirdParty = append(thirdParty, fmt.Sprintf("%s (×%d)", e.name, e.count))
		}
	}

	fmt.Printf("\n=== RENDER DEPENDENCIES (third-party = output needs install) ===\n")
	if len(thirdParty) == 0 {
		fmt.Println("  none — all effects native or Cycore-bundled (renders on stock AE)")
	} else {
		for _, t := range thirdParty {
			fmt.Printf("  ⚠ %s\n", t)
		}
	}

	// --- precomp nesting ---
	fmt.Printf("\n=== PRECOMP NESTING (comp -> precomp sources) ===\n")
	for _, c := range p.Compositions {
		var kids []string
		seen := map[string]bool{}
		for _, l := range c.Layers {
			if nm, ok := compName[l.SourceID]; ok && !seen[nm] {
				seen[nm] = true
				kids = append(kids, nm)
			}
		}
		if len(kids) > 0 {
			fmt.Printf("  %q -> %v\n", c.Name, kids)
		}
	}

	// --- per-comp detail ---
	fmt.Printf("\n=== PER-COMP DETAIL (✎=changed vs default · ?=no dict · ~=animated · act=[in→out]s) ===\n")
	for _, c := range p.Compositions {
		fmt.Printf("\n--- COMP %q (id=%d, %dx%d, %.0ffps, %.1fs, %d layers) ---\n",
			c.Name, c.ID, c.Width, c.Height, c.FrameRate, c.Duration, len(c.Layers))
		type animStart struct {
			name string
			t    float64
		}
		var starts []animStart
		for _, l := range c.Layers {
			src := "-"
			if l.SourceID != 0 {
				if nm, ok := compName[l.SourceID]; ok {
					src = "PRECOMP " + nm
				} else {
					src = fmt.Sprintf("footage(%d)", l.SourceID)
				}
			}
			fmt.Printf("  L[%d] %-18q %-10v blend=%-13s vis=%-5v src=%-20s act=[%.1f→%.1f] masks=%d fx=%d\n",
				l.Index, l.Name, l.Type, blend(int(l.BlendingMode)), l.Visible, src,
				l.InPoint(), l.OutPoint(), len(l.Masks), len(l.Effects))

			firstAnim := math.Inf(1)

			// transform / layer-level animated properties (effect params handled below)
			var animProps []*aep.Property
			walkAnimated(l.PropertyTree(), &animProps)
			for _, pr := range animProps {
				t0, t1 := kfSpan(pr.Keyframes)
				if t0 < firstAnim {
					firstAnim = t0
				}
				fmt.Printf("        ~ %-24s [ANIM %dkf %s  %.1f→%.1fs]\n",
					propName(pr.MatchName), len(pr.Keyframes), motionLabel(pr.Keyframes), t0, t1)
			}

			for _, fx := range l.Effects {
				de, haveEffect := dictEffect{}, false
				if dict != nil {
					de, haveEffect = dict.Effects[fx.MatchName]
				}
				ehuman := ""
				if haveEffect && de.Name != "" {
					ehuman = " \"" + de.Name + "\""
				}
				fmt.Printf("        FX %-26s%s [%s]\n", fx.MatchName, ehuman, classify(fx.MatchName))

				var tuned []string
				for _, pr := range fx.Parameters {
					anim := len(pr.Keyframes) > 0
					set := pr.StaticValue != nil
					if !anim && !set {
						continue
					}
					// resolve human name + default from dict
					pname, mark := pr.MatchName, "·"
					var defStr string
					if haveEffect {
						if dp, ok := de.Params[pr.MatchName]; ok {
							if dp.Name != "" {
								pname = dp.Name
							}
							if set && dp.Default != nil {
								if eq, cmp := equalsDefault(pr.StaticValue, dp.Default); cmp {
									if eq {
										mark = "=" // stored but equals default (structural/rare)
									} else {
										mark = "✎" // author changed it
										defStr = fmt.Sprintf("  (default %s)", fmtNum(dp.Default))
									}
								}
							}
						} else {
							mark = "?" // no dict entry for this slot (e.g. master slot 0000)
						}
					}
					line := fmt.Sprintf("         %s %-26s", mark, pname)
					if set {
						line += fmt.Sprintf(" = %s", fmtNum(pr.StaticValue))
					}
					line += defStr
					if anim {
						t0, t1 := kfSpan(pr.Keyframes)
						if t0 < firstAnim {
							firstAnim = t0
						}
						line += fmt.Sprintf("  [ANIM %dkf %s %.1f→%.1fs]", len(pr.Keyframes), motionLabel(pr.Keyframes), t0, t1)
					}
					if pr.Expression != "" {
						line += fmt.Sprintf("  expr=%q", pr.Expression)
					}
					fmt.Println(line)
					if mark == "✎" || anim {
						tuned = append(tuned, pname)
					}
				}
				if len(tuned) > 0 {
					fmt.Printf("           → tuned: %s\n", strings.Join(tuned, ", "))
				}
			}

			if !math.IsInf(firstAnim, 1) {
				starts = append(starts, animStart{l.Name, firstAnim})
			}
		}

		// stagger hint: ≥3 animated layers whose anim starts march forward at a
		// roughly constant offset = a classic MG stagger (错位启动).
		if len(starts) >= 3 {
			sort.Slice(starts, func(i, j int) bool { return starts[i].t < starts[j].t })
			var deltas []float64
			lo, hi := math.Inf(1), 0.0
			for i := 1; i < len(starts); i++ {
				d := starts[i].t - starts[i-1].t
				deltas = append(deltas, d)
				if d < lo {
					lo = d
				}
				if d > hi {
					hi = d
				}
			}
			if lo > 0.02 && hi < 3.0 && (hi-lo) < 0.2 {
				fmt.Printf("  ↳ stagger: %d layers, anim start offset ~%.2fs each (错位启动)\n", len(starts), (lo+hi)/2)
			}
		}
	}

	// --- dependency graph (parent / matte / source / expr-ref edges) ---
	fmt.Printf("\n=== DEPENDENCY GRAPH (parent / matte / source / expr-ref edges) ===\n")
	for _, c := range p.Compositions {
		idIndex := map[uint32]int{}
		for _, l := range c.Layers {
			idIndex[l.ID] = l.Index
		}
		var edges []string
		for _, l := range c.Layers {
			if l.ParentID != 0 {
				edges = append(edges, fmt.Sprintf("  L[%d] %-18q --parent--> L[%d]", l.Index, l.Name, idIndex[l.ParentID]))
			}
			if l.TrackMatteLayerID != 0 {
				edges = append(edges, fmt.Sprintf("  L[%d] %-18q --matte--> L[%d]", l.Index, l.Name, idIndex[l.TrackMatteLayerID]))
			}
			if nm, ok := compName[l.SourceID]; ok {
				edges = append(edges, fmt.Sprintf("  L[%d] %-18q --source--> precomp %q", l.Index, l.Name, nm))
			}
			refSeen := map[string]bool{}
			for _, fx := range l.Effects {
				for _, pr := range fx.Parameters {
					for _, r := range exprRefs(pr.Expression) {
						if !refSeen[r] {
							refSeen[r] = true
							edges = append(edges, fmt.Sprintf("  L[%d] %-18q --expr--> %q", l.Index, l.Name, r))
						}
					}
				}
			}
		}
		if len(edges) > 0 {
			fmt.Printf("\n--- COMP %q ---\n", c.Name)
			for _, e := range edges {
				fmt.Println(e)
			}
		}
	}
}
