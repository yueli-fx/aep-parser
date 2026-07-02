package aepmigrate

import "strings"

func inferMatrixLedgerDomain(recipeName string) string {
	switch {
	case strings.HasPrefix(recipeName, "minimal-text-"):
		return "text"
	case strings.HasPrefix(recipeName, "minimal-layer-"):
		return "layer"
	case strings.HasPrefix(recipeName, "minimal-shape-"):
		return "shape"
	case strings.HasPrefix(recipeName, "minimal-effect-"), strings.HasPrefix(recipeName, "minimal-adjustment-"):
		return "effect"
	case strings.HasPrefix(recipeName, "minimal-transform-"):
		return "transform"
	case strings.HasPrefix(recipeName, "minimal-comp-"):
		return "comp"
	case strings.HasPrefix(recipeName, "minimal-camera-"), strings.HasPrefix(recipeName, "minimal-light-"):
		return "camera-light"
	case strings.HasPrefix(recipeName, "minimal-precomp-"):
		return "precomp"
	case strings.HasPrefix(recipeName, "minimal-project-"):
		return "project"
	default:
		return "other"
	}
}

func matrixLedgerStatus(row MatrixLedgerRow) string {
	total := row.Passed + row.Blocked + row.Failed + row.Skipped
	switch {
	case row.Failed > 0:
		return "failed"
	case row.Blocked > 0:
		return "blocked"
	case row.Skipped > 0 && row.Passed == 0 && row.Skipped == total:
		return "skipped"
	case row.Skipped > 0:
		return "mixed"
	case row.Passed == total:
		return "pass"
	default:
		return "mixed"
	}
}
