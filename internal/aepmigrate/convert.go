package aepmigrate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aehost"
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/profilediff"
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

func verifyConvertedProfileBytes(report *Report, source *profile.Profile, targetPath string, targetBytes []byte) error {
	targetProject, err := aep.FromReader(bytes.NewReader(targetBytes))
	if err != nil {
		return fmt.Errorf("profile diff target open: %w", err)
	}
	target, err := profile.Build(targetProject, profile.Options{Path: targetPath})
	if err != nil {
		return fmt.Errorf("profile diff target profile: %w", err)
	}
	diffReport, err := profilediff.Compare(source, target, profilediff.Options{})
	if err != nil {
		return fmt.Errorf("profile diff compare: %w", err)
	}
	report.Verification.ProfileDiffCount = diffReport.DiffCount
	report.Verification.ProfileDiffIgnoredCount = diffReport.IgnoredCount
	report.Verification.ProfileDiffs = verificationDiffs(diffReport.Diffs)
	if diffReport.DiffCount == 0 {
		report.Verification.ProfileDiffStatus = "pass"
		return nil
	}
	report.Verification.ProfileDiffStatus = "fail"
	report.Entries = append(report.Entries, Entry{
		Path:          "verification.profile_diff",
		Class:         ClassBlocked,
		TargetVersion: report.Target.Version,
		Reason:        fmt.Sprintf("Profile diff verification found %d unexpected migrated output difference(s).", diffReport.DiffCount),
	})
	report.Summary = summarize(report.Entries)
	return nil
}

func verificationDiffs(diffs []profilediff.Diff) []VerificationDiff {
	out := make([]VerificationDiff, 0, len(diffs))
	for _, diff := range diffs {
		out = append(out, VerificationDiff{
			Path:       diff.Path,
			Kind:       string(diff.Kind),
			Severity:   string(diff.Severity),
			ActionType: string(diff.ActionType),
			Expected:   diff.Expected,
			Actual:     diff.Actual,
		})
	}
	return out
}

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
	footage := newConvertFootageIndex(prof)
	comps := newConvertCompIndex(prof)
	for _, comp := range prof.Comps {
		entries = append(entries, Entry{
			Path:          "comps[" + comp.Name + "]",
			Class:         ClassRetargeted,
			TargetVersion: target,
			Reason:        "Composition and stable composition settings are recreated through the target AE project template.",
		})
		for _, layer := range comp.Layers {
			if isSupportedDefaultNullLayer(layer, footage) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default null layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedDefaultSolidLayer(layer, footage) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default solid layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedDefaultAdjustmentLayer(layer, footage) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default adjustment layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedDefaultCameraLayer(layer) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default camera layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedDefaultLightLayer(layer) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default light layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedDefaultTextLayer(layer) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default text layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedDefaultShapeLayer(layer) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default empty shape layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedRectGraphicShapeLayer(layer) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Single parametric graphic shape layer is recreated from the stable profile shape properties.",
				})
				continue
			}
			if isSupportedDefaultPrecompLayer(layer, comps) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default precomp layer is recreated through the target AE project template.",
				})
				continue
			}
			entries = append(entries, Entry{
				Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
				Class:         ClassBlocked,
				TargetVersion: target,
				Reason:        "This convert slice only reconstructs no-layer comps, default null layers, default solid layers, default adjustment layers, default camera layers, default light layers, default text layers, default empty shape layers, single parametric graphic/filter shape layers, and default precomp layers; refusing output to avoid silent layer loss.",
			})
		}
	}
	return entries
}

func rebuildProject(target VersionLabel, prof *profile.Profile) (*aep.Project, error) {
	project := aep.NewProject(aepTarget(target))
	footage := newConvertFootageIndex(prof)
	comps := newConvertCompIndex(prof)
	targetComps := newConvertTargetCompIndex()
	for _, comp := range prof.Comps {
		next, err := aep.NewComposition(project, comp.Name, comp.Width, comp.Height, comp.FrameRate, comp.Duration)
		if err != nil {
			return nil, err
		}
		if err := applyStableCompSettings(next, comp); err != nil {
			return nil, err
		}
		targetComps.add(comp, next)
	}
	for _, comp := range prof.Comps {
		next, ok := targetComps.bySourceID[comp.ID]
		if !ok {
			next = targetComps.byName[comp.Name]
		}
		if next == nil {
			return nil, fmt.Errorf("comp %q: target comp missing after creation", comp.Name)
		}
		var createdLayers []convertLayerPair
		recordLayer := func(source profile.Layer, target *aep.Layer) {
			createdLayers = append(createdLayers, convertLayerPair{source: source, target: target})
		}
		for _, layer := range comp.Layers {
			switch {
			case isSupportedDefaultNullLayer(layer, footage):
				dstLayer, err := materializeNullLayer(next, layer, footage)
				if err != nil {
					return nil, fmt.Errorf("comp %q null layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q null layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q null layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q null layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q null layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultSolidLayer(layer, footage):
				solid, _ := footage.solidDetails(layer)
				dstLayer, err := aep.NewSolidLayer(next, layer.Name, int(solid.Width), int(solid.Height), *solid.SolidColor)
				if err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if err := materializeCenteredTransformSurface(next, dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultAdjustmentLayer(layer, footage):
				dstLayer, err := aep.NewAdjustmentLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultCameraLayer(layer):
				dstLayer, err := aep.NewCameraLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if err := materializeCameraLightTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeCameraOptions(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q options: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultLightLayer(layer):
				dstLayer, err := aep.NewLightLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q light layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q light layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q light layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if err := materializeCameraLightTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q light layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLightOptions(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q light layer %q options: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q light layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultTextLayer(layer):
				dstLayer, err := aep.NewTextLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q text layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q text layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q text layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if layer.Text != nil {
					if err := dstLayer.SetText(layer.Text.Text); err != nil {
						return nil, fmt.Errorf("comp %q text layer %q text: %w", comp.Name, layer.Name, err)
					}
					if err := materializeTextStyle(dstLayer, layer.Text); err != nil {
						return nil, fmt.Errorf("comp %q text layer %q text style: %w", comp.Name, layer.Name, err)
					}
				}
				if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q text layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q text layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultShapeLayer(layer):
				dstLayer, err := aep.NewShapeLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q shape layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer.Layer, layer); err != nil {
					return nil, fmt.Errorf("comp %q shape layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer.Layer)
			case isSupportedRectGraphicShapeLayer(layer):
				dstLayer, err := materializeRectGraphicShapeLayer(next, layer)
				if err != nil {
					return nil, fmt.Errorf("comp %q shape layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q shape layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			case isSupportedDefaultPrecompLayer(layer, comps):
				sourceComp, ok := targetComps.sourceComposition(layer)
				if !ok {
					return nil, fmt.Errorf("comp %q precomp layer %q source %q not found", comp.Name, layer.Name, layer.SourceRef.Name)
				}
				dstLayer, err := aep.NewPrecompLayer(next, sourceComp, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerMetadata(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q metadata: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerSwitchSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q switches: %w", comp.Name, layer.Name, err)
				}
				if err := materializePrecompTransformSurface(next, dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q timing: %w", comp.Name, layer.Name, err)
				}
				recordLayer(layer, dstLayer)
			default:
				return nil, fmt.Errorf("unsupported layer %q in comp %q", layer.Name, comp.Name)
			}
		}
		if err := materializeLayerRefs(createdLayers); err != nil {
			return nil, fmt.Errorf("comp %q layer refs: %w", comp.Name, err)
		}
	}
	return project, nil
}

type convertLayerPair struct {
	source profile.Layer
	target *aep.Layer
}

func materializeLayerRefs(layers []convertLayerPair) error {
	bySourceID := map[uint32]*aep.Layer{}
	byName := map[string]*aep.Layer{}
	for _, pair := range layers {
		if pair.source.ID != 0 {
			bySourceID[pair.source.ID] = pair.target
		}
		if _, exists := byName[pair.source.Name]; !exists {
			byName[pair.source.Name] = pair.target
		}
	}
	for _, pair := range layers {
		if pair.source.ParentRef == nil {
			continue
		}
		parent := bySourceID[pair.source.ParentRef.ID]
		if parent == nil && pair.source.ParentRef.Name != "" {
			parent = byName[pair.source.ParentRef.Name]
		}
		if parent == nil {
			return fmt.Errorf("layer %q parent %q not found", pair.source.Name, pair.source.ParentRef.Name)
		}
		if err := pair.target.SetParent(parent.ID); err != nil {
			return fmt.Errorf("layer %q parent %q: %w", pair.source.Name, pair.source.ParentRef.Name, err)
		}
	}
	for _, pair := range layers {
		if pair.source.LightSourceRef == nil {
			continue
		}
		source := bySourceID[pair.source.LightSourceRef.ID]
		if source == nil && pair.source.LightSourceRef.Name != "" {
			source = byName[pair.source.LightSourceRef.Name]
		}
		if source == nil {
			return fmt.Errorf("layer %q light source %q not found", pair.source.Name, pair.source.LightSourceRef.Name)
		}
		if err := pair.target.SetLightSource(source); err != nil {
			return fmt.Errorf("layer %q light source %q: %w", pair.source.Name, pair.source.LightSourceRef.Name, err)
		}
	}
	return nil
}

func materializeLayerMetadata(layer *aep.Layer, source profile.Layer) error {
	if source.Label != 0 && source.Label != layer.Label {
		if err := layer.SetLabel(source.Label); err != nil {
			return err
		}
	}
	if source.Comment != "" {
		if err := layer.SetComment(source.Comment); err != nil {
			return err
		}
	}
	return nil
}

func materializeNullLayer(comp *aep.Composition, source profile.Layer, footage convertFootageIndex) (*aep.Layer, error) {
	if solid, ok := footage.solidDetails(source); ok && solid.Width != 0 && solid.Height != 0 && solid.SolidColor != nil {
		layer, err := aep.NewSolidLayer(comp, source.Name, int(solid.Width), int(solid.Height), *solid.SolidColor)
		if err != nil {
			return nil, err
		}
		if err := layer.SetIsNull(true); err != nil {
			return nil, err
		}
		return layer, nil
	}
	return aep.NewNullLayer(comp, source.Name)
}

func materializeCameraOptions(layer *aep.Layer, source profile.Layer) error {
	scalars := []struct {
		matchName string
		set       func(float64) error
	}{
		{aep.MatchNameCameraZoom, layer.SetCameraZoom},
		{aep.MatchNameCameraDepthOfField, func(v float64) error { return layer.SetCameraDepthOfField(v != 0) }},
		{aep.MatchNameCameraFocusDistance, layer.SetCameraFocusDistance},
		{aep.MatchNameCameraAperture, layer.SetCameraAperture},
		{aep.MatchNameCameraBlurLevel, layer.SetCameraBlurLevel},
		{aep.MatchNameCameraIrisShape, layer.SetIrisShape},
		{aep.MatchNameCameraIrisRotation, layer.SetIrisRotation},
		{aep.MatchNameCameraIrisRoundness, layer.SetIrisRoundness},
		{aep.MatchNameCameraIrisAspectRatio, layer.SetIrisAspectRatio},
		{aep.MatchNameCameraIrisDiffractionFringe, layer.SetIrisDiffractionFringe},
		{aep.MatchNameCameraIrisHighlightGain, layer.SetIrisHighlightGain},
		{aep.MatchNameCameraIrisHighlightThreshold, layer.SetIrisHighlightThreshold},
		{aep.MatchNameCameraIrisHighlightSaturation, layer.SetIrisHighlightSaturation},
	}
	for _, scalar := range scalars {
		value, ok := propertyFloat(source.Properties, scalar.matchName)
		if !ok {
			continue
		}
		if err := scalar.set(value); err != nil {
			return fmt.Errorf("%s: %w", scalar.matchName, err)
		}
	}
	return nil
}

func materializeLightOptions(layer *aep.Layer, source profile.Layer) error {
	if source.LightKind != "" {
		kind, ok := convertLightKind(source.LightKind)
		if !ok {
			return fmt.Errorf("light_kind %q unsupported", source.LightKind)
		}
		if err := layer.SetLightKind(kind); err != nil {
			return fmt.Errorf("light_kind: %w", err)
		}
	}
	if value, ok := propertyVector(source.Properties, aep.MatchNameLightColor, 4); ok {
		if err := layer.SetLightColor(value); err != nil {
			return fmt.Errorf("%s: %w", aep.MatchNameLightColor, err)
		}
	}
	scalars := []struct {
		matchName string
		set       func(float64) error
	}{
		{aep.MatchNameLightIntensity, layer.SetLightIntensity},
		{aep.MatchNameLightConeAngle, layer.SetLightConeAngle},
		{aep.MatchNameLightConeFeather, layer.SetLightConeFeather},
		{aep.MatchNameLightFalloffType, layer.SetLightFalloffType},
		{aep.MatchNameLightFalloffStart, layer.SetLightFalloffStart},
		{aep.MatchNameLightFalloffDistance, layer.SetLightFalloffDistance},
		{aep.MatchNameLightCastsShadows, func(v float64) error { return layer.SetLightCastsShadows(v != 0) }},
		{aep.MatchNameLightShadowDarkness, layer.SetLightShadowDarkness},
		{aep.MatchNameLightShadowDiffusion, layer.SetLightShadowDiffusion},
	}
	for _, scalar := range scalars {
		value, ok := propertyFloat(source.Properties, scalar.matchName)
		if !ok {
			continue
		}
		if err := scalar.set(value); err != nil {
			return fmt.Errorf("%s: %w", scalar.matchName, err)
		}
	}
	return nil
}

func convertLightKind(value string) (aep.LightKind, bool) {
	switch strings.ToLower(value) {
	case "parallel":
		return aep.LightKindParallel, true
	case "spot":
		return aep.LightKindSpot, true
	case "point":
		return aep.LightKindPoint, true
	case "ambient":
		return aep.LightKindAmbient, true
	default:
		return 0, false
	}
}

func convertAutoOrient(value string) (aep.AutoOrientType, bool) {
	switch strings.ToLower(value) {
	case "none":
		return aep.AutoOrientNone, true
	case "along_path", "along-path":
		return aep.AutoOrientAlongPath, true
	case "camera_or_point_of_interest", "camera-or-point-of-interest":
		return aep.AutoOrientCameraOrPointOfInterest, true
	case "characters_toward_camera", "characters-toward-camera":
		return aep.AutoOrientCharactersTowardCamera, true
	default:
		return 0, false
	}
}

func materializeLayerSwitchSurface(layer *aep.Layer, source profile.Layer) error {
	flags := source.Flags
	if err := layer.SetVisible(flags.Visible); err != nil {
		return err
	}
	if err := layer.SetSolo(flags.Solo); err != nil {
		return err
	}
	if err := layer.SetShy(flags.Shy); err != nil {
		return err
	}
	if err := layer.SetLocked(flags.Locked); err != nil {
		return err
	}
	if err := layer.SetEffectsEnabled(flags.EffectsEnabled); err != nil {
		return err
	}
	if err := layer.SetAudioEnabled(flags.AudioEnabled); err != nil {
		return err
	}
	if err := layer.SetMotionBlur(flags.MotionBlur); err != nil {
		return err
	}
	if err := layer.SetFrameBlendEnabled(flags.FrameBlendEnabled); err != nil {
		return err
	}
	if err := layer.SetMarkersLocked(flags.MarkersLocked); err != nil {
		return err
	}
	if err := layer.SetCollapseTransform(flags.CollapseTransform); err != nil {
		return err
	}
	if err := layer.SetIs3D(flags.Is3D); err != nil {
		return err
	}
	if err := layer.SetIsAdjust(flags.IsAdjustment); err != nil {
		return err
	}
	if err := layer.SetIsGuide(flags.IsGuide); err != nil {
		return err
	}
	if err := layer.SetSamplingBicubic(flags.SamplingBicubic); err != nil {
		return err
	}
	if err := layer.SetFrameBlendPixelMotion(flags.FrameBlendPixelMotion); err != nil {
		return err
	}
	if err := layer.SetPreserveTransparency(flags.PreserveTransparency); err != nil {
		return err
	}
	if source.AutoOrient != "" {
		autoOrient, ok := convertAutoOrient(source.AutoOrient)
		if !ok {
			return fmt.Errorf("auto_orient %q unsupported", source.AutoOrient)
		}
		if err := layer.SetAutoOrient(autoOrient); err != nil {
			return err
		}
	}
	if flags.Blend != 0 {
		if err := layer.SetBlendingMode(aep.BlendingMode(flags.Blend)); err != nil {
			return err
		}
	}
	quality, ok := convertLayerQuality(source.Quality)
	if ok {
		if err := layer.SetQuality(quality); err != nil {
			return err
		}
	}
	return nil
}

func convertLayerQuality(value string) (aep.LayerQuality, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "wireframe":
		return aep.LayerQualityWireframe, true
	case "draft":
		return aep.LayerQualityDraft, true
	case "best":
		return aep.LayerQualityBest, true
	default:
		return 0, false
	}
}

func materializeDefaultTransformSurface(layer *aep.Layer, source profile.Layer) error {
	if !hasTransformProperties(source) {
		return nil
	}
	transform, err := layerTransformFromStaticProfile(source)
	if err != nil {
		return err
	}
	return aep.SetLayerTransform(layer, transform)
}

func materializeCenteredTransformSurface(comp *aep.Composition, layer *aep.Layer, source profile.Layer) error {
	if !hasTransformProperties(source) {
		return nil
	}
	if layer == nil {
		return fmt.Errorf("layer not found after creation")
	}
	if _, ok := propertyVectorAtLeast(source.Properties, "ADBE Position", 2); ok {
		return materializeDefaultTransformSurface(layer, source)
	}
	transform := aep.NewLayerTransform()
	if err := transform.Position().SetStaticValue([2]float64{float64(comp.Width) / 2, float64(comp.Height) / 2}); err != nil {
		return err
	}
	return aep.SetLayerTransform(layer, transform)
}

func materializeCameraLightTransformSurface(layer *aep.Layer, source profile.Layer) error {
	if hasProperty(source, "ADBE Position_2") {
		return nil
	}
	return materializeDefaultTransformSurface(layer, source)
}

func materializePrecompTransformSurface(comp *aep.Composition, layer *aep.Layer, source profile.Layer) error {
	if hasStaticPropertyVector(source, "ADBE Position", []float64{0, 0, 0}) {
		return materializeDefaultTransformSurface(layer, source)
	}
	return materializeCenteredTransformSurface(comp, layer, source)
}

func materializeRectGraphicShapeLayer(comp *aep.Composition, source profile.Layer) (*aep.Layer, error) {
	shapeLayer, err := aep.NewShapeLayer(comp, source.Name)
	if err != nil {
		return nil, err
	}
	if err := materializeShapePrimitive(shapeLayer, source.Shapes[0]); err != nil {
		return nil, err
	}
	if hasProperty(source, "ADBE Vector RoundCorner Radius") {
		if err := materializeShapeRoundCorners(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Offset Amount") {
		if err := materializeShapeOffsetPaths(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Trim Start") ||
		hasProperty(source, "ADBE Vector Trim End") ||
		hasProperty(source, "ADBE Vector Trim Offset") {
		if err := materializeShapeTrim(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Zigzag Size") ||
		hasProperty(source, "ADBE Vector Zigzag Detail") ||
		hasProperty(source, "ADBE Vector Zigzag Points") {
		if err := materializeShapeZigZag(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector PuckerBloat Amount") {
		if err := materializeShapePuckerBloat(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Twist Angle") ||
		hasProperty(source, "ADBE Vector Twist Center") {
		if err := materializeShapeTwist(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasWigglePathsFilter(source) {
		if err := materializeShapeWigglePaths(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasWiggleTransformFilter(source) {
		if err := materializeShapeWiggleTransform(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasRepeaterFilter(source) {
		if err := materializeShapeRepeater(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Merge Type") {
		if err := materializeShapeMergePaths(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasGradientFillGraphic(source) {
		if err := materializeShapeGradientFill(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Fill Color") {
		if err := materializeShapeFill(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Stroke Color") {
		if err := materializeShapeStroke(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	return shapeLayer.Layer, nil
}

func materializeShapePrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	switch shape.Kind {
	case "rect":
		return materializeRectPrimitive(shapeLayer, shape)
	case "ellipse":
		return materializeEllipsePrimitive(shapeLayer, shape)
	case "star":
		return materializeStarPrimitive(shapeLayer, shape)
	default:
		return fmt.Errorf("unsupported shape primitive %q", shape.Kind)
	}
}

func materializeRectPrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	rect, err := shapeLayer.RootGroup().AddRect()
	if err != nil {
		return err
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Rect Size", 2); ok {
		if err := rect.SetSize([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Rect Position", 2); ok {
		if err := rect.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Rect Roundness"); ok {
		if err := rect.SetRoundness(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Shape Direction"); ok {
		if err := rect.SetDirection(aep.ShapeDirection(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeEllipsePrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	ellipse, err := shapeLayer.RootGroup().AddEllipse()
	if err != nil {
		return err
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Ellipse Size", 2); ok {
		if err := ellipse.SetSize([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Ellipse Position", 2); ok {
		if err := ellipse.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Shape Direction"); ok {
		if err := ellipse.SetDirection(aep.ShapeDirection(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeStarPrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	star, err := shapeLayer.RootGroup().AddStar()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Type"); ok {
		if err := star.SetStarType(aep.StarType(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Points"); ok {
		if err := star.SetPoints(value); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Star Position", 2); ok {
		if err := star.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Rotation"); ok {
		if err := star.SetRotation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Inner Radius"); ok {
		if err := star.SetInnerRadius(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Outer Radius"); ok {
		if err := star.SetOuterRadius(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Inner Roundess"); ok {
		if err := star.SetInnerRoundness(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Outer Roundess"); ok {
		if err := star.SetOuterRoundness(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeRoundCorners(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	roundCorners, err := shapeLayer.RootGroup().AddRoundCorners()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector RoundCorner Radius"); ok {
		if err := roundCorners.SetRadius(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeOffsetPaths(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	offsetPaths, err := shapeLayer.RootGroup().AddOffsetPaths()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Amount"); ok {
		if err := offsetPaths.SetAmount(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Line Join"); ok {
		if err := offsetPaths.SetLineJoin(aep.StrokeLineJoin(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Miter Limit"); ok {
		if err := offsetPaths.SetMiterLimit(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Copies"); ok {
		if err := offsetPaths.SetCopies(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Copy Offset"); ok {
		if err := offsetPaths.SetCopyOffset(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeTrim(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	trim, err := shapeLayer.RootGroup().AddTrim()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Trim Start"); ok {
		if err := trim.SetStart(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Trim End"); ok {
		if err := trim.SetEnd(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Trim Offset"); ok {
		if err := trim.SetOffset(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Trim Type"); ok {
		if err := trim.SetType(aep.TrimType(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeZigZag(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	zigZag, err := shapeLayer.RootGroup().AddZigZag()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Zigzag Size"); ok {
		if err := zigZag.SetSize(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Zigzag Detail"); ok {
		if err := zigZag.SetDetail(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Zigzag Points"); ok {
		if err := zigZag.SetPoints(aep.ZigZagPoints(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapePuckerBloat(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	puckerBloat, err := shapeLayer.RootGroup().AddPuckerBloat()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector PuckerBloat Amount"); ok {
		if err := puckerBloat.SetAmount(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeTwist(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	twist, err := shapeLayer.RootGroup().AddTwist()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Twist Angle"); ok {
		if err := twist.SetAngle(value); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Twist Center", 2); ok {
		if err := twist.SetCenter([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeWigglePaths(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	wigglePaths, err := shapeLayer.RootGroup().AddWigglePaths()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Roughen Size"); ok {
		if err := wigglePaths.SetSize(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Roughen Detail"); ok {
		if err := wigglePaths.SetDetail(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Temporal Freq"); ok {
		if err := wigglePaths.SetWigglesPerSecond(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Random Seed"); ok {
		if err := wigglePaths.SetRandomSeed(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Roughen Points"); ok {
		if err := wigglePaths.SetPoints(aep.RoughenPoints(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Correlation"); ok {
		if err := wigglePaths.SetCorrelation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Temporal Phase"); ok {
		if err := wigglePaths.SetTemporalPhase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Spatial Phase"); ok {
		if err := wigglePaths.SetSpatialPhase(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeWiggleTransform(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	wiggleTransform, err := shapeLayer.RootGroup().AddWiggleTransform()
	if err != nil {
		return err
	}
	transform := wiggleTransform.Transform()
	if value, ok := propertyVector(source.Properties, "ADBE Vector Wiggler Anchor", 2); ok {
		if err := transform.SetAnchor([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Wiggler Position", 2); ok {
		if err := transform.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Wiggler Scale", 2); ok {
		if err := transform.SetScale([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Wiggler Rotation"); ok {
		if err := transform.SetRotation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Xform Temporal Freq"); ok {
		if err := wiggleTransform.SetWigglesPerSecond(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Random Seed"); ok {
		if err := wiggleTransform.SetRandomSeed(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Correlation"); ok {
		if err := wiggleTransform.SetCorrelation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Temporal Phase"); ok {
		if err := wiggleTransform.SetTemporalPhase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Spatial Phase"); ok {
		if err := wiggleTransform.SetSpatialPhase(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeRepeater(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	repeater, err := shapeLayer.RootGroup().AddRepeater()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Copies"); ok {
		if err := repeater.SetCopies(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Offset"); ok {
		if err := repeater.SetOffset(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Order"); ok {
		if err := repeater.SetOrder(aep.RepeaterOrder(int(value))); err != nil {
			return err
		}
	}
	transform := repeater.Transform()
	if value, ok := propertyVector(source.Properties, "ADBE Vector Repeater Anchor", 2); ok {
		if err := transform.SetAnchor([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Repeater Position", 2); ok {
		if err := transform.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Repeater Scale", 2); ok {
		if err := transform.SetScale([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Rotation"); ok {
		if err := transform.SetRotation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Opacity 1"); ok {
		if err := transform.SetStartOpacity(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Opacity 2"); ok {
		if err := transform.SetEndOpacity(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeMergePaths(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	mergePaths, err := shapeLayer.RootGroup().AddMergePaths()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Merge Type"); ok {
		if err := mergePaths.SetType(aep.MergeType(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeGradientFill(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	fill, err := shapeLayer.RootGroup().AddGradientFill()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Grad Type"); ok {
		if err := fill.SetGradientType(aep.GradientType(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Grad Start Pt", 2); ok {
		if err := fill.SetStartPoint([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Grad End Pt", 2); ok {
		if err := fill.SetEndPoint([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Grad HiLite Length"); ok {
		if err := fill.SetHighlightLength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Grad HiLite Angle"); ok {
		if err := fill.SetHighlightAngle(value); err != nil {
			return err
		}
	}
	if gradient, ok := propertyGradient(source.Properties, "ADBE Vector Grad Colors"); ok {
		if len(gradient.ColorStops) > 0 {
			if err := fill.SetColorStops(gradient.ColorStops); err != nil {
				return err
			}
		}
		if len(gradient.AlphaStops) > 0 {
			if err := fill.SetAlphaStops(gradient.AlphaStops); err != nil {
				return err
			}
		}
	}
	return nil
}

func materializeShapeFill(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	fill, err := shapeLayer.RootGroup().AddFill()
	if err != nil {
		return err
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Fill Color", 4); ok {
		if err := fill.SetColor(profileARGBToRGBA(value)); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Fill Opacity"); ok {
		if err := fill.SetOpacity(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Blend Mode"); ok {
		if err := fill.SetBlendMode(aep.ShapeBlendMode(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Composite Order"); ok {
		if err := fill.SetCompositeOrder(aep.ShapeCompositeOrder(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Fill Rule"); ok {
		if err := fill.SetFillRule(aep.FillRule(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeStroke(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	stroke, err := shapeLayer.RootGroup().AddStroke()
	if err != nil {
		return err
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Stroke Color", 4); ok {
		if err := stroke.SetColor(profileARGBToRGBA(value)); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Opacity"); ok {
		if err := stroke.SetOpacity(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Width"); ok {
		if err := stroke.SetWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Line Cap"); ok {
		if err := stroke.SetLineCap(aep.StrokeLineCap(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Line Join"); ok {
		if err := stroke.SetLineJoin(aep.StrokeLineJoin(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Miter Limit"); ok {
		if err := stroke.SetMiterLimit(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Blend Mode"); ok {
		if err := stroke.SetBlendMode(aep.ShapeBlendMode(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Composite Order"); ok {
		if err := stroke.SetCompositeOrder(aep.ShapeCompositeOrder(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Dash 1"); ok {
		if err := stroke.Dashes().SetDash(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Gap 1"); ok {
		if err := stroke.Dashes().SetGap(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Start Length"); ok {
		if err := stroke.Taper().SetStartLength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper End Length"); ok {
		if err := stroke.Taper().SetEndLength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Start Width"); ok {
		if err := stroke.Taper().SetStartWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper End Width"); ok {
		if err := stroke.Taper().SetEndWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Start Ease"); ok {
		if err := stroke.Taper().SetStartEase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper End Ease"); ok {
		if err := stroke.Taper().SetEndEase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Wave Amount"); ok {
		if err := stroke.Wave().SetAmount(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Wavelength"); ok {
		if err := stroke.Wave().SetWavelength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Wave Phase"); ok {
		if err := stroke.Wave().SetPhase(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeLayerTiming(layer *aep.Layer, source profile.Layer) error {
	if source.Timing.StartTime != 0 {
		if err := layer.SetStartTime(source.Timing.StartTime); err != nil {
			return err
		}
	}
	if source.Timing.InPoint != 0 {
		if err := layer.SetInPoint(source.Timing.InPoint); err != nil {
			return err
		}
	}
	if source.Timing.OutPoint != 0 && !isDefaultLayerOutPoint(source.Timing) {
		if err := layer.SetOutPoint(source.Timing.OutPoint); err != nil {
			return err
		}
	}
	if source.Timing.Stretch != 0 && source.Timing.Stretch != 1 {
		if err := layer.SetStretch(source.Timing.Stretch); err != nil {
			return err
		}
	}
	return nil
}

func isDefaultLayerOutPoint(timing profile.LayerTiming) bool {
	return timing.InPoint == 0 && math.Abs(timing.OutPoint-timing.Duration) < 1e-6
}

func materializeTextStyle(layer *aep.Layer, source *profile.TextSource) error {
	if source == nil {
		return nil
	}
	for i, run := range source.Runs {
		if err := layer.SetRunFontSize(i, run.FontSize); err != nil {
			return fmt.Errorf("run %d font_size: %w", i, err)
		}
		if err := layer.SetRunFillColor(i, run.FillColor); err != nil {
			return fmt.Errorf("run %d fill_color: %w", i, err)
		}
		if err := layer.SetRunAutoLeading(i, run.AutoLeading); err != nil {
			return fmt.Errorf("run %d auto_leading: %w", i, err)
		}
		if err := layer.SetRunLeading(i, run.Leading); err != nil {
			return fmt.Errorf("run %d leading: %w", i, err)
		}
		if err := layer.SetRunTracking(i, run.Tracking); err != nil {
			return fmt.Errorf("run %d tracking: %w", i, err)
		}
		if err := layer.SetRunBaselineShift(i, run.BaselineShift); err != nil {
			return fmt.Errorf("run %d baseline_shift: %w", i, err)
		}
		if err := layer.SetRunHorizontalScale(i, run.HorizontalScale); err != nil {
			return fmt.Errorf("run %d horizontal_scale: %w", i, err)
		}
		if err := layer.SetRunVerticalScale(i, run.VerticalScale); err != nil {
			return fmt.Errorf("run %d vertical_scale: %w", i, err)
		}
		if err := layer.SetRunTsume(i, run.Tsume); err != nil {
			return fmt.Errorf("run %d tsume: %w", i, err)
		}
		caps, err := textCapsOption(run.CapsOption)
		if err != nil {
			return fmt.Errorf("run %d caps_option: %w", i, err)
		}
		if err := layer.SetRunCapsOption(i, caps); err != nil {
			return fmt.Errorf("run %d caps_option: %w", i, err)
		}
		baseline, err := textBaselineOption(run.BaselineOption)
		if err != nil {
			return fmt.Errorf("run %d baseline_option: %w", i, err)
		}
		if err := layer.SetRunBaselineOption(i, baseline); err != nil {
			return fmt.Errorf("run %d baseline_option: %w", i, err)
		}
		kern, err := textAutoKernType(run.AutoKernType)
		if err != nil {
			return fmt.Errorf("run %d auto_kern_type: %w", i, err)
		}
		if err := layer.SetRunAutoKernType(i, kern); err != nil {
			return fmt.Errorf("run %d auto_kern_type: %w", i, err)
		}
		lineJoin, err := textLineJoinType(run.LineJoinType)
		if err != nil {
			return fmt.Errorf("run %d line_join_type: %w", i, err)
		}
		if err := layer.SetRunLineJoinType(i, lineJoin); err != nil {
			return fmt.Errorf("run %d line_join_type: %w", i, err)
		}
		digitSet, err := textDigitSet(run.DigitSet)
		if err != nil {
			return fmt.Errorf("run %d digit_set: %w", i, err)
		}
		if err := layer.SetRunDigitSet(i, digitSet); err != nil {
			return fmt.Errorf("run %d digit_set: %w", i, err)
		}
		if err := layer.SetRunNoBreak(i, run.NoBreak); err != nil {
			return fmt.Errorf("run %d no_break: %w", i, err)
		}
		if err := layer.SetRunFauxBold(i, run.FauxBold); err != nil {
			return fmt.Errorf("run %d faux_bold: %w", i, err)
		}
		if err := layer.SetRunFauxItalic(i, run.FauxItalic); err != nil {
			return fmt.Errorf("run %d faux_italic: %w", i, err)
		}
		if err := layer.SetRunApplyStroke(i, run.ApplyStroke); err != nil {
			return fmt.Errorf("run %d apply_stroke: %w", i, err)
		}
		if err := layer.SetRunStrokeColor(i, run.StrokeColor); err != nil {
			return fmt.Errorf("run %d stroke_color: %w", i, err)
		}
		if err := layer.SetRunStrokeWidth(i, run.StrokeWidth); err != nil {
			return fmt.Errorf("run %d stroke_width: %w", i, err)
		}
		if err := layer.SetRunStrokeOverFill(i, run.StrokeOverFill); err != nil {
			return fmt.Errorf("run %d stroke_over_fill: %w", i, err)
		}
	}
	for i, paragraph := range source.Paragraphs {
		justification, err := textJustification(paragraph.Justification)
		if err != nil {
			return fmt.Errorf("paragraph %d justification: %w", i, err)
		}
		if err := layer.SetParagraphJustification(i, justification); err != nil {
			return fmt.Errorf("paragraph %d justification: %w", i, err)
		}
		if err := layer.SetParagraphFirstLineIndent(i, paragraph.FirstLineIndent); err != nil {
			return fmt.Errorf("paragraph %d first_line_indent: %w", i, err)
		}
		if err := layer.SetParagraphStartIndent(i, paragraph.StartIndent); err != nil {
			return fmt.Errorf("paragraph %d start_indent: %w", i, err)
		}
		if err := layer.SetParagraphEndIndent(i, paragraph.EndIndent); err != nil {
			return fmt.Errorf("paragraph %d end_indent: %w", i, err)
		}
		if err := layer.SetParagraphSpaceBefore(i, paragraph.SpaceBefore); err != nil {
			return fmt.Errorf("paragraph %d space_before: %w", i, err)
		}
		if err := layer.SetParagraphSpaceAfter(i, paragraph.SpaceAfter); err != nil {
			return fmt.Errorf("paragraph %d space_after: %w", i, err)
		}
		if err := layer.SetParagraphAutoHyphenate(i, paragraph.AutoHyphenate); err != nil {
			return fmt.Errorf("paragraph %d auto_hyphenate: %w", i, err)
		}
		leadingType, err := textLeadingType(paragraph.LeadingType)
		if err != nil {
			return fmt.Errorf("paragraph %d leading_type: %w", i, err)
		}
		if err := layer.SetParagraphLeadingType(i, leadingType); err != nil {
			return fmt.Errorf("paragraph %d leading_type: %w", i, err)
		}
		if err := layer.SetParagraphHangingRoman(i, paragraph.HangingRoman); err != nil {
			return fmt.Errorf("paragraph %d hanging_roman: %w", i, err)
		}
		direction, err := textParagraphDirection(paragraph.Direction)
		if err != nil {
			return fmt.Errorf("paragraph %d direction: %w", i, err)
		}
		if err := layer.SetParagraphDirection(i, direction); err != nil {
			return fmt.Errorf("paragraph %d direction: %w", i, err)
		}
	}
	return nil
}

func normalizeTextEnum(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.ReplaceAll(value, "-", "_")
}

func textJustification(value string) (aep.TextJustification, error) {
	switch normalizeTextEnum(value) {
	case "", "left":
		return aep.TextJustifyLeft, nil
	case "right":
		return aep.TextJustifyRight, nil
	case "center":
		return aep.TextJustifyCenter, nil
	default:
		return 0, fmt.Errorf("unsupported justification %q", value)
	}
}

func textCapsOption(value string) (aep.TextCapsOption, error) {
	switch normalizeTextEnum(value) {
	case "", "normal":
		return aep.TextCapsNormal, nil
	case "small_caps":
		return aep.TextCapsSmall, nil
	case "all_caps":
		return aep.TextCapsAll, nil
	case "all_small_caps":
		return aep.TextCapsAllSmall, nil
	default:
		return 0, fmt.Errorf("unsupported caps_option %q", value)
	}
}

func textBaselineOption(value string) (aep.TextBaselineOption, error) {
	switch normalizeTextEnum(value) {
	case "", "normal":
		return aep.TextBaselineNormal, nil
	case "superscript":
		return aep.TextBaselineSuperscript, nil
	case "subscript":
		return aep.TextBaselineSubscript, nil
	default:
		return 0, fmt.Errorf("unsupported baseline_option %q", value)
	}
}

func textAutoKernType(value string) (aep.TextAutoKernType, error) {
	switch normalizeTextEnum(value) {
	case "no_auto":
		return aep.TextAutoKernNoAuto, nil
	case "", "metric":
		return aep.TextAutoKernMetric, nil
	case "optical":
		return aep.TextAutoKernOptical, nil
	default:
		return 0, fmt.Errorf("unsupported auto_kern_type %q", value)
	}
}

func textLineJoinType(value string) (aep.TextLineJoinType, error) {
	switch normalizeTextEnum(value) {
	case "", "miter":
		return aep.TextLineJoinMiter, nil
	case "round":
		return aep.TextLineJoinRound, nil
	case "bevel":
		return aep.TextLineJoinBevel, nil
	default:
		return 0, fmt.Errorf("unsupported line_join_type %q", value)
	}
}

func textDigitSet(value string) (aep.TextDigitSet, error) {
	switch normalizeTextEnum(value) {
	case "", "default":
		return aep.TextDigitSetDefault, nil
	case "arabic":
		return aep.TextDigitSetArabic, nil
	case "hindi":
		return aep.TextDigitSetHindi, nil
	case "farsi":
		return aep.TextDigitSetFarsi, nil
	case "arabic_rtl":
		return aep.TextDigitSetArabicRTL, nil
	default:
		return 0, fmt.Errorf("unsupported digit_set %q", value)
	}
}

func textLeadingType(value string) (aep.TextLeadingType, error) {
	switch normalizeTextEnum(value) {
	case "", "roman":
		return aep.TextLeadingRoman, nil
	case "japanese":
		return aep.TextLeadingJapanese, nil
	default:
		return 0, fmt.Errorf("unsupported leading_type %q", value)
	}
}

func textParagraphDirection(value string) (aep.TextParagraphDirection, error) {
	switch normalizeTextEnum(value) {
	case "", "ltr":
		return aep.TextDirectionLeftToRight, nil
	case "rtl":
		return aep.TextDirectionRightToLeft, nil
	default:
		return 0, fmt.Errorf("unsupported paragraph_direction %q", value)
	}
}

func propertyVector(properties []profile.Property, matchName string, length int) ([]float64, bool) {
	for _, property := range properties {
		if property.MatchName != matchName {
			continue
		}
		return staticVector(property.StaticValue, length)
	}
	return nil, false
}

func propertyByMatchName(properties []profile.Property, matchName string) (profile.Property, bool) {
	for _, property := range properties {
		if property.MatchName == matchName {
			return property, true
		}
	}
	return profile.Property{}, false
}

func propertyFloat(properties []profile.Property, matchName string) (float64, bool) {
	for _, property := range properties {
		if property.MatchName != matchName {
			continue
		}
		return propertyFloatValue(property.StaticValue)
	}
	return 0, false
}

func propertyFloatValue(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func propertyGradient(properties []profile.Property, matchName string) (*codec.Gradient, bool) {
	for _, property := range properties {
		if property.MatchName != matchName || property.Gradient == nil {
			continue
		}
		return cloneCodecGradient(property.Gradient), true
	}
	return nil, false
}

func cloneCodecGradient(g *codec.Gradient) *codec.Gradient {
	if g == nil {
		return nil
	}
	return &codec.Gradient{
		Version:    g.Version,
		ColorStops: append([]codec.GradientColorStop(nil), g.ColorStops...),
		AlphaStops: append([]codec.GradientAlphaStop(nil), g.AlphaStops...),
	}
}

func staticVector(value any, length int) ([]float64, bool) {
	var out []float64
	switch v := value.(type) {
	case []float64:
		out = append([]float64(nil), v...)
	case []any:
		out = make([]float64, 0, len(v))
		for _, item := range v {
			got, ok := item.(float64)
			if !ok {
				return nil, false
			}
			out = append(out, got)
		}
	case [2]float64:
		out = []float64{v[0], v[1]}
	case [3]float64:
		out = []float64{v[0], v[1], v[2]}
	case [4]float64:
		out = []float64{v[0], v[1], v[2], v[3]}
	default:
		return nil, false
	}
	if len(out) != length {
		return nil, false
	}
	return out, true
}

func profileARGBToRGBA(value []float64) [4]float64 {
	return [4]float64{
		colorByteToUnit(value[1]),
		colorByteToUnit(value[2]),
		colorByteToUnit(value[3]),
		colorByteToUnit(value[0]),
	}
}

func colorByteToUnit(value float64) float64 {
	if value > 1 {
		return value / 255
	}
	return value
}

func layerTransformFromStaticProfile(layer profile.Layer) (*aep.LayerTransform, error) {
	transform := aep.NewLayerTransform()
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Anchor Point"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformVectorKeyframes(transform.AnchorPoint(), property, profileVector2); err != nil {
				return nil, fmt.Errorf("anchor point keyframes: %w", err)
			}
		} else if value, ok := staticVectorAtLeast(property.StaticValue, 2); ok {
			if err := transform.AnchorPoint().SetStaticValue([2]float64{value[0], value[1]}); err != nil {
				return nil, err
			}
		}
	}
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Position"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformVectorKeyframes(transform.Position(), property, profileVector2); err != nil {
				return nil, fmt.Errorf("position keyframes: %w", err)
			}
		} else if value, ok := staticVectorAtLeast(property.StaticValue, 2); ok {
			if err := transform.Position().SetStaticValue([2]float64{value[0], value[1]}); err != nil {
				return nil, err
			}
		}
	}
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Scale"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformVectorKeyframes(transform.Scale(), property, profileScaleVector2); err != nil {
				return nil, fmt.Errorf("scale keyframes: %w", err)
			}
		} else if value, ok := staticVectorAtLeast(property.StaticValue, 2); ok {
			if err := transform.Scale().SetStaticValue([2]float64{profileScaleToWriter(value[0]), profileScaleToWriter(value[1])}); err != nil {
				return nil, err
			}
		}
	}
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Rotate Z"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformScalarKeyframes(transform.Rotation(), property, profileScalar); err != nil {
				return nil, fmt.Errorf("rotation keyframes: %w", err)
			}
		} else if value, ok := propertyFloatValue(property.StaticValue); ok {
			if err := transform.Rotation().SetStaticValue(value); err != nil {
				return nil, err
			}
		}
	}
	if property, ok := propertyByMatchName(layer.Properties, "ADBE Opacity"); ok {
		if len(property.Keyframes) != 0 {
			if err := addTransformScalarKeyframes(transform.Opacity(), property, profileOpacityScalar); err != nil {
				return nil, fmt.Errorf("opacity keyframes: %w", err)
			}
		} else if value, ok := propertyFloatValue(property.StaticValue); ok {
			if err := transform.Opacity().SetStaticValue(profileOpacityToWriter(value)); err != nil {
				return nil, err
			}
		}
	}
	return transform, nil
}

func addTransformVectorKeyframes(stream *codec.PropertyStream[[2]float64], property profile.Property, convert func(any) ([2]float64, bool)) error {
	if len(property.Keyframes) < 2 {
		return fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	for i, keyframe := range property.Keyframes {
		value, ok := convert(keyframe.Value)
		if !ok {
			return fmt.Errorf("keyframes[%d].value unsupported (%T)", i, keyframe.Value)
		}
		inEase := transformTemporalEase(keyframe.InTemporalEase)
		outEase := transformTemporalEase(keyframe.OutTemporalEase)
		if inEase != (aep.TemporalEase{}) || outEase != (aep.TemporalEase{}) {
			if err := stream.AddKeyframeWithEase(keyframe.Time, value, inEase, outEase); err != nil {
				return err
			}
			continue
		}
		if err := stream.AddKeyframeLinear(keyframe.Time, value); err != nil {
			return err
		}
	}
	return nil
}

func addTransformScalarKeyframes(stream *codec.PropertyStream[float64], property profile.Property, convert func(any) (float64, bool)) error {
	if len(property.Keyframes) < 2 {
		return fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	for i, keyframe := range property.Keyframes {
		value, ok := convert(keyframe.Value)
		if !ok {
			return fmt.Errorf("keyframes[%d].value unsupported (%T)", i, keyframe.Value)
		}
		inEase := transformTemporalEase(keyframe.InTemporalEase)
		outEase := transformTemporalEase(keyframe.OutTemporalEase)
		if inEase != (aep.TemporalEase{}) || outEase != (aep.TemporalEase{}) {
			if err := stream.AddKeyframeWithEase(keyframe.Time, value, inEase, outEase); err != nil {
				return err
			}
			continue
		}
		if err := stream.AddKeyframeLinear(keyframe.Time, value); err != nil {
			return err
		}
	}
	return nil
}

func profileVector2(value any) ([2]float64, bool) {
	vector, ok := staticVectorAtLeast(value, 2)
	if !ok {
		return [2]float64{}, false
	}
	return [2]float64{vector[0], vector[1]}, true
}

func profileScaleVector2(value any) ([2]float64, bool) {
	vector, ok := staticVectorAtLeast(value, 2)
	if !ok {
		return [2]float64{}, false
	}
	return [2]float64{profileScaleToWriter(vector[0]), profileScaleToWriter(vector[1])}, true
}

func profileScalar(value any) (float64, bool) {
	return propertyFloatValue(value)
}

func profileOpacityScalar(value any) (float64, bool) {
	valueFloat, ok := propertyFloatValue(value)
	if !ok {
		return 0, false
	}
	return profileOpacityToWriter(valueFloat), true
}

func transformTemporalEase(in []profile.TemporalEase) aep.TemporalEase {
	if len(in) == 0 {
		return aep.TemporalEase{}
	}
	return aep.TemporalEase{Speed: in[0].Speed, Influence: in[0].Influence}
}

func propertyVectorAtLeast(properties []profile.Property, matchName string, length int) ([]float64, bool) {
	for _, property := range properties {
		if property.MatchName != matchName {
			continue
		}
		return staticVectorAtLeast(property.StaticValue, length)
	}
	return nil, false
}

func staticVectorAtLeast(value any, length int) ([]float64, bool) {
	out, ok := staticVectorAny(value)
	if !ok || len(out) < length {
		return nil, false
	}
	return out, true
}

func staticVectorAny(value any) ([]float64, bool) {
	switch v := value.(type) {
	case []float64:
		return append([]float64(nil), v...), true
	case []any:
		out := make([]float64, 0, len(v))
		for _, item := range v {
			got, ok := item.(float64)
			if !ok {
				return nil, false
			}
			out = append(out, got)
		}
		return out, true
	case [2]float64:
		return []float64{v[0], v[1]}, true
	case [3]float64:
		return []float64{v[0], v[1], v[2]}, true
	case [4]float64:
		return []float64{v[0], v[1], v[2], v[3]}, true
	default:
		return nil, false
	}
}

func profileScaleToWriter(value float64) float64 {
	return value * 100
}

func profileOpacityToWriter(value float64) float64 {
	return value * 100
}

func hasTransformProperties(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Anchor Point") ||
		hasProperty(layer, "ADBE Position") ||
		hasProperty(layer, "ADBE Scale") ||
		hasProperty(layer, "ADBE Rotate Z") ||
		hasProperty(layer, "ADBE Opacity")
}

func hasProperty(layer profile.Layer, matchName string) bool {
	for _, property := range layer.Properties {
		if property.MatchName == matchName {
			return true
		}
	}
	return false
}

func hasStaticPropertyVector(layer profile.Layer, matchName string, want []float64) bool {
	for _, property := range layer.Properties {
		if property.MatchName != matchName {
			continue
		}
		return vectorEquals(property.StaticValue, want)
	}
	return false
}

func vectorEquals(value any, want []float64) bool {
	switch v := value.(type) {
	case []float64:
		if len(v) != len(want) {
			return false
		}
		for i := range want {
			if v[i] != want[i] {
				return false
			}
		}
		return true
	case []any:
		if len(v) != len(want) {
			return false
		}
		for i := range want {
			got, ok := v[i].(float64)
			if !ok || got != want[i] {
				return false
			}
		}
		return true
	case [2]float64:
		return len(want) == 2 && v[0] == want[0] && v[1] == want[1]
	case [3]float64:
		return len(want) == 3 && v[0] == want[0] && v[1] == want[1] && v[2] == want[2]
	default:
		return false
	}
}

func isSupportedDefaultNullLayer(layer profile.Layer, footage convertFootageIndex) bool {
	if layer.Type != "null" {
		return false
	}
	if layer.SourceRef != nil {
		solid, ok := footage.solidDetails(layer)
		if !ok || solid.Width == 0 || solid.Height == 0 || solid.SolidColor == nil {
			return false
		}
	}
	if layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if layer.Text != nil || len(layer.Effects) != 0 || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	flags := layer.Flags
	return flags.Blend == 2 &&
		flags.TrackMatte == 0 &&
		flags.IsNull &&
		flags.EffectsEnabled &&
		flags.AudioEnabled &&
		!flags.Is3D &&
		!flags.Solo &&
		!flags.Shy &&
		!flags.Locked &&
		!flags.IsAdjustment &&
		!flags.IsGuide &&
		!flags.MotionBlur &&
		!flags.FrameBlendEnabled &&
		!flags.MarkersLocked &&
		!flags.FrameBlendPixelMotion &&
		!flags.CollapseTransform &&
		!flags.SamplingBicubic &&
		!flags.PreserveTransparency
}

func isSupportedDefaultSolidLayer(layer profile.Layer, footage convertFootageIndex) bool {
	if layer.Type != "av" || layer.SourceRef == nil || layer.SourceRef.Kind != "footage" {
		return false
	}
	solid, ok := footage.solidDetails(layer)
	if !ok || solid.Width == 0 || solid.Height == 0 || solid.SolidColor == nil {
		return false
	}
	if layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if layer.Text != nil || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	if !isSupportedStaticEffectSurface(layer) {
		return false
	}
	flags := layer.Flags
	return flags.Visible &&
		flags.Blend == 2 &&
		flags.TrackMatte == 0 &&
		!flags.IsNull &&
		flags.EffectsEnabled &&
		flags.AudioEnabled &&
		!flags.Is3D &&
		!flags.Solo &&
		!flags.Shy &&
		!flags.Locked &&
		!flags.IsAdjustment &&
		!flags.IsGuide &&
		!flags.MotionBlur &&
		!flags.FrameBlendEnabled &&
		!flags.MarkersLocked &&
		!flags.FrameBlendPixelMotion &&
		!flags.CollapseTransform &&
		!flags.SamplingBicubic &&
		!flags.PreserveTransparency
}

func isSupportedDefaultAdjustmentLayer(layer profile.Layer, footage convertFootageIndex) bool {
	if layer.Type != "adjustment" || layer.SourceRef == nil || layer.SourceRef.Kind != "footage" {
		return false
	}
	source, ok := footage.solidDetails(layer)
	if !ok || source.Width == 0 || source.Height == 0 || source.SolidColor == nil {
		return false
	}
	if layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if layer.Text != nil || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	if !isSupportedStaticEffectSurface(layer) {
		return false
	}
	flags := layer.Flags
	return flags.Visible &&
		flags.Blend == 2 &&
		flags.TrackMatte == 0 &&
		!flags.IsNull &&
		flags.IsAdjustment &&
		flags.EffectsEnabled &&
		flags.AudioEnabled &&
		!flags.Is3D &&
		!flags.Solo &&
		!flags.Shy &&
		!flags.Locked &&
		!flags.IsGuide &&
		!flags.MotionBlur &&
		!flags.FrameBlendEnabled &&
		!flags.MarkersLocked &&
		!flags.FrameBlendPixelMotion &&
		!flags.CollapseTransform &&
		!flags.SamplingBicubic &&
		!flags.PreserveTransparency
}

func isSupportedDefaultCameraLayer(layer profile.Layer) bool {
	return isSupportedDefaultCameraOrLightLayer(layer, "camera")
}

func isSupportedDefaultLightLayer(layer profile.Layer) bool {
	return isSupportedDefaultCameraOrLightLayer(layer, "light")
}

func isSupportedDefaultCameraOrLightLayer(layer profile.Layer, typ string) bool {
	if layer.Type != typ || layer.SourceRef != nil {
		return false
	}
	if layer.ParentRef != nil || layer.MatteRef != nil {
		return false
	}
	if typ != "light" && layer.LightSourceRef != nil {
		return false
	}
	if layer.Text != nil || len(layer.Effects) != 0 || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	flags := layer.Flags
	return flags.Visible &&
		flags.TrackMatte == 0 &&
		!flags.IsNull &&
		!flags.IsAdjustment &&
		!flags.Solo &&
		!flags.Shy &&
		!flags.Locked &&
		!flags.IsGuide &&
		!flags.MotionBlur &&
		!flags.FrameBlendEnabled &&
		!flags.MarkersLocked &&
		!flags.FrameBlendPixelMotion &&
		!flags.CollapseTransform &&
		!flags.SamplingBicubic &&
		!flags.PreserveTransparency
}

func isSupportedDefaultTextLayer(layer profile.Layer) bool {
	if layer.Type != "text" || layer.Text == nil || layer.SourceRef != nil {
		return false
	}
	if layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	if !isSupportedStaticEffectSurface(layer) {
		return false
	}
	if layer.Text.IsBoxText {
		return false
	}
	flags := layer.Flags
	return flags.Blend != 0 &&
		flags.TrackMatte == 0 &&
		!flags.IsNull &&
		flags.CollapseTransform
}

func isSupportedDefaultShapeLayer(layer profile.Layer) bool {
	if layer.Type != "shape" || layer.SourceRef != nil || layer.Text != nil {
		return false
	}
	if layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if len(layer.Effects) != 0 || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	flags := layer.Flags
	return flags.Visible &&
		flags.Blend == 2 &&
		flags.TrackMatte == 0 &&
		!flags.IsNull &&
		flags.EffectsEnabled &&
		flags.AudioEnabled &&
		!flags.Is3D &&
		!flags.Solo &&
		!flags.Shy &&
		!flags.Locked &&
		!flags.IsAdjustment &&
		!flags.IsGuide &&
		!flags.MotionBlur &&
		!flags.FrameBlendEnabled &&
		!flags.MarkersLocked &&
		!flags.FrameBlendPixelMotion &&
		flags.CollapseTransform &&
		!flags.SamplingBicubic &&
		!flags.PreserveTransparency
}

func isSupportedRectGraphicShapeLayer(layer profile.Layer) bool {
	if !isSupportedShapeLayerBase(layer) {
		return false
	}
	if len(layer.Shapes) != 1 || !isSupportedParametricGraphicShape(layer.Shapes[0]) {
		return false
	}
	hasFill := hasProperty(layer, "ADBE Vector Fill Color")
	hasStroke := hasProperty(layer, "ADBE Vector Stroke Color")
	hasGradientFill := hasGradientFillGraphic(layer)
	if !hasFill && !hasStroke && !hasGradientFill && !hasSupportedShapeFilter(layer) {
		return false
	}
	if (hasProperty(layer, "ADBE Vector Grad Colors") && !hasGradientFill) ||
		hasProperty(layer, "ADBE Vector Stroke Dash 2") ||
		hasProperty(layer, "ADBE Vector Stroke Gap 2") ||
		hasProperty(layer, "ADBE Vector Stroke Offset") ||
		hasProperty(layer, "ADBE Vector Taper Wave Units") ||
		hasProperty(layer, "ADBE Vector Taper Wave Cycles") {
		return false
	}
	return true
}

func hasGradientFillGraphic(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Vector Grad Colors") &&
		hasProperty(layer, "ADBE Vector Grad Type") &&
		!hasProperty(layer, "ADBE Vector Stroke Width")
}

func hasSupportedShapeFilter(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Vector RoundCorner Radius") ||
		hasProperty(layer, "ADBE Vector Offset Amount") ||
		hasProperty(layer, "ADBE Vector Trim Start") ||
		hasProperty(layer, "ADBE Vector Trim End") ||
		hasProperty(layer, "ADBE Vector Trim Offset") ||
		hasProperty(layer, "ADBE Vector Zigzag Size") ||
		hasProperty(layer, "ADBE Vector Zigzag Detail") ||
		hasProperty(layer, "ADBE Vector Zigzag Points") ||
		hasProperty(layer, "ADBE Vector PuckerBloat Amount") ||
		hasProperty(layer, "ADBE Vector Twist Angle") ||
		hasProperty(layer, "ADBE Vector Twist Center") ||
		hasWigglePathsFilter(layer) ||
		hasWiggleTransformFilter(layer) ||
		hasRepeaterFilter(layer) ||
		hasProperty(layer, "ADBE Vector Merge Type")
}

func hasWigglePathsFilter(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Vector Roughen Size") ||
		hasProperty(layer, "ADBE Vector Roughen Detail") ||
		hasProperty(layer, "ADBE Vector Temporal Freq") ||
		hasProperty(layer, "ADBE Vector Roughen Points")
}

func hasWiggleTransformFilter(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Vector Wiggler Anchor") ||
		hasProperty(layer, "ADBE Vector Wiggler Position") ||
		hasProperty(layer, "ADBE Vector Wiggler Scale") ||
		hasProperty(layer, "ADBE Vector Wiggler Rotation") ||
		hasProperty(layer, "ADBE Vector Xform Temporal Freq")
}

func hasRepeaterFilter(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Vector Repeater Copies") ||
		hasProperty(layer, "ADBE Vector Repeater Offset") ||
		hasProperty(layer, "ADBE Vector Repeater Order") ||
		hasProperty(layer, "ADBE Vector Repeater Anchor") ||
		hasProperty(layer, "ADBE Vector Repeater Position") ||
		hasProperty(layer, "ADBE Vector Repeater Scale") ||
		hasProperty(layer, "ADBE Vector Repeater Rotation") ||
		hasProperty(layer, "ADBE Vector Repeater Opacity 1") ||
		hasProperty(layer, "ADBE Vector Repeater Opacity 2")
}

func isSupportedParametricGraphicShape(shape profile.Shape) bool {
	switch shape.Kind {
	case "rect":
		_, ok := propertyVector(shape.Properties, "ADBE Vector Rect Size", 2)
		return ok
	case "ellipse":
		_, ok := propertyVector(shape.Properties, "ADBE Vector Ellipse Size", 2)
		return ok
	case "star":
		_, hasPoints := propertyFloat(shape.Properties, "ADBE Vector Star Points")
		_, hasOuterRadius := propertyFloat(shape.Properties, "ADBE Vector Star Outer Radius")
		return hasPoints || hasOuterRadius
	default:
		return false
	}
}

func isSupportedShapeLayerBase(layer profile.Layer) bool {
	if layer.Type != "shape" || layer.SourceRef != nil || layer.Text != nil {
		return false
	}
	if layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if len(layer.Effects) != 0 || len(layer.Masks) != 0 || len(layer.Markers) != 0 {
		return false
	}
	flags := layer.Flags
	return flags.Visible &&
		flags.Blend == 2 &&
		flags.TrackMatte == 0 &&
		!flags.IsNull &&
		flags.EffectsEnabled &&
		flags.AudioEnabled &&
		!flags.Is3D &&
		!flags.Solo &&
		!flags.Shy &&
		!flags.Locked &&
		!flags.IsAdjustment &&
		!flags.IsGuide &&
		!flags.MotionBlur &&
		!flags.FrameBlendEnabled &&
		!flags.MarkersLocked &&
		!flags.FrameBlendPixelMotion &&
		flags.CollapseTransform &&
		!flags.SamplingBicubic &&
		!flags.PreserveTransparency
}

func isSupportedDefaultPrecompLayer(layer profile.Layer, comps convertCompIndex) bool {
	if layer.Type != "av" || layer.SourceRef == nil || layer.SourceRef.Kind != "composition" {
		return false
	}
	if _, ok := comps.sourceComposition(layer); !ok {
		return false
	}
	if layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if layer.Text != nil || len(layer.Effects) != 0 || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	flags := layer.Flags
	return flags.Visible &&
		flags.Blend == 2 &&
		flags.TrackMatte == 0 &&
		!flags.IsNull &&
		flags.EffectsEnabled &&
		flags.AudioEnabled &&
		!flags.Is3D &&
		!flags.Solo &&
		!flags.Shy &&
		!flags.Locked &&
		!flags.IsAdjustment &&
		!flags.IsGuide &&
		!flags.MotionBlur &&
		!flags.FrameBlendEnabled &&
		!flags.MarkersLocked &&
		!flags.FrameBlendPixelMotion &&
		!flags.CollapseTransform &&
		!flags.SamplingBicubic &&
		!flags.PreserveTransparency
}

type convertCompIndex struct {
	byID   map[uint32]profile.Composition
	byName map[string]profile.Composition
}

func newConvertCompIndex(prof *profile.Profile) convertCompIndex {
	out := convertCompIndex{
		byID:   map[uint32]profile.Composition{},
		byName: map[string]profile.Composition{},
	}
	if prof == nil {
		return out
	}
	for _, comp := range prof.Comps {
		if comp.ID != 0 {
			out.byID[comp.ID] = comp
		}
		if _, exists := out.byName[comp.Name]; !exists {
			out.byName[comp.Name] = comp
		}
	}
	return out
}

func (idx convertCompIndex) sourceComposition(layer profile.Layer) (profile.Composition, bool) {
	if layer.SourceRef == nil {
		return profile.Composition{}, false
	}
	var comp profile.Composition
	var ok bool
	if layer.SourceRef.ID != 0 {
		comp, ok = idx.byID[layer.SourceRef.ID]
	}
	if !ok && layer.SourceRef.Name != "" {
		comp, ok = idx.byName[layer.SourceRef.Name]
	}
	return comp, ok
}

type convertTargetCompIndex struct {
	bySourceID map[uint32]*aep.Composition
	byName     map[string]*aep.Composition
}

func newConvertTargetCompIndex() convertTargetCompIndex {
	return convertTargetCompIndex{
		bySourceID: map[uint32]*aep.Composition{},
		byName:     map[string]*aep.Composition{},
	}
}

func (idx convertTargetCompIndex) add(source profile.Composition, target *aep.Composition) {
	if source.ID != 0 {
		idx.bySourceID[source.ID] = target
	}
	if _, exists := idx.byName[source.Name]; !exists {
		idx.byName[source.Name] = target
	}
}

func (idx convertTargetCompIndex) sourceComposition(layer profile.Layer) (*aep.Composition, bool) {
	if layer.SourceRef == nil {
		return nil, false
	}
	var comp *aep.Composition
	var ok bool
	if layer.SourceRef.ID != 0 {
		comp, ok = idx.bySourceID[layer.SourceRef.ID]
	}
	if !ok && layer.SourceRef.Name != "" {
		comp, ok = idx.byName[layer.SourceRef.Name]
	}
	return comp, ok && comp != nil
}

type convertFootageIndex struct {
	byID   map[uint32]profile.Item
	byName map[string]profile.Item
}

func newConvertFootageIndex(prof *profile.Profile) convertFootageIndex {
	out := convertFootageIndex{
		byID:   map[uint32]profile.Item{},
		byName: map[string]profile.Item{},
	}
	if prof == nil {
		return out
	}
	for _, item := range prof.Items.Footage {
		if item.ID != 0 {
			out.byID[item.ID] = item
		}
		if _, exists := out.byName[item.Name]; !exists {
			out.byName[item.Name] = item
		}
	}
	return out
}

func (idx convertFootageIndex) solidDetails(layer profile.Layer) (profile.FootageDetails, bool) {
	if layer.SourceRef == nil {
		return profile.FootageDetails{}, false
	}
	var item profile.Item
	var ok bool
	if layer.SourceRef.ID != 0 {
		item, ok = idx.byID[layer.SourceRef.ID]
	}
	if !ok && layer.SourceRef.Name != "" {
		item, ok = idx.byName[layer.SourceRef.Name]
	}
	if !ok || item.Footage == nil || item.Footage.AssetType != "solid" {
		return profile.FootageDetails{}, false
	}
	return *item.Footage, true
}

func applyStableCompSettings(dst *aep.Composition, src profile.Composition) error {
	if err := dst.SetBGColor(src.BackgroundColor); err != nil {
		return fmt.Errorf("comp %q background_color: %w", src.Name, err)
	}
	if src.MotionGraphicsTemplateName != "" {
		if err := dst.SetMotionGraphicsTemplateName(src.MotionGraphicsTemplateName); err != nil {
			return fmt.Errorf("comp %q motion_graphics_template_name: %w", src.Name, err)
		}
	}
	if src.Renderer != "" {
		if err := aep.SetRenderer(dst, src.Renderer); err != nil {
			return fmt.Errorf("comp %q renderer: %w", src.Name, err)
		}
	}
	if err := dst.SetResolutionFactor(src.ResolutionFactor[0], src.ResolutionFactor[1]); err != nil {
		return fmt.Errorf("comp %q resolution_factor: %w", src.Name, err)
	}
	if err := dst.SetPixelAspect(src.PixelAspect); err != nil {
		return fmt.Errorf("comp %q pixel_aspect: %w", src.Name, err)
	}
	if err := dst.SetDisplayStartTime(src.DisplayStartTime); err != nil {
		return fmt.Errorf("comp %q display_start_time: %w", src.Name, err)
	}
	if err := dst.SetFrameBlending(src.FrameBlending); err != nil {
		return fmt.Errorf("comp %q frame_blending: %w", src.Name, err)
	}
	if err := dst.SetDraft3D(src.Draft3D); err != nil {
		return fmt.Errorf("comp %q draft_3d: %w", src.Name, err)
	}
	if err := dst.SetHideShyLayers(src.HideShyLayers); err != nil {
		return fmt.Errorf("comp %q hide_shy_layers: %w", src.Name, err)
	}
	if err := dst.SetPreserveNestedFrameRate(src.PreserveNestedFrameRate); err != nil {
		return fmt.Errorf("comp %q preserve_nested_frame_rate: %w", src.Name, err)
	}
	if err := dst.SetPreserveNestedResolution(src.PreserveNestedResolution); err != nil {
		return fmt.Errorf("comp %q preserve_nested_resolution: %w", src.Name, err)
	}
	if err := dst.SetCompMotionBlur(src.MotionBlur.Enabled); err != nil {
		return fmt.Errorf("comp %q motion_blur.enabled: %w", src.Name, err)
	}
	if err := dst.SetShutterAngle(src.MotionBlur.ShutterAngle); err != nil {
		return fmt.Errorf("comp %q motion_blur.shutter_angle: %w", src.Name, err)
	}
	if err := dst.SetShutterPhase(src.MotionBlur.ShutterPhase); err != nil {
		return fmt.Errorf("comp %q motion_blur.shutter_phase: %w", src.Name, err)
	}
	if err := dst.SetMotionBlurAdaptiveSampleLimit(src.MotionBlur.AdaptiveSampleLimit); err != nil {
		return fmt.Errorf("comp %q motion_blur.adaptive_sample_limit: %w", src.Name, err)
	}
	if err := dst.SetMotionBlurSamplesPerFrame(src.MotionBlur.SamplesPerFrame); err != nil {
		return fmt.Errorf("comp %q motion_blur.samples_per_frame: %w", src.Name, err)
	}
	if err := dst.SetWorkArea(src.WorkArea.Start, src.WorkArea.End); err != nil {
		return fmt.Errorf("comp %q work_area: %w", src.Name, err)
	}
	return nil
}

func applyNoLayerCompMetadata(project *aep.Project, prof *profile.Profile) (*aep.Project, error) {
	if !hasCompMetadata(prof) {
		return project, nil
	}
	var buf bytes.Buffer
	if err := project.WriteAEP(&buf); err != nil {
		return nil, fmt.Errorf("write metadata base: %w", err)
	}
	reopened, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("reopen metadata base: %w", err)
	}
	for i, src := range prof.Comps {
		if src.Label == 0 && src.Comment == "" {
			continue
		}
		if i >= len(reopened.Compositions) {
			return nil, fmt.Errorf("comp %q metadata: reopened project has %d comps, want index %d", src.Name, len(reopened.Compositions), i)
		}
		dst := reopened.Compositions[i]
		if src.Label != 0 {
			if err := dst.SetLabel(src.Label); err != nil {
				return nil, fmt.Errorf("comp %q label: %w", src.Name, err)
			}
		}
		if src.Comment != "" {
			if err := dst.SetComment(src.Comment); err != nil {
				return nil, fmt.Errorf("comp %q comment: %w", src.Name, err)
			}
		}
	}
	return reopened, nil
}

func hasCompMetadata(prof *profile.Profile) bool {
	if prof == nil {
		return false
	}
	for _, comp := range prof.Comps {
		if comp.Label != 0 || comp.Comment != "" {
			return true
		}
	}
	return false
}

func materializeProjectTransformExpressions(project *aep.Project, prof *profile.Profile) (*aep.Project, error) {
	if !hasProfileTransformExpressions(prof) {
		return project, nil
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("reopen for transform expressions: %w", err)
	}
	for compIndex, sourceComp := range prof.Comps {
		targetComp, err := targetCompBySource(reopened, compIndex, sourceComp)
		if err != nil {
			return nil, err
		}
		for layerIndex, sourceLayer := range sourceComp.Layers {
			if !hasLayerTransformExpressions(sourceLayer) {
				continue
			}
			if layerIndex >= len(targetComp.Layers) {
				return nil, fmt.Errorf("comp %q transform expressions: target layer index %d missing", sourceComp.Name, layerIndex)
			}
			if err := materializeLayerTransformExpressions(targetComp.Layers[layerIndex], sourceLayer); err != nil {
				return nil, fmt.Errorf("comp %q layer %q transform expressions: %w", sourceComp.Name, sourceLayer.Name, err)
			}
		}
	}
	return reopened, nil
}

func hasProfileTransformExpressions(prof *profile.Profile) bool {
	if prof == nil {
		return false
	}
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if hasLayerTransformExpressions(layer) {
				return true
			}
		}
	}
	return false
}

func hasLayerTransformExpressions(layer profile.Layer) bool {
	for _, property := range layer.Properties {
		if isTransformProperty(property.MatchName) && property.Expression != "" {
			return true
		}
		if isTransformProperty(property.MatchName) && property.ExpressionEnabled != nil {
			return true
		}
	}
	return false
}

func materializeLayerTransformExpressions(targetLayer *aep.Layer, sourceLayer profile.Layer) error {
	for _, property := range sourceLayer.Properties {
		if !isTransformProperty(property.MatchName) || !hasPropertyExpression(property) {
			continue
		}
		targetProperty := targetTransformProperty(targetLayer, property.MatchName)
		if targetProperty == nil {
			return fmt.Errorf("property %q not found", property.MatchName)
		}
		if property.Expression != "" {
			if err := targetProperty.SetExpression(property.Expression); err != nil {
				return fmt.Errorf("property %q expression source: %w", property.MatchName, err)
			}
		}
		if property.ExpressionEnabled != nil {
			if err := targetProperty.SetExpressionEnabled(*property.ExpressionEnabled); err != nil {
				return fmt.Errorf("property %q expression enabled: %w", property.MatchName, err)
			}
		}
	}
	return nil
}

func hasPropertyExpression(property profile.Property) bool {
	return property.Expression != "" || property.ExpressionEnabled != nil
}

func isTransformProperty(matchName string) bool {
	switch matchName {
	case "ADBE Anchor Point", "ADBE Position", "ADBE Scale", "ADBE Rotate Z", "ADBE Opacity":
		return true
	default:
		return false
	}
}

func targetTransformProperty(layer *aep.Layer, matchName string) *aep.Property {
	switch matchName {
	case "ADBE Anchor Point":
		return layer.AnchorPoint()
	case "ADBE Position":
		return layer.Position()
	case "ADBE Scale":
		return layer.Scale()
	case "ADBE Rotate Z":
		return layer.Rotation()
	case "ADBE Opacity":
		return layer.Opacity()
	default:
		return nil
	}
}

func materializeProjectTextAnimators(project *aep.Project, prof *profile.Profile) (*aep.Project, error) {
	if !hasProfileTextAnimators(prof) {
		return project, nil
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("reopen for text animators: %w", err)
	}
	for compIndex, sourceComp := range prof.Comps {
		targetComp, err := targetCompBySource(reopened, compIndex, sourceComp)
		if err != nil {
			return nil, err
		}
		for layerIndex, sourceLayer := range sourceComp.Layers {
			if !hasLayerTextAnimators(sourceLayer) {
				continue
			}
			if layerIndex >= len(targetComp.Layers) {
				return nil, fmt.Errorf("comp %q text animators: target layer index %d missing", sourceComp.Name, layerIndex)
			}
			if err := materializeLayerTextAnimators(targetComp.Layers[layerIndex], sourceLayer); err != nil {
				return nil, fmt.Errorf("comp %q layer %q text animators: %w", sourceComp.Name, sourceLayer.Name, err)
			}
		}
	}
	return reopened, nil
}

func hasProfileTextAnimators(prof *profile.Profile) bool {
	if prof == nil {
		return false
	}
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if hasLayerTextAnimators(layer) {
				return true
			}
		}
	}
	return false
}

func hasLayerTextAnimators(layer profile.Layer) bool {
	if layer.Type != "text" {
		return false
	}
	for _, property := range layer.Properties {
		if isTextAnimatorValueProperty(property.MatchName) {
			return true
		}
	}
	return false
}

func isTextAnimatorValueProperty(matchName string) bool {
	switch matchName {
	case "ADBE Text Opacity",
		"ADBE Text Position 3D",
		"ADBE Text Scale 3D",
		"ADBE Text Rotation",
		"ADBE Text Fill Color",
		"ADBE Text Stroke Color",
		"ADBE Text Tracking Amount",
		"ADBE Text Character Offset",
		"ADBE Text Fill Opacity",
		"ADBE Text Stroke Opacity",
		"ADBE Text Stroke Width",
		"ADBE Text Skew",
		"ADBE Text Rotation X",
		"ADBE Text Rotation Y":
		return true
	default:
		return false
	}
}

func materializeLayerTextAnimators(targetLayer *aep.Layer, sourceLayer profile.Layer) error {
	rangeValues, err := textAnimatorRangeValues(sourceLayer)
	if err != nil {
		return err
	}
	rangeOffsetAnimated := false
	for _, property := range sourceLayer.Properties {
		if !isTextAnimatorValueProperty(property.MatchName) {
			continue
		}
		if isImplicitTextRotationZProperty(sourceLayer, property) {
			continue
		}
		if err := materializeTextAnimator(targetLayer, property, rangeValues); err != nil {
			return fmt.Errorf("property %q: %w", property.MatchName, err)
		}
		if !rangeOffsetAnimated && len(rangeValues.offset.Keyframes) != 0 {
			if err := materializeTextRangeOffsetKeyframes(targetLayer, rangeValues.offset); err != nil {
				return fmt.Errorf("range offset keyframes: %w", err)
			}
			rangeOffsetAnimated = true
		}
		if len(property.Keyframes) != 0 {
			if err := materializeTextAnimatorValueKeyframes(targetLayer, property); err != nil {
				return fmt.Errorf("property %q keyframes: %w", property.MatchName, err)
			}
		}
	}
	return nil
}

func isImplicitTextRotationZProperty(layer profile.Layer, property profile.Property) bool {
	if property.MatchName != "ADBE Text Rotation" || len(property.Keyframes) != 0 || hasPropertyExpression(property) {
		return false
	}
	value, ok := propertyFloatValue(property.StaticValue)
	if !ok || value != 0 {
		return false
	}
	return hasProperty(layer, "ADBE Text Rotation X") || hasProperty(layer, "ADBE Text Rotation Y")
}

type textAnimatorRange struct {
	start  float64
	end    float64
	offset profile.Property
	value  float64
}

func textAnimatorRangeValues(layer profile.Layer) (textAnimatorRange, error) {
	startProperty, ok := propertyByMatchName(layer.Properties, "ADBE Text Percent Start")
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range start property missing")
	}
	start, ok := textAnimatorInitialFloatValue(startProperty)
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range start value unsupported (%T)", startProperty.StaticValue)
	}
	endProperty, ok := propertyByMatchName(layer.Properties, "ADBE Text Percent End")
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range end property missing")
	}
	end, ok := textAnimatorInitialFloatValue(endProperty)
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range end value unsupported (%T)", endProperty.StaticValue)
	}
	offsetProperty, ok := propertyByMatchName(layer.Properties, "ADBE Text Percent Offset")
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range offset property missing")
	}
	offset, ok := textAnimatorInitialFloatValue(offsetProperty)
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range offset value unsupported (%T)", offsetProperty.StaticValue)
	}
	return textAnimatorRange{
		start:  start,
		end:    end,
		offset: offsetProperty,
		value:  offset,
	}, nil
}

func materializeTextAnimator(targetLayer *aep.Layer, property profile.Property, rangeValues textAnimatorRange) error {
	switch property.MatchName {
	case "ADBE Text Opacity":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextOpacityAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Position 3D":
		value, ok := textAnimatorInitialVectorValue(property, 3)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextPositionAnimator(targetLayer, value[0], value[1], value[2], rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Scale 3D":
		value, ok := textAnimatorInitialVectorValue(property, 3)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextScaleAnimator(targetLayer, value[0], value[1], value[2], rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Rotation":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextRotationAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Fill Color":
		value, ok := textAnimatorInitialColorValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextColorAnimator(targetLayer, value[0], value[1], value[2], value[3], rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Stroke Color":
		value, ok := textAnimatorInitialColorValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextStrokeColorAnimator(targetLayer, value[0], value[1], value[2], value[3], rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Tracking Amount":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextTrackingAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Character Offset":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextCharacterOffsetAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Fill Opacity":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextFillOpacityAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Stroke Opacity":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextStrokeOpacityAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Stroke Width":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextStrokeWidthAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Skew":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextSkewAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Rotation X":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextRotationXAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	case "ADBE Text Rotation Y":
		value, ok := textAnimatorInitialFloatValue(property)
		if !ok {
			return fmt.Errorf("value unsupported (%T)", property.StaticValue)
		}
		_, err := aep.AddTextRotationYAnimator(targetLayer, value, rangeValues.start, rangeValues.end, rangeValues.value)
		return err
	default:
		return fmt.Errorf("unsupported text animator property")
	}
}

func materializeTextRangeOffsetKeyframes(targetLayer *aep.Layer, property profile.Property) error {
	keyframes, err := textAnimatorScalarKeyframes(property)
	if err != nil {
		return err
	}
	return aep.AnimateTextRangeOffset(targetLayer, 0, keyframes)
}

func materializeTextAnimatorValueKeyframes(targetLayer *aep.Layer, property profile.Property) error {
	switch property.MatchName {
	case "ADBE Text Opacity":
		keyframes, err := textAnimatorScalarKeyframes(property)
		if err != nil {
			return err
		}
		return aep.AnimateTextOpacity(targetLayer, 0, keyframes)
	case "ADBE Text Position 3D":
		keyframes, err := textAnimatorVectorKeyframes(property, 3)
		if err != nil {
			return err
		}
		return aep.AnimateTextPosition(targetLayer, 0, keyframes)
	case "ADBE Text Scale 3D":
		keyframes, err := textAnimatorVectorKeyframes(property, 3)
		if err != nil {
			return err
		}
		return aep.AnimateTextScale(targetLayer, 0, keyframes)
	case "ADBE Text Rotation":
		keyframes, err := textAnimatorScalarKeyframes(property)
		if err != nil {
			return err
		}
		return aep.AnimateTextRotation(targetLayer, 0, keyframes)
	case "ADBE Text Fill Color":
		keyframes, err := textAnimatorColorKeyframes(property)
		if err != nil {
			return err
		}
		return aep.AnimateTextColor(targetLayer, 0, keyframes)
	case "ADBE Text Tracking Amount":
		keyframes, err := textAnimatorScalarKeyframes(property)
		if err != nil {
			return err
		}
		return aep.AnimateTextTracking(targetLayer, 0, keyframes)
	case "ADBE Text Character Offset":
		keyframes, err := textAnimatorScalarKeyframes(property)
		if err != nil {
			return err
		}
		return aep.AnimateTextCharacterOffset(targetLayer, 0, keyframes)
	default:
		return fmt.Errorf("value keyframes are not supported")
	}
}

func textAnimatorInitialFloatValue(property profile.Property) (float64, bool) {
	if len(property.Keyframes) != 0 {
		return propertyFloatValue(property.Keyframes[0].Value)
	}
	return propertyFloatValue(property.StaticValue)
}

func textAnimatorInitialVectorValue(property profile.Property, length int) ([]float64, bool) {
	if len(property.Keyframes) != 0 {
		return staticVectorAtLeast(property.Keyframes[0].Value, length)
	}
	return staticVectorAtLeast(property.StaticValue, length)
}

func textAnimatorInitialColorValue(property profile.Property) ([4]float64, bool) {
	var value any = property.StaticValue
	if len(property.Keyframes) != 0 {
		value = property.Keyframes[0].Value
	}
	vector, ok := staticVectorAtLeast(value, 3)
	if !ok {
		return [4]float64{}, false
	}
	return profileTextColorToRGBA(vector), true
}

func textAnimatorScalarKeyframes(property profile.Property) ([]aep.ScalarKeyframe, error) {
	if len(property.Keyframes) < 2 {
		return nil, fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	out := make([]aep.ScalarKeyframe, 0, len(property.Keyframes))
	for i, keyframe := range property.Keyframes {
		value, ok := propertyFloatValue(keyframe.Value)
		if !ok {
			return nil, fmt.Errorf("keyframes[%d].value must be a number", i)
		}
		out = append(out, aep.ScalarKeyframe{
			Time:    keyframe.Time,
			Value:   value,
			InEase:  effectParamTemporalEase(keyframe.InTemporalEase),
			OutEase: effectParamTemporalEase(keyframe.OutTemporalEase),
		})
	}
	return out, nil
}

func textAnimatorVectorKeyframes(property profile.Property, length int) ([]aep.VectorKeyframe, error) {
	if len(property.Keyframes) < 2 {
		return nil, fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	out := make([]aep.VectorKeyframe, 0, len(property.Keyframes))
	for i, keyframe := range property.Keyframes {
		value, ok := staticVectorAtLeast(keyframe.Value, length)
		if !ok {
			return nil, fmt.Errorf("keyframes[%d].value must be a %d-number array", i, length)
		}
		out = append(out, aep.VectorKeyframe{
			Time:    keyframe.Time,
			Value:   append([]float64(nil), value[:length]...),
			InEase:  effectParamTemporalEase(keyframe.InTemporalEase),
			OutEase: effectParamTemporalEase(keyframe.OutTemporalEase),
		})
	}
	return out, nil
}

func textAnimatorColorKeyframes(property profile.Property) ([]aep.VectorKeyframe, error) {
	if len(property.Keyframes) < 2 {
		return nil, fmt.Errorf("need >= 2 keyframes, got %d", len(property.Keyframes))
	}
	out := make([]aep.VectorKeyframe, 0, len(property.Keyframes))
	for i, keyframe := range property.Keyframes {
		value, ok := staticVectorAtLeast(keyframe.Value, 3)
		if !ok {
			return nil, fmt.Errorf("keyframes[%d].value must be a color array", i)
		}
		color := profileTextColorToRGBA(value)
		out = append(out, aep.VectorKeyframe{
			Time:    keyframe.Time,
			Value:   []float64{color[0], color[1], color[2], color[3]},
			InEase:  effectParamTemporalEase(keyframe.InTemporalEase),
			OutEase: effectParamTemporalEase(keyframe.OutTemporalEase),
		})
	}
	return out, nil
}

func profileTextColorToRGBA(value []float64) [4]float64 {
	if len(value) >= 4 {
		return profileARGBToRGBA(value[:4])
	}
	return [4]float64{
		colorByteToUnit(value[0]),
		colorByteToUnit(value[1]),
		colorByteToUnit(value[2]),
		1,
	}
}

func materializeProjectEffects(project *aep.Project, prof *profile.Profile) (*aep.Project, error) {
	if !hasProfileEffects(prof) {
		return project, nil
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("reopen for effects: %w", err)
	}
	for compIndex, sourceComp := range prof.Comps {
		targetComp, err := targetCompBySource(reopened, compIndex, sourceComp)
		if err != nil {
			return nil, err
		}
		for layerIndex, sourceLayer := range sourceComp.Layers {
			if len(sourceLayer.Effects) == 0 {
				continue
			}
			if layerIndex >= len(targetComp.Layers) {
				return nil, fmt.Errorf("comp %q effects: target layer index %d missing", sourceComp.Name, layerIndex)
			}
			targetLayer := targetComp.Layers[layerIndex]
			for _, sourceEffect := range sourceLayer.Effects {
				targetEffect, err := aep.AddEffect(targetLayer, sourceEffect.MatchName)
				if err != nil {
					return nil, fmt.Errorf("comp %q layer %q add effect %q: %w", sourceComp.Name, sourceLayer.Name, sourceEffect.MatchName, err)
				}
				if err := materializeEffectParams(targetComp, targetLayer, sourceLayer, targetEffect, sourceEffect); err != nil {
					return nil, fmt.Errorf("comp %q layer %q effect %q params: %w", sourceComp.Name, sourceLayer.Name, sourceEffect.MatchName, err)
				}
			}
		}
	}
	return reopened, nil
}

func materializeEffectParams(targetComp *aep.Composition, targetLayer *aep.Layer, sourceLayer profile.Layer, targetEffect *aep.Effect, sourceEffect profile.Effect) error {
	for _, param := range sourceEffect.Params {
		var property *aep.Property
		if param.LayerRef != nil && !effectParamLayerRefIsSelf(param.LayerRef, sourceLayer) {
			target := targetLayerBySourceRef(targetComp, param.LayerRef)
			if target == nil {
				return fmt.Errorf("param %q target layer %q not found", param.MatchName, param.LayerRef.Name)
			}
			if err := aep.SetEffectLayerParam(targetLayer, targetEffect, param.MatchName, target); err != nil {
				return fmt.Errorf("param %q layer_ref: %w", param.MatchName, err)
			}
			continue
		}
		if len(param.Keyframes) != 0 {
			var err error
			property, err = materializeEffectParamKeyframes(targetLayer, targetEffect, param)
			if err != nil {
				return fmt.Errorf("param %q keyframes: %w", param.MatchName, err)
			}
		} else {
			value, ok := effectParamStaticValue(param.StaticValue)
			if !ok {
				if param.Changed {
					return fmt.Errorf("param %q static value unsupported (%T)", param.MatchName, param.StaticValue)
				}
			} else {
				var err error
				property, err = aep.SetEffectParam(targetLayer, targetEffect, param.MatchName, value)
				if err != nil {
					return fmt.Errorf("param %q static value: %w", param.MatchName, err)
				}
			}
		}
		if effectParamHasExpression(param) {
			if property == nil {
				return fmt.Errorf("param %q expression requires a materialized value/keyframes param", param.MatchName)
			}
			if param.Expression != "" {
				if err := property.SetExpression(param.Expression); err != nil {
					return fmt.Errorf("param %q expression source: %w", param.MatchName, err)
				}
			}
			if param.ExpressionEnabled != nil {
				if err := property.SetExpressionEnabled(*param.ExpressionEnabled); err != nil {
					return fmt.Errorf("param %q expression enabled: %w", param.MatchName, err)
				}
			}
		}
	}
	return nil
}

func materializeEffectParamKeyframes(layer *aep.Layer, effect *aep.Effect, param profile.Property) (*aep.Property, error) {
	if _, ok := effectParamNumber(param.Keyframes[0].Value); ok {
		keyframes, err := effectParamScalarKeyframes(param.Keyframes)
		if err != nil {
			return nil, err
		}
		return aep.AnimateEffectParam(layer, effect, param.MatchName, keyframes)
	}
	keyframes, err := effectParamVectorKeyframes(param.Keyframes)
	if err != nil {
		return nil, err
	}
	return aep.AnimateEffectParamVec(layer, effect, param.MatchName, keyframes)
}

func hasProfileEffects(prof *profile.Profile) bool {
	if prof == nil {
		return false
	}
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if len(layer.Effects) != 0 {
				return true
			}
		}
	}
	return false
}

func isSupportedStaticEffectSurface(layer profile.Layer) bool {
	for _, effect := range layer.Effects {
		if !isSupportedEffectMatchName(effect.MatchName) {
			return false
		}
		for _, param := range effect.Params {
			if param.LayerRef != nil {
				if len(param.Keyframes) != 0 || effectParamHasExpression(param) {
					return false
				}
				continue
			}
			if len(param.Keyframes) != 0 {
				if !isSupportedEffectParamKeyframes(param.Keyframes) {
					return false
				}
				continue
			}
			if param.StaticValue != nil {
				_, ok := effectParamStaticValue(param.StaticValue)
				if !ok {
					return false
				}
				continue
			}
			if effectParamHasExpression(param) || param.Changed {
				return false
			}
		}
	}
	return true
}

func effectParamHasExpression(param profile.Property) bool {
	return param.Expression != "" || param.ExpressionEnabled != nil
}

func isSupportedEffectParamKeyframes(keyframes []profile.Keyframe) bool {
	if len(keyframes) < 2 {
		return false
	}
	if _, ok := effectParamNumber(keyframes[0].Value); ok {
		_, err := effectParamScalarKeyframes(keyframes)
		return err == nil
	}
	_, err := effectParamVectorKeyframes(keyframes)
	return err == nil
}

func isSupportedEffectMatchName(matchName string) bool {
	for _, supported := range aep.SupportedEffects() {
		if supported == matchName {
			return true
		}
	}
	return false
}

func effectParamStaticValue(value any) (any, bool) {
	switch v := value.(type) {
	case nil:
		return nil, false
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case bool:
		if v {
			return 1.0, true
		}
		return 0.0, true
	case []float64:
		return append([]float64(nil), v...), true
	case []any:
		out := make([]float64, 0, len(v))
		for _, item := range v {
			n, ok := effectParamNumber(item)
			if !ok {
				return nil, false
			}
			out = append(out, n)
		}
		return out, true
	default:
		return nil, false
	}
}

func effectParamNumber(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func effectParamScalarKeyframes(in []profile.Keyframe) ([]aep.ScalarKeyframe, error) {
	out := make([]aep.ScalarKeyframe, 0, len(in))
	for i, kf := range in {
		value, ok := effectParamNumber(kf.Value)
		if !ok {
			return nil, fmt.Errorf("keyframes[%d].value must be a number", i)
		}
		out = append(out, aep.ScalarKeyframe{
			Time:    kf.Time,
			Value:   value,
			InEase:  effectParamTemporalEase(kf.InTemporalEase),
			OutEase: effectParamTemporalEase(kf.OutTemporalEase),
		})
	}
	return out, nil
}

func effectParamVectorKeyframes(in []profile.Keyframe) ([]aep.VectorKeyframe, error) {
	out := make([]aep.VectorKeyframe, 0, len(in))
	for i, kf := range in {
		value, ok := effectParamVectorValue(kf.Value)
		if !ok || len(value) < 2 || len(value) > 4 {
			return nil, fmt.Errorf("keyframes[%d].value must be a 2-, 3-, or 4-number array", i)
		}
		out = append(out, aep.VectorKeyframe{
			Time:    kf.Time,
			Value:   value,
			InEase:  effectParamTemporalEase(kf.InTemporalEase),
			OutEase: effectParamTemporalEase(kf.OutTemporalEase),
		})
	}
	return out, nil
}

func effectParamVectorValue(value any) ([]float64, bool) {
	switch v := value.(type) {
	case []float64:
		return append([]float64(nil), v...), true
	case []any:
		out := make([]float64, 0, len(v))
		for _, item := range v {
			n, ok := effectParamNumber(item)
			if !ok {
				return nil, false
			}
			out = append(out, n)
		}
		return out, true
	default:
		return nil, false
	}
}

func effectParamTemporalEase(in []profile.TemporalEase) aep.TemporalEase {
	if len(in) == 0 {
		return aep.TemporalEase{}
	}
	return aep.TemporalEase{Speed: in[0].Speed, Influence: in[0].Influence}
}

func effectParamLayerRefIsSelf(ref *profile.LayerRef, layer profile.Layer) bool {
	if ref == nil {
		return false
	}
	if ref.Name != "" && ref.Name == layer.Name {
		return true
	}
	if ref.ID != 0 && ref.ID == layer.ID {
		return true
	}
	if ref.Index != 0 && ref.Index == layer.Index {
		return true
	}
	return false
}

func targetCompBySource(project *aep.Project, index int, source profile.Composition) (*aep.Composition, error) {
	if index < len(project.Compositions) {
		return project.Compositions[index], nil
	}
	for _, comp := range project.Compositions {
		if comp.Name == source.Name {
			return comp, nil
		}
	}
	return nil, fmt.Errorf("comp %q effects: target comp missing", source.Name)
}

func targetLayerBySourceRef(comp *aep.Composition, ref *profile.LayerRef) *aep.Layer {
	if ref == nil {
		return nil
	}
	if ref.Name != "" {
		if layer := comp.LayerByName(ref.Name); layer != nil {
			return layer
		}
	}
	if ref.Index >= 0 && ref.Index < len(comp.Layers) {
		return comp.Layers[ref.Index]
	}
	if ref.Index > 0 && ref.Index <= len(comp.Layers) {
		return comp.Layers[ref.Index-1]
	}
	return nil
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
