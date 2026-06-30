// Package profile builds a stable, normalized project profile for understanding,
// structural diffing, and later generation workflows.
package profile

import (
	"fmt"
	"sort"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/projectindex"
)

const SchemaVersion = 1

type EvidenceLevel string

const (
	EvidenceL0Raw          EvidenceLevel = "L0_raw"
	EvidenceL1Parsed       EvidenceLevel = "L1_parsed"
	EvidenceL2Roundtrip    EvidenceLevel = "L2_roundtrip"
	EvidenceL3AEAccept     EvidenceLevel = "L3_ae_accept"
	EvidenceL4Render       EvidenceLevel = "L4_render"
	EvidenceL5UserVerified EvidenceLevel = "L5_user_verified"
)

type Options struct {
	Path string
	Dict *EffectDictionary
}

type Profile struct {
	SchemaVersion int           `json:"schema_version"`
	Meta          Meta          `json:"meta"`
	Fingerprint   Fingerprint   `json:"fingerprint"`
	Items         Items         `json:"items"`
	Comps         []Composition `json:"comps"`
	RenderQueue   *RenderQueue  `json:"render_queue,omitempty"`
	Unknowns      []Unknown     `json:"unknowns,omitempty"`
}

type Meta struct {
	Path                    string             `json:"path,omitempty"`
	ParseWarnings           []string           `json:"parse_warnings,omitempty"`
	StructuredParseWarnings []aep.ParseWarning `json:"structured_parse_warnings,omitempty"`
	BitsPerChannel          string             `json:"bits_per_channel,omitempty"`
}

type Fingerprint struct {
	CompCount          int            `json:"comp_count"`
	LayerCount         int            `json:"layer_count"`
	FootageCount       int            `json:"footage_count"`
	EffectUsage        map[string]int `json:"effect_usage,omitempty"`
	PluginDependencies []string       `json:"plugin_dependencies,omitempty"`
}

type Items struct {
	Comps   []Item `json:"comps,omitempty"`
	Footage []Item `json:"footage,omitempty"`
	Folders []Item `json:"folders,omitempty"`
}

type Item struct {
	ID       uint32   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Path     PathRef  `json:"path"`
	Evidence Evidence `json:"evidence"`
}

type Composition struct {
	ID                         uint32             `json:"id"`
	Name                       string             `json:"name"`
	Width                      uint16             `json:"width"`
	Height                     uint16             `json:"height"`
	FrameRate                  float64            `json:"frame_rate"`
	Duration                   float64            `json:"duration_seconds"`
	TickRate                   float64            `json:"tick_rate,omitempty"`
	Label                      uint8              `json:"label,omitempty"`
	Comment                    string             `json:"comment,omitempty"`
	BackgroundColor            [3]uint8           `json:"background_color"`
	ResolutionFactor           [2]uint16          `json:"resolution_factor"`
	PixelAspect                float64            `json:"pixel_aspect"`
	DisplayStartTime           float64            `json:"display_start_time"`
	Renderer                   string             `json:"renderer,omitempty"`
	Draft3D                    bool               `json:"draft_3d,omitempty"`
	FrameBlending              bool               `json:"frame_blending"`
	HideShyLayers              bool               `json:"hide_shy_layers"`
	PreserveNestedFrameRate    bool               `json:"preserve_nested_frame_rate"`
	PreserveNestedResolution   bool               `json:"preserve_nested_resolution"`
	WorkArea                   WorkArea           `json:"work_area"`
	MotionBlur                 MotionBlurSettings `json:"motion_blur"`
	Markers                    []Marker           `json:"markers,omitempty"`
	Guides                     []Guide            `json:"guides,omitempty"`
	MotionGraphicsTemplateName string             `json:"motion_graphics_template_name,omitempty"`
	EssentialGraphics          []EGController     `json:"essential_graphics,omitempty"`
	Layers                     []Layer            `json:"layers,omitempty"`
	Path                       PathRef            `json:"path"`
	Evidence                   Evidence           `json:"evidence"`
}

type WorkArea struct {
	Start float64 `json:"start_seconds"`
	End   float64 `json:"end_seconds"`
}

type MotionBlurSettings struct {
	Enabled             bool   `json:"enabled"`
	ShutterAngle        uint16 `json:"shutter_angle_degrees"`
	ShutterPhase        int32  `json:"shutter_phase"`
	AdaptiveSampleLimit int32  `json:"adaptive_sample_limit"`
	SamplesPerFrame     int32  `json:"samples_per_frame"`
}

type Layer struct {
	ID             uint32      `json:"id,omitempty"`
	Index          int         `json:"index"`
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	Label          uint8       `json:"label,omitempty"`
	Comment        string      `json:"comment,omitempty"`
	Quality        string      `json:"quality,omitempty"`
	BlendingMode   string      `json:"blending_mode,omitempty"`
	AutoOrient     string      `json:"auto_orient,omitempty"`
	LightKind      string      `json:"light_kind,omitempty"`
	SourceRef      *ItemRef    `json:"source_ref,omitempty"`
	LightSourceRef *LayerRef   `json:"light_source_ref,omitempty"`
	ParentRef      *LayerRef   `json:"parent_ref,omitempty"`
	MatteRef       *LayerRef   `json:"matte_ref,omitempty"`
	Timing         LayerTiming `json:"timing"`
	Flags          LayerFlags  `json:"flags"`
	Effects        []Effect    `json:"effects,omitempty"`
	Properties     []Property  `json:"properties,omitempty"`
	Markers        []Marker    `json:"markers,omitempty"`
	Masks          []Mask      `json:"masks,omitempty"`
	Shapes         []Shape     `json:"shapes,omitempty"`
	Text           *TextSource `json:"text,omitempty"`
	Path           PathRef     `json:"path"`
	Evidence       Evidence    `json:"evidence"`
}

type ItemRef struct {
	ID   uint32 `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name,omitempty"`
}

type LayerRef struct {
	ID    uint32 `json:"id"`
	Index int    `json:"index,omitempty"`
	Name  string `json:"name,omitempty"`
}

type LayerTiming struct {
	StartTime float64 `json:"start_time_seconds"`
	Duration  float64 `json:"duration_seconds"`
	InPoint   float64 `json:"in_point_seconds"`
	OutPoint  float64 `json:"out_point_seconds"`
	Stretch   float64 `json:"stretch"`
}

type LayerFlags struct {
	Visible               bool   `json:"visible"`
	Blend                 uint8  `json:"blend"`
	BlendName             string `json:"blend_name,omitempty"`
	TrackMatte            uint8  `json:"track_matte,omitempty"`
	TrackMatteName        string `json:"track_matte_name,omitempty"`
	Is3D                  bool   `json:"is_3d,omitempty"`
	Solo                  bool   `json:"solo,omitempty"`
	Shy                   bool   `json:"shy,omitempty"`
	Locked                bool   `json:"locked,omitempty"`
	IsAdjustment          bool   `json:"is_adjustment,omitempty"`
	IsNull                bool   `json:"is_null,omitempty"`
	IsGuide               bool   `json:"is_guide,omitempty"`
	MotionBlur            bool   `json:"motion_blur,omitempty"`
	EffectsEnabled        bool   `json:"effects_enabled,omitempty"`
	AudioEnabled          bool   `json:"audio_enabled,omitempty"`
	FrameBlendEnabled     bool   `json:"frame_blend_enabled,omitempty"`
	MarkersLocked         bool   `json:"markers_locked,omitempty"`
	FrameBlendPixelMotion bool   `json:"frame_blend_pixel_motion,omitempty"`
	CollapseTransform     bool   `json:"collapse_transform,omitempty"`
	SamplingBicubic       bool   `json:"sampling_bicubic,omitempty"`
	PreserveTransparency  bool   `json:"preserve_transparency,omitempty"`
}

type Effect struct {
	MatchName       string     `json:"match_name"`
	DisplayName     string     `json:"display_name,omitempty"`
	DependencyClass string     `json:"dependency_class"`
	Occurrence      int        `json:"occurrence"`
	Params          []Property `json:"params,omitempty"`
	TunedParams     []string   `json:"tuned_params,omitempty"`
	UnknownParams   []Unknown  `json:"unknown_params,omitempty"`
	Path            PathRef    `json:"path"`
	Evidence        Evidence   `json:"evidence"`
}

type Property struct {
	Name              string     `json:"name,omitempty"`
	MatchName         string     `json:"match_name,omitempty"`
	Occurrence        int        `json:"occurrence,omitempty"`
	StaticValue       any        `json:"static_value,omitempty"`
	LayerRef          *LayerRef  `json:"layer_ref,omitempty"`
	Default           any        `json:"default,omitempty"`
	Changed           bool       `json:"changed,omitempty"`
	Expression        string     `json:"expression,omitempty"`
	ExpressionEnabled *bool      `json:"expression_enabled,omitempty"`
	Keyframes         []Keyframe `json:"keyframes,omitempty"`
	Path              PathRef    `json:"path"`
	Evidence          Evidence   `json:"evidence"`
}

type Keyframe struct {
	Time              float64        `json:"time_seconds"`
	Value             any            `json:"value,omitempty"`
	InInterp          string         `json:"in_interp,omitempty"`
	OutInterp         string         `json:"out_interp,omitempty"`
	InSpatialTangent  []float64      `json:"in_spatial_tangent,omitempty"`
	OutSpatialTangent []float64      `json:"out_spatial_tangent,omitempty"`
	InTemporalEase    []TemporalEase `json:"in_temporal_ease,omitempty"`
	OutTemporalEase   []TemporalEase `json:"out_temporal_ease,omitempty"`
}

type Marker struct {
	Time         float64  `json:"time_seconds"`
	Duration     float64  `json:"duration_seconds,omitempty"`
	Label        uint8    `json:"label,omitempty"`
	Comment      string   `json:"comment,omitempty"`
	Chapter      string   `json:"chapter,omitempty"`
	URL          string   `json:"url,omitempty"`
	FrameTarget  string   `json:"frame_target,omitempty"`
	CuePointName string   `json:"cue_point_name,omitempty"`
	Path         PathRef  `json:"path"`
	Evidence     Evidence `json:"evidence"`
}

type Guide struct {
	Orientation string   `json:"orientation"`
	Position    float64  `json:"position"`
	Path        PathRef  `json:"path"`
	Evidence    Evidence `json:"evidence"`
}

type EGController struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	UUID     string   `json:"uuid,omitempty"`
	Path     PathRef  `json:"path"`
	Evidence Evidence `json:"evidence"`
}

type RenderQueue struct {
	NumItems int               `json:"num_items"`
	Items    []RenderQueueItem `json:"items,omitempty"`
	Path     PathRef           `json:"path"`
	Evidence Evidence          `json:"evidence"`
}

type RenderQueueItem struct {
	CompName          string         `json:"comp_name,omitempty"`
	Status            uint32         `json:"status"`
	Name              string         `json:"name,omitempty"`
	Comment           string         `json:"comment,omitempty"`
	TimeSpanStart     float64        `json:"time_span_start_seconds"`
	TimeSpanDuration  float64        `json:"time_span_duration_seconds"`
	OutputModuleCount int            `json:"output_module_count"`
	OutputModules     []OutputModule `json:"output_modules,omitempty"`
	Path              PathRef        `json:"path"`
	Evidence          Evidence       `json:"evidence"`
}

type OutputModule struct {
	Name         string   `json:"name,omitempty"`
	FileTemplate string   `json:"file_template,omitempty"`
	Path         PathRef  `json:"path"`
	Evidence     Evidence `json:"evidence"`
}

type TemporalEase struct {
	Speed     float64 `json:"speed"`
	Influence float64 `json:"influence"`
}

type Mask struct {
	Name           string             `json:"name,omitempty"`
	Index          uint32             `json:"index,omitempty"`
	Mode           string             `json:"mode,omitempty"`
	Inverted       bool               `json:"inverted,omitempty"`
	Locked         bool               `json:"locked,omitempty"`
	Color          []float64          `json:"color,omitempty"`
	MotionBlur     string             `json:"motion_blur,omitempty"`
	FeatherFalloff string             `json:"feather_falloff,omitempty"`
	Opacity        float64            `json:"opacity,omitempty"`
	Feather        []float64          `json:"feather,omitempty"`
	Expansion      float64            `json:"expansion,omitempty"`
	Closed         bool               `json:"closed"`
	Vertices       []MaskVertex       `json:"vertices,omitempty"`
	PathKeyframes  []MaskPathKeyframe `json:"path_keyframes,omitempty"`
	Path           PathRef            `json:"path"`
	Evidence       Evidence           `json:"evidence"`
}

type MaskPathKeyframe struct {
	Time            float64       `json:"time_seconds"`
	InInterp        string        `json:"in_interp,omitempty"`
	OutInterp       string        `json:"out_interp,omitempty"`
	InTemporalEase  *TemporalEase `json:"in_temporal_ease,omitempty"`
	OutTemporalEase *TemporalEase `json:"out_temporal_ease,omitempty"`
	Vertices        []MaskVertex  `json:"vertices,omitempty"`
}

type MaskVertex struct {
	Anchor     [2]float64 `json:"anchor"`
	InTangent  [2]float64 `json:"in_tangent"`
	OutTangent [2]float64 `json:"out_tangent"`
}

type Shape struct {
	Kind       string       `json:"kind"`
	Name       string       `json:"name,omitempty"`
	Closed     bool         `json:"closed,omitempty"`
	Vertices   []MaskVertex `json:"vertices,omitempty"`
	Properties []Property   `json:"properties,omitempty"`
	Path       PathRef      `json:"path"`
	Evidence   Evidence     `json:"evidence"`
}

type TextSource struct {
	Text          string          `json:"text"`
	Fonts         []string        `json:"fonts,omitempty"`
	Runs          []TextStyleRun  `json:"runs,omitempty"`
	Paragraphs    []TextParagraph `json:"paragraphs,omitempty"`
	Justification string          `json:"justification,omitempty"`
	IsBoxText     bool            `json:"is_box_text,omitempty"`
	BoxBounds     [4]float64      `json:"box_bounds,omitempty"`
	Path          PathRef         `json:"path"`
	Evidence      Evidence        `json:"evidence"`
}

type TextStyleRun struct {
	FontIndex       int        `json:"font_index"`
	FontName        string     `json:"font_name,omitempty"`
	FontSize        float64    `json:"font_size,omitempty"`
	FillColor       [4]float64 `json:"fill_color"`
	FauxBold        bool       `json:"faux_bold,omitempty"`
	FauxItalic      bool       `json:"faux_italic,omitempty"`
	AutoLeading     bool       `json:"auto_leading,omitempty"`
	Leading         float64    `json:"leading,omitempty"`
	Tracking        float64    `json:"tracking,omitempty"`
	BaselineShift   float64    `json:"baseline_shift,omitempty"`
	HorizontalScale float64    `json:"horizontal_scale,omitempty"`
	VerticalScale   float64    `json:"vertical_scale,omitempty"`
	Tsume           float64    `json:"tsume,omitempty"`
	CapsOption      string     `json:"caps_option,omitempty"`
	BaselineOption  string     `json:"baseline_option,omitempty"`
	AutoKernType    string     `json:"auto_kern_type,omitempty"`
	LineJoinType    string     `json:"line_join_type,omitempty"`
	DigitSet        string     `json:"digit_set,omitempty"`
	ApplyStroke     bool       `json:"apply_stroke,omitempty"`
	StrokeColor     [4]float64 `json:"stroke_color,omitempty"`
	StrokeWidth     float64    `json:"stroke_width,omitempty"`
	StrokeOverFill  bool       `json:"stroke_over_fill,omitempty"`
	NoBreak         bool       `json:"no_break,omitempty"`
}

type TextParagraph struct {
	Justification string `json:"justification,omitempty"`
}

type PathRef struct {
	Path        string         `json:"path"`
	DisplayPath string         `json:"display_path"`
	Identity    map[string]any `json:"identity,omitempty"`
}

type Evidence struct {
	Level      EvidenceLevel `json:"level"`
	Source     string        `json:"source"`
	Confidence string        `json:"confidence"`
	Notes      string        `json:"notes,omitempty"`
}

type Unknown struct {
	Path     string   `json:"path,omitempty"`
	Reason   string   `json:"reason"`
	Evidence Evidence `json:"evidence"`
}

func Build(project *aep.Project, opts Options) (*Profile, error) {
	if project == nil {
		return nil, fmt.Errorf("profile: nil project")
	}
	jp := project.ToJSON()
	idx := projectindex.Build(project)
	prof := &Profile{
		SchemaVersion: SchemaVersion,
		Meta: Meta{
			Path:                    opts.Path,
			ParseWarnings:           append([]string(nil), project.Warnings...),
			StructuredParseWarnings: append([]aep.ParseWarning(nil), project.ParseWarnings...),
			BitsPerChannel:          project.BitsPerChannel.String(),
		},
		Fingerprint: Fingerprint{
			CompCount:    len(project.Compositions),
			FootageCount: len(project.Footage),
			EffectUsage:  map[string]int{},
		},
	}

	for _, c := range jp.Compositions {
		prof.Items.Comps = append(prof.Items.Comps, Item{
			ID: c.ID, Name: c.Name, Type: "composition",
			Path:     itemPath("comp", c.ID, c.Name),
			Evidence: parsedEvidence(),
		})
	}
	for _, f := range jp.Footage {
		prof.Items.Footage = append(prof.Items.Footage, Item{
			ID: f.ID, Name: f.Name, Type: "footage",
			Path:     itemPath("footage", f.ID, f.Name),
			Evidence: parsedEvidence(),
		})
	}
	for _, f := range jp.Folders {
		prof.Items.Folders = append(prof.Items.Folders, Item{
			ID: f.ID, Name: f.Name, Type: "folder",
			Path:     itemPath("folder", f.ID, f.Name),
			Evidence: parsedEvidence(),
		})
	}
	if jp.RenderQueue != nil {
		prof.RenderQueue = buildRenderQueue(jp.RenderQueue)
	}

	pluginSeen := map[string]bool{}
	for ci, c := range jp.Compositions {
		var sceneComp *aep.Composition
		if ci < len(project.Compositions) {
			sceneComp = project.Compositions[ci]
		}
		var (
			backgroundColor  [3]uint8
			resolutionFactor = c.ResolutionFactor
			pixelAspect      float64
			displayStartTime float64
			label            uint8
			comment          string
		)
		if sceneComp != nil {
			backgroundColor = sceneComp.BGColor
			resolutionFactor = sceneComp.ResolutionFactor
			pixelAspect = sceneComp.PixelAspect
			displayStartTime = sceneComp.DisplayStartTime
			label = sceneComp.Label
			comment = sceneComp.Comment
		}
		cp := Composition{
			ID: c.ID, Name: c.Name, Width: c.Width, Height: c.Height,
			FrameRate: c.FrameRate, Duration: c.Duration, TickRate: c.TickRate,
			Label:                    label,
			Comment:                  comment,
			BackgroundColor:          backgroundColor,
			ResolutionFactor:         resolutionFactor,
			PixelAspect:              pixelAspect,
			DisplayStartTime:         displayStartTime,
			Renderer:                 c.Renderer,
			Draft3D:                  compositionDraft3D(sceneComp),
			FrameBlending:            compositionCdtaFlag(sceneComp, codec.CdtaFlagsByte8B, 0x10),
			HideShyLayers:            compositionCdtaFlag(sceneComp, codec.CdtaFlagsByte8B, 0x01),
			PreserveNestedFrameRate:  compositionCdtaFlag(sceneComp, codec.CdtaFlagsByte8B, 0x20),
			PreserveNestedResolution: compositionCdtaFlag(sceneComp, codec.CdtaFlagsByte8B, 0x80),
			WorkArea:                 WorkArea{Start: c.WorkAreaStart, End: c.WorkAreaEnd},
			MotionBlur: MotionBlurSettings{
				Enabled:             compositionCdtaFlag(sceneComp, codec.CdtaFlagsByte8B, 0x08),
				ShutterAngle:        c.ShutterAngle,
				ShutterPhase:        c.ShutterPhase,
				AdaptiveSampleLimit: c.MotionBlurAdaptiveSampleLimit,
				SamplesPerFrame:     c.MotionBlurSamplesPerFrame,
			},
			Path:     compPath(c),
			Evidence: parsedEvidence(),
		}
		for i, marker := range c.Markers {
			cp.Markers = append(cp.Markers, buildMarker(marker, markerPath(cp.Path.Path, cp.Path.DisplayPath, i)))
		}
		for i, guide := range c.Guides {
			cp.Guides = append(cp.Guides, buildGuide(guide, guidePath(cp.Path.Path, cp.Path.DisplayPath, i)))
		}
		cp.MotionGraphicsTemplateName = c.MotionGraphicsTemplateName
		for i, controller := range c.EssentialGraphics {
			cp.EssentialGraphics = append(cp.EssentialGraphics, buildEGController(controller, egControllerPath(cp.Path.Path, cp.Path.DisplayPath, i)))
		}
		layerByID := map[uint32]*aep.JSONLayer{}
		layerByIndex := map[int]*aep.JSONLayer{}
		for _, l := range c.Layers {
			layerByID[l.ID] = l
			layerByIndex[l.Index] = l
		}
		sceneLayerByID, sceneLayerByIndex := indexSceneLayers(sceneComp)
		for li, l := range c.Layers {
			sceneLayer := sceneLayerFor(l, li, sceneLayerByID, sceneLayerByIndex)
			lp := buildLayer(c, l, sceneLayer, layerByID, layerByIndex, idx, prof.Fingerprint.EffectUsage, pluginSeen, opts.Dict)
			prof.Fingerprint.LayerCount++
			cp.Layers = append(cp.Layers, lp)
		}
		prof.Comps = append(prof.Comps, cp)
	}
	for dep := range pluginSeen {
		prof.Fingerprint.PluginDependencies = append(prof.Fingerprint.PluginDependencies, dep)
	}
	sort.Strings(prof.Fingerprint.PluginDependencies)
	if len(prof.Fingerprint.EffectUsage) == 0 {
		prof.Fingerprint.EffectUsage = nil
	}
	return prof, nil
}

func compositionDraft3D(comp *aep.Composition) bool {
	return compositionCdtaFlag(comp, codec.CdtaFlagsByte8A, 0x01)
}

func compositionCdtaFlag(comp *aep.Composition, offset int, mask byte) bool {
	if comp == nil {
		return false
	}
	cdta := comp.CdtaRawBytes()
	return len(cdta) > offset && cdta[offset]&mask != 0
}

func buildLayer(
	c *aep.JSONComposition,
	l *aep.JSONLayer,
	sceneLayer *aep.Layer,
	layerByID map[uint32]*aep.JSONLayer,
	layerByIndex map[int]*aep.JSONLayer,
	idx *projectindex.Index,
	effectUsage map[string]int,
	pluginSeen map[string]bool,
	dict *EffectDictionary,
) Layer {
	lp := Layer{
		ID:           l.ID,
		Index:        l.Index,
		Name:         l.Name,
		Type:         l.Type,
		Label:        l.Label,
		Comment:      l.Comment,
		Quality:      qualityName(int(l.Quality)),
		BlendingMode: blendRecipeName(int(l.BlendingMode)),
		AutoOrient:   autoOrientName(l.AutoOrient),
		LightKind:    l.LightKind,
		Timing: LayerTiming{
			StartTime: l.StartTime,
			Duration:  l.Duration,
			InPoint:   layerInPoint(l, sceneLayer),
			OutPoint:  layerOutPoint(l, sceneLayer),
			Stretch:   l.Stretch,
		},
		Flags: LayerFlags{
			Visible:               l.Visible,
			Blend:                 l.BlendingMode,
			BlendName:             blendName(int(l.BlendingMode)),
			TrackMatte:            l.TrackMatte,
			TrackMatteName:        trackMatteName(int(l.TrackMatte)),
			Is3D:                  l.Is3D,
			Solo:                  l.Solo,
			Shy:                   l.Shy,
			Locked:                l.Locked,
			IsAdjustment:          l.IsAdjust,
			IsNull:                l.IsNull,
			IsGuide:               l.IsGuide,
			MotionBlur:            l.MotionBlur,
			EffectsEnabled:        l.EffectsEnabled,
			AudioEnabled:          l.AudioEnabled,
			FrameBlendEnabled:     l.FrameBlendEnabled,
			MarkersLocked:         l.MarkersLocked,
			FrameBlendPixelMotion: l.FrameBlendPixelMotion,
			CollapseTransform:     l.CollapseTransform,
			SamplingBicubic:       l.SamplingBicubic,
			PreserveTransparency:  l.PreserveTransparency,
		},
		Path:     layerPath(c, l),
		Evidence: parsedEvidence(),
	}
	if l.SourceID != 0 && l.Type != "light" {
		lp.SourceRef = sourceRef(l.SourceID, idx)
	}
	if sceneLayer != nil {
		lp.LightSourceRef = sceneLayerRef(sceneLayer.LightSource())
	}
	if l.ParentID != 0 {
		lp.ParentRef = layerRef(l.ParentID, layerByID)
	}
	if l.TrackMatte != 0 {
		lp.MatteRef = matteRef(l, sceneLayer, layerByID, layerByIndex)
	}
	for i, fx := range l.Effects {
		effectUsage[fx.MatchName]++
		class := classifyEffect(fx.MatchName)
		if class == "third_party" && !pluginSeen[fx.MatchName] {
			pluginSeen[fx.MatchName] = true
		}
		dictEffect, haveDictEffect := dict.effect(fx.MatchName)
		displayName := fx.Name
		if displayName == "" && haveDictEffect {
			displayName = dictEffect.Name
		}
		ep := Effect{
			MatchName:       fx.MatchName,
			DisplayName:     displayName,
			DependencyClass: class,
			Occurrence:      i,
			Path:            effectPath(c, l, fx.MatchName, i, displayName),
			Evidence:        parsedEvidence(),
		}
		paramSeen := map[string]int{}
		for _, p := range fx.Parameters {
			occ := paramSeen[p.MatchName]
			paramSeen[p.MatchName] = occ + 1
			pp := buildProperty(p, paramPath(ep.Path.Path, ep.Path.DisplayPath, p.MatchName, occ, p.Name), occ, layerByID)
			if pp.LayerRef != nil && pp.LayerRef.ID != l.ID {
				pp.Changed = true
			}
			if haveDictEffect {
				if dp, ok := dictEffect.Params[p.MatchName]; ok {
					if pp.Name == "" {
						pp.Name = dp.Name
					}
					pp.Default = dp.Default
					if propertyChanged(pp.StaticValue, dp.Default) || len(pp.Keyframes) > 0 {
						pp.Changed = true
					}
				}
			}
			if pp.Changed || len(pp.Keyframes) > 0 {
				ep.TunedParams = append(ep.TunedParams, pp.MatchName)
			}
			ep.Params = append(ep.Params, pp)
		}
		lp.Effects = append(lp.Effects, ep)
	}
	propSeen := map[string]int{}
	for _, p := range l.Properties {
		occ := propSeen[p.MatchName]
		propSeen[p.MatchName] = occ + 1
		lp.Properties = append(lp.Properties, buildProperty(p, propertyPath(lp.Path.Path, lp.Path.DisplayPath, p.MatchName, occ, p.Name), occ, layerByID))
	}
	for i, marker := range l.Markers {
		lp.Markers = append(lp.Markers, buildMarker(marker, markerPath(lp.Path.Path, lp.Path.DisplayPath, i)))
	}
	for i, m := range l.Masks {
		lp.Masks = append(lp.Masks, buildMask(c, l, m, i))
	}
	for i, sp := range l.ShapePaths {
		lp.Shapes = append(lp.Shapes, Shape{
			Kind:     "path",
			Name:     sp.Name,
			Closed:   sp.Closed,
			Vertices: maskVertices(sp.Vertices),
			Path:     shapePath(c, l, "path", i, sp.Name),
			Evidence: parsedEvidence(),
		})
	}
	for i, prim := range l.ShapePrimitives {
		lp.Shapes = append(lp.Shapes, buildShapePrimitive(c, l, prim, i, layerByID))
	}
	if l.TextSource != nil {
		lp.Text = buildText(c, l, l.TextSource)
	}
	return lp
}

func buildProperty(p *aep.JSONProperty, path PathRef, occurrence int, layerByID map[uint32]*aep.JSONLayer) Property {
	pp := Property{
		Name:        p.Name,
		MatchName:   p.MatchName,
		Occurrence:  occurrence,
		StaticValue: p.StaticValue,
		Expression:  p.Expression,
		Path:        path,
		Evidence:    parsedEvidence(),
	}
	if p.LayerRefID != 0 {
		pp.LayerRef = layerRef(p.LayerRefID, layerByID)
	}
	if p.ExpressionEnabled != nil {
		enabled := *p.ExpressionEnabled
		pp.ExpressionEnabled = &enabled
	}
	for _, kf := range p.Keyframes {
		pp.Keyframes = append(pp.Keyframes, Keyframe{
			Time:              kf.Time,
			Value:             kf.Value,
			InInterp:          kf.InInterp,
			OutInterp:         kf.OutInterp,
			InSpatialTangent:  append([]float64(nil), kf.InSpatialTangent...),
			OutSpatialTangent: append([]float64(nil), kf.OutSpatialTangent...),
			InTemporalEase:    temporalEaseSlice(kf.InTemporalEase),
			OutTemporalEase:   temporalEaseSlice(kf.OutTemporalEase),
		})
	}
	return pp
}

func buildMarker(m *aep.JSONMarker, path PathRef) Marker {
	if m == nil {
		return Marker{Path: path, Evidence: parsedEvidence()}
	}
	return Marker{
		Time:         m.Time,
		Duration:     m.Duration,
		Label:        m.Label,
		Comment:      m.Comment,
		Chapter:      m.Chapter,
		URL:          m.URL,
		FrameTarget:  m.FrameTarget,
		CuePointName: m.CuePointName,
		Path:         path,
		Evidence:     parsedEvidence(),
	}
}

func buildGuide(g *aep.JSONGuide, path PathRef) Guide {
	if g == nil {
		return Guide{Path: path, Evidence: parsedEvidence()}
	}
	return Guide{
		Orientation: g.Orientation,
		Position:    g.Position,
		Path:        path,
		Evidence:    parsedEvidence(),
	}
}

func buildEGController(ctrl *aep.JSONEGController, path PathRef) EGController {
	if ctrl == nil {
		return EGController{Path: path, Evidence: parsedEvidence()}
	}
	return EGController{
		Name:     ctrl.Name,
		Type:     ctrl.Type,
		UUID:     ctrl.UUID,
		Path:     path,
		Evidence: parsedEvidence(),
	}
}

func buildRenderQueue(rq *aep.JSONRenderQueue) *RenderQueue {
	out := &RenderQueue{
		NumItems: rq.NumItems,
		Path:     renderQueuePath(),
		Evidence: parsedEvidence(),
	}
	for i, item := range rq.Items {
		out.Items = append(out.Items, buildRenderQueueItem(item, i))
	}
	return out
}

func buildRenderQueueItem(item *aep.JSONRenderQueueItem, occurrence int) RenderQueueItem {
	path := renderQueueItemPath(occurrence)
	if item == nil {
		return RenderQueueItem{Path: path, Evidence: parsedEvidence()}
	}
	out := RenderQueueItem{
		CompName:          item.CompName,
		Status:            item.Status,
		Name:              item.Name,
		Comment:           item.Comment,
		TimeSpanStart:     item.TimeSpanStart,
		TimeSpanDuration:  item.TimeSpanDuration,
		OutputModuleCount: len(item.OutputModules),
		Path:              path,
		Evidence:          parsedEvidence(),
	}
	for i, output := range item.OutputModules {
		out.OutputModules = append(out.OutputModules, buildOutputModule(output, outputModulePath(path.Path, path.DisplayPath, i)))
	}
	return out
}

func buildOutputModule(output *aep.JSONOutputModule, path PathRef) OutputModule {
	if output == nil {
		return OutputModule{Path: path, Evidence: parsedEvidence()}
	}
	return OutputModule{
		Name:         output.Name,
		FileTemplate: output.FileTemplate,
		Path:         path,
		Evidence:     parsedEvidence(),
	}
}

func buildMask(c *aep.JSONComposition, l *aep.JSONLayer, m *aep.JSONMask, occurrence int) Mask {
	mp := Mask{
		Name:           m.Name,
		Index:          m.Index,
		Mode:           m.Mode,
		Inverted:       m.Inverted,
		Locked:         m.Locked,
		Color:          maskColor(m.Color),
		MotionBlur:     m.MotionBlur,
		FeatherFalloff: m.FeatherFalloff,
		Opacity:        m.Opacity,
		Feather:        []float64{m.Feather[0], m.Feather[1]},
		Expansion:      m.Expansion,
		Closed:         m.Closed,
		Vertices:       maskVertices(m.Vertices),
		Path:           maskPath(c, l, occurrence, m.Name),
		Evidence:       parsedEvidence(),
	}
	for _, kf := range m.PathKeyframes {
		mp.PathKeyframes = append(mp.PathKeyframes, MaskPathKeyframe{
			Time:            kf.Time,
			InInterp:        kf.InInterp,
			OutInterp:       kf.OutInterp,
			InTemporalEase:  temporalEasePtr(kf.InTemporalEase),
			OutTemporalEase: temporalEasePtr(kf.OutTemporalEase),
			Vertices:        maskVertices(kf.Vertices),
		})
	}
	return mp
}

func maskColor(value string) []float64 {
	var r, g, b int
	if _, err := fmt.Sscanf(value, "#%02X%02X%02X", &r, &g, &b); err != nil {
		return nil
	}
	return []float64{float64(r), float64(g), float64(b)}
}

func buildShapePrimitive(c *aep.JSONComposition, l *aep.JSONLayer, prim *aep.JSONShapePrimitive, occurrence int, layerByID map[uint32]*aep.JSONLayer) Shape {
	sp := Shape{
		Kind:     prim.Kind,
		Name:     prim.GroupName,
		Path:     shapePath(c, l, prim.Kind, occurrence, prim.GroupName),
		Evidence: parsedEvidence(),
	}
	props := []*aep.JSONProperty{
		prim.Size, prim.Position, prim.Roundness, prim.StarType, prim.Points,
		prim.Rotation, prim.InnerRadius, prim.OuterRadius, prim.InnerRoundness,
		prim.OuterRoundness,
	}
	seen := map[string]int{}
	for _, p := range props {
		if p == nil {
			continue
		}
		occ := seen[p.MatchName]
		seen[p.MatchName] = occ + 1
		sp.Properties = append(sp.Properties, buildProperty(p, propertyPath(sp.Path.Path, sp.Path.DisplayPath, p.MatchName, occ, p.Name), occ, layerByID))
	}
	return sp
}

func buildText(c *aep.JSONComposition, l *aep.JSONLayer, ts *aep.JSONTextSource) *TextSource {
	out := &TextSource{
		Text:          ts.Text,
		Fonts:         append([]string(nil), ts.Fonts...),
		Justification: ts.Justification,
		IsBoxText:     ts.IsBoxText,
		BoxBounds:     ts.BoxBounds,
		Path:          textPath(c, l),
		Evidence:      parsedEvidence(),
	}
	for _, r := range ts.Runs {
		out.Runs = append(out.Runs, TextStyleRun{
			FontIndex: r.FontIndex, FontName: r.FontName, FontSize: r.FontSize,
			FillColor: r.FillColor, FauxBold: r.FauxBold, FauxItalic: r.FauxItalic,
			AutoLeading: r.AutoLeading, Leading: r.Leading, Tracking: r.Tracking,
			BaselineShift: r.BaselineShift, HorizontalScale: r.HorizontalScale,
			VerticalScale: r.VerticalScale, Tsume: r.Tsume, ApplyStroke: r.ApplyStroke,
			CapsOption: r.CapsOption, BaselineOption: r.BaselineOption,
			AutoKernType: r.AutoKernType, LineJoinType: r.LineJoinType, DigitSet: r.DigitSet,
			StrokeColor: r.StrokeColor, StrokeWidth: r.StrokeWidth, StrokeOverFill: r.StrokeOverFill,
			NoBreak: r.NoBreak,
		})
	}
	for _, p := range ts.Paragraphs {
		out.Paragraphs = append(out.Paragraphs, TextParagraph{Justification: p.Justification})
	}
	return out
}

func parsedEvidence() Evidence {
	return Evidence{Level: EvidenceL1Parsed, Source: "internal/scene", Confidence: "high"}
}

func classifyEffect(matchName string) string {
	switch {
	case len(matchName) >= 5 && matchName[:5] == "ADBE ":
		return "native"
	case len(matchName) >= 3 && matchName[:3] == "CC ":
		return "cycore_bundled"
	default:
		return "third_party"
	}
}

func sourceRef(id uint32, idx *projectindex.Index) *ItemRef {
	switch item := idx.AVItemByID(id).(type) {
	case *aep.Composition:
		return &ItemRef{ID: id, Kind: "composition", Name: item.Name}
	case *aep.Footage:
		return &ItemRef{ID: id, Kind: "footage", Name: item.Name}
	}
	return &ItemRef{ID: id, Kind: "unknown"}
}

func indexSceneLayers(c *aep.Composition) (map[uint32]*aep.Layer, map[int]*aep.Layer) {
	byID := map[uint32]*aep.Layer{}
	byIndex := map[int]*aep.Layer{}
	if c == nil {
		return byID, byIndex
	}
	for _, l := range c.Layers {
		if l == nil {
			continue
		}
		if l.ID != 0 {
			byID[l.ID] = l
		}
		byIndex[l.Index] = l
	}
	return byID, byIndex
}

func sceneLayerFor(
	l *aep.JSONLayer,
	occurrence int,
	byID map[uint32]*aep.Layer,
	byIndex map[int]*aep.Layer,
) *aep.Layer {
	if l == nil {
		return nil
	}
	if l.ID != 0 {
		if layer := byID[l.ID]; layer != nil {
			return layer
		}
	}
	if layer := byIndex[l.Index]; layer != nil {
		return layer
	}
	return byIndex[occurrence]
}

func layerRef(id uint32, layers map[uint32]*aep.JSONLayer) *LayerRef {
	ref := &LayerRef{ID: id}
	if l, ok := layers[id]; ok {
		ref.Index = l.Index
		ref.Name = l.Name
	}
	return ref
}

func sceneLayerRef(l *aep.Layer) *LayerRef {
	if l == nil {
		return nil
	}
	return &LayerRef{ID: l.ID, Index: l.Index, Name: l.Name}
}

func jsonLayerRef(l *aep.JSONLayer) *LayerRef {
	if l == nil {
		return nil
	}
	return &LayerRef{ID: l.ID, Index: l.Index, Name: l.Name}
}

func matteRef(
	l *aep.JSONLayer,
	sceneLayer *aep.Layer,
	layerByID map[uint32]*aep.JSONLayer,
	layerByIndex map[int]*aep.JSONLayer,
) *LayerRef {
	if sceneLayer != nil && sceneLayer.TrackMatteLayerID != 0 {
		return layerRef(sceneLayer.TrackMatteLayerID, layerByID)
	}
	if l == nil || l.TrackMatte == 0 {
		return nil
	}
	return jsonLayerRef(layerByIndex[l.Index-1])
}

func layerInPoint(l *aep.JSONLayer, sceneLayer *aep.Layer) float64 {
	if sceneLayer != nil {
		in := sceneLayer.InPoint()
		out := sceneLayer.OutPoint()
		if out != 0 || in != 0 {
			return in
		}
	}
	return l.StartTime
}

func layerOutPoint(l *aep.JSONLayer, sceneLayer *aep.Layer) float64 {
	if sceneLayer != nil {
		in := sceneLayer.InPoint()
		out := sceneLayer.OutPoint()
		if out != 0 || in != 0 {
			return out
		}
	}
	return l.StartTime + l.Duration
}

func compPath(c *aep.JSONComposition) PathRef {
	return PathRef{
		Path:        fmt.Sprintf("comps.by_id[%d]", c.ID),
		DisplayPath: fmt.Sprintf("comps[%q]", c.Name),
		Identity:    map[string]any{"comp_id": c.ID},
	}
}

func itemPath(kind string, id uint32, name string) PathRef {
	return PathRef{
		Path:        fmt.Sprintf("items.%s.by_id[%d]", kind, id),
		DisplayPath: fmt.Sprintf("items.%s[%q]", kind, name),
		Identity:    map[string]any{"item_id": id, "item_type": kind},
	}
}

func renderQueuePath() PathRef {
	return PathRef{
		Path:        "render_queue",
		DisplayPath: "render_queue",
	}
}

func renderQueueItemPath(occurrence int) PathRef {
	return PathRef{
		Path:        fmt.Sprintf("render_queue.items[%d]", occurrence),
		DisplayPath: fmt.Sprintf("render_queue.items[%d]", occurrence),
		Identity:    map[string]any{"render_queue_item_occurrence": occurrence},
	}
}

func outputModulePath(parentPath, parentDisplay string, occurrence int) PathRef {
	return PathRef{
		Path:        fmt.Sprintf("%s.output_modules[%d]", parentPath, occurrence),
		DisplayPath: fmt.Sprintf("%s.output_modules[%d]", parentDisplay, occurrence),
		Identity:    map[string]any{"output_module_occurrence": occurrence},
	}
}

func layerPath(c *aep.JSONComposition, l *aep.JSONLayer) PathRef {
	path := fmt.Sprintf("%s.layers.by_index[%d]", compPath(c).Path, l.Index)
	identity := map[string]any{"comp_id": c.ID, "layer_index": l.Index}
	if l.ID != 0 {
		path = fmt.Sprintf("%s.layers.by_id[%d]", compPath(c).Path, l.ID)
		identity["layer_id"] = l.ID
	}
	return PathRef{
		Path:        path,
		DisplayPath: fmt.Sprintf("%s.layers[%d:%q]", compPath(c).DisplayPath, l.Index, l.Name),
		Identity:    identity,
	}
}

func effectPath(c *aep.JSONComposition, l *aep.JSONLayer, matchName string, occurrence int, display string) PathRef {
	parent := layerPath(c, l)
	label := display
	if label == "" {
		label = matchName
	}
	return PathRef{
		Path:        fmt.Sprintf("%s.effects.by_match_name[%q]#%d", parent.Path, matchName, occurrence),
		DisplayPath: fmt.Sprintf("%s.effects[%d:%q]", parent.DisplayPath, occurrence, label),
		Identity: map[string]any{
			"comp_id": c.ID, "layer_id": l.ID, "layer_index": l.Index,
			"effect_match_name": matchName, "effect_occurrence": occurrence,
		},
	}
}

func propertyPath(parentPath, parentDisplay, matchName string, occurrence int, name string) PathRef {
	label := name
	if label == "" {
		label = matchName
	}
	return PathRef{
		Path:        fmt.Sprintf("%s.properties.by_match_name[%q]#%d", parentPath, matchName, occurrence),
		DisplayPath: fmt.Sprintf("%s.properties[%d:%q]", parentDisplay, occurrence, label),
		Identity:    map[string]any{"property_match_name": matchName, "property_occurrence": occurrence},
	}
}

func markerPath(parentPath, parentDisplay string, occurrence int) PathRef {
	return PathRef{
		Path:        fmt.Sprintf("%s.markers[%d]", parentPath, occurrence),
		DisplayPath: fmt.Sprintf("%s.markers[%d]", parentDisplay, occurrence),
		Identity:    map[string]any{"marker_occurrence": occurrence},
	}
}

func guidePath(parentPath, parentDisplay string, occurrence int) PathRef {
	return PathRef{
		Path:        fmt.Sprintf("%s.guides[%d]", parentPath, occurrence),
		DisplayPath: fmt.Sprintf("%s.guides[%d]", parentDisplay, occurrence),
		Identity:    map[string]any{"guide_occurrence": occurrence},
	}
}

func egControllerPath(parentPath, parentDisplay string, occurrence int) PathRef {
	return PathRef{
		Path:        fmt.Sprintf("%s.essential_graphics[%d]", parentPath, occurrence),
		DisplayPath: fmt.Sprintf("%s.essential_graphics[%d]", parentDisplay, occurrence),
		Identity:    map[string]any{"essential_graphics_occurrence": occurrence},
	}
}

func paramPath(parentPath, parentDisplay, matchName string, occurrence int, name string) PathRef {
	label := name
	if label == "" {
		label = matchName
	}
	return PathRef{
		Path:        fmt.Sprintf("%s.params.by_match_name[%q]#%d", parentPath, matchName, occurrence),
		DisplayPath: fmt.Sprintf("%s.params[%d:%q]", parentDisplay, occurrence, label),
		Identity:    map[string]any{"property_match_name": matchName, "property_occurrence": occurrence},
	}
}

func maskPath(c *aep.JSONComposition, l *aep.JSONLayer, occurrence int, name string) PathRef {
	parent := layerPath(c, l)
	return PathRef{
		Path:        fmt.Sprintf("%s.masks.by_index[%d]", parent.Path, occurrence),
		DisplayPath: fmt.Sprintf("%s.masks[%d:%q]", parent.DisplayPath, occurrence, name),
		Identity:    map[string]any{"comp_id": c.ID, "layer_id": l.ID, "layer_index": l.Index, "mask_occurrence": occurrence},
	}
}

func shapePath(c *aep.JSONComposition, l *aep.JSONLayer, kind string, occurrence int, name string) PathRef {
	parent := layerPath(c, l)
	return PathRef{
		Path:        fmt.Sprintf("%s.shapes.%s#%d", parent.Path, kind, occurrence),
		DisplayPath: fmt.Sprintf("%s.shapes[%d:%q]", parent.DisplayPath, occurrence, name),
		Identity:    map[string]any{"comp_id": c.ID, "layer_id": l.ID, "layer_index": l.Index, "shape_kind": kind, "shape_occurrence": occurrence},
	}
}

func textPath(c *aep.JSONComposition, l *aep.JSONLayer) PathRef {
	parent := layerPath(c, l)
	return PathRef{
		Path:        parent.Path + ".text",
		DisplayPath: parent.DisplayPath + ".text",
		Identity:    map[string]any{"comp_id": c.ID, "layer_id": l.ID, "layer_index": l.Index},
	}
}

func temporalEaseSlice(in []aep.JSONTemporalEase) []TemporalEase {
	var out []TemporalEase
	for _, e := range in {
		out = append(out, TemporalEase{Speed: e.Speed, Influence: e.Influence})
	}
	return out
}

func temporalEasePtr(in *aep.JSONTemporalEase) *TemporalEase {
	if in == nil {
		return nil
	}
	return &TemporalEase{Speed: in.Speed, Influence: in.Influence}
}

func maskVertices(in []aep.JSONMaskVertex) []MaskVertex {
	var out []MaskVertex
	for _, v := range in {
		out = append(out, MaskVertex{Anchor: v.Anchor, InTangent: v.InTangent, OutTangent: v.OutTangent})
	}
	return out
}

var blendNames = map[int]string{
	0: "NormalCamera", 2: "Normal", 3: "Dissolve", 4: "Add", 5: "Multiply",
	6: "Screen", 7: "Overlay", 8: "SoftLight", 9: "HardLight", 10: "Darken",
	11: "Lighten", 12: "ClassicDiff", 13: "Hue", 14: "Saturation", 15: "Color",
	16: "Luminosity", 17: "StencilAlpha", 18: "StencilLuma", 19: "SilhouetteAlpha",
	20: "SilhouetteLuma", 21: "LuminescentPremul", 22: "AlphaAdd", 23: "ClassicColorDodge",
	24: "ClassicColorBurn", 25: "Exclusion", 26: "Difference", 27: "ColorDodge",
	28: "ColorBurn", 29: "LinearDodge", 30: "LinearBurn", 31: "LinearLight",
	32: "VividLight", 33: "PinLight", 34: "HardMix", 35: "LighterColor",
	36: "DarkerColor", 37: "Subtract", 38: "Divide",
}

func blendName(mode int) string {
	if name, ok := blendNames[mode]; ok {
		return name
	}
	return fmt.Sprintf("blend#%d", mode)
}

var blendRecipeNames = map[int]string{
	0: "normal_camera", 2: "normal", 3: "dissolve", 4: "add", 5: "multiply",
	6: "screen", 7: "overlay", 8: "soft_light", 9: "hard_light", 10: "darken",
	11: "lighten", 12: "classic_difference", 13: "hue", 14: "saturation", 15: "color",
	16: "luminosity", 17: "stencil_alpha", 18: "stencil_luma", 19: "silhouette_alpha",
	20: "silhouette_luma", 21: "luminescent_premul", 22: "alpha_add", 23: "classic_color_dodge",
	24: "classic_color_burn", 25: "exclusion", 26: "difference", 27: "color_dodge",
	28: "color_burn", 29: "linear_dodge", 30: "linear_burn", 31: "linear_light",
	32: "vivid_light", 33: "pin_light", 34: "hard_mix", 35: "lighter_color",
	36: "darker_color", 37: "subtract", 38: "divide",
}

func blendRecipeName(mode int) string {
	if name, ok := blendRecipeNames[mode]; ok {
		return name
	}
	return fmt.Sprintf("blend_%d", mode)
}

func qualityName(quality int) string {
	switch quality {
	case 0:
		return "wireframe"
	case 1:
		return "draft"
	case 2:
		return "best"
	default:
		return fmt.Sprintf("quality_%d", quality)
	}
}

func autoOrientName(value string) string {
	switch value {
	case "along-path":
		return "along_path"
	case "camera-or-point-of-interest":
		return "camera_or_point_of_interest"
	case "characters-toward-camera":
		return "characters_toward_camera"
	default:
		return value
	}
}

var trackMatteNames = map[int]string{
	1: "Alpha",
	2: "AlphaInverse",
	3: "Luma",
	4: "LumaInverse",
}

func trackMatteName(mode int) string {
	if mode == 0 {
		return ""
	}
	if name, ok := trackMatteNames[mode]; ok {
		return name
	}
	return fmt.Sprintf("track_matte#%d", mode)
}
