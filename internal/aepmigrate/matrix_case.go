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
		aeOpen, err := matrixAEOpenOptions(opts, c.AEOpenPath, caseDir)
		if err != nil {
			c.Status = MatrixStatusFailed
			c.Reason = err.Error()
			return c
		}
		convertOpts.AEOpen = aeOpen
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
	c.Status, c.Reason = matrixConvertStatus(convertReport)
	return c
}
