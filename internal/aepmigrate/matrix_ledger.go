package aepmigrate

import (
	"encoding/json"
	"os"
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
