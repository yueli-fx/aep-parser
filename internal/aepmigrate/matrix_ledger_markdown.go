package aepmigrate

import (
	"fmt"
	"strings"
)

func RenderMatrixLedgerMarkdown(ledger MatrixLedger) string {
	var b strings.Builder
	b.WriteString("# Matrix Coverage Ledger\n\n")
	if ledger.Source != "" {
		fmt.Fprintf(&b, "Source: `%s`\n\n", ledger.Source)
	}
	fmt.Fprintf(&b, "Summary: %d total, %d pass, %d blocked, %d failed, %d skipped\n\n",
		ledger.Summary.Total,
		ledger.Summary.Passed,
		ledger.Summary.Blocked,
		ledger.Summary.Failed,
		ledger.Summary.Skipped,
	)
	b.WriteString("| Domain | Recipe | Status | Pass | Blocked | Failed | Skipped | Sources | Targets | AE Hosts |\n")
	b.WriteString("| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- |\n")
	for _, row := range ledger.Rows {
		fmt.Fprintf(&b, "| %s | `%s` | %s | %d | %d | %d | %d | %s | %s | %s |\n",
			row.Domain,
			row.Recipe,
			row.Status,
			row.Passed,
			row.Blocked,
			row.Failed,
			row.Skipped,
			matrixLedgerList(row.SourceLabels),
			matrixLedgerList(row.TargetLabels),
			matrixLedgerList(row.AEOpenLabels),
		)
	}
	return b.String()
}

func matrixLedgerList(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ",")
}
