package aeoracle

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

const SchemaVersion = 1
const DefaultFrameTimeoutMS = 120000

type FrameOptions struct {
	CompName  string
	MaxFrames int
}

type FrameTarget struct {
	Frame   int     `json:"frame"`
	Seconds float64 `json:"seconds"`
	Tag     string  `json:"tag"`
	Reason  string  `json:"reason"`
}

type RenderRequest struct {
	SchemaVersion  int           `json:"schema_version"`
	AEPPath        string        `json:"aep_path"`
	CompName       string        `json:"comp_name,omitempty"`
	OutputDir      string        `json:"output_dir"`
	DonePath       string        `json:"done_path,omitempty"`
	MetadataPath   string        `json:"metadata_path,omitempty"`
	FrameTimeoutMS int           `json:"frame_timeout_ms,omitempty"`
	Frames         []FrameTarget `json:"frames"`
}

type RenderMetadata struct {
	SchemaVersion int                   `json:"schema_version"`
	AEVersion     string                `json:"ae_version,omitempty"`
	OS            string                `json:"os,omitempty"`
	AEPPath       string                `json:"aep_path"`
	CompName      string                `json:"comp_name"`
	OutputDir     string                `json:"output_dir"`
	Status        string                `json:"status"`
	Frames        []RenderedFrameRecord `json:"frames,omitempty"`
	Warnings      []string              `json:"warnings,omitempty"`
}

type RenderedFrameRecord struct {
	Frame      int     `json:"frame"`
	Seconds    float64 `json:"seconds"`
	Tag        string  `json:"tag"`
	Reason     string  `json:"reason"`
	OutputPath string  `json:"output_path"`
	Status     string  `json:"status"`
}

func SelectFrames(prof *profile.Profile, opts FrameOptions) ([]FrameTarget, error) {
	if prof == nil {
		return nil, fmt.Errorf("aeoracle: nil profile")
	}
	comp, err := selectComp(prof, opts.CompName)
	if err != nil {
		return nil, err
	}
	fps := comp.FrameRate
	if fps <= 0 {
		fps = 30
	}
	activeEnd := comp.Duration
	visibleOut := 0.0
	for _, layer := range comp.Layers {
		if layer.Flags.Visible && layer.Timing.OutPoint > visibleOut {
			visibleOut = layer.Timing.OutPoint
		}
	}
	if visibleOut > 0 {
		activeEnd = visibleOut
	}
	maxFrame := int(math.Round(activeEnd * fps))
	if maxFrame < 0 {
		maxFrame = 0
	}
	maxFrames := opts.MaxFrames
	if maxFrames <= 0 {
		maxFrames = 8
	}

	candidates := map[int]FrameTarget{}
	add := func(seconds float64, reason string) {
		if seconds < 0 {
			seconds = 0
		}
		frame := int(math.Round(seconds * fps))
		if frame > maxFrame {
			frame = maxFrame
		}
		if frame < 0 {
			frame = 0
		}
		if _, ok := candidates[frame]; ok {
			return
		}
		candidates[frame] = FrameTarget{
			Frame:   frame,
			Seconds: roundSeconds(float64(frame) / fps),
			Tag:     fmt.Sprintf("f%06d", frame),
			Reason:  reason,
		}
	}

	add(0, "comp_start")
	add(float64(maxFrame)/fps, "comp_end")
	for _, layer := range comp.Layers {
		if !layer.Flags.Visible {
			continue
		}
		if layer.Timing.InPoint > 0 {
			add(layer.Timing.InPoint, "layer_in")
		}
		if layer.Timing.OutPoint > 0 {
			add(layer.Timing.OutPoint, "layer_out")
		}
		for _, prop := range layer.Properties {
			addPropertyKeyframes(prop, add)
		}
		for _, effect := range layer.Effects {
			for _, param := range effect.Params {
				addPropertyKeyframes(param, add)
			}
		}
	}

	frames := sortedFrames(candidates)
	for i := 0; i+1 < len(frames); i++ {
		if frames[i+1].Frame-frames[i].Frame >= int(math.Round(fps)) {
			mid := (frames[i].Seconds + frames[i+1].Seconds) / 2
			add(mid, "quiet_midpoint")
		}
	}
	frames = sortedFrames(candidates)
	if len(frames) > maxFrames {
		frames = capFrames(frames, maxFrames)
	}
	return frames, nil
}

func NewRenderRequest(aepPath string, compName string, outputDir string, frames []FrameTarget) RenderRequest {
	return RenderRequest{
		SchemaVersion:  SchemaVersion,
		AEPPath:        aepPath,
		CompName:       compName,
		OutputDir:      outputDir,
		DonePath:       filepath.Join(outputDir, "aeoracle_render.done"),
		MetadataPath:   filepath.Join(outputDir, "metadata.json"),
		FrameTimeoutMS: DefaultFrameTimeoutMS,
		Frames:         append([]FrameTarget(nil), frames...),
	}
}

func CloneRenderRequest(source RenderRequest, aepPath string, compName string, outputDir string) RenderRequest {
	if compName == "" {
		compName = source.CompName
	}
	return NewRenderRequest(aepPath, compName, outputDir, source.Frames)
}

func WriteRequest(path string, req RenderRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func ReadRequest(path string) (RenderRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RenderRequest{}, err
	}
	var req RenderRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return RenderRequest{}, err
	}
	if err := req.Validate(); err != nil {
		return RenderRequest{}, err
	}
	return req, nil
}

func (r RenderRequest) Validate() error {
	if r.SchemaVersion != SchemaVersion {
		return fmt.Errorf("aeoracle: unsupported request schema_version %d", r.SchemaVersion)
	}
	if r.AEPPath == "" {
		return fmt.Errorf("aeoracle: request aep_path is required")
	}
	if r.OutputDir == "" {
		return fmt.Errorf("aeoracle: request output_dir is required")
	}
	if len(r.Frames) == 0 {
		return fmt.Errorf("aeoracle: request frames are required")
	}
	if r.FrameTimeoutMS < 0 {
		return fmt.Errorf("aeoracle: request frame_timeout_ms must be non-negative")
	}
	return nil
}

func selectComp(prof *profile.Profile, name string) (profile.Composition, error) {
	if name != "" {
		for _, comp := range prof.Comps {
			if comp.Name == name {
				return comp, nil
			}
		}
		return profile.Composition{}, fmt.Errorf("aeoracle: comp %q not found", name)
	}
	if len(prof.Comps) == 0 {
		return profile.Composition{}, fmt.Errorf("aeoracle: profile has no compositions")
	}
	return prof.Comps[0], nil
}

func addPropertyKeyframes(prop profile.Property, add func(float64, string)) {
	for _, kf := range prop.Keyframes {
		add(kf.Time, "keyframe")
	}
}

func sortedFrames(in map[int]FrameTarget) []FrameTarget {
	out := make([]FrameTarget, 0, len(in))
	for _, frame := range in {
		out = append(out, frame)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Frame < out[j].Frame
	})
	return out
}

func capFrames(frames []FrameTarget, maxFrames int) []FrameTarget {
	if maxFrames <= 0 || len(frames) <= maxFrames {
		return frames
	}
	if maxFrames == 1 {
		return []FrameTarget{frames[0]}
	}
	out := []FrameTarget{frames[0]}
	middleSlots := maxFrames - 2
	for i := 1; i < len(frames)-1 && len(out) < middleSlots+1; i++ {
		out = append(out, frames[i])
	}
	out = append(out, frames[len(frames)-1])
	return out
}

func roundSeconds(v float64) float64 {
	return math.Round(v*1000000) / 1000000
}
