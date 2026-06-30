package recipe

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/projectindex"
)

func CompileToFile(rec Recipe, outPath string, caps CapabilityIndex) (Report, error) {
	report := ValidateWithCapabilities(rec, caps)
	report.OutputPath = outPath
	if !report.Valid {
		return report, nil
	}
	if outPath == "" {
		report.Valid = false
		report.Refusals = append(report.Refusals, Refusal{Code: "missing_output_path", Path: "output_path", Message: "output path is required"})
		return report, nil
	}

	projectTarget, err := parseRecipeProjectTarget(rec.Project.TargetVersion)
	if err != nil {
		return report, nil
	}
	project := aep.NewProject(projectTarget.aepTarget())
	compSpec := rec.Comps[0]
	comp, err := aep.NewComposition(project, compSpec.Name, uint16(compSpec.Width), uint16(compSpec.Height), compSpec.FrameRate, compSpec.Duration)
	if err != nil {
		return report, fmt.Errorf("recipe: create comp: %w", err)
	}
	if len(compSpec.BackgroundColor) == 3 {
		if err := comp.SetBGColor(rgb8Color(compSpec.BackgroundColor)); err != nil {
			return report, fmt.Errorf("recipe: comp %q background_color: %w", compSpec.Name, err)
		}
	}
	if compSpec.Renderer != "" {
		if err := aep.SetRenderer(comp, compSpec.Renderer); err != nil {
			return report, fmt.Errorf("recipe: comp %q renderer: %w", compSpec.Name, err)
		}
	}
	if len(compSpec.ResolutionFactor) == 2 {
		if err := comp.SetResolutionFactor(uint16(compSpec.ResolutionFactor[0]), uint16(compSpec.ResolutionFactor[1])); err != nil {
			return report, fmt.Errorf("recipe: comp %q resolution_factor: %w", compSpec.Name, err)
		}
	}
	if compSpec.PixelAspect != nil {
		if err := comp.SetPixelAspect(*compSpec.PixelAspect); err != nil {
			return report, fmt.Errorf("recipe: comp %q pixel_aspect: %w", compSpec.Name, err)
		}
	}
	if compSpec.DisplayStartTime != nil {
		if err := comp.SetDisplayStartTime(*compSpec.DisplayStartTime); err != nil {
			return report, fmt.Errorf("recipe: comp %q display_start_time: %w", compSpec.Name, err)
		}
	}
	if compSpec.FrameBlending != nil {
		if err := comp.SetFrameBlending(*compSpec.FrameBlending); err != nil {
			return report, fmt.Errorf("recipe: comp %q frame_blending: %w", compSpec.Name, err)
		}
	}
	if compSpec.Draft3D != nil {
		if err := comp.SetDraft3D(*compSpec.Draft3D); err != nil {
			return report, fmt.Errorf("recipe: comp %q draft_3d: %w", compSpec.Name, err)
		}
	}
	if compSpec.HideShyLayers != nil {
		if err := comp.SetHideShyLayers(*compSpec.HideShyLayers); err != nil {
			return report, fmt.Errorf("recipe: comp %q hide_shy_layers: %w", compSpec.Name, err)
		}
	}
	if compSpec.PreserveNestedFrameRate != nil {
		if err := comp.SetPreserveNestedFrameRate(*compSpec.PreserveNestedFrameRate); err != nil {
			return report, fmt.Errorf("recipe: comp %q preserve_nested_frame_rate: %w", compSpec.Name, err)
		}
	}
	if compSpec.PreserveNestedResolution != nil {
		if err := comp.SetPreserveNestedResolution(*compSpec.PreserveNestedResolution); err != nil {
			return report, fmt.Errorf("recipe: comp %q preserve_nested_resolution: %w", compSpec.Name, err)
		}
	}
	if compSpec.MotionBlur != nil {
		if err := applyCompMotionBlur(comp, compSpec.MotionBlur); err != nil {
			return report, fmt.Errorf("recipe: comp %q motion_blur: %w", compSpec.Name, err)
		}
	}
	if compSpec.WorkArea != nil {
		if err := comp.SetWorkArea(*compSpec.WorkArea.Start, *compSpec.WorkArea.End); err != nil {
			return report, fmt.Errorf("recipe: comp %q work_area: %w", compSpec.Name, err)
		}
	}
	for _, layerSpec := range compSpec.Layers {
		if _, err := compileLayer(comp, layerSpec, compSpec); err != nil {
			return report, err
		}
	}
	idx := projectindex.Build(project)
	if err := applyLayerParents(compSpec, idx); err != nil {
		return report, err
	}
	if err := applyExplicitMattes(compSpec, idx); err != nil {
		return report, err
	}
	if err := applyLightSources(compSpec, idx); err != nil {
		return report, err
	}
	if hasMasks(compSpec) {
		project, err = materializeMasks(project, compSpec)
		if err != nil {
			return report, err
		}
	}
	if hasEffects(compSpec) {
		project, err = materializeEffects(project, compSpec)
		if err != nil {
			return report, err
		}
	}
	if hasTransformExpressions(compSpec) {
		project, err = materializeTransformExpressions(project, compSpec)
		if err != nil {
			return report, err
		}
	}
	project, err = applyCompItemSettings(project, compSpec)
	if err != nil {
		return report, fmt.Errorf("recipe: comp %q item settings: %w", compSpec.Name, err)
	}
	if hasExpectedProfile(rec.ExpectedProfile) {
		prof, err := buildWrittenProfile(project, outPath)
		if err != nil {
			return report, fmt.Errorf("recipe: build profile for expected_profile: %w", err)
		}
		report.ProfileChecks = checkExpectedProfile(rec.ExpectedProfile, prof)
		for _, check := range report.ProfileChecks {
			if !check.Passed {
				report.Valid = false
				report.Refusals = append(report.Refusals, Refusal{
					Code:    "profile_contract_mismatch",
					Path:    check.Path,
					Message: check.Message,
				})
			}
		}
		if !report.Valid {
			return report, nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return report, err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return report, err
	}
	defer out.Close()
	if err := project.WriteAEP(out); err != nil {
		return report, err
	}
	return report, nil
}

func buildWrittenProfile(project *aep.Project, outPath string) (*profile.Profile, error) {
	var buf bytes.Buffer
	if err := project.WriteAEP(&buf); err != nil {
		return nil, err
	}
	reopened, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, err
	}
	return profile.Build(reopened, profile.Options{Path: outPath})
}
