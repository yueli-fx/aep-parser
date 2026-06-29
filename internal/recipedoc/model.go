package recipedoc

type Document struct {
	SchemaVersion int          `json:"schema_version"`
	Fields        []FieldModel `json:"fields"`
}

type FieldModel struct {
	Path                   string          `json:"path"`
	JSONName               string          `json:"json_name"`
	SourceType             string          `json:"source_type"`
	SourceField            string          `json:"source_field"`
	Type                   TypeRef         `json:"type"`
	StructuralRequired     bool            `json:"structural_required"`
	ValidationRequired     bool            `json:"validation_required,omitempty"`
	Summary                string          `json:"summary,omitempty"`
	Validation             string          `json:"validation,omitempty"`
	Enum                   []string        `json:"enum,omitempty"`
	Capabilities           []CapabilityRef `json:"capabilities,omitempty"`
	Example                string          `json:"example,omitempty"`
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
