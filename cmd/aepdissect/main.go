// cmd/aepdissect — structured teardown of any After Effects .aep, for the
// technique-internalization pipeline (spec 2026-06-18-fx-technique-internalization,
// step PARSE). Generalizes the one-off tmp_debug probes (dump_fxchain / dump_tdmn /
// dump_layers) into a repo tool: given a reference template, print
//   - effect-usage histogram with native / Cycore-bundled / third-party class
//   - third-party plugin dependency list (what the output needs installed to render)
//   - precomp nesting map
//   - per-comp layer detail: blend mode (named), source, mask/effect counts,
//     and each effect's SET or ANIMATED params (untouched defaults elided)
//
// Usage:
//   go run ./cmd/aepdissect <file.aep>
//
// This is the mechanical step; turning the dump into role/technique knowledge is
// human/AI judgment (see checklists/build-good-fire.md for the fire example).
package main

import (
	"fmt"
	"os"
	"sort"

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
	case len(mn) >= 5 && mn[:5] == "ADBE ":
		return "native"
	case len(mn) >= 3 && mn[:3] == "CC ":
		return "Cycore(bundled)"
	default:
		return "⚠THIRD-PARTY"
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: aepdissect <file.aep>")
		os.Exit(2)
	}
	path := os.Args[1]
	p, err := aep.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		os.Exit(1)
	}

	compName := map[uint32]string{}
	for _, c := range p.Compositions {
		compName[c.ID] = c.Name
	}

	fmt.Printf("=== PROJECT: %s ===\n", path)
	fmt.Printf("compositions: %d\n", len(p.Compositions))

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
		fmt.Printf("  %4d  %-28s [%s]\n", e.count, e.name, cls)
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
	fmt.Printf("\n=== PER-COMP DETAIL ===\n")
	for _, c := range p.Compositions {
		fmt.Printf("\n--- COMP %q (id=%d, %dx%d, %.0ffps, %.1fs, %d layers) ---\n",
			c.Name, c.ID, c.Width, c.Height, c.FrameRate, c.Duration, len(c.Layers))
		for _, l := range c.Layers {
			src := "-"
			if l.SourceID != 0 {
				if nm, ok := compName[l.SourceID]; ok {
					src = "PRECOMP " + nm
				} else {
					src = fmt.Sprintf("footage(%d)", l.SourceID)
				}
			}
			fmt.Printf("  L[%d] %-18q %-10v blend=%-15s vis=%-5v src=%-22s masks=%d fx=%d\n",
				l.Index, l.Name, l.Type, blend(int(l.BlendingMode)), l.Visible, src, len(l.Masks), len(l.Effects))
			for _, fx := range l.Effects {
				fmt.Printf("        FX %-26s [%s]\n", fx.MatchName, classify(fx.MatchName))
				for _, pr := range fx.Parameters {
					anim := len(pr.Keyframes) > 0
					set := pr.StaticValue != nil
					if !anim && !set {
						continue
					}
					line := fmt.Sprintf("           · %-26s", pr.MatchName)
					if set {
						line += fmt.Sprintf(" = %v", pr.StaticValue)
					}
					if anim {
						line += fmt.Sprintf("  [ANIM %dkf]", len(pr.Keyframes))
					}
					if pr.Expression != "" {
						line += fmt.Sprintf("  expr=%q", pr.Expression)
					}
					fmt.Println(line)
				}
			}
		}
	}
}
