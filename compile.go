package aep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/yueli-fx/aep-parser/internal/recipe"
)

const CompileSchemaVersion = recipe.SchemaVersion

type CompileStatus string

const (
	CompileStatusVerified CompileStatus = "verified"
	CompileStatusRejected CompileStatus = "rejected"
	CompileStatusFailed   CompileStatus = "failed"
)

// CompileReport describes validation, capability downgrades, refusals and the
// final parse verification without exposing the internal Recipe compiler.
type CompileReport struct {
	SchemaVersion int                 `json:"schema_version"`
	Status        CompileStatus       `json:"status"`
	Capabilities  []CompileCapability `json:"capabilities,omitempty"`
	Downgraded    []CompileIssue      `json:"downgraded,omitempty"`
	Rejected      []CompileIssue      `json:"rejected,omitempty"`
	Verification  CompileVerification `json:"verification"`
	OutputBytes   int                 `json:"output_bytes,omitempty"`
}

type CompileCapability struct {
	Path     string   `json:"path,omitempty"`
	Query    string   `json:"query"`
	Status   string   `json:"status"`
	Symbol   string   `json:"symbol,omitempty"`
	Domain   string   `json:"domain,omitempty"`
	Tier     string   `json:"tier,omitempty"`
	Verify   string   `json:"verify,omitempty"`
	MinVer   string   `json:"minver,omitempty"`
	Boundary string   `json:"boundary,omitempty"`
	Gate     []string `json:"gate,omitempty"`
}

type CompileIssue struct {
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Query   string `json:"query,omitempty"`
	Message string `json:"message,omitempty"`
}

type CompileVerification struct {
	Reparsed bool `json:"reparsed"`
}

// Compile creates an AEP from a versioned Recipe JSON document. The Recipe is
// decoded strictly, compiled behind the public package seam and parsed again
// before any bytes are copied to w. Internal scene, Recipe and writer backing
// types are deliberately not part of this interface.
func Compile(ctx context.Context, recipeJSON []byte, w io.Writer) (CompileReport, error) {
	report := CompileReport{SchemaVersion: CompileSchemaVersion, Status: CompileStatusRejected}
	if ctx == nil {
		return report, fmt.Errorf("aep: nil context")
	}
	if w == nil {
		return report, fmt.Errorf("aep: nil writer")
	}
	if err := ctx.Err(); err != nil {
		return report, err
	}

	var specification recipe.Recipe
	decoder := json.NewDecoder(bytes.NewReader(recipeJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&specification); err != nil {
		report.Rejected = []CompileIssue{{Code: "invalid-json", Path: "$", Message: err.Error()}}
		return report, nil
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		report.Rejected = []CompileIssue{{Code: "trailing-json", Path: "$", Message: err.Error()}}
		return report, nil
	}

	var compiled bytes.Buffer
	internalReport, err := recipe.CompileToWriter(specification, &compiled, "compiled.aep", recipe.StaticCapabilities{})
	report = publicCompileReport(internalReport)
	if err != nil {
		report.Status = CompileStatusFailed
		return report, err
	}
	if !internalReport.Valid {
		return report, nil
	}
	report.Verification.Reparsed = true
	if err := ctx.Err(); err != nil {
		return report, err
	}
	written, err := io.Copy(w, &compiled)
	report.OutputBytes = int(written)
	if err != nil {
		report.Status = CompileStatusFailed
		return report, err
	}
	report.Status = CompileStatusVerified
	return report, nil
}

func publicCompileReport(source recipe.Report) CompileReport {
	report := CompileReport{SchemaVersion: CompileSchemaVersion, Status: CompileStatusRejected}
	for _, capability := range source.Capabilities {
		report.Capabilities = append(report.Capabilities, CompileCapability{
			Path: capability.Path, Query: capability.Query, Status: string(capability.Status),
			Symbol: capability.Symbol, Domain: capability.Domain, Tier: capability.Tier,
			Verify: capability.Verify, MinVer: capability.MinVer, Boundary: capability.Boundary,
			Gate: append([]string(nil), capability.Gate...),
		})
	}
	for _, downgrade := range source.Downgrades {
		report.Downgraded = append(report.Downgraded, CompileIssue{
			Code: downgrade.Code, Path: downgrade.Path, Query: downgrade.Query, Message: downgrade.Message,
		})
	}
	for _, refusal := range source.Refusals {
		report.Rejected = append(report.Rejected, CompileIssue{
			Code: refusal.Code, Path: refusal.Path, Message: refusal.Message,
		})
	}
	return report
}
