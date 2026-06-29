package recipedoc

import (
	"reflect"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/recipe"
)

func BuildStructuralModel() (Document, error) {
	var fields []FieldModel
	walkStruct(reflect.TypeOf(recipe.Recipe{}), "", &fields, map[reflect.Type]bool{})
	return Document{SchemaVersion: recipe.SchemaVersion, Fields: fields}, nil
}

func walkStruct(t reflect.Type, prefix string, out *[]FieldModel, seen map[reflect.Type]bool) {
	t = derefType(t)
	if t.Kind() != reflect.Struct {
		return
	}
	if t.PkgPath() != reflect.TypeOf(recipe.Recipe{}).PkgPath() {
		return
	}
	if seen[t] {
		return
	}
	nextSeen := cloneSeen(seen)
	nextSeen[t] = true

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue
		}
		name, omitempty, ok := jsonFieldName(sf)
		if !ok {
			continue
		}
		path := joinPath(prefix, name)
		fieldType := sf.Type
		typeRef := typeRefOf(fieldType)
		*out = append(*out, FieldModel{
			Path:               displayPath(path, fieldType),
			JSONName:           name,
			SourceType:         t.Name(),
			SourceField:        sf.Name,
			Type:               typeRef,
			StructuralRequired: structuralRequired(sf.Type, omitempty),
		})
		walkNested(fieldType, displayPath(path, fieldType), out, nextSeen)
	}
}

func walkNested(t reflect.Type, prefix string, out *[]FieldModel, seen map[reflect.Type]bool) {
	t = derefType(t)
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = derefType(t.Elem())
	}
	if t.Kind() != reflect.Struct {
		return
	}
	if t.PkgPath() != reflect.TypeOf(recipe.Recipe{}).PkgPath() {
		return
	}
	walkStruct(t, prefix, out, seen)
}

func jsonFieldName(sf reflect.StructField) (string, bool, bool) {
	tag := sf.Tag.Get("json")
	if tag == "-" {
		return "", false, false
	}
	parts := strings.Split(tag, ",")
	name := parts[0]
	if name == "" {
		name = sf.Name
	}
	omitempty := false
	for _, part := range parts[1:] {
		if part == "omitempty" {
			omitempty = true
			break
		}
	}
	return name, omitempty, true
}

func structuralRequired(t reflect.Type, omitempty bool) bool {
	if omitempty {
		return false
	}
	return t.Kind() != reflect.Pointer
}

func typeRefOf(t reflect.Type) TypeRef {
	original := t
	t = derefType(t)
	ref := TypeRef{GoType: original.String()}
	switch t.Kind() {
	case reflect.String:
		ref.Kind = "string"
	case reflect.Bool:
		ref.Kind = "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		ref.Kind = "number"
	case reflect.Slice, reflect.Array:
		elem := derefType(t.Elem())
		ref.Kind = "array"
		ref.Element = elem.String()
		if elem.Kind() == reflect.Struct {
			ref.ObjectType = elem.Name()
		}
	case reflect.Struct:
		ref.Kind = "object"
		ref.ObjectType = t.Name()
	case reflect.Interface:
		ref.Kind = "any"
	default:
		ref.Kind = t.Kind().String()
	}
	return ref
}

func displayPath(path string, t reflect.Type) string {
	t = derefType(t)
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		elem := derefType(t.Elem())
		if elem.Kind() == reflect.Struct && elem.PkgPath() == reflect.TypeOf(recipe.Recipe{}).PkgPath() {
			return path + "[]"
		}
	}
	return path
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func cloneSeen(in map[reflect.Type]bool) map[reflect.Type]bool {
	out := make(map[reflect.Type]bool, len(in)+1)
	for k, v := range in {
		out[k] = v
	}
	return out
}
