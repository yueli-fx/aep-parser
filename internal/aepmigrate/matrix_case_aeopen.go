package aepmigrate

import "path/filepath"

func matrixAEOpenOptions(opts MatrixOptions, aePath, caseDir string) (*AEOpenOptions, error) {
	jsPath := opts.AEOpenJSXPath
	if jsPath == "" {
		jsPath = filepath.Join("test_data", "generators", "verify_open.jsx")
	}
	jsPath, err := filepath.Abs(jsPath)
	if err != nil {
		return nil, err
	}
	argsPath, err := filepath.Abs(filepath.Join(caseDir, "ae_open.args.json"))
	if err != nil {
		return nil, err
	}
	donePath, err := filepath.Abs(filepath.Join(caseDir, "ae_open.done"))
	if err != nil {
		return nil, err
	}
	return &AEOpenOptions{
		Host:       opts.Host,
		AEPath:     aePath,
		JSXPath:    jsPath,
		ArgsPath:   argsPath,
		DonePath:   donePath,
		TimeoutSec: opts.AEOpenTimeoutSec,
	}, nil
}
