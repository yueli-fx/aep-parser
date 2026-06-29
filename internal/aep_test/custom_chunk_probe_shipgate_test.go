// internal/aep/custom_chunk_probe_shipgate_test.go
//
// EXPERIMENT (negative-finding probe), not a capability gate: does AE preserve
// an arbitrary unknown RIFX chunk across a save? Builds a minimal project,
// injects a custom leaf chunk ("uMD1") carrying author-style metadata at a
// chosen position, has AE open + resave, then re-parses to see whether the
// chunk survived. Positions probed:
//   - ROOT: direct child of the RIFX root (sibling of Fold).
//   - ITEM: inside the comp's Item LIST (sibling of idta/cdta/Layr — a
//     container AE fully models).
//
// Either outcome is a valid finding:
//   - SURVIVED → AE round-trips an unknown chunk there (a real metadata stash).
//   - DROPPED  → AE rebuilds from its object model and discards it.
//   - AE reject/crash → AE validates that container's chunk set (also a finding).
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

var probeChunkID = rifx.ChunkID{'u', 'M', 'D', '1'}

const probePayload = "AEPARSER_PROBE:author=YueLi;tool=aep-parser"

// injectRoot appends the custom chunk as a direct RIFX-root child.
func injectRoot(t *testing.T, root *rifx.Chunk) {
	root.Children = append(root.Children, &rifx.Chunk{ID: probeChunkID, Data: []byte(probePayload)})
}

// injectItem appends the custom chunk at the end of the first comp Item LIST.
func injectItem(t *testing.T, root *rifx.Chunk) {
	item := findShipListByForm(root, rifx.IDItem)
	if item == nil {
		t.Fatal("no Item LIST found to inject into")
	}
	item.Children = append(item.Children, &rifx.Chunk{ID: probeChunkID, Data: []byte(probePayload)})
}

func runCustomChunkProbe(t *testing.T, aeExe, ver, pos string, inject func(*testing.T, *rifx.Chunk), target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/chunk_probe_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_chunk_probe.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	// 1. Minimal openable project.
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "PROBE", 640, 360, 30, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "S", 100, 100, [3]float64{0.5, 0.5, 0.5}); err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}
	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}

	// 2. Parse our own output, inject a custom leaf chunk at the chosen position.
	root, err := rifx.Parse(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("rifx.Parse own output: %v", err)
	}
	inject(t, root)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "chunk_probe_in.aep")
	resavedAEP := filepath.Join(tempDir, "chunk_probe_resaved.aep")
	doneFile := filepath.Join(tempDir, "chunk_probe.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := root.Write(out); err != nil {
		out.Close()
		t.Fatalf("root.Write: %v", err)
	}
	out.Close()

	// 3. Sanity: our own parser sees the injected chunk in the written input.
	inRoot := parseAEP(t, inputAEP)
	if got := findShipChunk(inRoot, probeChunkID); got == nil {
		t.Fatalf("injection failed: custom chunk not in our own written input")
	} else if string(got.Data) != probePayload {
		t.Fatalf("injected payload mismatch: %q", got.Data)
	}
	t.Logf("%s/%s: injected custom chunk %q present in input (pre-AE)", ver, pos, probeChunkID)

	// 4. AE open + resave.
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 240)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s/%s AE open/resave: %s", ver, pos, strings.TrimSpace(string(content)))

	// 5. Verdict: did the chunk survive AE's resave?
	reRoot := parseAEP(t, resavedAEP)
	if got := findShipChunk(reRoot, probeChunkID); got != nil {
		t.Logf("RESULT %s/%s: SURVIVED — AE preserved custom chunk %q = %q", ver, pos, probeChunkID, got.Data)
	} else {
		t.Logf("RESULT %s/%s: DROPPED — AE discarded custom chunk %q on resave", ver, pos, probeChunkID)
	}
}

func TestCustomChunkProbe_Root_AE2025(t *testing.T) {
	runCustomChunkProbe(t, ae2025(), "AE2025", "root", injectRoot, aep.TargetAE2025)
}
func TestCustomChunkProbe_Root_AE2020(t *testing.T) {
	runCustomChunkProbe(t, ae2020(), "AE2020", "root", injectRoot, aep.TargetAE2020)
}
func TestCustomChunkProbe_Item_AE2025(t *testing.T) {
	runCustomChunkProbe(t, ae2025(), "AE2025", "item", injectItem, aep.TargetAE2025)
}
func TestCustomChunkProbe_Item_AE2020(t *testing.T) {
	runCustomChunkProbe(t, ae2020(), "AE2020", "item", injectItem, aep.TargetAE2020)
}
