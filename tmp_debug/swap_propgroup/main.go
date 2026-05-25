// tmp_debug/swap_propgroup/main.go
//
// Build N transplant variants. Each takes tolerance.aep as base + swaps the
// named property group(s) inside the user Layr's outer LIST(tdgp) with our
// minfail_v2.aep's version. Runs all variants through a single AE batch
// session via verify_open_batch.jsx, prints layers.length for each.
//
// Used to bisect which property group in our outer tdgp is the silent-drop
// trigger. After Layr-level + tdgp-level transplant confirmed the problem
// is in our outer tdgp content, this narrows to specific group(s).
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/example/aep-parser/internal/rifx"
)

const (
	aeExe       = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	jsxPath     = `E:/projects/tools/aep-parser/test_data/verify_open_batch.jsx`
	argsPath    = `e:/projects/tools/aep-parser/test_data/verify_open_batch_args.json`
	summaryPath = `e:/projects/tools/aep-parser/test_data/verify_open_batch_summary.txt`
	totalTO     = 3 * time.Minute
)

// Each variant lists the property-group matchNames to take from ours; rest
// come from tolerance. All swaps happen inside the user Layr's outer LIST(tdgp).
var variants = []struct {
	tag    string
	groups []string
}{
	{"01_user_content_groups", []string{"ADBE Root Vectors Group", "ADBE Transform Group", "ADBE Layer Styles"}},
	{"02_placeholder_groups", []string{"ADBE Extrsn Options Group", "ADBE Material Options Group", "ADBE Audio Group", "ADBE Layer Sets"}},
	{"03_transform_only", []string{"ADBE Transform Group"}},
	{"04_rootvectors_only", []string{"ADBE Root Vectors Group"}},
	{"05_layerstyles_only", []string{"ADBE Layer Styles"}},
}

func main() {
	baseRoot := load("test_data/v2_2_shape_tolerance.aep")
	donorRoot := load("tmp_debug/minfail_v2.aep")
	donorLayr := findFirstUserLayr(donorRoot)
	if donorLayr == nil {
		fmt.Println("donor has no user Layr")
		os.Exit(1)
	}
	donorTdgp := findOuterTdgp(donorLayr)
	if donorTdgp == nil {
		fmt.Println("donor user Layr has no outer LIST(tdgp)")
		os.Exit(1)
	}

	type entry struct{ aepPath, donePath string }
	var entries []entry

	for _, v := range variants {
		// Re-parse base for each variant (need fresh tree per swap).
		bRoot := load("test_data/v2_2_shape_tolerance.aep")
		bLayr := findFirstUserLayr(bRoot)
		bTdgp := findOuterTdgp(bLayr)
		swapped := swapGroups(bTdgp, donorTdgp, v.groups)
		_ = swapped
		aepPath := fmt.Sprintf("tmp_debug/swap_%s.aep", v.tag)
		donePath := fmt.Sprintf("tmp_debug/swap_%s.done", v.tag)
		f, _ := os.Create(aepPath)
		_ = bRoot.Write(f)
		f.Close()
		_ = os.Remove(donePath)
		entries = append(entries, entry{aepPath, donePath})
		fmt.Printf("built %s (swap groups: %v)\n", aepPath, v.groups)
	}
	_ = baseRoot

	// Write batch args.
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	var items []string
	for _, e := range entries {
		absA, _ := filepath.Abs(e.aepPath)
		absD, _ := filepath.Abs(e.donePath)
		items = append(items, fmt.Sprintf(`{"input":%q,"done":%q}`, toFwd(absA), toFwd(absD)))
	}
	argsJSON := fmt.Sprintf(`{"items":[%s]}`, strings.Join(items, ","))
	_ = os.WriteFile(argsPath, []byte(argsJSON), 0644)
	_ = os.Remove(summaryPath)

	fmt.Println("\nLaunching AE 2025 (one process, batch mode)...")
	cmd := exec.Command(aeExe, "-r", jsxPath)
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	deadline := time.Now().Add(totalTO)
	for _, e := range entries {
		for {
			if _, err := os.Stat(e.donePath); err == nil {
				break
			}
			if time.Now().After(deadline) {
				_ = cmd.Process.Kill()
				fmt.Println("TIMEOUT")
				os.Exit(1)
			}
			time.Sleep(1 * time.Second)
		}
	}
	_ = cmd.Wait()

	fmt.Println("\n--- results ---")
	for i, e := range entries {
		data, _ := os.ReadFile(e.donePath)
		// Extract layers.length line
		layersLine := "?"
		for _, ln := range strings.Split(string(data), "\n") {
			if strings.Contains(ln, "layers.length=") {
				layersLine = strings.TrimSpace(ln)
				break
			}
		}
		fmt.Printf("  %-30s  %s\n", variants[i].tag, layersLine)
	}
}

func load(path string) *rifx.Chunk {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	r, err := rifx.Parse(f)
	if err != nil {
		panic(err)
	}
	return r
}

func findFirstUserLayr(root *rifx.Chunk) *rifx.Chunk {
	var found *rifx.Chunk
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found != nil {
			return
		}
		for _, ch := range c.Children {
			if ch.IsList() && ch.FormType == rifx.IDLayr {
				found = ch
				return
			}
			if ch.IsList() {
				walk(ch)
				if found != nil {
					return
				}
			}
		}
	}
	walk(root)
	return found
}

func findOuterTdgp(layr *rifx.Chunk) *rifx.Chunk {
	for _, ch := range layr.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			return ch
		}
	}
	return nil
}

// swapGroups walks `target` (an outer tdgp chunk), finds each tdmn whose name
// matches one of `groupMatchNames`, and replaces the IMMEDIATELY-following
// LIST(tdgp) chunk with the corresponding one from `donor`. Returns count of
// groups swapped.
func swapGroups(target, donor *rifx.Chunk, groupMatchNames []string) int {
	want := map[string]bool{}
	for _, n := range groupMatchNames {
		want[n] = true
	}

	donorBodies := map[string]*rifx.Chunk{}
	dKids := donor.Children
	for i := 0; i < len(dKids); i++ {
		if dKids[i].ID == rifx.IDTdmn {
			name := trimNUL(string(dKids[i].Data))
			if want[name] && i+1 < len(dKids) && dKids[i+1].IsList() {
				donorBodies[name] = dKids[i+1]
			}
		}
	}

	swapped := 0
	tKids := target.Children
	for i := 0; i < len(tKids); i++ {
		if tKids[i].ID == rifx.IDTdmn {
			name := trimNUL(string(tKids[i].Data))
			if body, ok := donorBodies[name]; ok && i+1 < len(tKids) && tKids[i+1].IsList() {
				tKids[i+1] = body
				swapped++
			}
		}
	}
	return swapped
}

func trimNUL(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}
