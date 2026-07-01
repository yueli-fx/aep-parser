package aepmigrate

import (
	"bytes"
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
