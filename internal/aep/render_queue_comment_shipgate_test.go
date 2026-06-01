// internal/aep/render_queue_comment_shipgate_test.go
//
// AE ship gate for RenderQueueItem.SetComment (insert path). Opens an
// AE-2020-native render-queue fixture whose first item has NO comment, inserts
// one via SetComment (a fresh RCom wrapper chunk), WriteAEP, and has AE open the
// mutated file: proves AE ACCEPTS the inserted RCom (no data-loss / corrupt) and
// reads the comment back. Resaves so the Go side confirms the RCom survived.
//
// Gated by AE_SHIP_GATE. Base fixture built by test_data/build_rq_ae2020.jsx.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

// runRQCommentShipGate inserts a comment on baseFixture's first render-queue
// item via SetComment, then confirms the AE version under test opens the result
// (no corrupt/data-loss), reads the comment back, and keeps the RCom on resave.
// baseFixture must be openable by that AE version (AE refuses NEWER-saved files).
func runRQCommentShipGate(t *testing.T, aeExe, baseFixture string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/rq_comment_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_rq_comment.jsx`
	const expected = "ship gate insert comment"
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	proj, err := aep.Open(baseFixture)
	if err != nil {
		t.Skipf("%s not present: %v", baseFixture, err)
	}
	if proj.RenderQueue == nil || proj.RenderQueue.NumItems() < 1 {
		t.Fatalf("%s: expected >= 1 render queue item", baseFixture)
	}
	item := proj.RenderQueue.Items[0]
	if item.Comment != "" {
		t.Fatalf("%s: base item already has comment %q (want empty for insert path)", baseFixture, item.Comment)
	}
	if err := item.SetComment(expected); err != nil {
		t.Fatalf("SetComment: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "rq_comment_in.aep")
	resavedAEP := filepath.Join(tempDir, "rq_comment_resaved.aep")
	doneFile := filepath.Join(tempDir, "rq_comment.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := proj.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"expected":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), expected)
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("SetComment AE readback:\n%s", body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("rq comment ship gate FAIL:\n%s", body)
	}

	// Preservation proof: resaved file still carries the RCom, and its embedded
	// Utf8 decodes back to our comment — i.e. AE kept the inserted chunk rather
	// than silently dropping it. (There is no ScriptingAPI to read RQ comments,
	// so byte-level preservation is the strongest available verification.)
	root := parseAEP(t, resavedAEP)
	rcom := findShipChunk(root, rifx.IDRCom)
	if rcom == nil {
		t.Fatal("resaved: RCom chunk missing — AE dropped the inserted comment")
	}
	if got := decodeShipRCom(rcom.Data); got != expected {
		t.Errorf("resaved RCom comment = %q, want %q", got, expected)
	}
}

// decodeShipRCom extracts the comment from an RCom leaf's embedded Utf8 chunk
// ("Utf8" + big-endian u32 len + payload), mirroring the parser's decoder.
func decodeShipRCom(data []byte) string {
	if len(data) < 8 || string(data[0:4]) != "Utf8" {
		return ""
	}
	n := int(data[4])<<24 | int(data[5])<<16 | int(data[6])<<8 | int(data[7])
	if 8+n > len(data) {
		n = len(data) - 8
	}
	return trimShipNUL(string(data[8 : 8+n]))
}

func TestRenderQueueComment_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runRQCommentShipGate(t, aeExe, "../../test_data/rq_ae2020_base.aep")
}

func TestRenderQueueComment_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	// AE 2025 opens the AE-2020-native base via the auto-dismissed convert dialog.
	runRQCommentShipGate(t, aeExe, "../../test_data/rq_ae2020_base.aep")
}
