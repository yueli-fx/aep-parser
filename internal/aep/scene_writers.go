package aep

import "io"

// scene_writers.go — XWriter interfaces for M8 back-ref inversion.
//
// Each interface declares the writer-facing methods (plus decoded read
// accessors) that the serializer implements by holding the corresponding
// concrete back-ref struct. Scene types (Project, Composition, Layer, …) hold
// the interface in their `back` field and never name the concrete back-ref or
// touch rifx.Chunk — the prerequisite for the scene/serializer package split.
//
// Constraints:
//   - Set* method signatures mirror the existing setter signatures EXACTLY
//     (no extra `error` returns, no parameter changes).
//   - Structural ops (New*/Delete*/Insert*/Move*/Duplicate*/Add*) stay
//     serializer free-functions, reaching the concrete back-ref via the
//     unexported xBack accessor helpers.

// ProjectWriter is the writer interface for Project's back-ref operations.
// Beyond the Set* writers it exposes decoded read accessors so the scene
// layer can surface project-settings values without naming the concrete
// back-ref or touching rifx.Chunk. The read accessors return decoded values;
// nil/length handling lives in the back-ref implementation.
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

	xmpPacket() string
	revision() uint16
	versionString() string
	effectNames() []string
	compensateForSceneReferredProfiles() bool
	audioSampleRate() float64
	workingGamma() float64
	gpuAccelType() (string, bool)
	expressionEngine() (string, bool)
	nnhdByte(off int) (byte, bool)
	nnhdUint16(off int) (uint16, bool)
	cmsJSON() ([]byte, bool)
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

	// Decoded read accessors: the scene CdtaRawBytes / PrdaRawBytes accessors
	// read the live cdta / prda Data slices through these so they never name
	// the concrete back-ref or touch rifx.Chunk.
	cdtaData() []byte
	prdaData() []byte
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

	// Decoded read accessors: the scene layer reads the ldta dividend/divisor
	// frac and the alt-source slot presence through these so it never names
	// the concrete back-ref or touches rifx.Chunk.
	ldtaFrac(off int) (float64, bool)
	hasAlternateSourceSlot() bool
}

// PropertyWriter is the writer interface for Property's length-preserving /
// expression back-ref setters. The keyframe-stream ops (InsertKeyframe /
// DeleteKeyframe) and SetDimensionsSeparated are structural — they rebuild the
// scene Keyframes slice or add/remove follower Property nodes — so per §F D-U3
// they stay scene-method stopgaps reaching the concrete back-ref, not interface
// methods (the same treatment as RenderQueue AddItem/RemoveItem).
type PropertyWriter interface {
	SetStaticValue(any) error
	SetExpressionEnabled(bool) error
	SetExpression(string) error
	SetLockedRatio(bool) error

	// Decoded read accessors for the tdb4 / tdsb flag readers and the
	// min/max value chunks. The scene layer reads bits / bytes through these
	// so it never names the concrete back-ref or touches rifx.Chunk.
	tdb4Byte(off int) (byte, bool)
	tdsbByte(off int) (byte, bool)
	minValueBytes() []byte
	maxValueBytes() []byte
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

	// frameRate exposes the owning composition's FrameRate cached at parse
	// time so the scene FrameTime / SetFrameTime accessors avoid naming the
	// concrete back-ref.
	frameRate() float64
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

	// sspcData exposes the live sspc Data slice so the scene footage
	// convenience accessors (FootageMissing / HasAudio / StartFrame /
	// EndFrame) read their byte offsets without naming the concrete back-ref
	// or touching rifx.Chunk.
	sspcData() []byte
}

// RenderQueueItemWriter is the writer interface for RenderQueueItem's one
// back-ref operation: the length-variable SetComment (RCom insert/replace),
// which genuinely needs serializer chunk-tree access. Every other render-queue
// and output-module setter is a pure scene-buffer mutation (single source of
// truth) synced back by syncRenderQueue at WriteAEP time, so it carries no
// writer-interface method (§F D-U1). OutputModule and Guide need no interface
// at all for the same reason.
type RenderQueueItemWriter interface {
	SetComment(string) error
}

// RenderQueueWriter is the back-ref handle for a RenderQueue. The queue carries
// no scene-facing back-ref setter (the structural AddItem / RemoveItem ops are
// serializer free-functions reaching the concrete back-ref via renderQueueBack),
// so this is a nominal marker letting the scene field stay interface-typed and
// concrete-free (M8 P3.1-prep).
type RenderQueueWriter interface {
	isRenderQueueWriter()
}

// OutputModuleWriter is the back-ref handle for an OutputModule. Every
// output-module setter is a pure scene-buffer mutation synced at WriteAEP time,
// so the back exists only to locate the write-time sync targets; this nominal
// marker keeps the scene field interface-typed and concrete-free (M8 P3.1-prep).
type OutputModuleWriter interface {
	isOutputModuleWriter()
}

// PropertyGroupWriter is the back-ref handle for an AEPropertyGroup. The
// structural group ops (Remove / Duplicate / MoveTo / SetDimensionsSeparated)
// are serializer free-functions reaching the concrete back-ref's tdgp LIST via
// propertyGroupBack; the scene group carries no back-ref setter, so this nominal
// marker keeps the scene field interface-typed and concrete-free (M8 P3.1-prep).
type PropertyGroupWriter interface {
	isPropertyGroupWriter()
}
