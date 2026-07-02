package aepmigrate

import (
	"sort"
	"strings"
)

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
