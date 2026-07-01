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
			if isSupportedDefaultNullLayer(layer) {
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
					Reason:        "Single rect graphic shape layer is recreated from the stable profile shape properties.",
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
				Reason:        "This convert slice only reconstructs no-layer comps, default null layers, default solid layers, default adjustment layers, default camera layers, default light layers, default text layers, default empty shape layers, single rect fill/stroke shape layers, and default precomp layers; refusing output to avoid silent layer loss.",
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
		for _, layer := range comp.Layers {
			switch {
			case isSupportedDefaultNullLayer(layer):
				dstLayer, err := aep.NewNullLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q null layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q null layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q null layer %q timing: %w", comp.Name, layer.Name, err)
				}
			case isSupportedDefaultSolidLayer(layer, footage):
				solid, _ := footage.solidDetails(layer)
				dstLayer, err := aep.NewSolidLayer(next, layer.Name, int(solid.Width), int(solid.Height), *solid.SolidColor)
				if err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeCenteredTransformSurface(next, dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q solid layer %q timing: %w", comp.Name, layer.Name, err)
				}
			case isSupportedDefaultAdjustmentLayer(layer, footage):
				dstLayer, err := aep.NewAdjustmentLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q adjustment layer %q timing: %w", comp.Name, layer.Name, err)
				}
			case isSupportedDefaultCameraLayer(layer):
				dstLayer, err := aep.NewCameraLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeCameraLightTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q camera layer %q timing: %w", comp.Name, layer.Name, err)
				}
			case isSupportedDefaultLightLayer(layer):
				dstLayer, err := aep.NewLightLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q light layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeCameraLightTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q light layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q light layer %q timing: %w", comp.Name, layer.Name, err)
				}
			case isSupportedDefaultTextLayer(layer):
				dstLayer, err := aep.NewTextLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q text layer %q: %w", comp.Name, layer.Name, err)
				}
				if layer.Text != nil {
					if err := dstLayer.SetText(layer.Text.Text); err != nil {
						return nil, fmt.Errorf("comp %q text layer %q text: %w", comp.Name, layer.Name, err)
					}
				}
				if err := materializeDefaultTransformSurface(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q text layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q text layer %q timing: %w", comp.Name, layer.Name, err)
				}
			case isSupportedDefaultShapeLayer(layer):
				dstLayer, err := aep.NewShapeLayer(next, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q shape layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer.Layer, layer); err != nil {
					return nil, fmt.Errorf("comp %q shape layer %q timing: %w", comp.Name, layer.Name, err)
				}
			case isSupportedRectGraphicShapeLayer(layer):
				dstLayer, err := materializeRectGraphicShapeLayer(next, layer)
				if err != nil {
					return nil, fmt.Errorf("comp %q shape layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q shape layer %q timing: %w", comp.Name, layer.Name, err)
				}
			case isSupportedDefaultPrecompLayer(layer, comps):
				sourceComp, ok := targetComps.sourceComposition(layer)
				if !ok {
					return nil, fmt.Errorf("comp %q precomp layer %q source %q not found", comp.Name, layer.Name, layer.SourceRef.Name)
				}
				dstLayer, err := aep.NewPrecompLayer(next, sourceComp, layer.Name)
				if err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q: %w", comp.Name, layer.Name, err)
				}
				if err := materializePrecompTransformSurface(next, dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q transform: %w", comp.Name, layer.Name, err)
				}
				if err := materializeLayerTiming(dstLayer, layer); err != nil {
					return nil, fmt.Errorf("comp %q precomp layer %q timing: %w", comp.Name, layer.Name, err)
				}
			default:
				return nil, fmt.Errorf("unsupported layer %q in comp %q", layer.Name, comp.Name)
			}
		}
	}
	return project, nil
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
	rectShape := source.Shapes[0]
	rect, err := shapeLayer.RootGroup().AddRect()
	if err != nil {
		return nil, err
	}
	if value, ok := propertyVector(rectShape.Properties, "ADBE Vector Rect Size", 2); ok {
		if err := rect.SetSize([2]float64{value[0], value[1]}); err != nil {
			return nil, err
		}
	}
	if value, ok := propertyVector(rectShape.Properties, "ADBE Vector Rect Position", 2); ok {
		if err := rect.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return nil, err
		}
	}
	if value, ok := propertyFloat(rectShape.Properties, "ADBE Vector Rect Roundness"); ok {
		if err := rect.SetRoundness(value); err != nil {
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

func propertyVector(properties []profile.Property, matchName string, length int) ([]float64, bool) {
	for _, property := range properties {
		if property.MatchName != matchName {
			continue
		}
		return staticVector(property.StaticValue, length)
	}
	return nil, false
}

func propertyFloat(properties []profile.Property, matchName string) (float64, bool) {
	for _, property := range properties {
		if property.MatchName != matchName {
			continue
		}
		switch value := property.StaticValue.(type) {
		case float64:
			return value, true
		case int:
			return float64(value), true
		}
		return 0, false
	}
	return 0, false
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
	if value, ok := propertyVectorAtLeast(layer.Properties, "ADBE Anchor Point", 2); ok {
		if err := transform.AnchorPoint().SetStaticValue([2]float64{value[0], value[1]}); err != nil {
			return nil, err
		}
	}
	if value, ok := propertyVectorAtLeast(layer.Properties, "ADBE Position", 2); ok {
		if err := transform.Position().SetStaticValue([2]float64{value[0], value[1]}); err != nil {
			return nil, err
		}
	}
	if value, ok := propertyVectorAtLeast(layer.Properties, "ADBE Scale", 2); ok {
		if err := transform.Scale().SetStaticValue([2]float64{profileScaleToWriter(value[0]), profileScaleToWriter(value[1])}); err != nil {
			return nil, err
		}
	}
	if value, ok := propertyFloat(layer.Properties, "ADBE Rotate Z"); ok {
		if err := transform.Rotation().SetStaticValue(value); err != nil {
			return nil, err
		}
	}
	if value, ok := propertyFloat(layer.Properties, "ADBE Opacity"); ok {
		if err := transform.Opacity().SetStaticValue(profileOpacityToWriter(value)); err != nil {
			return nil, err
		}
	}
	return transform, nil
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

func isSupportedDefaultSolidLayer(layer profile.Layer, footage convertFootageIndex) bool {
	if layer.Type != "av" || layer.SourceRef == nil || layer.SourceRef.Kind != "footage" {
		return false
	}
	solid, ok := footage.solidDetails(layer)
	if !ok || solid.Width == 0 || solid.Height == 0 || solid.SolidColor == nil {
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
	if layer.Comment != "" || layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
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
	if layer.Comment != "" || layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
		return false
	}
	if len(layer.Effects) != 0 || len(layer.Masks) != 0 || len(layer.Shapes) != 0 || len(layer.Markers) != 0 {
		return false
	}
	if layer.Text.IsBoxText {
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

func isSupportedDefaultShapeLayer(layer profile.Layer) bool {
	if layer.Type != "shape" || layer.SourceRef != nil || layer.Text != nil {
		return false
	}
	if layer.Comment != "" || layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
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
	if len(layer.Shapes) != 1 || layer.Shapes[0].Kind != "rect" {
		return false
	}
	if _, ok := propertyVector(layer.Shapes[0].Properties, "ADBE Vector Rect Size", 2); !ok {
		return false
	}
	hasFill := hasProperty(layer, "ADBE Vector Fill Color")
	hasStroke := hasProperty(layer, "ADBE Vector Stroke Color")
	if !hasFill && !hasStroke {
		return false
	}
	if hasProperty(layer, "ADBE Vector Grad Colors") ||
		hasProperty(layer, "ADBE Vector Filter - Trim") ||
		hasProperty(layer, "ADBE Vector Stroke Dash 2") ||
		hasProperty(layer, "ADBE Vector Stroke Gap 2") ||
		hasProperty(layer, "ADBE Vector Stroke Offset") ||
		hasProperty(layer, "ADBE Vector Stroke Taper Start Length") ||
		hasProperty(layer, "ADBE Vector Stroke Wave Amount") {
		return false
	}
	return true
}

func isSupportedShapeLayerBase(layer profile.Layer) bool {
	if layer.Type != "shape" || layer.SourceRef != nil || layer.Text != nil {
		return false
	}
	if layer.Comment != "" || layer.ParentRef != nil || layer.MatteRef != nil || layer.LightSourceRef != nil {
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
