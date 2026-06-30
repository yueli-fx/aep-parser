package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

type Options struct {
	InputPath string
	Target    VersionLabel
}

func Assess(opts Options) (Report, error) {
	project, err := aep.Open(opts.InputPath)
	if err != nil {
		return Report{}, err
	}
	prof, err := profile.Build(project, profile.Options{Path: opts.InputPath})
	if err != nil {
		return Report{}, err
	}
	sourceVersion := NormalizeSourceVersion((&aep.Application{Project: project}).Version())
	entries := classifyProfile(sourceVersion.Label, opts.Target, prof)
	report := Report{
		SchemaVersion: SchemaVersion,
		Source: SourceInfo{
			Path:       opts.InputPath,
			Version:    sourceVersion.Label,
			VersionRaw: sourceVersion.Raw,
		},
		Target:  TargetInfo{Version: opts.Target},
		Summary: summarize(entries),
		Entries: entries,
		Verification: Verification{
			ProfileDiffStatus: "not_run",
			AEOpenStatus:      "not_run",
			RenderStatus:      "not_run",
		},
	}
	return report, nil
}

func summarize(entries []Entry) Summary {
	summary := Summary{Status: StatusPass}
	for _, entry := range entries {
		switch entry.Class {
		case ClassPreserved:
			summary.Preserved++
		case ClassRetargeted:
			summary.Retargeted++
		case ClassTranslated:
			summary.Translated++
		case ClassApproximated:
			summary.Approximated++
			if summary.Status == StatusPass {
				summary.Status = StatusWarn
			}
		case ClassDropped:
			summary.Dropped++
			if summary.Status == StatusPass {
				summary.Status = StatusWarn
			}
		case ClassBlocked:
			summary.Blocked++
			summary.Status = StatusBlocked
		default:
			summary.Unknown++
			if summary.Status != StatusBlocked {
				summary.Status = StatusWarn
			}
		}
	}
	return summary
}
