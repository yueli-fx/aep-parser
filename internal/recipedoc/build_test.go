package recipedoc

import (
	"strings"
	"testing"
)

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

func TestNoUnexpectedMissingSemanticSummaries(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	var missing []string
	for _, field := range doc.Fields {
		if field.MissingSemanticSummary {
			if _, ok := allowedMissingSemanticSummary[field.Path]; !ok {
				missing = append(missing, field.Path)
			}
		}
	}
	if len(missing) > 0 {
		t.Fatalf("unexpected missing semantic summaries:\n%s", strings.Join(missing, "\n"))
	}
}
