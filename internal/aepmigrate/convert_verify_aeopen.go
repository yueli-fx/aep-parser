package aepmigrate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aehost"
)

func verifyAEOpen(report *Report, opts ConvertOptions) error {
	if opts.AEOpen == nil {
		return nil
	}
	openOpts := *opts.AEOpen
	if openOpts.Host == nil {
		report.Verification.AEOpenStatus = "unavailable"
		report.Entries = append(report.Entries, Entry{
			Path:          "verification.ae_open",
			Class:         ClassBlocked,
			TargetVersion: report.Target.Version,
			Reason:        "AE open verification requested but no AE host is configured.",
		})
		report.Summary = summarize(report.Entries)
		return nil
	}
	if openOpts.TimeoutSec == 0 {
		openOpts.TimeoutSec = 180
	}
	argsAbs, err := filepath.Abs(openOpts.ArgsPath)
	if err != nil {
		return err
	}
	openOpts.ArgsPath = argsAbs
	if err := writeAEOpenArgs(openOpts.ArgsPath, report.Target.Path, openOpts.DonePath); err != nil {
		return err
	}
	_ = os.Remove(openOpts.DonePath)
	result, err := openOpts.Host.RunScript(context.Background(), aehost.ScriptRequest{
		AEPath:     openOpts.AEPath,
		JSXPath:    openOpts.JSXPath,
		DonePath:   openOpts.DonePath,
		TimeoutSec: openOpts.TimeoutSec,
		Env: map[string]string{
			"AE_OPEN_ARGS": filepath.ToSlash(openOpts.ArgsPath),
		},
	})
	if err != nil {
		return fmt.Errorf("ae open gate: %w", err)
	}
	report.Verification.AEOpenExitCode = &result.ExitCode
	done, readErr := os.ReadFile(openOpts.DonePath)
	if readErr != nil {
		report.Verification.AEOpenStatus = "fail"
		report.Verification.AEOpenLog = readErr.Error()
	} else {
		report.Verification.AEOpenLog = string(done)
		if result.ExitCode == 0 && strings.HasPrefix(string(done), "PASS") {
			report.Verification.AEOpenStatus = "pass"
			return nil
		}
		report.Verification.AEOpenStatus = "fail"
	}
	report.Entries = append(report.Entries, Entry{
		Path:          "verification.ae_open",
		Class:         ClassBlocked,
		TargetVersion: report.Target.Version,
		Reason:        "AE open verification failed.",
	})
	report.Summary = summarize(report.Entries)
	return nil
}

func writeAEOpenArgs(path, inputPath, donePath string) error {
	if path == "" {
		return fmt.Errorf("ae open gate: args path is required")
	}
	if donePath == "" {
		return fmt.Errorf("ae open gate: done path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	inputAbs, err := filepath.Abs(inputPath)
	if err != nil {
		return err
	}
	doneAbs, err := filepath.Abs(donePath)
	if err != nil {
		return err
	}
	payload := struct {
		Input string `json:"input"`
		Done  string `json:"done"`
	}{
		Input: filepath.ToSlash(inputAbs),
		Done:  filepath.ToSlash(doneAbs),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
