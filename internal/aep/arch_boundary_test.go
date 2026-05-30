package aep_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

const rifxImportPath = "github.com/example/aep-parser/internal/rifx"

func archStageOf(name string) string {
	for _, s := range []string{"scene", "codec", "parse", "lower", "write", "back", "mutate"} {
		if strings.HasPrefix(name, s+"_") {
			return s
		}
	}
	return ""
}

// sceneRifxWhitelist names scene_ files still permitted to import rifx.
// Each entry is a known boundary violation pending removal; the guard fails
// for any scene_ file NOT listed here that imports rifx.
var sceneRifxWhitelist = map[string]bool{
	// filled from this guard's first run after the scene_ rename
	"scene_features.go":        true,
	"scene_project_settings.go": true,
	"scene_project_views.go":   true,
	"scene_property_flags.go":  true,
	"scene_property_group.go":  true,
}

// sceneTypeNames are the runtime types a codec_ file must never reference —
// codec_ handles only value objects, byte streams, and rifx structures.
var sceneTypeNames = map[string]bool{
	"Project": true, "Composition": true, "Layer": true, "ShapeLayer": true,
	"Property": true, "ShapeNode": true, "VectorGroup": true, "Footage": true,
	"LayerTransform": true, "RectNode": true, "EllipseNode": true, "PathNode": true,
	"FillNode": true, "StrokeNode": true,
}

func archSrcFiles(t *testing.T) []string {
	t.Helper()
	all, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, f := range all {
		if !strings.HasSuffix(f, "_test.go") {
			out = append(out, f)
		}
	}
	return out
}

func TestArchBoundary_SceneNoRifxImport(t *testing.T) {
	fset := token.NewFileSet()
	for _, f := range archSrcFiles(t) {
		if archStageOf(f) != "scene" {
			continue
		}
		af, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		for _, imp := range af.Imports {
			if strings.Trim(imp.Path.Value, `"`) == rifxImportPath {
				if sceneRifxWhitelist[f] {
					t.Logf("WHITELIST: %s imports rifx — pending removal", f)
					continue
				}
				t.Errorf("scene_ file %s imports rifx: scene model must stay chunk-free", f)
			}
		}
	}
}

func TestArchBoundary_CodecNoSceneRef(t *testing.T) {
	fset := token.NewFileSet()
	for _, f := range archSrcFiles(t) {
		if archStageOf(f) != "codec" {
			continue
		}
		af, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		ast.Inspect(af, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if ok && sceneTypeNames[id.Name] {
				t.Errorf("codec_ file %s references scene type %q: codec must stay scene-free", f, id.Name)
			}
			return true
		})
	}
}
