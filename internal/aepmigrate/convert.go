package aepmigrate

import (
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

type ConvertOptions struct {
	InputPath  string
	OutputPath string
	Target     VersionLabel
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
	targetProject, err := rebuildNoLayerProject(opts.Target, prof)
	if err != nil {
		return Report{}, err
	}
	out, err := os.Create(opts.OutputPath)
	if err != nil {
		return Report{}, err
	}
	defer out.Close()
	if err := targetProject.WriteAEP(out); err != nil {
		return Report{}, err
	}
	report.Target.Path = opts.OutputPath
	return report, nil
}

func convertAssessEntries(entries []Entry) []Entry {
	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if entry.Path == "project" && entry.Class == ClassPreserved && entry.CapabilityKey == "" {
			continue
		}
		out = append(out, entry)
	}
	return out
}

func convertScopeEntries(target VersionLabel, prof *profile.Profile) []Entry {
	if prof == nil {
		return nil
	}
	var entries []Entry
	for _, comp := range prof.Comps {
		if len(comp.Layers) == 0 {
			entries = append(entries, Entry{
				Path:          "comps[" + comp.Name + "]",
				Class:         ClassRetargeted,
				TargetVersion: target,
				Reason:        "Composition skeleton is recreated through the target AE project template.",
			})
			continue
		}
		for _, layer := range comp.Layers {
			entries = append(entries, Entry{
				Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
				Class:         ClassBlocked,
				TargetVersion: target,
				Reason:        "The first convert slice does not reconstruct layers; refusing output to avoid silent layer loss.",
			})
		}
	}
	return entries
}

func rebuildNoLayerProject(target VersionLabel, prof *profile.Profile) (*aep.Project, error) {
	project := aep.NewProject(aepTarget(target))
	for _, comp := range prof.Comps {
		if _, err := aep.NewComposition(project, comp.Name, comp.Width, comp.Height, comp.FrameRate, comp.Duration); err != nil {
			return nil, err
		}
	}
	return project, nil
}

func aepTarget(target VersionLabel) aep.AETarget {
	switch target {
	case VersionAE2022:
		return aep.TargetAE2022
	case VersionAE2025:
		return aep.TargetAE2025
	default:
		return aep.TargetAE2020
	}
}
