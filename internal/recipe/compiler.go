package recipe

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
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
	if rec.Project.BitsPerChannel != "" {
		bpc, err := projectBitsPerChannel(rec.Project.BitsPerChannel)
		if err != nil {
			return report, nil
		}
		if err := project.SetBitsPerChannel(bpc); err != nil {
			return report, fmt.Errorf("recipe: project bits_per_channel: %w", err)
		}
	}
	if rec.Project.LinearBlending != nil {
		if err := project.SetLinearBlending(*rec.Project.LinearBlending); err != nil {
			return report, fmt.Errorf("recipe: project linear_blending: %w", err)
		}
	}
	if rec.Project.LinearizeWorkingSpace != nil {
		if err := project.SetLinearizeWorkingSpace(*rec.Project.LinearizeWorkingSpace); err != nil {
			return report, fmt.Errorf("recipe: project linearize_working_space: %w", err)
		}
	}
	if rec.Project.TimeDisplayType != "" {
		v, err := projectTimeDisplayType(rec.Project.TimeDisplayType)
		if err != nil {
			return report, nil
		}
		if err := project.SetTimeDisplayType(v); err != nil {
			return report, fmt.Errorf("recipe: project time_display_type: %w", err)
		}
	}
	if rec.Project.FramesCountType != "" {
		v, err := projectFramesCountType(rec.Project.FramesCountType)
		if err != nil {
			return report, nil
		}
		if err := project.SetFramesCountType(v); err != nil {
			return report, fmt.Errorf("recipe: project frames_count_type: %w", err)
		}
	}
	if rec.Project.FramesUseFeetFrames != nil {
		if err := project.SetFramesUseFeetFrames(*rec.Project.FramesUseFeetFrames); err != nil {
			return report, fmt.Errorf("recipe: project frames_use_feet_frames: %w", err)
		}
	}
	if rec.Project.FeetFramesFilmType != "" {
		v, err := projectFeetFramesFilmType(rec.Project.FeetFramesFilmType)
		if err != nil {
			return report, nil
		}
		if err := project.SetFeetFramesFilmType(v); err != nil {
			return report, fmt.Errorf("recipe: project feet_frames_film_type: %w", err)
		}
	}
	if rec.Project.FootageTimecodeDisplayStartType != "" {
		v, err := projectFootageTimecodeDisplayStartType(rec.Project.FootageTimecodeDisplayStartType)
		if err != nil {
			return report, nil
		}
		if err := project.SetFootageTimecodeDisplayStartType(v); err != nil {
			return report, fmt.Errorf("recipe: project footage_timecode_display_start_type: %w", err)
		}
	}
	if rec.Project.ExpressionEngine != "" {
		v, err := projectExpressionEngine(rec.Project.ExpressionEngine)
		if err != nil {
			return report, nil
		}
		if err := project.SetExpressionEngine(v); err != nil {
			return report, fmt.Errorf("recipe: project expression_engine: %w", err)
		}
	}
	if rec.Project.AudioSampleRate != nil {
		if err := project.SetAudioSampleRate(*rec.Project.AudioSampleRate); err != nil {
			return report, fmt.Errorf("recipe: project audio_sample_rate: %w", err)
		}
	}
	if rec.Project.WorkingGamma != nil {
		if err := project.SetWorkingGamma(*rec.Project.WorkingGamma); err != nil {
			return report, fmt.Errorf("recipe: project working_gamma: %w", err)
		}
	}
	if rec.Project.CompensateForSceneReferredProfiles != nil {
		if err := project.SetCompensateForSceneReferredProfiles(*rec.Project.CompensateForSceneReferredProfiles); err != nil {
			return report, fmt.Errorf("recipe: project compensate_for_scene_referred_profiles: %w", err)
		}
	}
	if rec.Project.TimecodeDefaultBase != nil {
		if err := project.SetTimecodeDefaultBase(*rec.Project.TimecodeDefaultBase); err != nil {
			return report, fmt.Errorf("recipe: project timecode_default_base: %w", err)
		}
	}
	if rec.Project.TransparencyGridThumbnails != nil {
		if err := project.SetTransparencyGridThumbnails(*rec.Project.TransparencyGridThumbnails); err != nil {
			return report, fmt.Errorf("recipe: project transparency_grid_thumbnails: %w", err)
		}
	}
	compsByName := map[string]*aep.Composition{}
	for _, compSpec := range rec.Comps {
		comp, err := aep.NewComposition(project, compSpec.Name, uint16(compSpec.Width), uint16(compSpec.Height), compSpec.FrameRate, compSpec.Duration)
		if err != nil {
			return report, fmt.Errorf("recipe: create comp %q: %w", compSpec.Name, err)
		}
		compsByName[compSpec.Name] = comp
	}
	for _, compSpec := range rec.Comps {
		comp := compsByName[compSpec.Name]
		if err := applyCompSettings(comp, compSpec); err != nil {
			return report, fmt.Errorf("recipe: comp %q: %w", compSpec.Name, err)
		}
		for _, layerSpec := range compSpec.Layers {
			if _, err := compileLayerWithSources(comp, layerSpec, compSpec, compsByName); err != nil {
				return report, err
			}
		}
	}
	for _, compSpec := range rec.Comps {
		comp := compsByName[compSpec.Name]
		if err := applyLayerParents(comp, compSpec); err != nil {
			return report, err
		}
		if err := applyExplicitMattes(comp, compSpec); err != nil {
			return report, err
		}
		if err := applyLightSources(comp, compSpec); err != nil {
			return report, err
		}
	}
	for _, compSpec := range rec.Comps {
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

func applyCompSettings(comp *aep.Composition, compSpec CompSpec) error {
	if len(compSpec.BackgroundColor) == 3 {
		if err := comp.SetBGColor(rgb8Color(compSpec.BackgroundColor)); err != nil {
			return fmt.Errorf("background_color: %w", err)
		}
	}
	if compSpec.MotionGraphicsTemplateName != "" {
		if err := comp.SetMotionGraphicsTemplateName(compSpec.MotionGraphicsTemplateName); err != nil {
			return fmt.Errorf("motion_graphics_template_name: %w", err)
		}
	}
	if compSpec.Renderer != "" {
		if err := aep.SetRenderer(comp, compSpec.Renderer); err != nil {
			return fmt.Errorf("renderer: %w", err)
		}
	}
	if len(compSpec.ResolutionFactor) == 2 {
		if err := comp.SetResolutionFactor(uint16(compSpec.ResolutionFactor[0]), uint16(compSpec.ResolutionFactor[1])); err != nil {
			return fmt.Errorf("resolution_factor: %w", err)
		}
	}
	if compSpec.PixelAspect != nil {
		if err := comp.SetPixelAspect(*compSpec.PixelAspect); err != nil {
			return fmt.Errorf("pixel_aspect: %w", err)
		}
	}
	if compSpec.DisplayStartTime != nil {
		if err := comp.SetDisplayStartTime(*compSpec.DisplayStartTime); err != nil {
			return fmt.Errorf("display_start_time: %w", err)
		}
	}
	if compSpec.FrameBlending != nil {
		if err := comp.SetFrameBlending(*compSpec.FrameBlending); err != nil {
			return fmt.Errorf("frame_blending: %w", err)
		}
	}
	if compSpec.Draft3D != nil {
		if err := comp.SetDraft3D(*compSpec.Draft3D); err != nil {
			return fmt.Errorf("draft_3d: %w", err)
		}
	}
	if compSpec.HideShyLayers != nil {
		if err := comp.SetHideShyLayers(*compSpec.HideShyLayers); err != nil {
			return fmt.Errorf("hide_shy_layers: %w", err)
		}
	}
	if compSpec.PreserveNestedFrameRate != nil {
		if err := comp.SetPreserveNestedFrameRate(*compSpec.PreserveNestedFrameRate); err != nil {
			return fmt.Errorf("preserve_nested_frame_rate: %w", err)
		}
	}
	if compSpec.PreserveNestedResolution != nil {
		if err := comp.SetPreserveNestedResolution(*compSpec.PreserveNestedResolution); err != nil {
			return fmt.Errorf("preserve_nested_resolution: %w", err)
		}
	}
	if compSpec.MotionBlur != nil {
		if err := applyCompMotionBlur(comp, compSpec.MotionBlur); err != nil {
			return fmt.Errorf("motion_blur: %w", err)
		}
	}
	if compSpec.WorkArea != nil {
		if err := comp.SetWorkArea(*compSpec.WorkArea.Start, *compSpec.WorkArea.End); err != nil {
			return fmt.Errorf("work_area: %w", err)
		}
	}
	return nil
}
