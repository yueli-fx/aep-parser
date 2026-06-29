package recipedoc

import (
	"encoding/json"
	"strings"
)

type schemaDocument struct {
	Schema     string                    `json:"$schema"`
	ID         string                    `json:"$id"`
	Title      string                    `json:"title"`
	Type       string                    `json:"type"`
	Properties map[string]schemaProperty `json:"properties"`
	Required   []string                  `json:"required,omitempty"`
	XAEPFields []schemaField             `json:"x-aep-fields"`
}

type schemaProperty struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	XAEPPath    string `json:"x-aep-path"`
}

type schemaField struct {
	Path                   string          `json:"x-aep-path"`
	JSONName               string          `json:"json_name"`
	SourceType             string          `json:"source_type"`
	SourceField            string          `json:"source_field"`
	Type                   string          `json:"type"`
	GoType                 string          `json:"go_type"`
	Element                string          `json:"element,omitempty"`
	ObjectType             string          `json:"object_type,omitempty"`
	StructuralRequired     bool            `json:"x-aep-structural-required,omitempty"`
	ValidationRequired     bool            `json:"x-aep-validation-required,omitempty"`
	Description            string          `json:"description,omitempty"`
	Validation             string          `json:"x-aep-validation,omitempty"`
	Enum                   []string        `json:"enum,omitempty"`
	XAEPCapabilities       []CapabilityRef `json:"x-aep-capabilities,omitempty"`
	Example                string          `json:"x-aep-example,omitempty"`
	MissingSemanticSummary bool            `json:"x-aep-missing-semantic-summary,omitempty"`
}

func RenderJSONSchema(doc Document) (string, error) {
	schema := schemaDocument{
		Schema:     "https://json-schema.org/draft/2020-12/schema",
		ID:         "https://github.com/yueli-fx/aep-parser/docs/recipe_schema.json",
		Title:      "aep-parser recipe schema",
		Type:       "object",
		Properties: map[string]schemaProperty{},
	}
	for _, field := range doc.Fields {
		if isTopLevelPath(field.Path) {
			schema.Properties[field.JSONName] = schemaProperty{
				Type:        jsonSchemaType(field.Type),
				Description: field.Summary,
				XAEPPath:    field.Path,
			}
			if field.StructuralRequired {
				schema.Required = append(schema.Required, field.JSONName)
			}
		}
		schema.XAEPFields = append(schema.XAEPFields, schemaField{
			Path:                   field.Path,
			JSONName:               field.JSONName,
			SourceType:             field.SourceType,
			SourceField:            field.SourceField,
			Type:                   jsonSchemaType(field.Type),
			GoType:                 field.Type.GoType,
			Element:                field.Type.Element,
			ObjectType:             field.Type.ObjectType,
			StructuralRequired:     field.StructuralRequired,
			ValidationRequired:     field.ValidationRequired,
			Description:            field.Summary,
			Validation:             field.Validation,
			Enum:                   append([]string(nil), field.Enum...),
			XAEPCapabilities:       append([]CapabilityRef(nil), field.Capabilities...),
			Example:                field.Example,
			MissingSemanticSummary: field.MissingSemanticSummary,
		})
	}
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out) + "\n", nil
}

func isTopLevelPath(path string) bool {
	return !strings.Contains(path, ".") && !strings.Contains(path, "[]")
}

func jsonSchemaType(ref TypeRef) string {
	switch ref.Kind {
	case "string", "boolean", "array", "object":
		return ref.Kind
	case "number":
		return "number"
	case "any":
		return "object"
	default:
		return "string"
	}
}
