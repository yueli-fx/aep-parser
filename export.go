package aep

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"reflect"

	internal "github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/serializer"
)

const ExportSchemaVersion = 1

const ExportPreservationModeByteExact = "byte-exact-outside-claimed-ranges"

type ExportOperation string

const ExportSetStaticValue ExportOperation = "set-static-value"

type ExportStatus string

const (
	ExportStatusVerified ExportStatus = "verified"
	ExportStatusRejected ExportStatus = "rejected"
	ExportStatusFailed   ExportStatus = "failed"
)

type ExportOperationStatus string

const (
	ExportOperationApplied    ExportOperationStatus = "applied"
	ExportOperationDowngraded ExportOperationStatus = "downgraded"
	ExportOperationRejected   ExportOperationStatus = "rejected"
)

// WriteTarget identifies one property only within the source document named
// by DocumentRef. It is deliberately separate from ProjectJSON property_ref,
// which remains a detached-snapshot association and is not a write locator.
type WriteTarget struct {
	DocumentRef   string `json:"document_ref"`
	CompositionID uint32 `json:"composition_id"`
	LayerID       uint32 `json:"layer_id"`
	CanonicalPath string `json:"canonical_path"`
}

// ExportRequest is a versioned, document-scoped change set. The first public
// version supports length-preserving numeric static-property edits.
type ExportRequest struct {
	SchemaVersion int            `json:"schema_version"`
	Changes       []ExportChange `json:"changes"`
}

type ExportChange struct {
	ID        string          `json:"id"`
	Operation ExportOperation `json:"operation"`
	Target    WriteTarget     `json:"target"`
	// Value always uses component form: one number for a scalar, N numbers for
	// an N-dimensional property.
	Value []float64 `json:"value"`
}

type ExportReport struct {
	SchemaVersion int                     `json:"schema_version"`
	Status        ExportStatus            `json:"status"`
	DocumentRef   string                  `json:"document_ref,omitempty"`
	Operations    []ExportOperationResult `json:"operations,omitempty"`
	Preservation  ExportPreservation      `json:"preservation"`
	Verification  ExportVerification      `json:"verification"`
	OutputBytes   int                     `json:"output_bytes,omitempty"`
}

type ExportOperationResult struct {
	ID      string                `json:"id"`
	Status  ExportOperationStatus `json:"status"`
	Code    string                `json:"code,omitempty"`
	Message string                `json:"message,omitempty"`
}

type ExportPreservation struct {
	Verified     bool   `json:"verified"`
	Mode         string `json:"mode,omitempty"`
	SHA256       string `json:"sha256,omitempty"`
	NodeCount    int    `json:"node_count,omitempty"`
	DataBytes    int    `json:"data_bytes,omitempty"`
	ClaimedBytes int    `json:"claimed_bytes,omitempty"`
}

type ExportVerification struct {
	Reparsed          bool `json:"reparsed"`
	OperationsMatched int  `json:"operations_matched"`
}

type preparedExportChange struct {
	request  ExportChange
	property *internal.Property
	value    any
}

// Export applies a versioned change set to a private clone of d, serializes
// it, reparses it, verifies every declared result and proves that all bytes
// outside the claimed mutation ranges are unchanged. Output is committed to w
// only after every verification succeeds; d itself is never mutated.
func (d *Document) Export(ctx context.Context, request ExportRequest, w io.Writer) (ExportReport, error) {
	report := ExportReport{SchemaVersion: ExportSchemaVersion, Status: ExportStatusRejected}
	if d == nil || d.project == nil {
		return report, fmt.Errorf("aep: nil document")
	}
	if ctx == nil {
		return report, fmt.Errorf("aep: nil context")
	}
	if w == nil {
		return report, fmt.Errorf("aep: nil writer")
	}
	if err := ctx.Err(); err != nil {
		return report, err
	}

	original, documentRef, err := serializeDocument(d.project)
	if err != nil {
		return report, err
	}
	report.DocumentRef = documentRef
	if request.SchemaVersion != ExportSchemaVersion {
		report.Operations = rejectAllChanges(request.Changes, "unsupported-schema-version", fmt.Sprintf("schema_version must be %d", ExportSchemaVersion))
		return report, nil
	}
	if len(request.Changes) == 0 {
		report.Operations = []ExportOperationResult{{Status: ExportOperationRejected, Code: "empty-change-set", Message: "at least one change is required"}}
		return report, nil
	}

	clone, err := internal.FromReader(bytes.NewReader(original))
	if err != nil {
		return report, fmt.Errorf("aep: clone source document: %w", err)
	}
	prepared := make([]preparedExportChange, 0, len(request.Changes))
	report.Operations = make([]ExportOperationResult, len(request.Changes))
	ids := make(map[string]bool, len(request.Changes))
	targets := make(map[WriteTarget]bool, len(request.Changes))
	rejected := false
	for index, change := range request.Changes {
		result := ExportOperationResult{ID: change.ID, Status: ExportOperationRejected}
		report.Operations[index] = result
		if change.ID == "" {
			rejectExportOperation(&report.Operations[index], "missing-operation-id", "change id is required")
			rejected = true
			continue
		}
		if ids[change.ID] {
			rejectExportOperation(&report.Operations[index], "duplicate-operation-id", "change id must be unique")
			rejected = true
			continue
		}
		ids[change.ID] = true
		if change.Operation != ExportSetStaticValue {
			rejectExportOperation(&report.Operations[index], "unsupported-operation", fmt.Sprintf("operation %q is not supported", change.Operation))
			rejected = true
			continue
		}
		if change.Target.DocumentRef == "" || change.Target.DocumentRef != documentRef {
			rejectExportOperation(&report.Operations[index], "document-ref-mismatch", "write target does not belong to this document")
			rejected = true
			continue
		}
		if targets[change.Target] {
			rejectExportOperation(&report.Operations[index], "duplicate-write-target", "a write target may appear only once in a change set")
			rejected = true
			continue
		}
		targets[change.Target] = true
		property := resolveWriteTarget(clone, change.Target)
		if property == nil {
			rejectExportOperation(&report.Operations[index], "write-target-not-found", "write target cannot be resolved in this document")
			rejected = true
			continue
		}
		capabilities := property.MutationCapabilities()
		if !capabilities.Known || !capabilities.StaticValue || property.StaticValue == nil || len(property.Keyframes) > 0 {
			rejectExportOperation(&report.Operations[index], "static-value-write-unsupported", "property has no verified static-value writer")
			rejected = true
			continue
		}
		if len(change.Value) != property.Components {
			rejectExportOperation(&report.Operations[index], "value-dimension-mismatch", fmt.Sprintf("value has %d components; property requires %d", len(change.Value), property.Components))
			rejected = true
			continue
		}
		finite := true
		for _, value := range change.Value {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				finite = false
				break
			}
		}
		if !finite {
			rejectExportOperation(&report.Operations[index], "non-finite-value", "property values must be finite numbers")
			rejected = true
			continue
		}
		var value any = append([]float64(nil), change.Value...)
		if property.Components == 1 {
			value = change.Value[0]
		}
		prepared = append(prepared, preparedExportChange{request: change, property: property, value: value})
	}
	if rejected {
		for index := range report.Operations {
			if report.Operations[index].Code == "" {
				rejectExportOperation(&report.Operations[index], "batch-rejected", "change was not applied because another change in the batch was rejected")
			}
		}
		return report, nil
	}
	if err := ctx.Err(); err != nil {
		return report, err
	}

	properties := make([]*internal.Property, 0, len(prepared))
	for _, change := range prepared {
		properties = append(properties, change.property)
	}
	before, err := serializer.StaticValuePreservationDigest(clone, properties)
	if err != nil {
		return report, err
	}
	for _, change := range prepared {
		if err := change.property.SetStaticValue(change.value); err != nil {
			report.Status = ExportStatusFailed
			return report, fmt.Errorf("aep: apply change %q: %w", change.request.ID, err)
		}
	}

	var output bytes.Buffer
	if err := clone.WriteAEP(&output); err != nil {
		report.Status = ExportStatusFailed
		return report, fmt.Errorf("aep: write changed document: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return report, err
	}
	verified, err := internal.FromReader(bytes.NewReader(output.Bytes()))
	if err != nil {
		report.Status = ExportStatusFailed
		return report, fmt.Errorf("aep: reparse changed document: %w", err)
	}
	report.Verification.Reparsed = true
	verifiedProperties := make([]*internal.Property, 0, len(prepared))
	for _, change := range prepared {
		property := resolveWriteTarget(verified, change.request.Target)
		if property == nil || !reflect.DeepEqual(property.StaticValue, change.value) {
			report.Status = ExportStatusFailed
			return report, fmt.Errorf("aep: verification mismatch for change %q", change.request.ID)
		}
		verifiedProperties = append(verifiedProperties, property)
		report.Verification.OperationsMatched++
	}
	after, err := serializer.StaticValuePreservationDigest(verified, verifiedProperties)
	if err != nil {
		report.Status = ExportStatusFailed
		return report, err
	}
	if before != after {
		report.Status = ExportStatusFailed
		return report, fmt.Errorf("aep: preservation verification failed outside claimed mutation bytes")
	}

	if _, err := io.Copy(w, bytes.NewReader(output.Bytes())); err != nil {
		report.Status = ExportStatusFailed
		return report, err
	}
	for index := range report.Operations {
		report.Operations[index].Status = ExportOperationApplied
	}
	report.Status = ExportStatusVerified
	report.OutputBytes = output.Len()
	report.Preservation = ExportPreservation{
		Verified: true, Mode: ExportPreservationModeByteExact, SHA256: after.SHA256,
		NodeCount: after.NodeCount, DataBytes: after.DataBytes, ClaimedBytes: after.RedactedBytes,
	}
	return report, nil
}

func serializeDocument(project *internal.Project) ([]byte, string, error) {
	var data bytes.Buffer
	if err := project.WriteAEP(&data); err != nil {
		return nil, "", err
	}
	digest := sha256.Sum256(data.Bytes())
	return data.Bytes(), fmt.Sprintf("sha256:%x", digest[:]), nil
}

func projectDocumentRef(project *internal.Project) (string, error) {
	_, ref, err := serializeDocument(project)
	return ref, err
}

func resolveWriteTarget(project *internal.Project, target WriteTarget) *internal.Property {
	if project == nil || target.CompositionID == 0 || target.LayerID == 0 || target.CanonicalPath == "" {
		return nil
	}
	composition := project.CompositionByID(target.CompositionID)
	if composition == nil {
		return nil
	}
	layer := composition.LayerByID(target.LayerID)
	if layer == nil || layer.PropertyTree() == nil {
		return nil
	}
	return findPropertyAtCanonicalPath(layer.PropertyTree(), "", target.CanonicalPath)
}

func findPropertyAtCanonicalPath(group *internal.AEPropertyGroup, parentPath, targetPath string) *internal.Property {
	occurrences := map[string]int{}
	for index := 0; index < group.NumProperties(); index++ {
		child := group.ChildByIndex(index)
		if child == nil {
			continue
		}
		matchName := child.PropertyMatchName()
		occurrence := occurrences[matchName]
		occurrences[matchName] = occurrence + 1
		path := canonicalPropertyPath(parentPath, index, matchName, occurrence)
		switch typed := child.(type) {
		case *internal.Property:
			if path == targetPath {
				return typed
			}
		case *internal.AEPropertyGroup:
			if property := findPropertyAtCanonicalPath(typed, path, targetPath); property != nil {
				return property
			}
		}
	}
	return nil
}

func rejectAllChanges(changes []ExportChange, code, message string) []ExportOperationResult {
	if len(changes) == 0 {
		return []ExportOperationResult{{Status: ExportOperationRejected, Code: code, Message: message}}
	}
	result := make([]ExportOperationResult, len(changes))
	for index, change := range changes {
		result[index] = ExportOperationResult{ID: change.ID, Status: ExportOperationRejected, Code: code, Message: message}
	}
	return result
}

func rejectExportOperation(result *ExportOperationResult, code, message string) {
	result.Status = ExportOperationRejected
	result.Code = code
	result.Message = message
}
