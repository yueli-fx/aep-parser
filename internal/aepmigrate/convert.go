package aepmigrate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aehost"
	"github.com/yueli-fx/aep-parser/internal/aep"
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
	for _, comp := range prof.Comps {
		entries = append(entries, Entry{
			Path:          "comps[" + comp.Name + "]",
			Class:         ClassRetargeted,
			TargetVersion: target,
			Reason:        "Composition and stable composition settings are recreated through the target AE project template.",
		})
		for _, layer := range comp.Layers {
			if isSupportedDefaultNullLayer(layer) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default null layer is recreated through the target AE project template.",
				})
				continue
			}
			entries = append(entries, Entry{
				Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
				Class:         ClassBlocked,
				TargetVersion: target,
				Reason:        "This convert slice only reconstructs no-layer comps and default null layers; refusing output to avoid silent layer loss.",
			})
		}
	}
	return entries
}

func rebuildProject(target VersionLabel, prof *profile.Profile) (*aep.Project, error) {
	project := aep.NewProject(aepTarget(target))
	for _, comp := range prof.Comps {
		next, err := aep.NewComposition(project, comp.Name, comp.Width, comp.Height, comp.FrameRate, comp.Duration)
		if err != nil {
			return nil, err
		}
		if err := applyStableCompSettings(next, comp); err != nil {
			return nil, err
		}
		for _, layer := range comp.Layers {
			if !isSupportedDefaultNullLayer(layer) {
				return nil, fmt.Errorf("unsupported layer %q in comp %q", layer.Name, comp.Name)
			}
			dstLayer, err := aep.NewNullLayer(next, layer.Name)
			if err != nil {
				return nil, fmt.Errorf("comp %q null layer %q: %w", comp.Name, layer.Name, err)
			}
			if hasTransformProperties(layer) {
				if err := aep.SetLayerTransform(dstLayer, aep.NewLayerTransform()); err != nil {
					return nil, fmt.Errorf("comp %q null layer %q transform: %w", comp.Name, layer.Name, err)
				}
			}
		}
	}
	return project, nil
}

func hasTransformProperties(layer profile.Layer) bool {
	for _, property := range layer.Properties {
		switch property.MatchName {
		case "ADBE Anchor Point", "ADBE Position", "ADBE Scale", "ADBE Rotate Z", "ADBE Opacity":
			return true
		}
	}
	return false
}

func isSupportedDefaultNullLayer(layer profile.Layer) bool {
	if layer.Type != "null" {
		return false
	}
	if layer.Comment != "" || layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if layer.Text != nil || len(layer.Effects) != 0 || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	flags := layer.Flags
	return flags.Visible &&
		flags.Blend == 2 &&
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
