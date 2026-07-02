package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/recipe"

func matrixCompileInvalidReason(report recipe.Report) (MatrixStatus, string) {
	if len(report.Refusals) == 1 && report.Refusals[0].Code == "explicit_matte_requires_ae2025" {
		return MatrixStatusSkipped, "source_contract_unsupported"
	}
	return MatrixStatusBlocked, "recipe_compile_blocked"
}

func matrixConvertStatus(report Report) (MatrixStatus, string) {
	switch report.Summary.Status {
	case StatusPass:
		return MatrixStatusPass, ""
	case StatusBlocked:
		return MatrixStatusBlocked, "convert_blocked"
	case StatusError:
		return MatrixStatusFailed, "convert_error"
	default:
		return MatrixStatusFailed, "convert_status_" + string(report.Summary.Status)
	}
}
