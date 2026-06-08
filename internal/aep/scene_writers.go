package aep

import "io"

// scene_writers.go — XWriter interfaces for M8 back-ref inversion (P2).
//
// Each interface declares the writer-facing methods that the serializer
// (P3) will implement by holding the corresponding *Backrefs struct.
// The scene types (Project, Composition, Layer, …) will swap their
// concrete `back *xBackrefs` fields for these interfaces in P2.2.
//
// Constraints:
//   - Method signatures mirror the existing setter signatures EXACTLY
//     (no extra `error` returns, no parameter changes).
//   - Structural ops (New*/Delete*/Insert*/Move*/Duplicate*/Add*) are
//     excluded — they become serializer free-functions per spec §2.4.
//   - PropertyGroupWriter is deferred to its P2.2 step and omitted here.

// ProjectWriter is the writer interface for Project's back-ref operations.
type ProjectWriter interface {
	SetCompensateForSceneReferredProfiles(bool) error
	SetAudioSampleRate(float64) error
	SetWorkingGamma(float64) error
	SetGpuAccelType(string) error
	SetExpressionEngine(string) error
	SetFeetFramesFilmType(FeetFramesFilmType) error
	SetFootageTimecodeDisplayStartType(FootageTimecodeDisplayStartType) error
	SetTimecodeDefaultBase(int) error
	SetFramesCountType(FramesCountType) error
	SetDisplayStartFrame(int) error
	SetFramesUseFeetFrames(bool) error
	SetTimeDisplayType(TimeDisplayType) error
	SetTransparencyGridThumbnails(bool) error
	SetColorManagementSystem(ColorManagementSystem) error
	SetLutInterpolationMethod(LutInterpolationMethod) error
	SetOcioConfigurationFile(string) error
	SetBitsPerChannel(BitsPerChannel) error
	SetLinearBlending(bool) error
	SetLinearizeWorkingSpace(bool) error
	WriteAEP(w io.Writer) error
}

// CompositionWriter is the writer interface for Composition's back-ref operations.
type CompositionWriter interface {
	SetRenderer(string) error
	SetBGColor([3]uint8) error
	SetSize(width, height uint16) error
	SetResolutionFactor(x, y uint16) error
	SetShutterAngle(uint16) error
	SetShutterPhase(int32) error
	SetMotionBlurAdaptiveSampleLimit(int32) error
	SetMotionBlurSamplesPerFrame(int32) error
	SetName(string) error
	SetFrameRate(float64) error
	SetDuration(float64) error
	SetHideShyLayers(bool) error
	SetCompMotionBlur(bool) error
	SetPreserveNestedFrameRate(bool) error
	SetDraft3D(bool) error
	SetFrameBlending(bool) error
	SetPreserveNestedResolution(bool) error
	SetPixelAspect(float64) error
	SetWorkArea(startSeconds, endSeconds float64) error
	SetDisplayStartTime(float64) error
	SetDisplayStartFrame(int) error
	SetComment(string) error
	SetLabel(uint8) error
}

// LayerWriter is the writer interface for Layer's back-ref operations.
type LayerWriter interface {
	// write_layer.go — ldta / name / comment / matte / light / alternate source
	SetVisible(bool) error
	SetSolo(bool) error
	SetShy(bool) error
	SetLocked(bool) error
	SetEffectsEnabled(bool) error
	SetMotionBlur(bool) error
	SetAudioEnabled(bool) error
	SetFrameBlendEnabled(bool) error
	SetCollapseTransform(bool) error
	SetIs3D(bool) error
	SetIsAdjust(bool) error
	SetIsGuide(bool) error
	SetIsNull(bool) error
	SetMarkersLocked(bool) error
	SetSamplingBicubic(bool) error
	SetFrameBlendPixelMotion(bool) error
	SetBlendingMode(BlendingMode) error
	SetTrackMatte(TrackMatteType) error
	SetLabel(uint8) error
	SetQuality(LayerQuality) error
	SetParent(parentID uint32) error
	SetSource(sourceID uint32) error
	SetAutoOrient(AutoOrientType) error
	SetPreserveTransparency(bool) error
	SetStartTime(float64) error
	SetInPoint(float64) error
	SetOutPoint(float64) error
	SetStretch(float64) error
	SetName(string) error
	SetComment(string) error
	SetTrackMatteLayer(sourceID uint32, mode TrackMatteType) error
	SetLightKind(LightKind) error
	SetLightSource(*Layer) error
	SetAlternateSource(AVItem) error
	SetText(string) error
	// write_text.go — per-run and per-paragraph text style setters
	SetRunFontSize(runIdx int, sizePts float64) error
	SetRunTracking(runIdx int, tracking float64) error
	SetRunBaselineShift(runIdx int, shift float64) error
	SetRunLeading(runIdx int, leading float64) error
	SetRunAutoLeading(runIdx int, auto bool) error
	SetRunFontIndex(runIdx, fontIdx int) error
	SetRunFauxBold(runIdx int, on bool) error
	SetRunFauxItalic(runIdx int, on bool) error
	SetRunHorizontalScale(runIdx int, scale float64) error
	SetRunVerticalScale(runIdx int, scale float64) error
	SetRunTsume(runIdx int, tsume float64) error
	SetRunFillColor(runIdx int, rgba [4]float64) error
	SetRunStrokeColor(runIdx int, rgba [4]float64) error
	SetRunApplyStroke(runIdx int, apply bool) error
	SetRunStrokeWidth(runIdx int, width float64) error
	SetRunCapsOption(runIdx int, caps TextCapsOption) error
	SetRunBaselineOption(runIdx int, base TextBaselineOption) error
	SetRunStrokeOverFill(runIdx int, over bool) error
	SetRunAutoKernType(runIdx int, kt TextAutoKernType) error
	SetRunNoBreak(runIdx int, on bool) error
	SetRunLineJoinType(runIdx int, j TextLineJoinType) error
	SetRunDigitSet(runIdx int, d TextDigitSet) error
	SetParagraphJustification(paraIdx int, j TextJustification) error
	SetParagraphFirstLineIndent(paraIdx int, v float64) error
	SetParagraphStartIndent(paraIdx int, v float64) error
	SetParagraphEndIndent(paraIdx int, v float64) error
	SetParagraphSpaceBefore(paraIdx int, v float64) error
	SetParagraphSpaceAfter(paraIdx int, v float64) error
	SetParagraphAutoHyphenate(paraIdx int, on bool) error
	SetParagraphLeadingType(paraIdx int, lt TextLeadingType) error
	SetParagraphHangingRoman(paraIdx int, on bool) error
	SetParagraphDirection(paraIdx int, d TextParagraphDirection) error
	SetManualKerning(values []int) error
	// non-Set* A-class method
	AddFont(string) (int, error)
}

// PropertyWriter is the writer interface for Property's back-ref operations.
type PropertyWriter interface {
	SetStaticValue(any) error
	SetExpressionEnabled(bool) error
	SetExpression(string) error
	InsertKeyframe(float64, any) (*Keyframe, int, error)
	DeleteKeyframe(int) error
}

// KeyframeWriter is the writer interface for Keyframe's back-ref operations.
type KeyframeWriter interface {
	SetTime(float64) error
	SetValue(any) error
	SetInInterp(InterpType) error
	SetOutInterp(InterpType) error
	SetInTemporalEase([]TemporalEase) error
	SetOutTemporalEase([]TemporalEase) error
	SetInSpatialTangent([]float64) error
	SetOutSpatialTangent([]float64) error
}

// MarkerWriter is the writer interface for Marker's back-ref operations.
type MarkerWriter interface {
	SetTime(float64) error
	SetDuration(float64) error
	SetLabel(uint8) error
	SetComment(string) error
	SetChapter(string) error
	SetURL(string) error
	SetFrameTarget(string) error
	SetCuePointName(string) error
}

// MaskWriter is the writer interface for Mask's back-ref operations.
type MaskWriter interface {
	SetMode(MaskMode) error
	SetInverted(bool) error
	SetColor([3]uint8) error
	SetLocked(bool) error
	SetMaskMotionBlur(MaskMotionBlurMode) error
	SetClosed(bool) error
}

// FootageWriter is the writer interface for Footage's back-ref operations.
type FootageWriter interface {
	SetPath(string) error
	SetComment(string) error
	SetLabel(uint8) error
}

// RenderQueueWriter is the writer interface for RenderQueueItem and
// OutputModule back-ref operations (A1 + A2 alias-field setters).
type RenderQueueWriter interface {
	// RenderQueueItem — A1 (via back chunks)
	SetComment(string) error
	// RenderQueueItem — A2 (alias settingsBlock / roouData; no error return)
	SetQuality(int)
	SetColorDepth(int)
	SetEffects(int)
	SetFieldRender(int)
	SetPulldown(int)
	SetFrameBlending(int)
	SetMotionBlur(int)
	SetProxyUse(int)
	SetSoloSwitches(int)
	SetGuideLayers(int)
	SetDiskCache(int)
	SetFrameRate(int)
	SetResolution(x, y int)
	SetSkipExistingFiles(bool)
	SetName(string)
	SetQueueItemNotify(bool)
	SetLogType(uint16)
	SetTimeSpanStart(float64)
	SetTimeSpanDuration(float64)
	// OutputModule — A2 (alias settingsBlock / roouData; no error return)
	SetChannels(int)
	SetResizeQuality(int)
	SetResize(bool)
	SetLockAspectRatio(bool)
	SetCrop(bool)
	SetCropTop(int)
	SetCropLeft(int)
	SetCropBottom(int)
	SetCropRight(int)
	SetIncludeProjectLink(bool)
	SetPostRenderAction(uint32)
	SetUseCompFrameNumber(bool)
	SetUseRegionOfInterest(bool)
	SetIncludeSourceXMP(bool)
	SetPreserveRGB(bool)
	SetDepth(int)
	SetStartingNumber(uint32)
}

// GuideWriter is the writer interface for Guide's alias-block operations.
// (Guide has no *Backrefs struct; its write path aliases the ldat slot directly.)
type GuideWriter interface {
	SetPosition(px float64)
	SetOrientation(o GuideOrientation)
}
