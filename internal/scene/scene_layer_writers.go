package scene

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/codec"
)

// Layer setters covered in this file are all length-preserving single-byte
// or single-bit edits inside the layer's ldta chunk. Each setter:
//
//   - validates the layer has an underlying ldta chunk (else returns an
//     error — happens for layers built outside the parser);
//   - delegates the byte-patch work to layerBackrefs (LayerWriter);
//   - keeps the matching Go field in sync so later reads see the new value.
//
// The next call to Project.WriteAEP serializes the change.

// @summary     Set the layer's video (visibility) switch
// @param       v  the new visibility state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x27
// @alias       visible,visibility,显示,隐藏,视频开关,eye toggle
func (l *Layer) SetVisible(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetVisible(v); err != nil {
		return err
	}
	l.Visible = v
	return nil
}

// @summary     Set the layer's Solo flag
// @param       v  the new solo state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x26
// @alias       solo,独奏,单独显示
func (l *Layer) SetSolo(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetSolo(v); err != nil {
		return err
	}
	l.Solo = v
	return nil
}

// @summary     Set the layer's Shy flag
// @description Shy hides the layer from the shy-filter view in the timeline.
// @param       v  the new shy state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x27
// @alias       shy,羞涩,隐藏图层,shy filter
func (l *Layer) SetShy(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetShy(v); err != nil {
		return err
	}
	l.Shy = v
	return nil
}

// @summary     Set the layer's Lock flag
// @description When locked, AE refuses edits in the timeline UI; the
//   underlying file is still mutable through this API.
// @param       v  the new lock state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x27
// @alias       locked,lock,锁定,锁,图层锁定
func (l *Layer) SetLocked(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetLocked(v); err != nil {
		return err
	}
	l.Locked = v
	return nil
}

// @summary     Set the layer's effects (fx) switch
// @param       v  whether effects render on this layer
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerFlags2_AEShipGate_AE2020,TestLayerFlags2_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x27
// @alias       effects enabled,fx switch,特效开关,效果启用
func (l *Layer) SetEffectsEnabled(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetEffectsEnabled(v); err != nil {
		return err
	}
	l.EffectsEnabled = v
	return nil
}

// @summary     Set the layer's motion-blur switch
// @param       v  the new motion-blur state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x27
// @alias       motion blur,运动模糊,模糊开关
func (l *Layer) SetMotionBlur(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetMotionBlur(v); err != nil {
		return err
	}
	l.MotionBlur = v
	return nil
}

// @summary     Set the layer's audio switch
// @param       v  the new audio-enabled state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAudio_AEShipGate_AE2020,TestLayerAudio_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x27; this is a
//   non-visual switch so ae-accept verification is read back from the DOM
//   audioEnabled property rather than a rendered pixel
// @alias       audio enabled,audio switch,音频开关,静音
func (l *Layer) SetAudioEnabled(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetAudioEnabled(v); err != nil {
		return err
	}
	l.AudioEnabled = v
	return nil
}

// @summary     Set the layer's frame-blend switch
// @param       v  the new frame-blend state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerFrameBlend_AEShipGate_AE2020,TestLayerFrameBlend_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x27; only has an
//   effect on layers with time-based frames (video footage or nested comps)
// @alias       frame blend,帧混合,帧融合,frame blending
func (l *Layer) SetFrameBlendEnabled(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetFrameBlendEnabled(v); err != nil {
		return err
	}
	l.FrameBlendEnabled = v
	return nil
}

// @summary     Set "Collapse Transformations" / "Continuously Rasterize"
// @description Collapse Transformations applies to nested-comp layers;
//   Continuously Rasterize applies to Illustrator or shape layers. Both
//   share the same underlying switch bit.
// @param       v  the new collapse/rasterize state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerBool_AEShipGate_AE2020,TestLayerBool_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x27
// @alias       collapse transform,continuously rasterize,折叠变换,连续栅格化
func (l *Layer) SetCollapseTransform(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetCollapseTransform(v); err != nil {
		return err
	}
	l.CollapseTransform = v
	return nil
}

// @summary     Set the layer's 3D-layer switch
// @param       v  the new 3D state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayer3DEnable_AEShipGate_AE2020,TestLayer3DEnable_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x26
// @alias       3D layer,3D开关,三维图层,enable 3D
func (l *Layer) SetIs3D(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetIs3D(v); err != nil {
		return err
	}
	l.Is3D = v
	return nil
}

// @summary     Set the layer's "Adjustment Layer" switch
// @param       v  the new adjustment-layer state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x26
// @alias       adjustment layer,调整图层,调节层
func (l *Layer) SetIsAdjust(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetIsAdjust(v); err != nil {
		return err
	}
	l.IsAdjust = v
	return nil
}

// @summary     Set the layer's "Guide Layer" switch
// @description A guide layer renders in the comp viewer but is excluded
//   from rendered output.
// @param       v  the new guide-layer state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x25
// @alias       guide layer,参考线图层,辅助线
func (l *Layer) SetIsGuide(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetIsGuide(v); err != nil {
		return err
	}
	l.IsGuide = v
	return nil
}

// @summary     Set the Null-Object marker bit
// @description AE's UI normally creates null layers via its own dedicated
//   command; flipping this bit post-hoc is accepted by AE but produces
//   uncommon results — only set it when you understand the consequences.
// @param       v  the new null-object state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerFlags2_AEShipGate_AE2020,TestLayerFlags2_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x26
// @alias       null layer,空对象,null object
func (l *Layer) SetIsNull(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetIsNull(v); err != nil {
		return err
	}
	l.IsNull = v
	return nil
}

// @summary     Set "Lock markers" on the layer
// @param       v  the new markers-locked state
// @domain      layer-set
// @stability   stable
// @verify      roundtrip
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x26
// @alias       markers locked,标记锁定,lock markers
func (l *Layer) SetMarkersLocked(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetMarkersLocked(v); err != nil {
		return err
	}
	l.MarkersLocked = v
	return nil
}

// @summary     Switch the layer's sampling between Bilinear and Bicubic
// @param       v  true selects Bicubic, false selects Bilinear
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x25
// @alias       sampling bicubic,bicubic,bilinear,采样方式,图像质量
func (l *Layer) SetSamplingBicubic(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetSamplingBicubic(v); err != nil {
		return err
	}
	l.SamplingBicubic = v
	return nil
}

// @summary     Switch frame-blend mode between Frame Mix and Pixel Motion
// @description Only takes effect when FrameBlendEnabled is also true.
// @param       v  true selects Pixel Motion, false selects Frame Mix
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerFrameBlend_AEShipGate_AE2020,TestLayerFrameBlend_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at ldta offset 0x25; requires
//   FrameBlendEnabled=true to have any visible effect
// @alias       pixel motion,frame mix,帧混合模式,像素运动
func (l *Layer) SetFrameBlendPixelMotion(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetFrameBlendPixelMotion(v); err != nil {
		return err
	}
	l.FrameBlendPixelMotion = v
	return nil
}

// SetShy / Solo / Locked / etc. cover bool toggles. Multi-value byte
// fields follow.

// @summary     Set the layer's blending mode
// @param       m  the new blending mode
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single byte at ldta offset 0x63
// @alias       blending mode,混合模式,叠加模式,blend mode
func (l *Layer) SetBlendingMode(m BlendingMode) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.runtime.back.SetBlendingMode(m); err != nil {
		return err
	}
	l.BlendingMode = m
	return nil
}

// @summary     Set the layer's classic track-matte type
// @description AE additionally requires the matte source layer to sit
//   immediately above this layer in the comp; this setter only changes
//   the mode byte and does not reorder layers.
// @param       t  the new track-matte type
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestTrackMatteClassic_AEShipGate_AE2020,TestTrackMatteClassic_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single byte at ldta offset 0x6B; AE 2025
//   auto-migrates classic track mattes to the explicit form on load
// @alias       track matte,matte type,遮罩,轨道遮罩,alpha matte,luma matte
func (l *Layer) SetTrackMatte(t TrackMatteType) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.runtime.back.SetTrackMatte(t); err != nil {
		return err
	}
	l.TrackMatte = t
	return nil
}

// @summary     Set the layer's timeline label-color index
// @description Indices outside the normal 0..16 range are written
//   verbatim; AE shows index 0 for any unknown value but the byte itself
//   is preserved on round-trip.
// @param       index  the new label-color index (0..16)
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single byte at ldta offset 0x3D
// @alias       label,label color,标签,颜色标签,时间线颜色
func (l *Layer) SetLabel(index uint8) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.runtime.back.SetLabel(index); err != nil {
		return err
	}
	l.Label = index
	return nil
}

// @summary     Set the layer's render-quality (Wireframe/Draft/Best)
// @param       q  the new render quality
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a big-endian uint16 at ldta offset 0x04
// @alias       quality,render quality,渲染质量,线框,草图,最佳质量
func (l *Layer) SetQuality(q LayerQuality) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.runtime.back.SetQuality(q); err != nil {
		return err
	}
	l.Quality = q
	return nil
}

// @summary     Set the layer's parent-layer ID
// @description Pass 0 to clear the parent ("no parent", shown as "None"
//   in the timeline). When the layer lives inside a composition, the new
//   parent ID is validated against same-comp layers — an ID that doesn't
//   resolve returns an error and leaves the bytes untouched. Cross-comp
//   parenting isn't allowed. Setting parentID equal to the layer's own ID
//   is rejected since it would create a self-parent cycle.
// @param       parentID  the new parent layer ID (0 clears the parent)
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFields3_AEShipGate_AE2020,TestLayerAVFields3_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a 4-byte field at ldta offset 0x84
// @alias       parent,parent layer,父图层,父级,layer parenting
func (l *Layer) SetParent(parentID uint32) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if parentID != 0 && parentID == l.ID {
		return fmt.Errorf("layer %q: self-parenting (parentID == own ID = %d) not allowed", l.Name, l.ID)
	}
	if parentID != 0 && l.runtime.comp != nil {
		if l.runtime.comp.LayerByID(parentID) == nil {
			return fmt.Errorf("layer %q: parentID %d not found in comp %q", l.Name, parentID, l.runtime.comp.Name)
		}
	}
	if err := l.runtime.back.SetParent(parentID); err != nil {
		return err
	}
	l.ParentID = parentID
	return nil
}

// @summary     Set the layer's source-item ID
// @description For regular AV layers this is the footage or pre-comp ID.
//   When the owning project is known, sourceID is validated to match an
//   existing Composition or Footage item — an ID that doesn't resolve
//   returns an error and leaves the bytes untouched. Pass 0 to clear
//   (rare; usually turns the layer into a Null-like "no source" stub).
//   AE caches some source-derived metadata (Width/Height) on the layer at
//   render time rather than in ldta, so changing the source preserves the
//   rest of ldta verbatim.
// @param       sourceID  the new source-item ID (0 clears it)
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerSource_AEShipGate_AE2020,TestLayerSource_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a 4-byte field at ldta offset 0x28
// @alias       source,source id,footage source,替换素材,层来源
func (l *Layer) SetSource(sourceID uint32) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if sourceID != 0 && l.runtime.comp != nil && l.runtime.comp.proj != nil {
		p := l.runtime.comp.proj
		if p.CompositionByID(sourceID) == nil && !footageWithID(p, sourceID) {
			return fmt.Errorf("layer %q: sourceID %d not found in project items", l.Name, sourceID)
		}
	}
	if err := l.runtime.back.SetSource(sourceID); err != nil {
		return err
	}
	l.SourceID = sourceID
	return nil
}

// footageWithID reports whether any Footage item has the given ID.
// Used by SetSource for source-ID validation.
func footageWithID(p *Project, id uint32) bool {
	for _, f := range p.Footage {
		if f.ID == id {
			return true
		}
	}
	return false
}

// @summary     Set the layer's auto-orient mode
// @description Clears the three mutually-exclusive auto-orient bits
//   spread across ldta offsets 0x25/0x26 and sets the one matching the
//   requested mode. Unknown enum values are treated as None (a no-op
//   clearing of all three bits). The AlongPath mode only has a visible
//   effect when the layer's position has a motion path (keyframes).
// @param       t  the new auto-orient mode
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerBool_AEShipGate_AE2020,TestLayerBool_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; touches a mutually-exclusive bit group
//   spread across ldta offsets 0x25 and 0x26
// @alias       auto orient,自动旋转,沿路径旋转,朝向摄像机
func (l *Layer) SetAutoOrient(t AutoOrientType) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.runtime.back.SetAutoOrient(t); err != nil {
		return err
	}
	l.AutoOrient = t
	return nil
}

// @summary     Set "Preserve Underlying Transparency"
// @param       v  the new preserve-transparency state
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single byte at ldta offset 0x67
// @alias       preserve transparency,保留透明度,alpha preservation
func (l *Layer) SetPreserveTransparency(v bool) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.runtime.back.SetPreserveTransparency(v); err != nil {
		return err
	}
	l.PreserveTransparency = v
	return nil
}

// ──────────────────────────────────────────────────────────────────
// Time-field setters — dividend/divisor pairs in ldta. All length-
// preserving (8 bytes each, fixed offsets). We keep the existing
// divisor when non-zero (typical: comp's tick rate) and fall back to
// 600 when missing, matching how AE writes these fields.
// ──────────────────────────────────────────────────────────────────

// @summary     Set the layer's start time
// @description AE allows negative start times (pre-roll). Mirrors the
//   change into Layer.StartTime.
// @param       seconds  the new start time, in seconds
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFields3_AEShipGate_AE2020,TestLayerAVFields3_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, an 8-byte dividend/divisor pair at ldta
//   offsets 0x0C/0x10; negative values (pre-roll) are supported
// @alias       start time,开始时间,图层入点,layer offset
func (l *Layer) SetStartTime(seconds float64) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.runtime.back.SetStartTime(seconds); err != nil {
		return err
	}
	l.StartTime = l.readLdtaFrac(0x0C)
	return nil
}

// @summary     Set the layer's source-media in-point
// @description Refreshes Layer.Duration (= out minus in) after the write.
// @param       seconds  the new in-point, in seconds
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, an 8-byte dividend/divisor pair at ldta
//   offsets 0x14/0x18
// @alias       in point,入点,开始帧,trim start
func (l *Layer) SetInPoint(seconds float64) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.runtime.back.SetInPoint(seconds); err != nil {
		return err
	}
	l.recomputeDurationFromLdta()
	return nil
}

// @summary     Set the layer's source-media out-point
// @description Refreshes Layer.Duration after the write.
// @param       seconds  the new out-point, in seconds
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, an 8-byte dividend/divisor pair at ldta
//   offsets 0x1C/0x20
// @alias       out point,出点,结束帧,trim end
func (l *Layer) SetOutPoint(seconds float64) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.runtime.back.SetOutPoint(seconds); err != nil {
		return err
	}
	l.recomputeDurationFromLdta()
	return nil
}

// recomputeDurationFromLdta re-reads in/out points from ldta and
// updates `Layer.Duration`. Called by SetInPoint / SetOutPoint so the
// in-memory Duration stays consistent with the underlying bytes.
func (l *Layer) recomputeDurationFromLdta() {
	if l.runtime.back == nil {
		return
	}
	in, okIn := l.runtime.back.LdtaFrac(0x14)
	out, okOut := l.runtime.back.LdtaFrac(0x1C)
	if !okIn || !okOut {
		return
	}
	l.Duration = out - in
}

// @summary     Set the layer's time-stretch ratio
// @description 1.0 is normal speed, 2.0 is 2x slow. Unlike the other time
//   fields, the dividend/divisor pair is split across ldta — the dividend
//   lives near the top, the divisor in the trailing section.
// @param       ratio  the new time-stretch ratio (1.0 = normal speed)
// @domain      layer-set
// @stability   alpha
// @verify      roundtrip
// @since       AE2020
// @boundary    confirmed false-green: the dividend/divisor pair round-trips
//   through this library, but AE recomputes the stretch value from the
//   in/out point span rather than trusting these bytes directly, so the
//   written ratio is not honored unless the in/out points are coordinated
//   to match. See incident layer-setstretch-ae-recomputes-span.
// @incident    layer-setstretch-ae-recomputes-span
// @alias       stretch,time stretch,速度,时间拉伸,慢动作,快放
func (l *Layer) SetStretch(ratio float64) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.runtime.back.SetStretch(ratio); err != nil {
		return err
	}
	if v, ok := l.runtime.back.StretchFrac(); ok {
		l.Stretch = v
	}
	return nil
}

// @summary     Set the layer's display name
// @description Returns an error when the parsed layer has no Utf8 name
//   chunk (unusual — most AE-written layers have one even for default
//   names).
// @param       newName  the new display name shown in the AE timeline
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFields3_AEShipGate_AE2020,TestLayerAVFields3_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable: the underlying Utf8 chunk's data slice is
//   replaced wholesale and ancestor LIST sizes are recomputed on write
// @alias       layer name,图层名,重命名,rename layer
func (l *Layer) SetName(newName string) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no Utf8 name chunk to mutate", l.Name)
	}
	if err := l.runtime.back.SetName(newName); err != nil {
		return err
	}
	l.Name = newName
	return nil
}

// @summary     Set the layer's comment text
// @description Mirrors AE's "Comments" timeline column / Layer Settings
//   dialog. Encoding mirrors what AE writes: LF becomes CRLF plus a
//   single NUL terminator.
// @param       comment  the new comment text
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAVFields3_AEShipGate_AE2020,TestLayerAVFields3_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable: the underlying cmta chunk's data is
//   replaced, or a new cmta chunk is inserted when none existed; both the
//   cmta double-NUL terminator and the ldta 0x3C has-comment flag must be
//   set together, otherwise AE reads back an empty comment despite the
//   bytes being correct
// @incident    layer-setcomment-cmta-append-position
// @alias       comment,注释,备注,layer comment,图层注释
func (l *Layer) SetComment(comment string) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no chunk backrefs (built outside parser?)", l.Name)
	}
	if err := l.runtime.back.SetComment(comment); err != nil {
		return err
	}
	l.Comment = comment
	return nil
}

// @summary     Set the layer's explicit track-matte source and mode
// @description Pass sourceID=0 with mode=TrackMatteNone to clear the
//   matte. When the layer lives inside a parsed composition, sourceID is
//   validated against same-comp layers — an ID that doesn't resolve
//   returns an error and leaves the bytes untouched; the matte source
//   must live in the same comp. Setting sourceID equal to the layer's own
//   ID is rejected since self-matte is a no-op in AE and usually
//   indicates a mistake. Passing mode=TrackMatteNone together with a
//   non-zero sourceID means "preserve the target but apply no matte yet"
//   — the source pointer stays while the matte channel stays disabled.
//   On older project files whose ldta layout is shorter, this call errors
//   because the explicit track-matte slot doesn't exist; re-saving the
//   file through a newer AE version first extends ldta so the setter can
//   succeed.
// @param       sourceID  the new track-matte source layer ID (0 clears it)
// @param       mode      the new track-matte mode
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestTrackMatteExplicit_AEShipGate_AE2025
// @since       AE2025
// @boundary    length-preserving, a 4-byte field plus a 1-byte mode field
//   at ldta offset 0xA0/0x6B, present only on AE 23+ project layouts; the
//   ship gate is single-version because no intermediate AE target both
//   produces this layout and survives the next version's forward
//   compatibility check, so a two-version gate is not reachable
// @alias       track matte layer,matte source,遮罩来源,轨道遮罩图层
func (l *Layer) SetTrackMatteLayer(sourceID uint32, mode TrackMatteType) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if sourceID != 0 && sourceID == l.ID {
		return fmt.Errorf("layer %q: self-matte (sourceID == own ID = %d) not allowed", l.Name, l.ID)
	}
	if sourceID != 0 && l.runtime.comp != nil {
		if l.runtime.comp.LayerByID(sourceID) == nil {
			return fmt.Errorf("layer %q: sourceID %d not found in comp %q", l.Name, sourceID, l.runtime.comp.Name)
		}
	}
	if err := l.runtime.back.SetTrackMatteLayer(sourceID, mode); err != nil {
		return err
	}
	l.TrackMatteLayerID = sourceID
	l.TrackMatte = mode
	return nil
}

// @summary     Clear the layer's explicit track-matte source and mode
// @description Shorthand for SetTrackMatteLayer with a zero source ID
//   and TrackMatteNone mode. Removes the explicit matte source and the
//   matte mode in one call.
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestTrackMatteExplicit_AEShipGate_AE2025
// @since       AE2025
// @boundary    present only on AE 23+ project layouts, same single-version
//   gate constraint as SetTrackMatteLayer
// @alias       clear matte,remove matte,清除遮罩,取消轨道遮罩
func (l *Layer) ClearTrackMatteLayer() error {
	return l.SetTrackMatteLayer(0, TrackMatteNone)
}

// @summary     Set a Light layer's light kind
// @description Only meaningful when the layer's Type is Light; AE
//   silently accepts the byte change on non-light layers but it has no
//   visual effect. The field is present on older project layouts too, so
//   this setter works across the full supported AE range.
// @param       k  the new light kind
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayer3DLight_AEShipGate_AE2020,TestLayer3DLight_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a big-endian uint32 at ldta offset 0x88
// @alias       light kind,light type,灯光类型,平行光,点光源,聚光灯,环境光
func (l *Layer) SetLightKind(k LightKind) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.runtime.back.SetLightKind(k); err != nil {
		return err
	}
	l.LightKind = k
	return nil
}

// @summary     Set the environment-light source layer for a Light layer
// @description Stored in the same ldta slot used for the AV SourceID;
//   the field is repurposed depending on Layer.Type. Pass nil to clear,
//   which writes the clear sentinel. Validation requires: the layer must
//   be a Light layer; the target, if non-nil, must be in the same
//   composition; the target must not itself be a Light or Camera layer;
//   the target must not be a 3D layer (3D AV layers cannot drive
//   environment lights); and the target must not be the layer itself.
//   On any validation error the bytes are left unmodified.
// @param       target  the new environment-light source layer (nil clears it)
// @domain      layer-set
// @stability   stable
// @verify      roundtrip
// @since       AE2025
// @boundary    length-preserving, a 4-byte field at ldta offset 0x28;
//   only applies to environment lights on newer AE versions
// @alias       light source,environment light,环境光源,灯光来源
func (l *Layer) SetLightSource(target *Layer) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if l.Type != LayerTypeLight {
		return fmt.Errorf("layer %q: SetLightSource only valid for Light layers (Type=%q)", l.Name, l.Type)
	}
	if target == nil {
		if err := l.runtime.back.SetLightSource(nil); err != nil {
			return err
		}
		l.SourceID = codec.LightSourceUndefined
		return nil
	}
	if target == l || (target.ID != 0 && target.ID == l.ID) {
		return fmt.Errorf("layer %q: SetLightSource self-source not allowed", l.Name)
	}
	if target.ID == 0 {
		return fmt.Errorf("layer %q: SetLightSource target has zero ID", l.Name)
	}
	if l.runtime.comp != nil && target.runtime.comp != nil && target.runtime.comp != l.runtime.comp {
		return fmt.Errorf("layer %q: SetLightSource target must be in same composition", l.Name)
	}
	if target.Type == LayerTypeLight || target.Type == LayerTypeCamera {
		return fmt.Errorf("layer %q: SetLightSource target cannot be Light/Camera (got %q)", l.Name, target.Type)
	}
	if target.Is3D {
		return fmt.Errorf("layer %q: SetLightSource target cannot be a 3D layer", l.Name)
	}
	if err := l.runtime.back.SetLightSource(target); err != nil {
		return err
	}
	l.SourceID = target.ID
	return nil
}

// @summary     Override the layer's source via Essential Properties
// @description Overrides this layer's source through the Essential
//   Properties media-replacement slot. Pass nil (or a zero-id item) to
//   clear the override — equivalent to ClearAlternateSource. Requires the
//   layer to already have an Essential Properties slot: AE only persists
//   the underlying chunk after the source-side layer has been promoted to
//   a motion-graphics template inside its wrapper precomp. If the slot is
//   missing, this errors rather than attempting a structural chunk
//   insertion that would break length-preserving round-trip. When the
//   layer's owning project is known, the item's ID is validated against
//   the project's AV items; an unknown ID returns an error and leaves the
//   bytes untouched. AE additionally rejects items that are not
//   media-replacement compatible (for example a still image standing in
//   for video) at script time; this setter does not replicate that
//   check, and the resulting file still parses cleanly in AE regardless.
// @param       item  the new alternate source item (nil clears the override)
// @domain      layer-set
// @stability   stable
// @verify      roundtrip
// @since       AE2020
// @boundary    length-preserving, a 4-byte field; requires a pre-existing
//   Essential Properties media-replacement slot or the call errors
// @alias       alternate source,media replacement,素材替换,动态图形模板,essential properties,EG 替换
func (l *Layer) SetAlternateSource(item AVItem) error {
	if l.runtime.back == nil {
		return fmt.Errorf("layer %q: no Essential Properties media-replacement slot (call addToMotionGraphicsTemplateAs in AE first)", l.Name)
	}
	var newID uint32
	if item != nil {
		newID = item.ItemID()
	}
	if newID != 0 && l.runtime.comp != nil && l.runtime.comp.proj != nil {
		if l.runtime.comp.proj.AVItemByID(newID) == nil {
			return fmt.Errorf("layer %q: alternate source item id %d not in project", l.Name, newID)
		}
	}
	if err := l.runtime.back.SetAlternateSource(item); err != nil {
		return err
	}
	l.AlternateSourceID = newID
	return nil
}

// @summary     Clear the layer's media-replacement override
// @description Removes the media-replacement override so AE falls back
//   to the wrapper precomp's default source. Equivalent to calling
//   SetAlternateSource with a nil item.
// @domain      layer-set
// @stability   stable
// @verify      roundtrip
// @since       AE2020
// @boundary    requires a pre-existing Essential Properties
//   media-replacement slot, same as SetAlternateSource
// @alias       clear alternate source,clear media replacement,清除替换素材
func (l *Layer) ClearAlternateSource() error { return l.SetAlternateSource(nil) }

// @summary     Replace a text layer's user-visible text
// @description The new text may be any length and span any number of
//   paragraphs (split on line breaks): the PostScript string is spliced,
//   the paragraph array is rebuilt with one count-patched entry per
//   paragraph, the style-run array is collapsed to a single run carrying
//   the total count (keeping the first run's style, matching how AE
//   handles a whole-text replace), and the layout cache is left for AE to
//   recompute on load. Empty text is supported and becomes a single empty
//   paragraph. Replacements that keep both the encoded byte length and
//   the per-paragraph UTF-16 counts unchanged are written in place and
//   preserve all runs. A per-character manual-kerning table is dropped on
//   a length-changing replacement, again matching AE's own behavior; the
//   only error case for a well-formed text layer is when the layer isn't
//   actually a text layer. Encoding parity with AE: input is split on
//   line-feed characters (each segment becomes a paragraph terminated by
//   AE's carriage-return convention), then encoded as big-endian UTF-16
//   with a leading byte-order mark, with PostScript special characters
//   escaped at the byte level. After a successful call Layer.TextSource
//   is re-decoded so subsequent reads reflect the new value.
// @param       newText  the new text content
// @domain      text
// @stability   stable
// @verify      ae-accept
// @gate        TestSetTextVariable_AEShipGate_AE2020,TestSetTextVariable_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable: the PostScript string is spliced and the
//   paragraph/run arrays are rebuilt; supports multiple paragraphs and
//   multiple style runs collapsing to one; manual per-character kerning
//   is dropped on a length-changing replacement
// @alias       set text,text content,文本内容,替换文字,改文字,文字修改
func (l *Layer) SetText(newText string) error {
	if l.TextSource == nil || l.TextSourceRaw == nil {
		return fmt.Errorf("layer %q: not a text layer (or text source failed to decode)", l.Name)
	}
	if l.runtime.back != nil {
		if err := l.runtime.back.SetText(newText); err != nil {
			return fmt.Errorf("layer %q: %w", l.Name, err)
		}
		l.resyncTextSource()
		return nil
	}
	// No back-ref (layer not chunk-backed): legacy length-preserving in-place
	// write against the recorded string offsets.
	if l.TextSource.textStringEnd <= l.TextSource.textStringStart {
		return fmt.Errorf("layer %q: original text-string offset not recorded; can't splice", l.Name)
	}
	encoded := codec.EncodeAEPSText(newText)
	oldLen := l.TextSource.textStringEnd - l.TextSource.textStringStart
	if len(encoded) != oldLen {
		return fmt.Errorf("layer %q: SetText length mismatch (new=%d bytes, old=%d bytes — length-preserving only; pad input to match)",
			l.Name, len(encoded), oldLen)
	}
	copy(l.TextSourceRaw[l.TextSource.textStringStart:l.TextSource.textStringEnd], encoded)
	ts, _ := DecodeTextSource(l.TextSourceRaw)
	if ts != nil {
		l.TextSource = ts
	}
	return nil
}

// TextEncodedByteLen returns the encoded byte length that SetText
// would produce for s, so callers can check if a candidate value fits
// the length-preserving budget without trial-and-error. Use it like:
//
//	if aep.TextEncodedByteLen(candidate) != aep.TextEncodedByteLen(layer.TextSource.Text) {
//	    // doesn't fit — pad/truncate or skip
//	}
func TextEncodedByteLen(s string) int {
	return len(codec.EncodeAEPSText(s))
}
