package aepmigrate

import (
	"strings"
	"testing"
)

func TestRenderMatrixLedgerMarkdownIncludesStableTable(t *testing.T) {
	ledger := MatrixLedger{
		Source: "tmp/matrix/matrix.json",
		Summary: MatrixSummary{
			Total:   2,
			Passed:  1,
			Blocked: 1,
		},
		Rows: []MatrixLedgerRow{
			{
				Domain:       "layer",
				Recipe:       "minimal-layer-explicit-matte",
				Status:       "blocked",
				Blocked:      1,
				SourceLabels: []string{"AE2025"},
				TargetLabels: []string{"AE2020"},
			},
			{
				Domain:       "text",
				Recipe:       "minimal-text-animator-skew",
				Status:       "pass",
				Passed:       1,
				SourceLabels: []string{"AE2020"},
				TargetLabels: []string{"AE2020"},
				AEOpenLabels: []string{"AE2020", "AE2021"},
			},
		},
	}

	md := RenderMatrixLedgerMarkdown(ledger)

	for _, want := range []string{
		"# Matrix Coverage Ledger",
		"Source: `tmp/matrix/matrix.json`",
		"Summary: 2 total, 1 pass, 1 blocked, 0 failed, 0 skipped",
		"| Domain | Recipe | Status | Pass | Blocked | Failed | Skipped | Sources | Targets | AE Hosts |",
		"| layer | `minimal-layer-explicit-matte` | blocked | 0 | 1 | 0 | 0 | AE2025 | AE2020 | - |",
		"| text | `minimal-text-animator-skew` | pass | 1 | 0 | 0 | 0 | AE2020 | AE2020 | AE2020,AE2021 |",
	} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown missing %q:\n%s", want, md)
		}
	}
}
