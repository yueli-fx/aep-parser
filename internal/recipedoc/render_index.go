package recipedoc

import "encoding/json"

type recipeIndexDocument struct {
	SchemaVersion int                   `json:"schema_version"`
	Paths         []string              `json:"paths"`
	Fields        map[string]indexField `json:"fields"`
	BySourceType  map[string][]string   `json:"by_source_type"`
}

type indexField struct {
	Path                   string   `json:"path"`
	JSONName               string   `json:"json_name"`
	SourceType             string   `json:"source_type"`
	SourceField            string   `json:"source_field"`
	Type                   string   `json:"type"`
	GoType                 string   `json:"go_type"`
	Requiredness           []string `json:"requiredness,omitempty"`
	Summary                string   `json:"summary,omitempty"`
	Validation             string   `json:"validation,omitempty"`
	Enum                   []string `json:"enum,omitempty"`
	CapabilityKeys         []string `json:"capability_keys,omitempty"`
	Example                string   `json:"example,omitempty"`
	MissingSemanticSummary bool     `json:"missing_semantic_summary,omitempty"`
}

func RenderRecipeIndex(doc Document) (string, error) {
	index := recipeIndexDocument{
		SchemaVersion: doc.SchemaVersion,
		Fields:        map[string]indexField{},
		BySourceType:  map[string][]string{},
	}
	for _, field := range doc.Fields {
		index.Paths = append(index.Paths, field.Path)
		index.BySourceType[field.SourceType] = append(index.BySourceType[field.SourceType], field.Path)
		index.Fields[field.Path] = indexField{
			Path:                   field.Path,
			JSONName:               field.JSONName,
			SourceType:             field.SourceType,
			SourceField:            field.SourceField,
			Type:                   typeLabel(field.Type),
			GoType:                 field.Type.GoType,
			Requiredness:           requirednessValues(field),
			Summary:                field.Summary,
			Validation:             field.Validation,
			Enum:                   append([]string(nil), field.Enum...),
			CapabilityKeys:         capabilityKeys(field.Capabilities),
			Example:                field.Example,
			MissingSemanticSummary: field.MissingSemanticSummary,
		}
	}
	out, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out) + "\n", nil
}

func requirednessValues(field FieldModel) []string {
	var out []string
	if field.StructuralRequired {
		out = append(out, "structural")
	}
	if field.ValidationRequired {
		out = append(out, "validation")
	}
	return out
}

func capabilityKeys(caps []CapabilityRef) []string {
	out := make([]string, 0, len(caps))
	for _, cap := range caps {
		out = append(out, cap.Key)
	}
	return out
}
