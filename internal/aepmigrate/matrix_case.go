package aepmigrate

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/recipe"
)

func runMatrixCase(opts MatrixOptions, hosts map[string]string, recipePath, sourceLabel, targetLabel, aeOpenLabel string, explicitAEOpenLabel bool) MatrixCase {
	recipeName := strings.TrimSuffix(filepath.Base(recipePath), filepath.Ext(recipePath))
	caseName := sourceLabel + "_to_" + targetLabel
	aeOpenVersion := ""
	if opts.AEOpen {
		if explicitAEOpenLabel {
			aeOpenVersion = normalizeAEHostLabel(aeOpenLabel)
			caseName += "_open_" + aeOpenVersion
		} else {
			aeOpenVersion = normalizeAEHostLabel(targetLabel)
		}
	}
	caseDir := filepath.Join(opts.OutRoot, sanitizeMatrixName(recipeName), sanitizeMatrixName(caseName))
	c := MatrixCase{
		RecipePath:    recipePath,
		RecipeName:    recipeName,
		SourceVersion: sourceLabel,
		TargetVersion: targetLabel,
		AEOpenVersion: aeOpenVersion,
		CaseDir:       caseDir,
	}
	sourceWriterLabel, sourceWriterOK := matrixSourceWriterLabel(sourceLabel)
	if !sourceWriterOK {
		c.Status = MatrixStatusSkipped
		c.Reason = "unsupported_source_writer"
		return c
	}
	target, targetWriterOK := matrixWriterTarget(targetLabel)
	if !targetWriterOK {
		c.Status = MatrixStatusSkipped
		c.Reason = "unsupported_writer_target"
		return c
	}
	if opts.AEOpen {
		aePath := hosts[aeOpenVersion]
		if aePath == "" {
			c.Status = MatrixStatusSkipped
			c.Reason = "missing_ae_host"
			return c
		}
		c.AEOpenPath = aePath
		if opts.Host == nil {
			c.Status = MatrixStatusSkipped
			c.Reason = "missing_ae_host_runner"
			return c
		}
	}
	if err := os.MkdirAll(caseDir, 0o755); err != nil {
		c.Status = MatrixStatusFailed
		c.Reason = err.Error()
		return c
	}
	sourcePath := filepath.Join(caseDir, "source.aep")
	compileReportPath := filepath.Join(caseDir, "compile_report.json")
	c.SourcePath = sourcePath
	c.CompileReportPath = compileReportPath
	rec, err := readMatrixRecipe(recipePath)
	if err != nil {
		c.Status = MatrixStatusFailed
		c.Reason = err.Error()
		return c
	}
	if sourceWriterLabel != "" {
		rec.Project.TargetVersion = sourceWriterLabel
	}
	compileReport, err := recipe.CompileToFile(rec, sourcePath, opts.Capabilities)
	if writeErr := writeMatrixJSON(compileReportPath, compileReport); writeErr != nil && err == nil {
		err = writeErr
	}
	if err != nil {
		c.Status = MatrixStatusFailed
		c.Reason = err.Error()
		return c
	}
	if !compileReport.Valid {
		c.Status, c.Reason = matrixCompileInvalidReason(compileReport)
		return c
	}
	outputPath := filepath.Join(caseDir, "target.aep")
	convertReportPath := filepath.Join(caseDir, "convert_report.json")
	c.OutputPath = outputPath
	c.ConvertReportPath = convertReportPath
	convertOpts := ConvertOptions{
		InputPath:  sourcePath,
		OutputPath: outputPath,
		Target:     target,
	}
	if opts.AEOpen {
		jsPath := opts.AEOpenJSXPath
		if jsPath == "" {
			jsPath = filepath.Join("test_data", "generators", "verify_open.jsx")
		}
		jsPath, err = filepath.Abs(jsPath)
		if err != nil {
			c.Status = MatrixStatusFailed
			c.Reason = err.Error()
			return c
		}
		argsPath, err := filepath.Abs(filepath.Join(caseDir, "ae_open.args.json"))
		if err != nil {
			c.Status = MatrixStatusFailed
			c.Reason = err.Error()
			return c
		}
		donePath, err := filepath.Abs(filepath.Join(caseDir, "ae_open.done"))
		if err != nil {
			c.Status = MatrixStatusFailed
			c.Reason = err.Error()
			return c
		}
		convertOpts.AEOpen = &AEOpenOptions{
			Host:       opts.Host,
			AEPath:     c.AEOpenPath,
			JSXPath:    jsPath,
			ArgsPath:   argsPath,
			DonePath:   donePath,
			TimeoutSec: opts.AEOpenTimeoutSec,
		}
	}
	convertReport, err := Convert(convertOpts)
	if writeErr := writeMatrixJSON(convertReportPath, convertReport); writeErr != nil && err == nil {
		err = writeErr
	}
	if err != nil {
		c.Status = MatrixStatusFailed
		c.Reason = err.Error()
		return c
	}
	switch convertReport.Summary.Status {
	case StatusPass:
		c.Status = MatrixStatusPass
	case StatusBlocked:
		c.Status = MatrixStatusBlocked
		c.Reason = "convert_blocked"
	case StatusError:
		c.Status = MatrixStatusFailed
		c.Reason = "convert_error"
	default:
		c.Status = MatrixStatusFailed
		c.Reason = "convert_status_" + string(convertReport.Summary.Status)
	}
	return c
}

func matrixCompileInvalidReason(report recipe.Report) (MatrixStatus, string) {
	if len(report.Refusals) == 1 && report.Refusals[0].Code == "explicit_matte_requires_ae2025" {
		return MatrixStatusSkipped, "source_contract_unsupported"
	}
	return MatrixStatusBlocked, "recipe_compile_blocked"
}
