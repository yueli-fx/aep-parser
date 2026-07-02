package aepmigrate

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
)

type MatrixLedger struct {
	Source  string            `json:"source,omitempty"`
	Summary MatrixSummary     `json:"summary"`
	Rows    []MatrixLedgerRow `json:"rows"`
}

type MatrixLedgerRow struct {
	Domain        string   `json:"domain"`
	Recipe        string   `json:"recipe"`
	Status        string   `json:"status"`
	Passed        int      `json:"passed"`
	Blocked       int      `json:"blocked"`
	Failed        int      `json:"failed"`
	Skipped       int      `json:"skipped"`
	SourceLabels  []string `json:"source_labels,omitempty"`
	TargetLabels  []string `json:"target_labels,omitempty"`
	AEOpenLabels  []string `json:"ae_open_labels,omitempty"`
	BlockReasons  []string `json:"block_reasons,omitempty"`
	FailureReason []string `json:"failure_reasons,omitempty"`
	SkipReasons   []string `json:"skip_reasons,omitempty"`
}

func ReadMatrixLedger(path string) (MatrixLedger, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return MatrixLedger{}, err
	}
	var report MatrixReport
	if err := json.Unmarshal(data, &report); err != nil {
		return MatrixLedger{}, err
	}
	ledger := BuildMatrixLedger(report)
	ledger.Source = path
	return ledger, nil
}

func BuildMatrixLedger(report MatrixReport) MatrixLedger {
	rowsByRecipe := map[string]*MatrixLedgerRow{}
	for _, c := range report.Cases {
		recipeName := c.RecipeName
		if recipeName == "" {
			recipeName = strings.TrimSpace(c.RecipePath)
		}
		row := rowsByRecipe[recipeName]
		if row == nil {
			row = &MatrixLedgerRow{
				Domain: inferMatrixLedgerDomain(recipeName),
				Recipe: recipeName,
			}
			rowsByRecipe[recipeName] = row
		}
		switch c.Status {
		case MatrixStatusPass:
			row.Passed++
		case MatrixStatusBlocked:
			row.Blocked++
			addUnique(&row.BlockReasons, c.Reason)
		case MatrixStatusFailed:
			row.Failed++
			addUnique(&row.FailureReason, c.Reason)
		case MatrixStatusSkipped:
			row.Skipped++
			addUnique(&row.SkipReasons, c.Reason)
		}
		addUnique(&row.SourceLabels, c.SourceVersion)
		addUnique(&row.TargetLabels, c.TargetVersion)
		addUnique(&row.AEOpenLabels, c.AEOpenVersion)
	}

	rows := make([]MatrixLedgerRow, 0, len(rowsByRecipe))
	for _, row := range rowsByRecipe {
		sort.Strings(row.SourceLabels)
		sort.Strings(row.TargetLabels)
		sort.Strings(row.AEOpenLabels)
		sort.Strings(row.BlockReasons)
		sort.Strings(row.FailureReason)
		sort.Strings(row.SkipReasons)
		row.Status = matrixLedgerStatus(*row)
		rows = append(rows, *row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Domain != rows[j].Domain {
			return rows[i].Domain < rows[j].Domain
		}
		return rows[i].Recipe < rows[j].Recipe
	})

	return MatrixLedger{
		Summary: report.Summary,
		Rows:    rows,
	}
}

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

func addUnique(values *[]string, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	for _, existing := range *values {
		if existing == value {
			return
		}
	}
	*values = append(*values, value)
}
