package apidoc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func goodAnn() *Annotation {
	return &Annotation{
		Summary: "Add a vector mask to a layer", Domain: "mask", Stability: "stable",
		Verify: "ae-accept", Since: "AE2020",
		Gate:    []string{"TestAddMask_AEShipGate_AE2020"},
		Params:  []Param{{"layer", "owning layer reference"}, {"path", "closed Bezier outline path"}},
		Returns: "the created mask", HasTags: true,
	}
}

func goodSym() Symbol {
	return Symbol{Name: "AddMask", Pos: "facade.go:10", Params: []string{"layer", "path"}, ReturnsNonError: true, WriteSurface: true}
}

func ctx() Context {
	return Context{Gates: map[string]bool{"TestAddMask_AEShipGate_AE2020": false}}
}

func msgs(errs []error) string {
	var b strings.Builder
	for _, e := range errs {
		b.WriteString(e.Error())
		b.WriteByte('\n')
	}
	return b.String()
}

func TestValidate_Clean(t *testing.T) {
	if errs := Validate(goodAnn(), goodSym(), ctx(), ModeStrict); len(errs) != 0 {
		t.Fatalf("want clean, got:\n%s", msgs(errs))
	}
}

func TestValidate_SummaryRules(t *testing.T) {
	a := goodAnn()
	a.Summary = strings.Repeat("x", 81)
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "80 chars") {
		t.Errorf("want 80-char error, got:\n%s", msgs(errs))
	}
	a = goodAnn()
	a.Summary = "Adds a mask."
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "period") {
		t.Errorf("want period error, got:\n%s", msgs(errs))
	}
}

func TestValidate_ParamMismatch(t *testing.T) {
	a := goodAnn()
	a.Params = []Param{{"layer", "owning layer reference"}} // missing "path"
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), `missing @param "path"`) {
		t.Errorf("want missing-param error, got:\n%s", msgs(errs))
	}
	a = goodAnn()
	a.Params = append(a.Params, Param{"bogus", "extra param desc here"})
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), `"bogus" is not a parameter`) {
		t.Errorf("want extra-param error, got:\n%s", msgs(errs))
	}
}

func TestValidate_ParamDescRules(t *testing.T) {
	a := goodAnn()
	a.Params[0].Desc = "TODO"
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "TODO") {
		t.Errorf("want TODO error, got:\n%s", msgs(errs))
	}
	a = goodAnn()
	a.Params[0].Desc = "owning layer"
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "3 words") {
		t.Errorf("want 3-word error, got:\n%s", msgs(errs))
	}
}

func TestValidate_Enums(t *testing.T) {
	a := goodAnn()
	a.Domain = "bogus"
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), `invalid @domain "bogus"`) {
		t.Errorf("want domain enum error, got:\n%s", msgs(errs))
	}
	a = goodAnn()
	a.Since = "AE99"
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "@since") {
		t.Errorf("want since error, got:\n%s", msgs(errs))
	}
}

func TestValidate_GateRequiredAndExists(t *testing.T) {
	a := goodAnn()
	a.Gate = nil
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "requires @gate") {
		t.Errorf("want gate-required error, got:\n%s", msgs(errs))
	}
	a = goodAnn()
	a.Gate = []string{"TestNonexistent"}
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "not found") {
		t.Errorf("want gate-existence error, got:\n%s", msgs(errs))
	}
}

func TestValidate_IncidentUnderKnowledgeTree(t *testing.T) {
	root := t.TempDir()
	knowledgeDir := filepath.Join(root, "docs", "knowledge", "effects")
	if err := os.MkdirAll(knowledgeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(knowledgeDir, "known-incident.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	a := goodAnn()
	a.Incident = []string{"known-incident"}
	c := ctx()
	c.IncidentsDir = filepath.Join(root, "docs", "incidents")

	if errs := Validate(a, goodSym(), c, ModeStrict); len(errs) != 0 {
		t.Fatalf("expected knowledge-tree incident to validate, got:\n%s", msgs(errs))
	}
}

func TestValidate_WarnSkipsUntagged(t *testing.T) {
	a := &Annotation{HasTags: false}
	if errs := Validate(a, goodSym(), ctx(), ModeWarn); len(errs) != 0 {
		t.Errorf("warn mode should skip untagged symbol, got:\n%s", msgs(errs))
	}
	if errs := Validate(a, goodSym(), ctx(), ModeStrict); !strings.Contains(msgs(errs), "missing @tag") {
		t.Errorf("strict mode should flag untagged write-surface symbol, got:\n%s", msgs(errs))
	}
}
