package recipedoc

import (
	"fmt"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/capindex"
)

type registries struct {
	Summary            map[string]string
	Validation         map[string]FieldMeta
	Examples           map[string]string
	CapabilitiesByPath map[string][]string
	Capabilities       map[string]CapabilityMeta
}

func BuildDocument() (Document, error) {
	return buildDocumentWithRegistries(registries{})
}

func BuildDocumentWithCapabilities(capabilitiesPath string) (Document, error) {
	doc, err := BuildDocument()
	if err != nil {
		return Document{}, err
	}
	idx, err := capindex.Load(capabilitiesPath)
	if err != nil {
		return Document{}, err
	}
	metaByKey := capabilityRegistry
	for i := range doc.Fields {
		for j := range doc.Fields[i].Capabilities {
			capRef := &doc.Fields[i].Capabilities[j]
			meta, ok := metaByKey[capRef.Key]
			if !ok {
				return Document{}, fmt.Errorf("recipe doc capability key %q is not registered", capRef.Key)
			}
			got := idx.Lookup(meta.Query)
			if got.Entry.Symbol == "" {
				if meta.AllowUnknown {
					continue
				}
				return Document{}, fmt.Errorf("recipe doc capability query %q for key %q not found", meta.Query, capRef.Key)
			}
			capRef.Query = got.Query
			capRef.Status = string(got.Status)
			capRef.Symbol = got.Entry.Symbol
			if got.Entry.Recv != "" {
				capRef.Symbol = strings.TrimPrefix(got.Entry.Recv, "*") + "." + got.Entry.Symbol
			}
			capRef.Domain = got.Entry.Cap.Domain
			capRef.Verify = got.Entry.Cap.Verify
			capRef.MinVer = got.Entry.Cap.MinVer
			capRef.Boundary = got.Entry.Cap.Boundary
			capRef.Gate = append([]string(nil), got.Entry.Cap.Gate...)
		}
	}
	return doc, nil
}

func buildDocumentWithRegistries(overrides registries) (Document, error) {
	doc, err := BuildStructuralModel()
	if err != nil {
		return Document{}, err
	}
	reg := defaultRegistries()
	mergeRegistries(&reg, overrides)

	fieldByPath := map[string]int{}
	for i, field := range doc.Fields {
		fieldByPath[field.Path] = i
	}
	if err := validateRegistryPaths(fieldByPath, reg); err != nil {
		return Document{}, err
	}
	for i := range doc.Fields {
		field := &doc.Fields[i]
		field.Summary = reg.Summary[field.Path]
		if meta, ok := reg.Validation[field.Path]; ok {
			field.Validation = meta.Validation
			field.Enum = append([]string(nil), meta.Enum...)
			if meta.Summary != "" && field.Summary == "" {
				field.Summary = meta.Summary
			}
			if meta.Example != "" {
				field.Example = meta.Example
			}
		}
		if example, ok := reg.Examples[field.Path]; ok {
			field.Example = example
		}
		for _, key := range reg.CapabilitiesByPath[field.Path] {
			meta, ok := reg.Capabilities[key]
			if !ok {
				return Document{}, fmt.Errorf("recipe doc capability key %q for %s is not registered", key, field.Path)
			}
			field.Capabilities = append(field.Capabilities, CapabilityRef{
				Key:   key,
				Query: meta.Query,
			})
		}
		field.MissingSemanticSummary = field.Summary == ""
	}
	return doc, nil
}

func defaultRegistries() registries {
	return registries{
		Summary:            cloneStringMap(fieldSummaries),
		Validation:         cloneFieldMetaMap(fieldValidation),
		Examples:           cloneStringMap(fieldExamples),
		CapabilitiesByPath: cloneStringSliceMap(capabilitiesByPath),
		Capabilities:       cloneCapabilityMetaMap(capabilityRegistry),
	}
}

func mergeRegistries(dst *registries, src registries) {
	for k, v := range src.Summary {
		dst.Summary[k] = v
	}
	for k, v := range src.Validation {
		dst.Validation[k] = v
	}
	for k, v := range src.Examples {
		dst.Examples[k] = v
	}
	for k, v := range src.CapabilitiesByPath {
		dst.CapabilitiesByPath[k] = append([]string(nil), v...)
	}
	for k, v := range src.Capabilities {
		dst.Capabilities[k] = v
	}
}

func validateRegistryPaths(fieldByPath map[string]int, reg registries) error {
	for path := range reg.Summary {
		if _, ok := fieldByPath[path]; !ok {
			return fmt.Errorf("recipe doc summary path %q does not exist", path)
		}
	}
	for path := range reg.Validation {
		if _, ok := fieldByPath[path]; !ok {
			return fmt.Errorf("recipe doc validation path %q does not exist", path)
		}
	}
	for path := range reg.Examples {
		if _, ok := fieldByPath[path]; !ok {
			return fmt.Errorf("recipe doc example path %q does not exist", path)
		}
	}
	for path, keys := range reg.CapabilitiesByPath {
		if _, ok := fieldByPath[path]; !ok {
			return fmt.Errorf("recipe doc capability path %q does not exist", path)
		}
		for _, key := range keys {
			if _, ok := reg.Capabilities[key]; !ok {
				return fmt.Errorf("recipe doc capability key %q for %s is not registered", key, path)
			}
		}
	}
	return nil
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneStringSliceMap(in map[string][]string) map[string][]string {
	out := make(map[string][]string, len(in))
	for k, v := range in {
		out[k] = append([]string(nil), v...)
	}
	return out
}

func cloneFieldMetaMap(in map[string]FieldMeta) map[string]FieldMeta {
	out := make(map[string]FieldMeta, len(in))
	for k, v := range in {
		v.Enum = append([]string(nil), v.Enum...)
		out[k] = v
	}
	return out
}

func cloneCapabilityMetaMap(in map[string]CapabilityMeta) map[string]CapabilityMeta {
	out := make(map[string]CapabilityMeta, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
