package selfhost

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func TestAuditSampleShellEffectTemplateCandidatesFlagsMaterializedState(t *testing.T) {
	root := t.TempDir()
	candidate := filepath.Join(root, "effect_adbe_demo.bin")
	writeEffectTemplateAuditCandidate(t, candidate)

	result, err := AuditSampleShellEffectTemplateCandidates(SampleShellEffectTemplateAuditOptions{
		CandidatesDir: root,
	})
	if err != nil {
		t.Fatalf("AuditSampleShellEffectTemplateCandidates returned error: %v", err)
	}

	if result.Summary.Total != 1 || result.Summary.Reject != 1 {
		t.Fatalf("summary = %+v", result.Summary)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("candidates = %+v", result.Candidates)
	}
	got := result.Candidates[0]
	if got.MatchName != "ADBE Demo" || got.ParamValueGroups != 2 || got.CdatValues != 1 ||
		got.KeyframeValues != 1 || got.Expressions != 1 || got.LayerRefs != 1 ||
		got.Recommendation != "reject_candidate" {
		t.Fatalf("audit = %+v", got)
	}
	if _, err := os.Stat(result.ResultPath); err != nil {
		t.Fatalf("audit result missing: %v", err)
	}
}

func writeEffectTemplateAuditCandidate(t *testing.T, path string) {
	t.Helper()
	wrapper := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
		{ID: rifx.IDTdmn, Data: []byte("ADBE Demo")},
		{ID: rifx.IDList, FormType: rifx.IDSspc, Children: []*rifx.Chunk{
			{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
				{ID: rifx.IDTdmn, Data: []byte("Param One")},
				{ID: rifx.IDList, FormType: rifx.IDTdbs, Children: []*rifx.Chunk{
					{ID: rifx.IDCdat, Data: make([]byte, 8)},
					{ID: rifx.IDUtf8, Data: []byte("time")},
					{ID: rifx.IDTdpi, Data: []byte{0, 0, 0, 7}},
				}},
				{ID: rifx.IDTdmn, Data: []byte("Param Two")},
				{ID: rifx.IDList, FormType: rifx.IDTdbs, Children: []*rifx.Chunk{
					{ID: rifx.IDList, FormType: rifx.IDkfl, Children: []*rifx.Chunk{
						{ID: rifx.IDLhd3, Data: make([]byte, 52)},
						{ID: rifx.IDLdat, Data: make([]byte, 16)},
					}},
				}},
			}},
		}},
	}}
	var buf bytes.Buffer
	if err := wrapper.Write(&buf); err != nil {
		t.Fatalf("write wrapper: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write candidate: %v", err)
	}
}
