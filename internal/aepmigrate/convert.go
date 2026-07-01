package aepmigrate

import (
	"bytes"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aehost"
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

type ConvertOptions struct {
	InputPath  string
	OutputPath string
	Target     VersionLabel
	AEOpen     *AEOpenOptions
}

type AEOpenOptions struct {
	Host       aehost.Host
	AEPath     string
	JSXPath    string
	ArgsPath   string
	DonePath   string
	TimeoutSec int
}

func Convert(opts ConvertOptions) (Report, error) {
	project, err := aep.Open(opts.InputPath)
	if err != nil {
		return Report{}, err
	}
	prof, err := profile.Build(project, profile.Options{Path: opts.InputPath})
	if err != nil {
		return Report{}, err
	}
	sourceVersion := NormalizeSourceVersion((&aep.Application{Project: project}).Version())
	entries := convertAssessEntries(classifyProfile(sourceVersion.Label, opts.Target, prof))
	entries = append(entries, Entry{
		Path:          "project",
		Class:         ClassRetargeted,
		TargetVersion: opts.Target,
		Reason:        "Project skeleton is recreated through the requested target AE template.",
	})
	entries = append(entries, convertScopeEntries(opts.Target, prof)...)
	report := Report{
		SchemaVersion: SchemaVersion,
		Source: SourceInfo{
			Path:       opts.InputPath,
			Version:    sourceVersion.Label,
			VersionRaw: sourceVersion.Raw,
		},
		Target:  TargetInfo{Version: opts.Target},
		Entries: entries,
		Verification: Verification{
			ProfileDiffStatus: "not_run",
			AEOpenStatus:      "not_run",
			RenderStatus:      "not_run",
		},
	}
	report.Summary = summarize(entries)
	if report.Summary.Status == StatusBlocked {
		return report, nil
	}
	targetProject, err := rebuildProject(opts.Target, prof)
	if err != nil {
		return Report{}, err
	}
	targetProject, err = applyNoLayerCompMetadata(targetProject, prof)
	if err != nil {
		return Report{}, err
	}
	targetProject, err = materializeProjectTransformExpressions(targetProject, prof)
	if err != nil {
		return Report{}, err
	}
	targetProject, err = materializeProjectTextAnimators(targetProject, prof)
	if err != nil {
		return Report{}, err
	}
	targetProject, err = materializeProjectEffects(targetProject, prof)
	if err != nil {
		return Report{}, err
	}
	var out bytes.Buffer
	if err := targetProject.WriteAEP(&out); err != nil {
		return Report{}, err
	}
	report.Target.Path = opts.OutputPath
	if err := verifyConvertedProfileBytes(&report, prof, opts.OutputPath, out.Bytes()); err != nil {
		return Report{}, err
	}
	if report.Summary.Status == StatusBlocked {
		return report, nil
	}
	if err := os.WriteFile(opts.OutputPath, out.Bytes(), 0o644); err != nil {
		return Report{}, err
	}
	if err := verifyAEOpen(&report, opts); err != nil {
		return Report{}, err
	}
	return report, nil
}
