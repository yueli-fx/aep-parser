# Recipe Docgen Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Implement the first slice of a dedicated recipe documentation generator that emits `docs/recipe_schema.json` and `docs/recipe.md` from a canonical recipe field model.

**Architecture:** `internal/recipedoc` owns reflection, stable recipe paths, split metadata registries, capability-key mapping, and renderers. `cmd/recipedocgen` is a thin CLI that writes generated docs and provides the drift gate. Markdown and JSON schema are both rendered from the same canonical `FieldModel` list.

**Tech Stack:** Go reflection, deterministic JSON/Markdown rendering, existing `internal/recipe` schema structs, existing `internal/capindex` lookup over `docs/capabilities.json`, standard `go test` drift gates.

---

## File Structure

- Create `internal/recipedoc/model.go`
  - Canonical model types: `Document`, `FieldModel`, `TypeRef`, `FieldMeta`, `CapabilityRef`, `CapabilityMeta`.
- Create `internal/recipedoc/reflect.go`
  - Reflection walker over selected `internal/recipe` root structs.
  - Produces structural facts only: path, type, array/object shape, source type/field, JSON name, structural requiredness, order.
- Create `internal/recipedoc/reflect_test.go`
  - Locks stable path generation and structural requiredness semantics.
- Create `internal/recipedoc/summary_registry.go`
  - Field summaries keyed by stable recipe path.
- Create `internal/recipedoc/validation_registry.go`
  - Validation notes, enum values, numeric ranges, array lengths, validation-required markers.
- Create `internal/recipedoc/capability_registry.go`
  - Field path -> stable recipe capability keys.
  - Stable recipe capability key -> current capindex query string.
- Create `internal/recipedoc/example_registry.go`
  - Inline examples or example path references.
- Create `internal/recipedoc/build.go`
  - Joins reflection output with registries and capindex details into `Document`.
  - Validates registry keys and capability mappings.
- Create `internal/recipedoc/build_test.go`
  - Tests invalid registry paths, invalid capability keys, and capindex lookup failures.
- Create `internal/recipedoc/render_json.go`
  - Renders JSON Schema-like output with `x-aep-*` extensions.
- Create `internal/recipedoc/render_json_test.go`
  - Golden-ish unit test for JSON schema shape and deterministic output.
- Create `internal/recipedoc/render_md.go`
  - Renders human field reference from the canonical model.
- Create `internal/recipedoc/render_md_test.go`
  - Unit test for headings, field rows, capability display, and missing metadata markers.
- Create `cmd/recipedocgen/main.go`
  - `go run ./cmd/recipedocgen -out docs`
  - `go:generate go run . -out ../../docs`
- Create `cmd/recipedocgen/docs_uptodate_test.go`
  - Drift gate comparing generated output against committed docs.
- Create generated `docs/recipe_schema.json`
- Create generated `docs/recipe.md`
- Modify `flightdeck/work/recipe-docgen/index.md`
  - Mark implementation plan written and point to this file.

## Task 1: Canonical Structural Model

**Files:**
- Create: `internal/recipedoc/model.go`
- Create: `internal/recipedoc/reflect.go`
- Create: `internal/recipedoc/reflect_test.go`

- [x] **Step 1: Write structural model tests**

Create `internal/recipedoc/reflect_test.go` with tests that prove stable path
generation, source type names, JSON names, array handling, and structural
requiredness. The first test should include at least these paths:

```go
package recipedoc

import "testing"

func TestReflectRecipeFieldsBuildsStablePaths(t *testing.T) {
	doc, err := BuildStructuralModel()
	if err != nil {
		t.Fatal(err)
	}
	requireField(t, doc, "schema_version")
	requireField(t, doc, "project.name")
	requireField(t, doc, "comps[].name")
	requireField(t, doc, "comps[].layers[].type")
	requireField(t, doc, "comps[].layers[].transform.position_keyframes[].in_ease.influence")
	requireField(t, doc, "comps[].layers[].shape.gradient_fill.color_stops[].offset")
	requireField(t, doc, "expected_profile.keyframes[].keyframes[].value")
}

func TestReflectRecipeFieldsMarksStructuralRequiredness(t *testing.T) {
	doc, err := BuildStructuralModel()
	if err != nil {
		t.Fatal(err)
	}
	width := requireField(t, doc, "comps[].width")
	if !width.StructuralRequired {
		t.Fatalf("comps[].width should be structurally required")
	}
	background := requireField(t, doc, "comps[].background_color")
	if background.StructuralRequired {
		t.Fatalf("comps[].background_color should not be structurally required")
	}
}

func requireField(t *testing.T, doc Document, path string) FieldModel {
	t.Helper()
	for _, field := range doc.Fields {
		if field.Path == path {
			return field
		}
	}
	t.Fatalf("field %q not found", path)
	return FieldModel{}
}
```

- [x] **Step 2: Run test to verify it fails**

Run:

```powershell
go test ./internal/recipedoc -run TestReflectRecipeFields -count=1
```

Expected: FAIL because `internal/recipedoc` and `BuildStructuralModel` do not exist.

- [x] **Step 3: Implement model types**

Create `internal/recipedoc/model.go`:

```go
package recipedoc

type Document struct {
	SchemaVersion int          `json:"schema_version"`
	Fields        []FieldModel `json:"fields"`
}

type FieldModel struct {
	Path                  string          `json:"path"`
	JSONName              string          `json:"json_name"`
	SourceType            string          `json:"source_type"`
	SourceField           string          `json:"source_field"`
	Type                  TypeRef         `json:"type"`
	StructuralRequired    bool            `json:"structural_required"`
	ValidationRequired    bool            `json:"validation_required,omitempty"`
	Summary               string          `json:"summary,omitempty"`
	Validation            string          `json:"validation,omitempty"`
	Enum                  []string        `json:"enum,omitempty"`
	Capabilities          []CapabilityRef `json:"capabilities,omitempty"`
	Example               string          `json:"example,omitempty"`
	MissingSemanticSummary bool            `json:"missing_semantic_summary,omitempty"`
}

type TypeRef struct {
	Kind       string `json:"kind"`
	GoType     string `json:"go_type"`
	Element    string `json:"element,omitempty"`
	ObjectType string `json:"object_type,omitempty"`
}

type FieldMeta struct {
	Summary    string
	Validation string
	Enum       []string
	Example    string
}

type CapabilityRef struct {
	Key      string   `json:"key"`
	Query    string   `json:"query"`
	Status   string   `json:"status,omitempty"`
	Symbol   string   `json:"symbol,omitempty"`
	Domain   string   `json:"domain,omitempty"`
	Verify   string   `json:"verify,omitempty"`
	MinVer   string   `json:"minver,omitempty"`
	Boundary string   `json:"boundary,omitempty"`
	Gate     []string `json:"gate,omitempty"`
}

type CapabilityMeta struct {
	Query        string
	Summary      string
	AllowUnknown bool
}
```

- [x] **Step 4: Implement reflection walker**

Create `internal/recipedoc/reflect.go`. It should import
`github.com/yueli-fx/aep-parser/internal/recipe`, walk only the explicit root
type list, and derive paths from JSON tags.

Required behavior:

- `Recipe.Comps []CompSpec` becomes `comps[]`
- `Layer.Transform Transform` becomes `comps[].layers[].transform`
- `VectorKeyframe.InEase *TemporalEase` becomes
  `comps[].layers[].transform.position_keyframes[].in_ease`
- `omitempty` makes `StructuralRequired=false`
- pointers make `StructuralRequired=false`
- slices make `StructuralRequired=false` for the field itself unless the JSON
  tag lacks `omitempty`, while item child paths still exist under `[]`

Core function signatures:

```go
func BuildStructuralModel() (Document, error)
```

```go
func walkStruct(t reflect.Type, prefix string, out *[]FieldModel, seen map[reflect.Type]bool)
```

Use stable source order by iterating struct fields in declaration order. Do not
sort fields in the reflection layer.

- [x] **Step 5: Run structural tests**

Run:

```powershell
go test ./internal/recipedoc -run TestReflectRecipeFields -count=1
```

Expected: PASS.

- [x] **Step 6: Commit structural model**

Run:

```powershell
git add internal/recipedoc/model.go internal/recipedoc/reflect.go internal/recipedoc/reflect_test.go
git commit -m "feat: add recipe doc structural model"
```

## Task 2: Split Registries And Canonical Build

**Files:**
- Create: `internal/recipedoc/summary_registry.go`
- Create: `internal/recipedoc/validation_registry.go`
- Create: `internal/recipedoc/capability_registry.go`
- Create: `internal/recipedoc/example_registry.go`
- Create: `internal/recipedoc/build.go`
- Create: `internal/recipedoc/build_test.go`

- [x] **Step 1: Write registry validation tests**

Create `internal/recipedoc/build_test.go`:

```go
package recipedoc

import "testing"

func TestBuildDocumentRejectsUnknownRegistryPath(t *testing.T) {
	_, err := buildDocumentWithRegistries(registries{
		Summary: map[string]string{
			"comps[].does_not_exist": "bad path",
		},
	})
	if err == nil {
		t.Fatal("expected unknown registry path error")
	}
}

func TestBuildDocumentRejectsUnknownCapabilityKey(t *testing.T) {
	_, err := buildDocumentWithRegistries(registries{
		CapabilitiesByPath: map[string][]string{
			"comps[].background_color": {"missing.capability_key"},
		},
	})
	if err == nil {
		t.Fatal("expected unknown capability key error")
	}
}

func TestBuildDocumentJoinsFieldMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}
	field := requireField(t, doc, "comps[].background_color")
	if field.Summary == "" {
		t.Fatal("summary should be joined")
	}
	if len(field.Capabilities) == 0 || field.Capabilities[0].Key != "comp.set_background_color" {
		t.Fatalf("capabilities not joined: %+v", field.Capabilities)
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run:

```powershell
go test ./internal/recipedoc -run TestBuildDocument -count=1
```

Expected: FAIL because registry and build functions do not exist.

- [x] **Step 3: Implement split registries**

Create small first-slice registries. Start with high-value fields and allow
missing summaries explicitly in `build.go`; do not attempt all 300+ fields in
this task.

Required starting entries:

```go
var fieldSummaries = map[string]string{
	"schema_version":                 "Recipe schema version.",
	"project.name":                   "Project display name.",
	"comps[].name":                   "Composition display name.",
	"comps[].width":                  "Composition width in pixels.",
	"comps[].height":                 "Composition height in pixels.",
	"comps[].frame_rate":             "Composition frame rate in frames per second.",
	"comps[].duration":               "Composition duration in seconds.",
	"comps[].background_color":       "Composition background color as RGB channels.",
	"comps[].layers[].type":          "Layer creation type.",
	"comps[].layers[].name":          "Layer display name.",
	"comps[].layers[].transform":     "Layer transform block.",
	"comps[].layers[].effects[]":     "Built-in effect instance to add to the layer.",
	"expected_profile.keyframes[]":   "Expected keyframes for a layer property.",
}
```

Capability field mapping:

```go
var capabilitiesByPath = map[string][]string{
	"comps[].background_color": {"comp.set_background_color"},
	"comps[].renderer":         {"comp.set_renderer"},
	"comps[].layers[].type":    {"layer.create_text", "layer.create_shape", "layer.create_solid", "layer.create_camera", "layer.create_light", "layer.create_null", "layer.create_adjustment"},
	"comps[].layers[].transform": {"layer.set_transform"},
	"comps[].layers[].effects[]": {"effect.add_builtin"},
	"comps[].layers[].effects[].params[]": {"effect.set_param"},
	"comps[].layers[].effects[].params[].expression.source": {"property.set_expression"},
}
```

Stable key to capindex query mapping:

```go
var capabilityRegistry = map[string]CapabilityMeta{
	"comp.set_background_color": {Query: "SetBGColor", Summary: "Set composition background color."},
	"comp.set_renderer":         {Query: "SetRenderer", Summary: "Set composition renderer."},
	"layer.create_text":         {Query: "NewTextLayer", Summary: "Create a text layer."},
	"layer.create_shape":        {Query: "NewShapeLayer", Summary: "Create a shape layer."},
	"layer.create_solid":        {Query: "NewSolidLayer", Summary: "Create a solid layer."},
	"layer.create_camera":       {Query: "NewCameraLayer", Summary: "Create a camera layer."},
	"layer.create_light":        {Query: "NewLightLayer", Summary: "Create a light layer."},
	"layer.create_null":         {Query: "NewNullLayer", Summary: "Create a null layer."},
	"layer.create_adjustment":   {Query: "NewAdjustmentLayer", Summary: "Create an adjustment layer."},
	"layer.set_transform":       {Query: "SetLayerTransform", Summary: "Set layer transform values."},
	"effect.add_builtin":        {Query: "AddEffect", Summary: "Add a supported built-in effect."},
	"effect.set_param":          {Query: "SetEffectParam", Summary: "Set an effect parameter value."},
	"property.set_expression":   {Query: "Property.SetExpression", Summary: "Set a property expression."},
}
```

- [x] **Step 4: Implement canonical build**

Create `internal/recipedoc/build.go` with:

```go
func BuildDocument() (Document, error)
```

Also add an unexported test seam:

```go
type registries struct {
	Summary            map[string]string
	Validation         map[string]FieldMeta
	Examples           map[string]string
	CapabilitiesByPath map[string][]string
	Capabilities       map[string]CapabilityMeta
}

func buildDocumentWithRegistries(overrides registries) (Document, error)
```

`BuildDocument` should call reflection, then join default registries. It must
reject:

- registry path not found in reflected paths
- capability path not found in reflected paths
- capability key not found in capability registry

For first slice, capindex lookup can be joined in Task 5 from CLI; here the
document should at least carry key/query pairs.

- [x] **Step 5: Run registry tests**

Run:

```powershell
go test ./internal/recipedoc -run 'TestBuildDocument|TestReflectRecipeFields' -count=1
```

Expected: PASS.

- [x] **Step 6: Commit registries and canonical build**

Run:

```powershell
git add internal/recipedoc
git commit -m "feat: add recipe doc metadata registries"
```

## Task 3: JSON Schema-Like Renderer

**Files:**
- Create: `internal/recipedoc/render_json.go`
- Create: `internal/recipedoc/render_json_test.go`

- [x] **Step 1: Write JSON renderer tests**

Create `internal/recipedoc/render_json_test.go`:

```go
package recipedoc

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderJSONSchemaUsesJSONSchemaShape(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}
	out, err := RenderJSONSchema(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"$schema"`) {
		t.Fatal("missing $schema")
	}
	if !strings.Contains(out, `"x-aep-path": "comps[].background_color"`) {
		t.Fatal("missing x-aep-path for background color")
	}
	if !strings.Contains(out, `"x-aep-capabilities"`) {
		t.Fatal("missing capability extension")
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestRenderJSONSchemaIsDeterministic(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}
	a, err := RenderJSONSchema(doc)
	if err != nil {
		t.Fatal(err)
	}
	b, err := RenderJSONSchema(doc)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatal("schema output is not deterministic")
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run:

```powershell
go test ./internal/recipedoc -run TestRenderJSONSchema -count=1
```

Expected: FAIL because `RenderJSONSchema` does not exist.

- [x] **Step 3: Implement JSON renderer**

Create `internal/recipedoc/render_json.go`:

```go
func RenderJSONSchema(doc Document) (string, error)
```

Required output rules:

- top-level object contains `$schema`, `$id`, `title`, `type`, `properties`
- nested recipe structs appear under `$defs`
- fields include `description` from summary when present
- recipe-specific metadata uses `x-aep-*`
- structural required fields populate JSON Schema `required`
- validation required fields populate `x-aep-validation-required`
- stable path is always present in `x-aep-path`
- capability refs appear in `x-aep-capabilities`

Use `json.MarshalIndent` over deterministic structs and slices. Do not marshal
maps directly when field order matters; prebuild sorted key lists or stable
struct slices where needed.

- [x] **Step 4: Run JSON renderer tests**

Run:

```powershell
go test ./internal/recipedoc -run TestRenderJSONSchema -count=1
```

Expected: PASS.

- [x] **Step 5: Commit JSON renderer**

Run:

```powershell
git add internal/recipedoc/render_json.go internal/recipedoc/render_json_test.go
git commit -m "feat: render recipe JSON schema"
```

## Task 4: Markdown Renderer

**Files:**
- Create: `internal/recipedoc/render_md.go`
- Create: `internal/recipedoc/render_md_test.go`

- [x] **Step 1: Write Markdown renderer tests**

Create `internal/recipedoc/render_md_test.go`:

```go
package recipedoc

import (
	"strings"
	"testing"
)

func TestRenderMarkdownIncludesFieldReference(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}
	out := RenderMarkdown(doc)
	requireSubstring(t, out, "# Recipe Field Reference")
	requireSubstring(t, out, "## CompSpec")
	requireSubstring(t, out, "`comps[].background_color`")
	requireSubstring(t, out, "`comp.set_background_color`")
	requireSubstring(t, out, "structural")
}

func TestRenderMarkdownMarksMissingSemanticSummary(t *testing.T) {
	doc := Document{Fields: []FieldModel{{
		Path:                  "comps[].unknown_future_field",
		JSONName:              "unknown_future_field",
		SourceType:            "CompSpec",
		SourceField:           "UnknownFutureField",
		Type:                  TypeRef{Kind: "number", GoType: "float64"},
		MissingSemanticSummary: true,
	}}}
	out := RenderMarkdown(doc)
	requireSubstring(t, out, "missing semantic summary")
}

func requireSubstring(t *testing.T, s, want string) {
	t.Helper()
	if !strings.Contains(s, want) {
		t.Fatalf("missing %q in:\n%s", want, s)
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run:

```powershell
go test ./internal/recipedoc -run TestRenderMarkdown -count=1
```

Expected: FAIL because `RenderMarkdown` does not exist.

- [x] **Step 3: Implement Markdown renderer**

Create `internal/recipedoc/render_md.go`:

```go
func RenderMarkdown(doc Document) string
```

Required rendering:

- header: `# Recipe Field Reference`
- generated warning line
- group sections by `SourceType`
- table columns: Field Path, Type, Requiredness, Summary, Validation, Capability
- requiredness cell distinguishes `structural`, `validation`, both, or empty
- missing summary appears as `missing semantic summary`
- capability cell shows stable recipe key and capindex query if present

Keep renderer deterministic by preserving source order from `Document.Fields`
within each source type section.

- [x] **Step 4: Run Markdown renderer tests**

Run:

```powershell
go test ./internal/recipedoc -run TestRenderMarkdown -count=1
```

Expected: PASS.

- [x] **Step 5: Commit Markdown renderer**

Run:

```powershell
git add internal/recipedoc/render_md.go internal/recipedoc/render_md_test.go
git commit -m "feat: render recipe field reference"
```

## Task 5: CLI, Capindex Join, And Drift Gate

**Files:**
- Create: `cmd/recipedocgen/main.go`
- Create: `cmd/recipedocgen/docs_uptodate_test.go`
- Modify: `internal/recipedoc/build.go`
- Modify: `internal/recipedoc/model.go`

- [x] **Step 1: Write CLI drift test**

Create `cmd/recipedocgen/docs_uptodate_test.go`:

```go
package main

import (
	"os"
	"strings"
	"testing"
)

func TestRecipeDocsUpToDate(t *testing.T) {
	schema, markdown, err := generate("../../docs")
	if err != nil {
		t.Fatal(err)
	}
	assertFileCurrent(t, "../../docs/recipe_schema.json", schema)
	assertFileCurrent(t, "../../docs/recipe.md", markdown)
}

func assertFileCurrent(t *testing.T, path, got string) {
	t.Helper()
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	norm := func(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
	if norm(string(want)) != norm(got) {
		t.Fatalf("%s is stale; run go generate ./cmd/recipedocgen", path)
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run:

```powershell
go test ./cmd/recipedocgen -run TestRecipeDocsUpToDate -count=1
```

Expected: FAIL because `cmd/recipedocgen` and generated docs do not exist.

- [x] **Step 3: Implement CLI**

Create `cmd/recipedocgen/main.go`:

```go
package main

//go:generate go run . -out ../../docs

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yueli-fx/aep-parser/internal/recipedoc"
)

func main() {
	outDir := flag.String("out", "docs", "output docs directory")
	flag.Parse()
	schema, markdown, err := generate(*outDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "recipedocgen:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "recipedocgen:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(*outDir, "recipe_schema.json"), []byte(schema), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "recipedocgen:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(*outDir, "recipe.md"), []byte(markdown), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "recipedocgen:", err)
		os.Exit(1)
	}
}

func generate(docsDir string) (string, string, error) {
	doc, err := recipedoc.BuildDocumentWithCapabilities(filepath.Join(docsDir, "capabilities.json"))
	if err != nil {
		return "", "", err
	}
	schema, err := recipedoc.RenderJSONSchema(doc)
	if err != nil {
		return "", "", err
	}
	return schema, recipedoc.RenderMarkdown(doc), nil
}
```

- [x] **Step 4: Implement capindex join**

Add to `internal/recipedoc/build.go`:

```go
func BuildDocumentWithCapabilities(capabilitiesPath string) (Document, error)
```

This should load `internal/capindex.Load(capabilitiesPath)` and enrich each
`CapabilityRef` with:

- `Status`
- `Symbol`
- `Domain`
- `Verify`
- `MinVer`
- `Boundary`
- `Gate`

If the capindex lookup does not resolve and `CapabilityMeta.AllowUnknown` is
false, return an error.

- [x] **Step 5: Generate docs**

Run:

```powershell
go generate ./cmd/recipedocgen
```

Expected:

- `docs/recipe_schema.json` exists
- `docs/recipe.md` exists

- [x] **Step 6: Run CLI and drift tests**

Run:

```powershell
go test ./cmd/recipedocgen -count=1
go test ./internal/recipedoc -count=1
```

Expected: PASS.

- [x] **Step 7: Commit CLI and generated docs**

Run:

```powershell
git add cmd/recipedocgen internal/recipedoc docs/recipe_schema.json docs/recipe.md
git commit -m "feat: add recipe doc generator"
```

## Task 6: First-Slice Metadata Coverage And Final Verification

**Files:**
- Modify: `internal/recipedoc/summary_registry.go`
- Modify: `internal/recipedoc/validation_registry.go`
- Modify: `internal/recipedoc/capability_registry.go`
- Modify: `internal/recipedoc/example_registry.go`
- Modify: `docs/recipe_schema.json`
- Modify: `docs/recipe.md`
- Modify: `flightdeck/work/recipe-docgen/index.md`

- [x] **Step 1: Expand metadata for first-slice fields**

Cover at least these groups:

- top-level recipe fields
- composition basics and settings
- layer creation and common switches
- transform static values and transform keyframes
- effects and effect params
- shape primitives, stroke/fill, gradients, and shape filters
- text style fields
- `expected_profile` checks

Keep missing metadata explicit through a visible allowlist:

```go
var allowedMissingSemanticSummary = map[string]string{
	"expected_profile.masks[].path_keyframes[].vertices": "documented in later mask profile pass",
}
```

Do not allow broad prefixes in this map. Every entry must be a full stable path.

- [x] **Step 2: Add coverage test**

Add to `internal/recipedoc/build_test.go`:

```go
func TestNoUnexpectedMissingSemanticSummaries(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range doc.Fields {
		if field.MissingSemanticSummary {
			if _, ok := allowedMissingSemanticSummary[field.Path]; !ok {
				t.Fatalf("unexpected missing semantic summary for %s", field.Path)
			}
		}
	}
}
```

- [x] **Step 3: Regenerate docs**

Run:

```powershell
go generate ./cmd/recipedocgen
```

Expected: `docs/recipe_schema.json` and `docs/recipe.md` update deterministically.

- [x] **Step 4: Run full verification**

Run:

```powershell
go test ./cmd/recipedocgen -count=1
go test ./internal/recipedoc -count=1
go test ./internal/recipe -count=1
go test ./...
go vet ./...
git diff --check
```

Expected:

- all Go tests pass
- vet exits 0
- diff check has no whitespace errors; CRLF warnings are acceptable on Windows

- [x] **Step 5: Update work index**

In `flightdeck/work/recipe-docgen/index.md`, update:

- `Progress` Done: add "First implementation slice landed."
- `Current`: "Ready for next metadata expansion or editor integration."

- [x] **Step 6: Commit final metadata pass**

Run:

```powershell
git add internal/recipedoc docs/recipe_schema.json docs/recipe.md flightdeck/work/recipe-docgen/index.md
git commit -m "docs: expand recipe doc metadata"
```

## Final Acceptance

The implementation is complete for the first slice when these commands pass:

```powershell
go test ./cmd/recipedocgen -count=1
go test ./internal/recipedoc -count=1
go test ./internal/recipe -count=1
go test ./...
go vet ./...
git diff --check
```

The final response should report:

- commits created
- generated files
- verification commands and outcomes
- remaining optional scope: richer examples and editor integration
