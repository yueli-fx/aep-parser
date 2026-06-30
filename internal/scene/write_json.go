package scene

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// ──────────────────────────────────────────────
// JSON-friendly view types (snake_case, omitempty)
// ──────────────────────────────────────────────

// JSONProject is the JSON representation of a Project.
type JSONProject struct {
	Compositions []*JSONComposition `json:"compositions"`
	Footage      []*JSONFootage     `json:"footage"`
	Folders      []*JSONFolder      `json:"folders"`
	RenderQueue  *JSONRenderQueue   `json:"render_queue,omitempty"`
}

// JSONRenderQueue is the JSON representation of a RenderQueue (read-only).
type JSONRenderQueue struct {
	NumItems int                    `json:"num_items"`
	Items    []*JSONRenderQueueItem `json:"items,omitempty"`
}

// JSONRenderQueueItem is the JSON representation of a RenderQueueItem.
type JSONRenderQueueItem struct {
	CompName         string              `json:"comp_name,omitempty"`
	Status           uint32              `json:"status"`
	Name             string              `json:"name,omitempty"`
	Comment          string              `json:"comment,omitempty"`
	LogType          uint16              `json:"log_type,omitempty"`
	QueueItemNotify  bool                `json:"queue_item_notify,omitempty"`
	ElapsedSeconds   uint32              `json:"elapsed_seconds,omitempty"`
	TimeSpanStart    float64             `json:"time_span_start_seconds"`
	TimeSpanDuration float64             `json:"time_span_duration_seconds"`
	RenderSettings   JSONRenderSettings  `json:"render_settings"`
	OutputModules    []*JSONOutputModule `json:"output_modules,omitempty"`
}

// JSONRenderSettings mirrors RenderSettings (values follow the reference
// parser's NUMBER semantics: -1 = current settings).
type JSONRenderSettings struct {
	Quality           int    `json:"quality"`
	ColorDepth        int    `json:"color_depth"`
	Effects           int    `json:"effects"`
	FieldRender       int    `json:"field_render"`
	Pulldown          int    `json:"pulldown"`
	FrameBlending     int    `json:"frame_blending"`
	MotionBlur        int    `json:"motion_blur"`
	ProxyUse          int    `json:"proxy_use"`
	SoloSwitches      int    `json:"solo_switches"`
	GuideLayers       int    `json:"guide_layers"`
	DiskCache         int    `json:"disk_cache"`
	FrameRate         int    `json:"frame_rate"`
	Resolution        [2]int `json:"resolution"`
	SkipExistingFiles bool   `json:"skip_existing_files"`
}

// JSONOutputModule is the JSON representation of an OutputModule.
type JSONOutputModule struct {
	Name          string                   `json:"name,omitempty"`
	FileTemplate  string                   `json:"file_template,omitempty"`
	FullPath      string                   `json:"full_path,omitempty"`
	Settings      JSONOutputModuleSettings `json:"settings"`
	FormatOptions *JSONFormatOptions       `json:"format_options,omitempty"`
}

// JSONFormatOptions mirrors FormatOptions (only the active Kind's fields are
// meaningful).
type JSONFormatOptions struct {
	Kind                  string  `json:"kind"`
	TenBitBlackPoint      int     `json:"ten_bit_black_point,omitempty"`
	TenBitWhitePoint      int     `json:"ten_bit_white_point,omitempty"`
	ConvertedBlackPoint   float64 `json:"converted_black_point,omitempty"`
	ConvertedWhitePoint   float64 `json:"converted_white_point,omitempty"`
	CurrentGamma          float64 `json:"current_gamma,omitempty"`
	HighlightExpansion    int     `json:"highlight_expansion,omitempty"`
	LogarithmicConversion bool    `json:"logarithmic_conversion,omitempty"`
	CineonFileFormat      int     `json:"cineon_file_format,omitempty"`
	Quality               int     `json:"quality,omitempty"`
	ThirtyTwoBitFloat     bool    `json:"thirty_two_bit_float,omitempty"`
	LuminanceChroma       bool    `json:"luminance_chroma,omitempty"`
	BitsPerPixel          int     `json:"bits_per_pixel,omitempty"`
	RLECompression        bool    `json:"rle_compression,omitempty"`
	LZWCompression        bool    `json:"lzw_compression,omitempty"`
	IBMPCByteOrder        bool    `json:"ibm_pc_byte_order,omitempty"`
	Width                 int     `json:"width,omitempty"`
	Height                int     `json:"height,omitempty"`
	BitDepth              int     `json:"bit_depth,omitempty"`
	Compression           int     `json:"compression,omitempty"`
}

// JSONOutputModuleSettings mirrors OutputModuleSettings.
type JSONOutputModuleSettings struct {
	Channels            int     `json:"channels"`
	ResizeQuality       int     `json:"resize_quality"`
	Resize              bool    `json:"resize"`
	LockAspectRatio     bool    `json:"lock_aspect_ratio"`
	Crop                bool    `json:"crop"`
	CropTop             int     `json:"crop_top"`
	CropLeft            int     `json:"crop_left"`
	CropBottom          int     `json:"crop_bottom"`
	CropRight           int     `json:"crop_right"`
	OutputAudio         int     `json:"output_audio"`
	IncludeProjectLink  bool    `json:"include_project_link"`
	PostRenderAction    uint32  `json:"post_render_action"`
	ConvertToLinear     int     `json:"convert_to_linear"`
	UseCompFrameNumber  bool    `json:"use_comp_frame_number"`
	UseRegionOfInterest bool    `json:"use_region_of_interest"`
	IncludeSourceXMP    bool    `json:"include_source_xmp"`
	PreserveRGB         bool    `json:"preserve_rgb"`
	VideoCodec          string  `json:"video_codec,omitempty"`
	FormatID            string  `json:"format_id,omitempty"`
	StartingNumber      uint32  `json:"starting_number"`
	Width               int     `json:"width"`
	Height              int     `json:"height"`
	Depth               int     `json:"depth"`
	VideoOutput         bool    `json:"video_output"`
	AudioSampleRate     float64 `json:"audio_sample_rate"`
	AudioBitDepth       int     `json:"audio_bit_depth"`
	AudioChannels       int     `json:"audio_channels"`
	AudioEnabled        bool    `json:"audio_enabled"`
}

// JSONComposition is the JSON representation of a Composition.
type JSONComposition struct {
	ID                            uint32              `json:"id"`
	Name                          string              `json:"name"`
	Width                         uint16              `json:"width"`
	Height                        uint16              `json:"height"`
	FrameRate                     float64             `json:"frame_rate"`
	Duration                      float64             `json:"duration_seconds"`
	TickRate                      float64             `json:"tick_rate,omitempty"`
	BGColor                       string              `json:"bg_color"`
	ResolutionFactor              [2]uint16           `json:"resolution_factor,omitempty"`
	Renderer                      string              `json:"renderer,omitempty"`
	WorkAreaStart                 float64             `json:"work_area_start_seconds"`
	WorkAreaEnd                   float64             `json:"work_area_end_seconds"`
	ShutterAngle                  uint16              `json:"shutter_angle_degrees"`
	ShutterPhase                  int32               `json:"shutter_phase"`
	MotionBlurAdaptiveSampleLimit int32               `json:"motion_blur_adaptive_sample_limit"`
	MotionBlurSamplesPerFrame     int32               `json:"motion_blur_samples_per_frame"`
	Layers                        []*JSONLayer        `json:"layers,omitempty"`
	Markers                       []*JSONMarker       `json:"markers,omitempty"` // composition-level markers
	Guides                        []*JSONGuide        `json:"guides,omitempty"`  // composition ruler guides
	MotionGraphicsTemplateName    string              `json:"motion_graphics_template_name,omitempty"`
	EssentialGraphics             []*JSONEGController `json:"essential_graphics,omitempty"` // EG panel controllers
}

// JSONGuide is the JSON representation of a composition ruler guide.
type JSONGuide struct {
	Orientation string  `json:"orientation"` // "horizontal" / "vertical"
	Position    float64 `json:"position"`    // pixels from top (horizontal) / left (vertical)
}

// JSONEGController is the JSON representation of an Essential Graphics controller.
type JSONEGController struct {
	Name string `json:"name"`
	Type string `json:"type"` // "slider" / "color" / "checkbox" / ...
	UUID string `json:"uuid"`
}

// JSONLayer is the JSON representation of a Layer.
type JSONLayer struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	ID         uint32 `json:"id,omitempty"`
	ParentID   uint32 `json:"parent_id,omitempty"`
	ParentName string `json:"parent_name,omitempty"`
	SourceID   uint32 `json:"source_id,omitempty"`

	StartTime float64 `json:"start_time_seconds"`
	Duration  float64 `json:"duration_seconds"`
	Stretch   float64 `json:"stretch"`

	Quality              uint16 `json:"quality,omitempty"`
	Label                uint8  `json:"label,omitempty"`
	BlendingMode         uint8  `json:"blending_mode,omitempty"`
	TrackMatte           uint8  `json:"track_matte,omitempty"`
	LightKind            string `json:"light_kind,omitempty"`
	PreserveTransparency bool   `json:"preserve_transparency,omitempty"`
	AutoOrient           string `json:"auto_orient,omitempty"`
	Comment              string `json:"comment,omitempty"`

	Is3D                  bool                  `json:"is_3d,omitempty"`
	Solo                  bool                  `json:"solo,omitempty"`
	Shy                   bool                  `json:"shy,omitempty"`
	Locked                bool                  `json:"locked,omitempty"`
	Visible               bool                  `json:"visible"`
	IsAdjust              bool                  `json:"is_adjustment,omitempty"`
	IsNull                bool                  `json:"is_null,omitempty"`
	IsGuide               bool                  `json:"is_guide,omitempty"`
	MarkersLocked         bool                  `json:"markers_locked,omitempty"`
	MotionBlur            bool                  `json:"motion_blur,omitempty"`
	EffectsEnabled        bool                  `json:"effects_enabled,omitempty"`
	AudioEnabled          bool                  `json:"audio_enabled,omitempty"`
	FrameBlendEnabled     bool                  `json:"frame_blend_enabled,omitempty"`
	FrameBlendPixelMotion bool                  `json:"frame_blend_pixel_motion,omitempty"`
	CollapseTransform     bool                  `json:"collapse_transform,omitempty"`
	SamplingBicubic       bool                  `json:"sampling_bicubic,omitempty"`
	IsShapeLayer          bool                  `json:"is_shape_layer,omitempty"`
	Properties            []*JSONProperty       `json:"properties,omitempty"`
	Effects               []*JSONEffect         `json:"effects,omitempty"`
	Markers               []*JSONMarker         `json:"markers,omitempty"`
	Masks                 []*JSONMask           `json:"masks,omitempty"`
	ShapePaths            []*JSONShapePath      `json:"shape_paths,omitempty"`
	ShapePrimitives       []*JSONShapePrimitive `json:"shape_primitives,omitempty"`
	HasTextSource         bool                  `json:"has_text_source,omitempty"`
	TextSource            *JSONTextSource       `json:"text_source,omitempty"`
}

// JSONTextSource is the JSON view of a decoded TextSource.
type JSONTextSource struct {
	Text          string              `json:"text"`
	Fonts         []string            `json:"fonts,omitempty"`
	FontAxes      [][]float64         `json:"font_axes,omitempty"` // parallel to Fonts; per-font variable-axis values (empty for non-variable fonts)
	Runs          []JSONTextStyleRun  `json:"runs,omitempty"`
	Paragraphs    []JSONTextParagraph `json:"paragraphs,omitempty"`
	Justification string              `json:"justification,omitempty"`
	IsBoxText     bool                `json:"is_box_text,omitempty"`
	BoxBounds     [4]float64          `json:"box_bounds,omitempty"` // [xmin, ymin, xmax, ymax]
}

// JSONTextStyleRun mirrors TextStyleRun for JSON output.
type JSONTextStyleRun struct {
	FontIndex       int        `json:"font_index"`
	FontName        string     `json:"font_name,omitempty"`
	FontSize        float64    `json:"font_size,omitempty"`
	FillColor       [4]float64 `json:"fill_color"` // [R, G, B, A]
	FauxBold        bool       `json:"faux_bold,omitempty"`
	FauxItalic      bool       `json:"faux_italic,omitempty"`
	AutoLeading     bool       `json:"auto_leading,omitempty"`
	Leading         float64    `json:"leading,omitempty"`
	Tracking        float64    `json:"tracking,omitempty"`
	BaselineShift   float64    `json:"baseline_shift,omitempty"`
	HorizontalScale float64    `json:"horizontal_scale,omitempty"`
	VerticalScale   float64    `json:"vertical_scale,omitempty"`
	Tsume           float64    `json:"tsume,omitempty"`
	ApplyStroke     bool       `json:"apply_stroke,omitempty"`
	StrokeColor     [4]float64 `json:"stroke_color,omitempty"`
	StrokeWidth     float64    `json:"stroke_width,omitempty"`
	StrokeOverFill  bool       `json:"stroke_over_fill,omitempty"`
	NoBreak         bool       `json:"no_break,omitempty"`
}

// JSONTextParagraph mirrors TextParagraph for JSON output.
type JSONTextParagraph struct {
	Justification string `json:"justification,omitempty"`
}

// JSONProperty is the JSON representation of a Property.
type JSONProperty struct {
	Name              string          `json:"name"`
	MatchName         string          `json:"match_name,omitempty"`
	Keyframes         []*JSONKeyframe `json:"keyframes,omitempty"`
	StaticValue       interface{}     `json:"static_value,omitempty"`
	LayerRefID        uint32          `json:"layer_ref_id,omitempty"`
	Expression        string          `json:"expression,omitempty"`
	ExpressionEnabled *bool           `json:"expression_enabled,omitempty"`
}

// JSONKeyframe is the JSON representation of a Keyframe (extended).

// JSONEffect is the JSON representation of an Effect.
type JSONEffect struct {
	MatchName  string          `json:"match_name"`
	Name       string          `json:"name,omitempty"`
	Parameters []*JSONProperty `json:"parameters,omitempty"`
}

// JSONMarker is the JSON representation of a Marker.
type JSONMarker struct {
	Time         float64 `json:"time_seconds"`
	Duration     float64 `json:"duration_seconds,omitempty"`
	Label        uint8   `json:"label,omitempty"`
	Comment      string  `json:"comment,omitempty"`
	Chapter      string  `json:"chapter,omitempty"`
	URL          string  `json:"url,omitempty"`
	FrameTarget  string  `json:"frame_target,omitempty"`
	CuePointName string  `json:"cue_point_name,omitempty"`
}

// JSONMask is the JSON representation of a Mask.
type JSONMask struct {
	Name           string                  `json:"name,omitempty"`
	Index          uint32                  `json:"index,omitempty"`
	Mode           string                  `json:"mode,omitempty"`
	Inverted       bool                    `json:"inverted,omitempty"`
	Locked         bool                    `json:"locked,omitempty"`
	Color          string                  `json:"color,omitempty"` // "#RRGGBB"
	MotionBlur     string                  `json:"motion_blur,omitempty"`
	FeatherFalloff string                  `json:"feather_falloff,omitempty"`
	Closed         bool                    `json:"closed"`
	Feather        [2]float64              `json:"feather,omitempty"`
	Opacity        float64                 `json:"opacity,omitempty"`
	Expansion      float64                 `json:"expansion,omitempty"`
	Vertices       []JSONMaskVertex        `json:"vertices,omitempty"`
	PathKeyframes  []*JSONMaskPathKeyframe `json:"path_keyframes,omitempty"`
}

// JSONMaskPathKeyframe mirrors MaskPathKeyframe.
type JSONMaskPathKeyframe struct {
	Time            float64           `json:"time_seconds"`
	InInterp        string            `json:"in_interp,omitempty"`
	OutInterp       string            `json:"out_interp,omitempty"`
	InTemporalEase  *JSONTemporalEase `json:"in_temporal_ease,omitempty"`
	OutTemporalEase *JSONTemporalEase `json:"out_temporal_ease,omitempty"`
	Vertices        []JSONMaskVertex  `json:"vertices,omitempty"`
}

// JSONTemporalEase mirrors TemporalEase.
type JSONTemporalEase struct {
	Speed     float64 `json:"speed"`
	Influence float64 `json:"influence"`
}

// JSONMaskVertex mirrors MaskVertex.
type JSONMaskVertex struct {
	Anchor     [2]float64 `json:"anchor"`
	InTangent  [2]float64 `json:"in_tangent"`
	OutTangent [2]float64 `json:"out_tangent"`
}

// JSONShapePath is the JSON representation of a ShapePath.
// JSONShapePrimitive mirrors ShapePrimitive for JSON output. Nil
// sub-property fields are omitted; present ones serialize as nested
// JSONProperty records.
type JSONShapePrimitive struct {
	Kind           string        `json:"kind"`
	GroupName      string        `json:"group_name,omitempty"`
	Size           *JSONProperty `json:"size,omitempty"`
	Position       *JSONProperty `json:"position,omitempty"`
	Roundness      *JSONProperty `json:"roundness,omitempty"`
	StarType       *JSONProperty `json:"star_type,omitempty"`
	Points         *JSONProperty `json:"points,omitempty"`
	Rotation       *JSONProperty `json:"rotation,omitempty"`
	InnerRadius    *JSONProperty `json:"inner_radius,omitempty"`
	OuterRadius    *JSONProperty `json:"outer_radius,omitempty"`
	InnerRoundness *JSONProperty `json:"inner_roundness,omitempty"`
	OuterRoundness *JSONProperty `json:"outer_roundness,omitempty"`
}

type JSONShapePath struct {
	Name     string           `json:"name,omitempty"`
	Closed   bool             `json:"closed"`
	Vertices []JSONMaskVertex `json:"vertices,omitempty"`
}

// JSONKeyframe is the JSON representation of a Keyframe.
type JSONKeyframe struct {
	Time              float64            `json:"time_seconds"`
	Value             interface{}        `json:"value"`
	InInterp          string             `json:"in_interp,omitempty"`
	OutInterp         string             `json:"out_interp,omitempty"`
	InSpatialTangent  []float64          `json:"in_spatial_tangent,omitempty"`
	OutSpatialTangent []float64          `json:"out_spatial_tangent,omitempty"`
	InTemporalEase    []JSONTemporalEase `json:"in_temporal_ease,omitempty"`
	OutTemporalEase   []JSONTemporalEase `json:"out_temporal_ease,omitempty"`
}

// JSONFootage is the JSON representation of Footage.
type JSONFootage struct {
	ID        uint32  `json:"id"`
	Name      string  `json:"name"`
	Path      string  `json:"path,omitempty"`
	Width     uint16  `json:"width,omitempty"`
	Height    uint16  `json:"height,omitempty"`
	FrameRate float64 `json:"frame_rate,omitempty"`
	Duration  float64 `json:"duration_seconds,omitempty"`
	IsStill   bool    `json:"is_still,omitempty"`
	IsSolid   bool    `json:"is_solid,omitempty"`
}

// JSONFolder is the JSON representation of a Folder.
type JSONFolder struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

// ──────────────────────────────────────────────
// Conversion functions
// ──────────────────────────────────────────────

// ToJSON converts the Project to a JSON-serializable view.
func (p *Project) ToJSON() *JSONProject {
	jp := &JSONProject{}

	for _, c := range p.Compositions {
		jp.Compositions = append(jp.Compositions, compToJSON(c))
	}
	for _, f := range p.Footage {
		jp.Footage = append(jp.Footage, footageToJSON(f))
	}
	for _, f := range p.Folders {
		jp.Folders = append(jp.Folders, &JSONFolder{ID: f.ID, Name: f.Name})
	}
	if p.RenderQueue != nil {
		jp.RenderQueue = renderQueueToJSON(p.RenderQueue)
	}

	return jp
}

func formatOptionsToJSON(fo *FormatOptions) *JSONFormatOptions {
	if fo == nil {
		return nil
	}
	return &JSONFormatOptions{
		Kind:                  fo.Kind,
		TenBitBlackPoint:      fo.TenBitBlackPoint,
		TenBitWhitePoint:      fo.TenBitWhitePoint,
		ConvertedBlackPoint:   fo.ConvertedBlackPoint,
		ConvertedWhitePoint:   fo.ConvertedWhitePoint,
		CurrentGamma:          fo.CurrentGamma,
		HighlightExpansion:    fo.HighlightExpansion,
		LogarithmicConversion: fo.LogarithmicConversion,
		CineonFileFormat:      fo.CineonFileFormat,
		Quality:               fo.Quality,
		ThirtyTwoBitFloat:     fo.ThirtyTwoBitFloat,
		LuminanceChroma:       fo.LuminanceChroma,
		BitsPerPixel:          fo.BitsPerPixel,
		RLECompression:        fo.RLECompression,
		LZWCompression:        fo.LZWCompression,
		IBMPCByteOrder:        fo.IBMPCByteOrder,
		Width:                 fo.Width,
		Height:                fo.Height,
		BitDepth:              fo.BitDepth,
		Compression:           fo.Compression,
	}
}

func renderQueueToJSON(rq *RenderQueue) *JSONRenderQueue {
	jrq := &JSONRenderQueue{NumItems: rq.NumItems()}
	for _, it := range rq.Items {
		ji := &JSONRenderQueueItem{
			Status:           it.Status,
			Name:             it.Name,
			Comment:          it.Comment,
			LogType:          it.LogType,
			QueueItemNotify:  it.QueueItemNotify,
			ElapsedSeconds:   it.ElapsedSeconds,
			TimeSpanStart:    roundFloat(it.TimeSpanStart, 4),
			TimeSpanDuration: roundFloat(it.TimeSpanDuration, 4),
			RenderSettings: JSONRenderSettings{
				Quality:           it.RenderSettings.Quality,
				ColorDepth:        it.RenderSettings.ColorDepth,
				Effects:           it.RenderSettings.Effects,
				FieldRender:       it.RenderSettings.FieldRender,
				Pulldown:          it.RenderSettings.Pulldown,
				FrameBlending:     it.RenderSettings.FrameBlending,
				MotionBlur:        it.RenderSettings.MotionBlur,
				ProxyUse:          it.RenderSettings.ProxyUse,
				SoloSwitches:      it.RenderSettings.SoloSwitches,
				GuideLayers:       it.RenderSettings.GuideLayers,
				DiskCache:         it.RenderSettings.DiskCache,
				FrameRate:         it.RenderSettings.FrameRate,
				Resolution:        it.RenderSettings.Resolution,
				SkipExistingFiles: it.RenderSettings.SkipExistingFiles,
			},
		}
		if it.Comp != nil {
			ji.CompName = it.Comp.Name
		}
		for _, om := range it.OutputModules {
			ji.OutputModules = append(ji.OutputModules, &JSONOutputModule{
				Name:         om.Name,
				FileTemplate: om.FileTemplate,
				FullPath:     om.FullPath,
				Settings: JSONOutputModuleSettings{
					Channels:            om.Settings.Channels,
					ResizeQuality:       om.Settings.ResizeQuality,
					Resize:              om.Settings.Resize,
					LockAspectRatio:     om.Settings.LockAspectRatio,
					Crop:                om.Settings.Crop,
					CropTop:             om.Settings.CropTop,
					CropLeft:            om.Settings.CropLeft,
					CropBottom:          om.Settings.CropBottom,
					CropRight:           om.Settings.CropRight,
					OutputAudio:         om.Settings.OutputAudio,
					IncludeProjectLink:  om.Settings.IncludeProjectLink,
					PostRenderAction:    om.Settings.PostRenderAction,
					ConvertToLinear:     om.Settings.ConvertToLinear,
					UseCompFrameNumber:  om.Settings.UseCompFrameNumber,
					UseRegionOfInterest: om.Settings.UseRegionOfInterest,
					IncludeSourceXMP:    om.Settings.IncludeSourceXMP,
					PreserveRGB:         om.Settings.PreserveRGB,
					VideoCodec:          om.Settings.VideoCodec,
					FormatID:            om.Settings.FormatID,
					StartingNumber:      om.Settings.StartingNumber,
					Width:               om.Settings.Width,
					Height:              om.Settings.Height,
					Depth:               om.Settings.Depth,
					VideoOutput:         om.Settings.VideoOutput,
					AudioSampleRate:     om.Settings.AudioSampleRate,
					AudioBitDepth:       om.Settings.AudioBitDepth,
					AudioChannels:       om.Settings.AudioChannels,
					AudioEnabled:        om.Settings.AudioEnabled,
				},
				FormatOptions: formatOptionsToJSON(om.FormatOptions),
			})
		}
		jrq.Items = append(jrq.Items, ji)
	}
	return jrq
}

func compToJSON(c *Composition) *JSONComposition {
	jc := &JSONComposition{
		ID:                            c.ID,
		Name:                          c.Name,
		Width:                         c.Width,
		Height:                        c.Height,
		FrameRate:                     roundFloat(c.FrameRate, 3),
		Duration:                      roundFloat(c.Duration, 4),
		TickRate:                      c.TickRate,
		BGColor:                       fmt.Sprintf("#%02X%02X%02X", c.BGColor[0], c.BGColor[1], c.BGColor[2]),
		ResolutionFactor:              c.ResolutionFactor,
		Renderer:                      c.Renderer,
		WorkAreaStart:                 roundFloat(c.WorkAreaStart, 4),
		WorkAreaEnd:                   roundFloat(c.WorkAreaEnd, 4),
		ShutterAngle:                  c.ShutterAngle,
		ShutterPhase:                  c.ShutterPhase,
		MotionBlurAdaptiveSampleLimit: c.MotionBlurAdaptiveSampleLimit,
		MotionBlurSamplesPerFrame:     c.MotionBlurSamplesPerFrame,
	}
	for _, l := range c.Layers {
		jc.Layers = append(jc.Layers, layerToJSON(l))
	}
	for _, m := range c.Markers {
		jc.Markers = append(jc.Markers, &JSONMarker{
			Time:         roundFloat(m.Time, 4),
			Duration:     roundFloat(m.Duration, 4),
			Label:        m.Label,
			Comment:      m.Comment,
			Chapter:      m.Chapter,
			URL:          m.URL,
			FrameTarget:  m.FrameTarget,
			CuePointName: m.CuePointName,
		})
	}
	for _, g := range c.Guides {
		jc.Guides = append(jc.Guides, &JSONGuide{
			Orientation: g.Orientation.String(),
			Position:    roundFloat(g.Position, 4),
		})
	}
	jc.MotionGraphicsTemplateName = c.MotionGraphicsTemplateName
	for _, ctrl := range c.EssentialGraphicsControllers {
		jc.EssentialGraphics = append(jc.EssentialGraphics, &JSONEGController{
			Name: ctrl.Name,
			Type: ctrl.Type.String(),
			UUID: ctrl.UUID,
		})
	}
	return jc
}

func layerToJSON(l *Layer) *JSONLayer {
	jl := &JSONLayer{
		Index:    l.Index,
		Name:     l.Name,
		Type:     string(l.Type),
		ID:       l.ID,
		ParentID: l.ParentID,
		SourceID: l.SourceID,
		ParentName: func() string {
			if p := l.Parent(); p != nil {
				return p.Name
			}
			return ""
		}(),
		StartTime:    roundFloat(l.StartTime, 4),
		Duration:     roundFloat(l.Duration, 4),
		Stretch:      roundFloat(l.Stretch, 4),
		Quality:      uint16(l.Quality),
		Label:        l.Label,
		BlendingMode: uint8(l.BlendingMode),
		TrackMatte:   uint8(l.TrackMatte),
		LightKind: func() string {
			if l.Type != LayerTypeLight {
				return ""
			}
			return l.LightKind.String()
		}(),
		PreserveTransparency: l.PreserveTransparency,
		Comment:              l.Comment,
		AutoOrient: func() string {
			if l.AutoOrient == AutoOrientNone {
				return ""
			}
			return l.AutoOrient.String()
		}(),
		Is3D:                  l.Is3D,
		Solo:                  l.Solo,
		Shy:                   l.Shy,
		Locked:                l.Locked,
		Visible:               l.Visible,
		IsAdjust:              l.IsAdjust,
		IsNull:                l.IsNull,
		IsGuide:               l.IsGuide,
		MarkersLocked:         l.MarkersLocked,
		MotionBlur:            l.MotionBlur,
		EffectsEnabled:        l.EffectsEnabled,
		AudioEnabled:          l.AudioEnabled,
		FrameBlendEnabled:     l.FrameBlendEnabled,
		FrameBlendPixelMotion: l.FrameBlendPixelMotion,
		CollapseTransform:     l.CollapseTransform,
		SamplingBicubic:       l.SamplingBicubic,
		IsShapeLayer:          l.IsShapeLayer,
		HasTextSource:         l.TextSourceRaw != nil,
	}
	if l.TextSource != nil {
		jts := &JSONTextSource{
			Text:          l.TextSource.Text,
			Fonts:         l.TextSource.Fonts,
			FontAxes:      l.TextSource.FontAxes,
			Justification: l.TextSource.Justification.String(),
			IsBoxText:     l.TextSource.IsBoxText,
			BoxBounds:     l.TextSource.BoxBounds,
		}
		for _, r := range l.TextSource.Runs {
			jts.Runs = append(jts.Runs, JSONTextStyleRun{
				FontIndex:       r.FontIndex,
				FontName:        r.FontName,
				FontSize:        r.FontSize,
				FillColor:       r.FillColor,
				FauxBold:        r.FauxBold,
				FauxItalic:      r.FauxItalic,
				AutoLeading:     r.AutoLeading,
				Leading:         r.Leading,
				Tracking:        r.Tracking,
				BaselineShift:   r.BaselineShift,
				HorizontalScale: r.HorizontalScale,
				VerticalScale:   r.VerticalScale,
				Tsume:           r.Tsume,
				ApplyStroke:     r.ApplyStroke,
				StrokeColor:     r.StrokeColor,
				StrokeWidth:     r.StrokeWidth,
				StrokeOverFill:  r.StrokeOverFill,
				NoBreak:         r.NoBreak,
			})
		}
		for _, p := range l.TextSource.Paragraphs {
			jts.Paragraphs = append(jts.Paragraphs, JSONTextParagraph{
				Justification: p.Justification.String(),
			})
		}
		jl.TextSource = jts
	}
	for _, m := range l.Masks {
		jm := &JSONMask{
			Name:           m.Name,
			Index:          m.Index,
			Mode:           m.Mode.String(),
			Inverted:       m.Inverted,
			Locked:         m.Locked,
			Color:          fmt.Sprintf("#%02X%02X%02X", m.Color[0], m.Color[1], m.Color[2]),
			MotionBlur:     maskMotionBlurName(m.MotionBlur),
			FeatherFalloff: maskFeatherFalloffName(m.FeatherFalloff),
			Closed:         m.Closed,
			Feather:        m.Feather,
			Opacity:        m.Opacity,
			Expansion:      m.Expansion,
		}
		for _, v := range m.Vertices {
			jm.Vertices = append(jm.Vertices, JSONMaskVertex{
				Anchor: v.Anchor, InTangent: v.InTangent, OutTangent: v.OutTangent,
			})
		}
		for _, kf := range m.PathKeyframes {
			jpk := &JSONMaskPathKeyframe{
				Time:      roundFloat(kf.Time, 4),
				InInterp:  kf.InInterp.String(),
				OutInterp: kf.OutInterp.String(),
			}
			if kf.InInterp == InterpBezier || kf.OutInterp == InterpBezier {
				jpk.InTemporalEase = &JSONTemporalEase{Speed: kf.InTemporalEase.Speed, Influence: kf.InTemporalEase.Influence}
				jpk.OutTemporalEase = &JSONTemporalEase{Speed: kf.OutTemporalEase.Speed, Influence: kf.OutTemporalEase.Influence}
			}
			for _, v := range kf.Vertices {
				jpk.Vertices = append(jpk.Vertices, JSONMaskVertex{
					Anchor: v.Anchor, InTangent: v.InTangent, OutTangent: v.OutTangent,
				})
			}
			jm.PathKeyframes = append(jm.PathKeyframes, jpk)
		}
		jl.Masks = append(jl.Masks, jm)
	}
	for _, sp := range l.ShapePaths {
		jsp := &JSONShapePath{Name: sp.Name, Closed: sp.Closed}
		for _, v := range sp.Vertices {
			jsp.Vertices = append(jsp.Vertices, JSONMaskVertex{
				Anchor: v.Anchor, InTangent: v.InTangent, OutTangent: v.OutTangent,
			})
		}
		jl.ShapePaths = append(jl.ShapePaths, jsp)
	}
	for _, prim := range l.ShapePrimitives {
		jp := &JSONShapePrimitive{Kind: string(prim.Kind), GroupName: prim.GroupName}
		setIfNonNil := func(dst **JSONProperty, src *Property) {
			if src != nil {
				*dst = propertyToJSON(src)
			}
		}
		setIfNonNil(&jp.Size, prim.Size)
		setIfNonNil(&jp.Position, prim.Position)
		setIfNonNil(&jp.Roundness, prim.Roundness)
		setIfNonNil(&jp.StarType, prim.StarType)
		setIfNonNil(&jp.Points, prim.Points)
		setIfNonNil(&jp.Rotation, prim.Rotation)
		setIfNonNil(&jp.InnerRadius, prim.InnerRadius)
		setIfNonNil(&jp.OuterRadius, prim.OuterRadius)
		setIfNonNil(&jp.InnerRoundness, prim.InnerRoundness)
		setIfNonNil(&jp.OuterRoundness, prim.OuterRoundness)
		jl.ShapePrimitives = append(jl.ShapePrimitives, jp)
	}
	for _, prop := range l.Properties {
		jl.Properties = append(jl.Properties, propertyToJSON(prop))
	}
	for _, fx := range l.Effects {
		je := &JSONEffect{MatchName: fx.MatchName, Name: fx.Name}
		for _, prop := range fx.Parameters {
			je.Parameters = append(je.Parameters, propertyToJSON(prop))
		}
		jl.Effects = append(jl.Effects, je)
	}
	for _, m := range l.Markers {
		jl.Markers = append(jl.Markers, &JSONMarker{
			Time:         roundFloat(m.Time, 4),
			Duration:     roundFloat(m.Duration, 4),
			Label:        m.Label,
			Comment:      m.Comment,
			Chapter:      m.Chapter,
			URL:          m.URL,
			FrameTarget:  m.FrameTarget,
			CuePointName: m.CuePointName,
		})
	}
	return jl
}

func propertyToJSON(prop *Property) *JSONProperty {
	jp := &JSONProperty{
		Name:        prop.Name,
		MatchName:   prop.MatchName,
		StaticValue: prop.StaticValue,
		LayerRefID:  prop.LayerRefID,
		Expression:  prop.Expression,
	}
	if prop.Expression != "" {
		enabled := prop.ExpressionEnabled
		jp.ExpressionEnabled = &enabled
	}
	for _, kf := range prop.Keyframes {
		jkf := &JSONKeyframe{
			Time:              roundFloat(kf.Time, 4),
			Value:             kf.Value,
			InInterp:          kf.InInterp.String(),
			OutInterp:         kf.OutInterp.String(),
			InSpatialTangent:  kf.InSpatialTangent,
			OutSpatialTangent: kf.OutSpatialTangent,
		}
		for _, e := range kf.InTemporalEase {
			jkf.InTemporalEase = append(jkf.InTemporalEase, JSONTemporalEase{Speed: e.Speed, Influence: e.Influence})
		}
		for _, e := range kf.OutTemporalEase {
			jkf.OutTemporalEase = append(jkf.OutTemporalEase, JSONTemporalEase{Speed: e.Speed, Influence: e.Influence})
		}
		jp.Keyframes = append(jp.Keyframes, jkf)
	}
	return jp
}

func maskMotionBlurName(value MaskMotionBlurMode) string {
	switch value {
	case MaskMotionBlurOn:
		return "on"
	case MaskMotionBlurOff:
		return "off"
	default:
		return "same_as_layer"
	}
}

func maskFeatherFalloffName(value MaskFeatherFalloff) string {
	switch value {
	case MaskFeatherFalloffLinear:
		return "linear"
	default:
		return "smooth"
	}
}

func footageToJSON(f *Footage) *JSONFootage {
	return &JSONFootage{
		ID:        f.ID,
		Name:      f.Name,
		Path:      f.Path,
		Width:     f.Width,
		Height:    f.Height,
		FrameRate: roundFloat(f.FrameRate, 3),
		Duration:  roundFloat(f.Duration, 4),
		IsStill:   f.IsStill,
		IsSolid:   f.IsSolid,
	}
}

// MarshalJSON serializes the Project as JSON.
func (p *Project) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.ToJSON())
}

// @summary     Write the project as indented JSON
// @param       w  the destination writer
// @domain      io
// @stability   stable
// @verify      roundtrip
// @since       AE2020
// @boundary    one-way export — there is no corresponding ReadJSON. Output is deterministic and checked against golden fixtures; this path does not involve AE.
// @alias       write json,json 导出,export json,序列化
func (p *Project) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(p.ToJSON())
}

func roundFloat(f float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(f*pow) / pow
}
