package aepmigrate

import (
	"strings"
	"testing"
)

func TestBuildMatrixLedgerGroupsRecipesAndInfersDomains(t *testing.T) {
	report := MatrixReport{
		OutRoot: "tmp/matrix",
		Summary: MatrixSummary{
			Total:   5,
			Passed:  2,
			Blocked: 1,
			Failed:  1,
			Skipped: 1,
		},
		Cases: []MatrixCase{
			{
				RecipeName:    "minimal-text-animator-skew",
				SourceVersion: "AE2020",
				TargetVersion: "AE2020",
				AEOpenVersion: "AE2020",
				Status:        MatrixStatusPass,
			},
			{
				RecipeName:    "minimal-text-animator-skew",
				SourceVersion: "AE2020",
				TargetVersion: "AE2020",
				AEOpenVersion: "AE2021",
				Status:        MatrixStatusPass,
			},
			{
				RecipeName:    "minimal-layer-explicit-matte",
				SourceVersion: "AE2025",
				TargetVersion: "AE2020",
				Status:        MatrixStatusBlocked,
				Reason:        "convert_blocked",
			},
			{
				RecipeName:    "minimal-shape-gradient-stroke",
				SourceVersion: "AE2020",
				TargetVersion: "AE2025",
				Status:        MatrixStatusFailed,
				Reason:        "open_failed",
			},
			{
				RecipeName:    "minimal-custom-fixture",
				SourceVersion: "AE2020",
				TargetVersion: "AE2021",
				Status:        MatrixStatusSkipped,
				Reason:        "unsupported_writer_target",
			},
		},
	}

	ledger := BuildMatrixLedger(report)

	if ledger.Summary != report.Summary {
		t.Fatalf("summary = %+v, want %+v", ledger.Summary, report.Summary)
	}
	if len(ledger.Rows) != 4 {
		t.Fatalf("rows = %d, want 4: %+v", len(ledger.Rows), ledger.Rows)
	}
	assertLedgerRow(t, ledger.Rows[0], MatrixLedgerRow{
		Domain:        "layer",
		Recipe:        "minimal-layer-explicit-matte",
		Status:        "blocked",
		Blocked:       1,
		SourceLabels:  []string{"AE2025"},
		TargetLabels:  []string{"AE2020"},
		AEOpenLabels:  nil,
		BlockReasons:  []string{"convert_blocked"},
		FailureReason: nil,
	})
	assertLedgerRow(t, ledger.Rows[1], MatrixLedgerRow{
		Domain:       "other",
		Recipe:       "minimal-custom-fixture",
		Status:       "skipped",
		Skipped:      1,
		SourceLabels: []string{"AE2020"},
		TargetLabels: []string{"AE2021"},
		SkipReasons:  []string{"unsupported_writer_target"},
	})
	assertLedgerRow(t, ledger.Rows[2], MatrixLedgerRow{
		Domain:        "shape",
		Recipe:        "minimal-shape-gradient-stroke",
		Status:        "failed",
		Failed:        1,
		SourceLabels:  []string{"AE2020"},
		TargetLabels:  []string{"AE2025"},
		FailureReason: []string{"open_failed"},
		AEOpenLabels:  nil,
		BlockReasons:  nil,
		SkipReasons:   nil,
	})
	assertLedgerRow(t, ledger.Rows[3], MatrixLedgerRow{
		Domain:       "text",
		Recipe:       "minimal-text-animator-skew",
		Status:       "pass",
		Passed:       2,
		SourceLabels: []string{"AE2020"},
		TargetLabels: []string{"AE2020"},
		AEOpenLabels: []string{"AE2020", "AE2021"},
	})
}

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

func assertLedgerRow(t *testing.T, got MatrixLedgerRow, want MatrixLedgerRow) {
	t.Helper()
	if got.Domain != want.Domain ||
		got.Recipe != want.Recipe ||
		got.Status != want.Status ||
		got.Passed != want.Passed ||
		got.Blocked != want.Blocked ||
		got.Failed != want.Failed ||
		got.Skipped != want.Skipped ||
		strings.Join(got.SourceLabels, ",") != strings.Join(want.SourceLabels, ",") ||
		strings.Join(got.TargetLabels, ",") != strings.Join(want.TargetLabels, ",") ||
		strings.Join(got.AEOpenLabels, ",") != strings.Join(want.AEOpenLabels, ",") ||
		strings.Join(got.BlockReasons, ",") != strings.Join(want.BlockReasons, ",") ||
		strings.Join(got.FailureReason, ",") != strings.Join(want.FailureReason, ",") ||
		strings.Join(got.SkipReasons, ",") != strings.Join(want.SkipReasons, ",") {
		t.Fatalf("row = %+v, want %+v", got, want)
	}
}
