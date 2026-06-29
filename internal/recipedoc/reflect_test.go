package recipedoc

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

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

func TestReflectRecipeFieldsCoversInputSchemaTypes(t *testing.T) {
	doc, err := BuildStructuralModel()
	if err != nil {
		t.Fatal(err)
	}

	sourceFields := map[string]map[string]bool{}
	for _, field := range doc.Fields {
		if sourceFields[field.SourceType] == nil {
			sourceFields[field.SourceType] = map[string]bool{}
		}
		sourceFields[field.SourceType][field.SourceField] = true
	}

	structFields := parseRecipeSchemaStructFields(t)
	var missing []string
	for structName, fields := range structFields {
		if reportOnlyRecipeStructs[structName] {
			if sourceFields[structName] != nil {
				missing = append(missing, structName+" should stay out of recipe docs")
			}
			continue
		}
		if sourceFields[structName] == nil {
			missing = append(missing, structName)
			continue
		}
		for _, fieldName := range fields {
			if !sourceFields[structName][fieldName] {
				missing = append(missing, structName+"."+fieldName)
			}
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("structural model does not cover recipe schema fields:\n%s", strings.Join(missing, "\n"))
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

var reportOnlyRecipeStructs = map[string]bool{
	"CapabilityUse": true,
	"Downgrade":     true,
	"ProfileCheck":  true,
	"Refusal":       true,
	"Report":        true,
}

func parseRecipeSchemaStructFields(t *testing.T) map[string][]string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../recipe/schema_types.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	out := map[string][]string{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 || field.Tag == nil {
					continue
				}
				tag, err := strconv.Unquote(field.Tag.Value)
				if err != nil {
					t.Fatal(err)
				}
				jsonTag := reflect.StructTag(tag).Get("json")
				if jsonTag == "" || jsonTag == "-" {
					continue
				}
				out[typeSpec.Name.Name] = append(out[typeSpec.Name.Name], field.Names[0].Name)
			}
		}
	}
	return out
}
