package scene

import (
	"fmt"

	"github.com/example/aep-parser/internal/codec"
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

// SetVisible toggles the layer's video switch (the 👁️ icon).
// length-preserving (single bit @ldta 0x27).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025 boundary="length-preserving 低风险;单 bit @ldta 0x27;双版本 AE gated(layer-av-flags 批量 fixture)" alias="visible,visibility,显示,隐藏,视频开关,eye toggle"
func (l *Layer) SetVisible(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetVisible(v); err != nil {
		return err
	}
	l.Visible = v
	return nil
}

// SetSolo toggles the layer's Solo flag.
// length-preserving (single bit @ldta 0x26).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025 boundary="length-preserving 低风险;单 bit @ldta 0x26;双版本 AE gated(layer-av-flags 批量 fixture)" alias="solo,独奏,单独显示"
func (l *Layer) SetSolo(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetSolo(v); err != nil {
		return err
	}
	l.Solo = v
	return nil
}

// SetShy toggles the layer's Shy flag (hides from the shy-filter view).
// length-preserving (single bit @ldta 0x27).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025 boundary="length-preserving 低风险;单 bit @ldta 0x27;双版本 AE gated(layer-av-flags 批量 fixture)" alias="shy,羞涩,隐藏图层,shy filter"
func (l *Layer) SetShy(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetShy(v); err != nil {
		return err
	}
	l.Shy = v
	return nil
}

// SetLocked toggles the layer's Lock flag (🔒). When locked, AE refuses
// edits in the timeline UI; the AEP file itself is still mutable.
// length-preserving (single bit @ldta 0x27).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025 boundary="length-preserving 低风险;单 bit @ldta 0x27;双版本 AE gated(layer-av-flags 批量 fixture)" alias="locked,lock,锁定,锁,图层锁定"
func (l *Layer) SetLocked(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetLocked(v); err != nil {
		return err
	}
	l.Locked = v
	return nil
}

// SetEffectsEnabled toggles the layer's fx switch (whether effects render).
// length-preserving (single bit @ldta 0x27).
//
//aep:cap domain=layer-set tier=stable verify=roundtrip boundary="length-preserving 低风险;单 bit @ldta 0x27;无专门 AE gate→round-trip" alias="effects enabled,fx switch,特效开关,效果启用"
func (l *Layer) SetEffectsEnabled(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetEffectsEnabled(v); err != nil {
		return err
	}
	l.EffectsEnabled = v
	return nil
}

// SetMotionBlur toggles the layer's motion-blur switch.
// length-preserving (single bit @ldta 0x27).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025 boundary="length-preserving 低风险;单 bit @ldta 0x27;双版本 AE gated(layer-av-flags 批量 fixture)" alias="motion blur,运动模糊,模糊开关"
func (l *Layer) SetMotionBlur(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetMotionBlur(v); err != nil {
		return err
	}
	l.MotionBlur = v
	return nil
}

// SetAudioEnabled toggles the layer's audio switch.
// length-preserving (single bit @ldta 0x27).
//
//aep:cap domain=layer-set tier=stable verify=roundtrip boundary="length-preserving 低风险;单 bit @ldta 0x27;无专门 AE gate→round-trip" alias="audio enabled,audio switch,音频开关,静音"
func (l *Layer) SetAudioEnabled(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetAudioEnabled(v); err != nil {
		return err
	}
	l.AudioEnabled = v
	return nil
}

// SetFrameBlendEnabled toggles the layer's frame-blend switch.
// length-preserving (single bit @ldta 0x27).
//
//aep:cap domain=layer-set tier=stable verify=roundtrip boundary="length-preserving 低风险;单 bit @ldta 0x27;无专门 AE gate→round-trip" alias="frame blend,帧混合,帧融合,frame blending"
func (l *Layer) SetFrameBlendEnabled(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetFrameBlendEnabled(v); err != nil {
		return err
	}
	l.FrameBlendEnabled = v
	return nil
}

// SetCollapseTransform toggles "Collapse Transformations" (for nested
// comps) or "Continuously Rasterize" (for Illustrator / shape layers).
// length-preserving (single bit @ldta 0x27).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerBool_AEShipGate_AE2020,TestLayerBool_AEShipGate_AE2025 boundary="length-preserving 低风险;单 bit @ldta 0x27;双版本 AE gated(layer-bool,shape 层=continuously rasterize,DOM collapseTransformation=true)" alias="collapse transform,continuously rasterize,折叠变换,连续栅格化"
func (l *Layer) SetCollapseTransform(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetCollapseTransform(v); err != nil {
		return err
	}
	l.CollapseTransform = v
	return nil
}

// SetIs3D toggles the layer's 3D-layer switch.
// length-preserving (single bit @ldta 0x26).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayer3DEnable_AEShipGate_AE2020,TestLayer3DEnable_AEShipGate_AE2025 alias="3D layer,3D开关,三维图层,enable 3D"
func (l *Layer) SetIs3D(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetIs3D(v); err != nil {
		return err
	}
	l.Is3D = v
	return nil
}

// SetIsAdjust toggles the layer's "Adjustment Layer" switch.
// length-preserving (single bit @ldta 0x26).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025 boundary="length-preserving 低风险;单 bit @ldta 0x26;双版本 AE gated(layer-av-fields2 批量 fixture)" alias="adjustment layer,调整图层,调节层"
func (l *Layer) SetIsAdjust(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetIsAdjust(v); err != nil {
		return err
	}
	l.IsAdjust = v
	return nil
}

// SetIsGuide toggles the layer's "Guide Layer" switch (AE renders it
// in the comp viewer but excludes it from output).
// length-preserving (single bit @ldta 0x25).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025 boundary="length-preserving 低风险;单 bit @ldta 0x25;双版本 AE gated(layer-av-fields2 批量 fixture)" alias="guide layer,参考线图层,辅助线"
func (l *Layer) SetIsGuide(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetIsGuide(v); err != nil {
		return err
	}
	l.IsGuide = v
	return nil
}

// SetIsNull toggles the Null-Object marker bit. AE's UI creates null
// layers via Layer > New > Null Object; flipping this post-hoc is
// supported by the byte but produces uncommon AE behavior — set only
// when you understand the consequences.
// length-preserving (single bit @ldta 0x26).
//
//aep:cap domain=layer-set tier=stable verify=roundtrip boundary="length-preserving 低风险;单 bit @ldta 0x26;post-hoc flip 产生非常见行为;无专门 AE gate→round-trip" alias="null layer,空对象,null object"
func (l *Layer) SetIsNull(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetIsNull(v); err != nil {
		return err
	}
	l.IsNull = v
	return nil
}

// SetMarkersLocked toggles "Lock markers" on the layer.
// length-preserving (single bit @ldta 0x26).
//
//aep:cap domain=layer-set tier=stable verify=roundtrip boundary="length-preserving 低风险;单 bit @ldta 0x26;无专门 AE gate→round-trip" alias="markers locked,标记锁定,lock markers"
func (l *Layer) SetMarkersLocked(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetMarkersLocked(v); err != nil {
		return err
	}
	l.MarkersLocked = v
	return nil
}

// SetSamplingBicubic switches between Bilinear (false) and Bicubic (true)
// sampling for the layer.
// length-preserving (single bit @ldta 0x25).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025 boundary="length-preserving 低风险;单 bit @ldta 0x25;双版本 AE gated(layer-av-fields2 批量 fixture,验 Bicubic)" alias="sampling bicubic,bicubic,bilinear,采样方式,图像质量"
func (l *Layer) SetSamplingBicubic(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetSamplingBicubic(v); err != nil {
		return err
	}
	l.SamplingBicubic = v
	return nil
}

// SetFrameBlendPixelMotion switches frame blending mode between
// Frame Mix (false) and Pixel Motion (true). Only relevant when
// FrameBlendEnabled is also true.
// length-preserving (single bit @ldta 0x25).
//
//aep:cap domain=layer-set tier=stable verify=roundtrip boundary="length-preserving 低风险;单 bit @ldta 0x25;需配合 FrameBlendEnabled=true 才生效;无专门 AE gate→round-trip" alias="pixel motion,frame mix,帧混合模式,像素运动"
func (l *Layer) SetFrameBlendPixelMotion(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if err := l.back.SetFrameBlendPixelMotion(v); err != nil {
		return err
	}
	l.FrameBlendPixelMotion = v
	return nil
}

// SetShy / Solo / Locked / etc. cover bool toggles. Multi-value byte
// fields follow.

// SetBlendingMode writes a new blending-mode enum byte to ldta @0x63.
// length-preserving (single byte).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025 boundary="length-preserving 低风险;单字节 @ldta 0x63;双版本 AE gated(layer-av-flags 批量 fixture,验 Multiply)" alias="blending mode,混合模式,叠加模式,blend mode"
func (l *Layer) SetBlendingMode(m BlendingMode) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.back.SetBlendingMode(m); err != nil {
		return err
	}
	l.BlendingMode = m
	return nil
}

// SetTrackMatte writes a new track-matte type byte to ldta @0x6B.
// length-preserving (single byte).
//
// AE additionally requires the matte source layer to sit immediately
// above this layer in the comp; this setter only flips the mode byte
// and does NOT reorder layers.
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestTrackMatteClassic_AEShipGate_AE2020,TestTrackMatteClassic_AEShipGate_AE2025 boundary="length-preserving 低风险;单字节 @ldta 0x6B;仅写模式字节,matte 须为正上方层(不重排);双版本 AE gated(track_matte_classic;AE2025 自迁移 classic→显式,DOM trackMatteType=ALPHA)" alias="track matte,matte type,遮罩,轨道遮罩,alpha matte,luma matte"
func (l *Layer) SetTrackMatte(t TrackMatteType) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.back.SetTrackMatte(t); err != nil {
		return err
	}
	l.TrackMatte = t
	return nil
}

// SetLabel writes a new timeline label-color index (0..16) to ldta @0x3D.
// Indices outside 0..16 are written verbatim (AE shows index 0 for any
// unknown value but the byte is preserved on round-trip).
// length-preserving (single byte).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025 boundary="length-preserving 低风险;单字节 @ldta 0x3D;范围 0..16;双版本 AE gated(layer-av-fields2 批量 fixture,验 9)" alias="label,label color,标签,颜色标签,时间线颜色"
func (l *Layer) SetLabel(index uint8) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.back.SetLabel(index); err != nil {
		return err
	}
	l.Label = index
	return nil
}

// SetQuality writes a new render-quality enum (Wireframe / Draft / Best)
// to ldta @0x04 (uint16 BE).
// length-preserving (2 bytes).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFlags_AEShipGate_AE2020,TestLayerAVFlags_AEShipGate_AE2025 boundary="length-preserving 低风险;2 字节 uint16 BE @ldta 0x04;双版本 AE gated(layer-av-flags 批量 fixture,验 Draft)" alias="quality,render quality,渲染质量,线框,草图,最佳质量"
func (l *Layer) SetQuality(q LayerQuality) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.back.SetQuality(q); err != nil {
		return err
	}
	l.Quality = q
	return nil
}

// SetParent rewrites the layer's parent-layer ID (ldta @0x84) to
// `parentID`. Pass 0 to clear the parent (= "no parent", AE shows
// "None" in the timeline). length-preserving (4 bytes).
//
// When the layer was created via the parser and lives inside a
// composition, the new parentID is validated against same-comp layers
// — passing an ID that doesn't resolve returns an error and doesn't
// touch the bytes. Cross-comp parenting isn't allowed by AE.
//
// `parentID == l.ID` is rejected (would create a self-parent cycle).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFields3_AEShipGate_AE2020,TestLayerAVFields3_AEShipGate_AE2025 boundary="length-preserving 低风险;4 字节 @ldta 0x84;同 comp 内验证 parentID;自 parent 拒绝;双版本 AE gated(layer-av-fields3 两层 fixture,验 AE DOM parent 指向正确层)" alias="parent,parent layer,父图层,父级,layer parenting"
func (l *Layer) SetParent(parentID uint32) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if parentID != 0 && parentID == l.ID {
		return fmt.Errorf("layer %q: self-parenting (parentID == own ID = %d) not allowed", l.Name, l.ID)
	}
	if parentID != 0 && l.comp != nil {
		if l.comp.LayerByID(parentID) == nil {
			return fmt.Errorf("layer %q: parentID %d not found in comp %q", l.Name, parentID, l.comp.Name)
		}
	}
	if err := l.back.SetParent(parentID); err != nil {
		return err
	}
	l.ParentID = parentID
	return nil
}

// SetSource rewrites the layer's source-item ID (ldta @0x28). For
// regular AV layers this is the footage or pre-comp ID. length-
// preserving (4 bytes).
//
// When the layer was created via the parser and the owning Project is
// known, `sourceID` is validated to match an existing Composition /
// Footage item — passing an ID that doesn't resolve returns an error
// and doesn't touch the bytes. Pass 0 to clear (rare; usually means
// the layer becomes a Null-like "no source" stub).
//
// Note: AE caches some source-derived metadata (Width/Height) on the
// layer at render time, not in ldta. Changing source preserves the
// rest of ldta verbatim, which is what we want.
//
//aep:cap domain=layer-set tier=stable verify=roundtrip boundary="length-preserving 低风险;4 字节 @ldta 0x28;项目内验证 sourceID;无专门 AE gate→round-trip" alias="source,source id,footage source,替换素材,层来源"
func (l *Layer) SetSource(sourceID uint32) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if sourceID != 0 && l.comp != nil && l.comp.proj != nil {
		p := l.comp.proj
		if p.CompositionByID(sourceID) == nil && !footageWithID(p, sourceID) {
			return fmt.Errorf("layer %q: sourceID %d not found in project items", l.Name, sourceID)
		}
	}
	if err := l.back.SetSource(sourceID); err != nil {
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

// SetAutoOrient writes the layer's auto-orient mode by clearing the
// three mutually-exclusive bits across ldta @0x25/@0x26 and setting
// the one matching the requested AutoOrientType. length-preserving
// (touches 2 bytes).
//
// Bit positions (mirroring parse_layer.go's decode):
//   - CharactersTowardCamera → 0x25 bit 4
//   - CameraOrPointOfInterest → 0x26 bit 5
//   - AlongPath → 0x26 bit 0
//   - None → all three cleared
//
// Unknown enum values are treated as None (= no-op clearing).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerBool_AEShipGate_AE2020,TestLayerBool_AEShipGate_AE2025 boundary="length-preserving 低风险;互斥 bit 组跨 @0x25/0x26;双版本 AE gated(layer-bool,AlongPath 需运动路径=position 关键帧,DOM autoOrient=ALONG_PATH)" alias="auto orient,自动旋转,沿路径旋转,朝向摄像机"
func (l *Layer) SetAutoOrient(t AutoOrientType) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.back.SetAutoOrient(t); err != nil {
		return err
	}
	l.AutoOrient = t
	return nil
}

// SetPreserveTransparency toggles "Preserve Underlying Transparency"
// (ldta @0x67, single byte 0/1).
// length-preserving (single byte).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025 boundary="length-preserving 低风险;单字节 @ldta 0x67;双版本 AE gated(layer-av-fields2 批量 fixture)" alias="preserve transparency,保留透明度,alpha preservation"
func (l *Layer) SetPreserveTransparency(v bool) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.back.SetPreserveTransparency(v); err != nil {
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

// SetStartTime writes the layer's start time (seconds) to
// ldta @0x0C/@0x10. length-preserving (8 bytes).
//
// AE allows negative start times (pre-roll). Mirrors the change to
// `Layer.StartTime`.
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFields3_AEShipGate_AE2020,TestLayerAVFields3_AEShipGate_AE2025 boundary="length-preserving 低风险;8 字节分数对 @ldta 0x0C/0x10;支持负值(pre-roll);双版本 AE gated(layer-av-fields3 批量 fixture,验 0.5)" alias="start time,开始时间,图层入点,layer offset"
func (l *Layer) SetStartTime(seconds float64) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.back.SetStartTime(seconds); err != nil {
		return err
	}
	l.StartTime = l.readLdtaFrac(0x0C)
	return nil
}

// SetInPoint writes the layer's source-media in-point (seconds) to
// ldta @0x14/@0x18, then refreshes `Layer.Duration` (= out − in).
// length-preserving (8 bytes).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025 boundary="length-preserving 低风险;8 字节分数对 @ldta 0x14/0x18;写后自动更新 Duration;双版本 AE gated(layer-av-fields2 批量 fixture)" alias="in point,入点,开始帧,trim start"
func (l *Layer) SetInPoint(seconds float64) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.back.SetInPoint(seconds); err != nil {
		return err
	}
	l.recomputeDurationFromLdta()
	return nil
}

// SetOutPoint writes the layer's source-media out-point (seconds) to
// ldta @0x1C/@0x20, then refreshes `Layer.Duration`.
// length-preserving (8 bytes).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFields2_AEShipGate_AE2020,TestLayerAVFields2_AEShipGate_AE2025 boundary="length-preserving 低风险;8 字节分数对 @ldta 0x1C/0x20;写后自动更新 Duration;双版本 AE gated(layer-av-fields2 批量 fixture)" alias="out point,出点,结束帧,trim end"
func (l *Layer) SetOutPoint(seconds float64) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.back.SetOutPoint(seconds); err != nil {
		return err
	}
	l.recomputeDurationFromLdta()
	return nil
}

// recomputeDurationFromLdta re-reads in/out points from ldta and
// updates `Layer.Duration`. Called by SetInPoint / SetOutPoint so the
// in-memory Duration stays consistent with the underlying bytes.
func (l *Layer) recomputeDurationFromLdta() {
	if l.back == nil {
		return
	}
	in, okIn := l.back.LdtaFrac(0x14)
	out, okOut := l.back.LdtaFrac(0x1C)
	if !okIn || !okOut {
		return
	}
	l.Duration = out - in
}

// SetStretch writes the layer's time-stretch ratio (1.0 = normal,
// 2.0 = 2× slow) to ldta @0x08 (dividend) / @0x6C (divisor). Unlike
// the other time fields the dividend/divisor pair is **split across
// the ldta** — dividend lives near the top, divisor in the trailing
// section.
// length-preserving (8 bytes total in two 4-byte writes).
//
//aep:cap domain=layer-set tier=alpha verify=roundtrip boundary="⚠ 确认 false-green:写 @ldta 0x08/0x6C 分子/分母 Go round-trips,但 AE 读 layer.stretch=100 不认(precomp 源也一样)——AE 按 in/out span 重算 stretch,本 setter 不调 in/out。须协调 outPoint 或 RE 正确字段。详 incident layer-setstretch-ae-recomputes-span" alias="stretch,time stretch,速度,时间拉伸,慢动作,快放"
func (l *Layer) SetStretch(ratio float64) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.back.SetStretch(ratio); err != nil {
		return err
	}
	if v, ok := l.back.StretchFrac(); ok {
		l.Stretch = v
	}
	return nil
}

// SetName rewrites the layer's display name (the AE timeline label).
// length-variable — the underlying Utf8 chunk's data slice is replaced
// wholesale; WriteAEP recomputes ancestor LIST sizes.
//
// Returns an error when the parsed layer has no Utf8 name chunk
// (unusual — most AE-written layers have one even for default names).
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFields3_AEShipGate_AE2020,TestLayerAVFields3_AEShipGate_AE2025 boundary="length-variable(Utf8 整片替换 + 父 LIST size 重算,CLAUDE.md #1 例外);双版本 AE gated(layer-av-fields3 批量 fixture,验 AE 接受 resize + DOM name)" alias="layer name,图层名,重命名,rename layer"
func (l *Layer) SetName(newName string) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no Utf8 name chunk to mutate", l.Name)
	}
	if err := l.back.SetName(newName); err != nil {
		return err
	}
	l.Name = newName
	return nil
}

// SetComment rewrites the layer's comment text (AE's "Comments"
// timeline column / Layer Settings dialog).
//
// length-variable — the underlying cmta chunk's data is replaced (or a
// new cmta chunk is inserted into the Layr LIST when none existed).
// Encoding mirrors what AE writes: LF → CRLF + a single NUL terminator.
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayerAVFields3_AEShipGate_AE2020,TestLayerAVFields3_AEShipGate_AE2025 boundary="length-variable;须 cmta double-NUL 终止 + ldta @0x3C has-comment flag=1(二者皆须,缺 flag 时 AE 字节全对仍读回空,已修);双版本 AE gated(layer-av-fields3 from-scratch 层)" incident=layer-setcomment-cmta-append-position alias="comment,注释,备注,layer comment,图层注释"
func (l *Layer) SetComment(comment string) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no chunk backrefs (built outside parser?)", l.Name)
	}
	if err := l.back.SetComment(comment); err != nil {
		return err
	}
	l.Comment = comment
	return nil
}

// SetTrackMatteLayer rewrites this layer's explicit track matte source
// (ldta @0xA0, AE 23+) to `sourceID` and updates the track matte mode
// (ldta @0x6B) to `mode`. Pass sourceID=0 with mode=TrackMatteNone to
// clear the matte. length-preserving (4 bytes + 1 byte).
//
// When the layer lives inside a parsed composition, `sourceID` is
// validated against same-comp layers — passing an ID that doesn't
// resolve returns an error and doesn't touch the bytes. AE 23+ requires
// the matte source layer to live in the same comp.
//
// `sourceID == l.ID` is rejected (self-matte is a no-op in AE; usually
// indicates a programmer error). Pass mode=TrackMatteNone with non-zero
// sourceID for "preserve target, no matte applied yet" — AE allows that
// (the source pointer stays but the matte channel is disabled).
//
// On AE 2020 / 2022 files (ldta 160 bytes), this call errors — the
// @0xA0 slot doesn't exist. Re-save the file through AE 23+ first to
// extend ldta, then this setter works.
//
//aep:cap domain=layer-set tier=stable verify=roundtrip minver=2023 boundary="length-preserving;AE 23+ ldta 专属(@0xA0 不存在于 AE2020/2022 → 报错);同 comp 验证;自 matte 拒绝;AE2025 单版本已验(TestTrackMatteExplicit_AEShipGate_AE2025,DOM trackMatteLayer+type 读回对);双版本不可达:仅 TargetAE2025 产 @0xA0,其 fingerprint 被 AE2024 forward-compat 拒,无中间 target → 按惯例留 roundtrip" alias="track matte layer,matte source,遮罩来源,轨道遮罩图层"
func (l *Layer) SetTrackMatteLayer(sourceID uint32, mode TrackMatteType) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if sourceID != 0 && sourceID == l.ID {
		return fmt.Errorf("layer %q: self-matte (sourceID == own ID = %d) not allowed", l.Name, l.ID)
	}
	if sourceID != 0 && l.comp != nil {
		if l.comp.LayerByID(sourceID) == nil {
			return fmt.Errorf("layer %q: sourceID %d not found in comp %q", l.Name, sourceID, l.comp.Name)
		}
	}
	if err := l.back.SetTrackMatteLayer(sourceID, mode); err != nil {
		return err
	}
	l.TrackMatteLayerID = sourceID
	l.TrackMatte = mode
	return nil
}

// ClearTrackMatteLayer is a shorthand for SetTrackMatteLayer(0, TrackMatteNone).
// Removes the explicit matte source AND the matte mode in one call.
//
//aep:cap domain=layer-set tier=stable verify=roundtrip minver=2023 boundary="委托 SetTrackMatteLayer(0, None);AE 23+ ldta 专属;AE2025 单版本已验(track_matte_explicit,clear→DOM NO_TRACK_MATTE);双版本不可达(同 SetTrackMatteLayer,fingerprint)→ 留 roundtrip" alias="clear matte,remove matte,清除遮罩,取消轨道遮罩"
func (l *Layer) ClearTrackMatteLayer() error {
	return l.SetTrackMatteLayer(0, TrackMatteNone)
}

// SetLightKind rewrites the light layer's kind (ldta @0x88, 4 bytes BE
// uint32). length-preserving. Only meaningful when Type == LayerTypeLight;
// AE silently accepts the byte change on non-light layers but it has
// no visual effect.
//
// On AE 22 / 2020 ldta (160 bytes), @0x88 is still present (parent ID
// at @0x84 + 4 bytes after = @0x88), so this setter works on older
// fixtures too.
//
//aep:cap domain=layer-set tier=stable verify=ae-accept gate=TestLayer3DLight_AEShipGate_AE2020,TestLayer3DLight_AEShipGate_AE2025 alias="light kind,light type,灯光类型,平行光,点光源,聚光灯,环境光"
func (l *Layer) SetLightKind(k LightKind) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if err := l.back.SetLightKind(k); err != nil {
		return err
	}
	l.LightKind = k
	return nil
}

// SetLightSource writes the environment-light source layer ID for a Light
// layer (AE 24+). Stored in ldta @0x28 (same slot as AV SourceID; the
// field is repurposed per Layer.Type). Pass nil to clear — writes sentinel
// 0xFFFFFFFF. length-preserving (4 bytes).
//
// Validation (matches py-aep LightLayer.light_source.setter):
//   - layer must be a Light layer (Type == LayerTypeLight)
//   - target (if non-nil) must be in the same composition
//   - target must not be Light or Camera
//   - target must not be 3D (py-aep: 3D AV layers can't drive env lights)
//   - target must not be self
//
// On error the bytes are not modified.
//
//aep:cap domain=layer-set tier=stable verify=roundtrip minver=2024 boundary="length-preserving;AE 24+ 环境灯专用;Light 层 only;3D/Light/Camera 目标拒绝;无专门 AE gate→round-trip" alias="light source,environment light,环境光源,灯光来源"
func (l *Layer) SetLightSource(target *Layer) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if l.Type != LayerTypeLight {
		return fmt.Errorf("layer %q: SetLightSource only valid for Light layers (Type=%q)", l.Name, l.Type)
	}
	if target == nil {
		if err := l.back.SetLightSource(nil); err != nil {
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
	if l.comp != nil && target.comp != nil && target.comp != l.comp {
		return fmt.Errorf("layer %q: SetLightSource target must be in same composition", l.Name)
	}
	if target.Type == LayerTypeLight || target.Type == LayerTypeCamera {
		return fmt.Errorf("layer %q: SetLightSource target cannot be Light/Camera (got %q)", l.Name, target.Type)
	}
	if target.Is3D {
		return fmt.Errorf("layer %q: SetLightSource target cannot be a 3D layer", l.Name)
	}
	if err := l.back.SetLightSource(target); err != nil {
		return err
	}
	l.SourceID = target.ID
	return nil
}

// SetAlternateSource overrides this layer's source via the Essential
// Properties → Media Replacement slot (4-byte blsi write, length-
// preserving). Pass nil (or a zero-id item) to clear the override —
// equivalent to ClearAlternateSource. Mirrors AE script's
// Property.setAlternateSource.
//
// Requires the layer to already have an Essential Properties slot —
// AE only persists the blsi chunk after the source-side layer has been
// promoted via `AVLayer.addToMotionGraphicsTemplateAs()` in the wrapper
// precomp. If the slot is missing (HasAlternateSourceSlot() == false),
// this errors instead of attempting structural chunk insertion (which
// would break length-preserving roundtrip).
//
// When the layer's owning project is wired up, `item.ItemID()` is
// validated against the project's AV items; an id unknown to the
// project returns an error and doesn't touch bytes. AE rejects items
// whose `isMediaReplacementCompatible == false` (e.g., still images vs.
// video) at script time — we don't replicate that check; the rendered
// .aep is still parsed cleanly by AE regardless.
//
//aep:cap domain=layer-set tier=stable verify=roundtrip boundary="length-preserving;须预先有 blsi slot(无 slot → 报错);项目内验证 itemID;无专门 AE gate→round-trip" alias="alternate source,media replacement,素材替换,动态图形模板,essential properties,EG 替换"
func (l *Layer) SetAlternateSource(item AVItem) error {
	if l.back == nil {
		return fmt.Errorf("layer %q: no Essential Properties media-replacement slot (call addToMotionGraphicsTemplateAs in AE first)", l.Name)
	}
	var newID uint32
	if item != nil {
		newID = item.ItemID()
	}
	if newID != 0 && l.comp != nil && l.comp.proj != nil {
		if l.comp.proj.AVItemByID(newID) == nil {
			return fmt.Errorf("layer %q: alternate source item id %d not in project", l.Name, newID)
		}
	}
	if err := l.back.SetAlternateSource(item); err != nil {
		return err
	}
	l.AlternateSourceID = newID
	return nil
}

// ClearAlternateSource removes the media-replacement override (blsi
// = 0) so AE falls back to the wrapper precomp's default source.
// Equivalent to SetAlternateSource(nil).
//
//aep:cap domain=layer-set tier=stable verify=roundtrip boundary="委托 SetAlternateSource(nil);须预先有 blsi slot;无专门 AE gate→round-trip" alias="clear alternate source,clear media replacement,清除替换素材"
func (l *Layer) ClearAlternateSource() error { return l.SetAlternateSource(nil) }

// SetText replaces a text layer's user-visible text. The new text may be any
// length and span any number of paragraphs ('\n' / '\r' line breaks): the
// PostScript string is spliced, the paragraph array is rebuilt with one
// count-patched entry per paragraph, the style-run array is collapsed to a
// single run carrying the total count (keeping the first run's style, as AE
// does on a whole-text replace), and the btdk layout cache is left for AE to
// recompute on load. Empty text ("") is supported (it becomes a single empty
// paragraph). Replacements that keep both the encoded byte length and the
// per-paragraph UTF-16 counts are written in place and preserve all runs.
//
// A per-character manual-kerning table is dropped on a length-changing
// replacement (as AE does on a whole-text replace); the only error from a
// well-formed text layer is when the layer isn't actually a text layer.
//
// Encoding parity with AE: input is split on '\n' (each segment becomes a
// paragraph terminated by AE's '\r' convention), then encoded as UTF-16BE
// with a leading FE FF BOM, with PostScript specials ( ) \ escaped at the
// byte level.
//
// After a successful call Layer.TextSource is re-decoded so subsequent
// reads reflect the new value; Layer.WriteAEP serializes the change.
//
//aep:cap domain=text tier=stable verify=ae-accept gate=TestSetTextVariable_AEShipGate_AE2020,TestSetTextVariable_AEShipGate_AE2025 boundary="length-variable(PostScript 字符串拼接 + 段落/run 数组重建);多段落/多 run 支持;手动字距在长度变化时丢弃" alias="set text,text content,文本内容,替换文字,改文字,文字修改"
func (l *Layer) SetText(newText string) error {
	if l.TextSource == nil || l.TextSourceRaw == nil {
		return fmt.Errorf("layer %q: not a text layer (or text source failed to decode)", l.Name)
	}
	if l.back != nil {
		if err := l.back.SetText(newText); err != nil {
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
