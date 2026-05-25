// tmp_debug/bisect_v2_2/main.go
//
// V2.2 Phase 5 ship-gate bisection runner (batch mode, iter-5b).
//
// Builds 6 variants of increasing complexity, then launches AE 2025 ONCE
// to verify all of them in sequence via verify_open_batch.jsx. The script
// opens each .aep, dumps comp.layers.length + names to a per-variant .done
// file, closes the project, moves to the next. AE quits at the end.
//
// Previous flow re-launched AE per variant — ~9 minutes for 6 variants.
// Batch flow is bounded by AE startup (~10s) + per-variant open+verify
// (~5-8s) ≈ 1 minute total.
//
// Matrix:
//   #2: empty NewShapeLayer (no mutation)
//   #3: + AddRect() (default values)
//   #4: + rect.SetSize (static)
//   #5: + AddFill (static color)
//   #6: + rect.Size keyframed (non-spatial 2D)
//   #7: + L.Position keyframed (spatial 2D)
//
// Run: go run tmp_debug/bisect_v2_2/main.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	aep "github.com/example/aep-parser/internal/aep"
)

const (
	aeExe       = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	jsxPath     = `E:/projects/tools/aep-parser/test_data/verify_open_batch.jsx`
	argsPath    = `e:/projects/tools/aep-parser/test_data/verify_open_batch_args.json`
	summaryPath = `e:/projects/tools/aep-parser/test_data/verify_open_batch_summary.txt`
	totalTO     = 6 * time.Minute // safety upper bound for all 6 variants
)

type result struct {
	variant int
	desc    string
	status  string // "PASS" / "FAIL" / "TIMEOUT" / "BUILD_ERR"
	detail  string
}

func main() {
	// 1. Build all 6 variants up front.
	type plan struct {
		variant  int
		desc     string
		aepPath  string
		donePath string
	}
	var plans []plan
	results := make(map[int]*result)
	for v := 2; v <= 7; v++ {
		aepPath := filepath.Join("tmp_debug", fmt.Sprintf("minfail_v%d.aep", v))
		donePath := filepath.Join("tmp_debug", fmt.Sprintf("minfail_v%d.done", v))
		if err := buildVariant(v, aepPath); err != nil {
			results[v] = &result{variant: v, desc: variantDesc(v), status: "BUILD_ERR", detail: err.Error()}
			continue
		}
		plans = append(plans, plan{variant: v, desc: variantDesc(v), aepPath: aepPath, donePath: donePath})
		_ = os.Remove(donePath)
	}

	// 2. Write args.json batch payload.
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	var entries []string
	for _, pl := range plans {
		absAep, _ := filepath.Abs(pl.aepPath)
		absDone, _ := filepath.Abs(pl.donePath)
		entries = append(entries, fmt.Sprintf(`{"input":%q,"done":%q}`, toFwd(absAep), toFwd(absDone)))
	}
	argsJSON := fmt.Sprintf(`{"items":[%s]}`, strings.Join(entries, ","))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "write args:", err)
		os.Exit(1)
	}
	_ = os.Remove(summaryPath)

	// 3. Launch AE once.
	fmt.Printf("Launching AE 2025 (one process for %d variants)...\n", len(plans))
	cmd := exec.Command(aeExe, "-r", jsxPath)
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "AE start:", err)
		os.Exit(1)
	}

	// 4. Poll for .done files in order; report each as it lands.
	deadline := time.Now().Add(totalTO)
	for _, pl := range plans {
		seen := false
		for !seen {
			if _, err := os.Stat(pl.donePath); err == nil {
				seen = true
				break
			}
			if _, err := os.Stat(summaryPath); err == nil {
				// JSX finished early (possibly an early error) — break inner
				// loop; outer loop will treat any missing .done as TIMEOUT.
				break
			}
			if time.Now().After(deadline) {
				_ = cmd.Process.Kill()
				results[pl.variant] = &result{variant: pl.variant, desc: pl.desc, status: "TIMEOUT", detail: fmt.Sprintf("batch deadline %s exceeded", totalTO)}
				goto teardown
			}
			time.Sleep(1 * time.Second)
		}
		if !seen {
			results[pl.variant] = &result{variant: pl.variant, desc: pl.desc, status: "MISSING", detail: "no .done before JSX summary written"}
			continue
		}
		results[pl.variant] = readDone(pl)
		fmt.Printf("=== variant %d %-30s → %s\n", pl.variant, pl.desc, results[pl.variant].status)
		if results[pl.variant].detail != "" {
			for _, line := range strings.Split(results[pl.variant].detail, "\n") {
				fmt.Printf("    %s\n", line)
			}
		}
	}

teardown:
	_ = cmd.Wait()

	// 5. Print final matrix.
	fmt.Println()
	fmt.Println("--- bisection matrix (batch single-launch) ---")
	for v := 2; v <= 7; v++ {
		r, ok := results[v]
		if !ok {
			fmt.Printf("  #%d %-32s (no result)\n", v, variantDesc(v))
			continue
		}
		fmt.Printf("  #%d %-32s %s\n", r.variant, r.desc, r.status)
	}
}

func variantDesc(v int) string {
	switch v {
	case 2:
		return "empty ShapeLayer"
	case 3:
		return "+ AddRect (default)"
	case 4:
		return "+ rect.SetSize (static)"
	case 5:
		return "+ AddFill (static color)"
	case 6:
		return "+ rect.Size keyframed"
	case 7:
		return "+ L.Position keyframed"
	}
	return "?"
}

func readDone(pl struct {
	variant  int
	desc     string
	aepPath  string
	donePath string
}) *result {
	data, err := os.ReadFile(pl.donePath)
	if err != nil {
		return &result{variant: pl.variant, desc: pl.desc, status: "BUILD_ERR", detail: "read .done: " + err.Error()}
	}
	lines := strings.SplitN(string(data), "\n", 2)
	status := strings.TrimSpace(lines[0])
	detail := ""
	if len(lines) > 1 {
		detail = strings.TrimSpace(lines[1])
	}
	return &result{variant: pl.variant, desc: pl.desc, status: status, detail: detail}
}

// buildVariant constructs the aep for a given variant and writes it to out.
// V2.1 baseline = NewProject + 1 NewComposition; variant complexity is layered
// onto that baseline via NewShapeLayer + AddX + setters / keyframes.
func buildVariant(variant int, out string) error {
	p := aep.NewProject(aep.TargetAE2025)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		return err
	}
	// All variants ≥ 2 add a single ShapeLayer.
	l, err := comp.NewShapeLayer("L")
	if err != nil {
		return err
	}
	var rect *aep.RectNode
	if variant >= 3 {
		rect, err = l.RootGroup().AddRect()
		if err != nil {
			return err
		}
	}
	if variant >= 4 {
		if err := rect.SetSize([2]float64{200, 200}); err != nil {
			return err
		}
	}
	if variant >= 5 {
		fill, err := l.RootGroup().AddFill()
		if err != nil {
			return err
		}
		if err := fill.SetColor([4]float64{1, 0, 0, 1}); err != nil {
			return err
		}
	}
	if variant >= 6 {
		if err := rect.Size().AddKeyframeLinear(0, [2]float64{100, 100}); err != nil {
			return err
		}
		if err := rect.Size().AddKeyframeLinear(1, [2]float64{200, 200}); err != nil {
			return err
		}
	}
	if variant >= 7 {
		if err := l.Position().AddKeyframeLinear(0, [2]float64{0, 0}); err != nil {
			return err
		}
		if err := l.Position().AddKeyframeLinear(1, [2]float64{500, 300}); err != nil {
			return err
		}
	}

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	return p.WriteAEP(f)
}
