// tmp_debug/bisect_v2_2/main.go
//
// V2.2 Phase 5 ship-gate iter 4 — minimum-failing bisection runner.
//
// Builds 6 variants of increasing complexity, invokes AE 2025 on each via
// verify_open.jsx, records PASS/FAIL + error message. The first FAIL variant
// localizes the failure category (layer skel / shape emit / static value /
// keyframe encoding / spatial-specific).
//
// Matrix (board.md "Next session" 第 3 节):
//   #2: empty NewShapeLayer (no mutation)
//   #3: + AddRect() (default values)
//   #4: + rect.SetSize (static)
//   #5: + AddFill() (static color)
//   #6: + rect.Size keyframed (non-spatial 2D)
//   #7: + L.Position keyframed (spatial 2D)
//
// Each variant writes tmp_debug/minfail_v<N>.aep, launches AE, waits for
// .done file (90s timeout), parses first line, prints matrix row.
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
	jsxPath     = `E:/projects/tools/aep-parser/test_data/verify_open.jsx`
	argsPath    = `e:/projects/tools/aep-parser/test_data/verify_open_args.json`
	repoRoot    = `e:/projects/tools/aep-parser`
	openTimeout = 90 * time.Second
)

type result struct {
	variant int
	desc    string
	status  string // "PASS" / "FAIL" / "TIMEOUT" / "BUILD_ERR"
	detail  string
}

func main() {
	results := []result{}
	for v := 2; v <= 7; v++ {
		r := runVariant(v)
		results = append(results, r)
		fmt.Printf("=== variant %d %-30s → %s\n", r.variant, r.desc, r.status)
		if r.detail != "" {
			fmt.Printf("    %s\n", truncate(r.detail, 200))
		}
		// First-FAIL is the diagnostic answer; continuing past it adds no
		// signal (we already know AE rejects). But print remaining variants
		// as SKIPPED for the matrix display, then exit.
		if r.status != "PASS" {
			for skip := v + 1; skip <= 7; skip++ {
				results = append(results, result{variant: skip, desc: variantDesc(skip), status: "SKIPPED"})
			}
			break
		}
	}

	fmt.Println()
	fmt.Println("--- bisection matrix ---")
	for _, r := range results {
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

func runVariant(v int) result {
	desc := variantDesc(v)
	aepPath := filepath.Join("tmp_debug", fmt.Sprintf("minfail_v%d.aep", v))
	donePath := filepath.Join("tmp_debug", fmt.Sprintf("minfail_v%d.done", v))

	if err := buildVariant(v, aepPath); err != nil {
		return result{variant: v, desc: desc, status: "BUILD_ERR", detail: err.Error()}
	}

	absAep, _ := filepath.Abs(aepPath)
	absDone, _ := filepath.Abs(donePath)
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q}`, toFwd(absAep), toFwd(absDone))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		return result{variant: v, desc: desc, status: "BUILD_ERR", detail: "write args: " + err.Error()}
	}
	os.Remove(donePath)

	cmd := exec.Command(aeExe, "-r", jsxPath)
	if err := cmd.Start(); err != nil {
		return result{variant: v, desc: desc, status: "BUILD_ERR", detail: "AE start: " + err.Error()}
	}

	deadline := time.Now().Add(openTimeout)
	for {
		if _, err := os.Stat(donePath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = cmd.Process.Kill()
			return result{variant: v, desc: desc, status: "TIMEOUT", detail: fmt.Sprintf("no .done after %s", openTimeout)}
		}
		time.Sleep(2 * time.Second)
	}

	// Wait for AE to actually quit before next variant (avoids singleton conflict).
	_ = cmd.Wait()

	data, err := os.ReadFile(donePath)
	if err != nil {
		return result{variant: v, desc: desc, status: "BUILD_ERR", detail: "read .done: " + err.Error()}
	}
	lines := strings.SplitN(string(data), "\n", 2)
	status := strings.TrimSpace(lines[0])
	detail := ""
	if len(lines) > 1 {
		detail = strings.TrimSpace(lines[1])
	}
	return result{variant: v, desc: desc, status: status, detail: detail}
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
