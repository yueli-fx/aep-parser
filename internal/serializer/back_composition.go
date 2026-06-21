package serializer

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// compositionBackrefs holds the rifx.Chunk references that power
// Composition's length-preserving write paths (SetBGColor / SetShutterAngle
// / SetMotionBlur* / SetWorkArea cdta writes, length-variable name writes,
// item-level Comment / Label edits) and NewComposition's re-parse closed
// loop.
//
// Lifecycle:
//   - Populated by parseComposition + parseItem when a Composition is built
//     from a parsed .aep file.
//   - Nil for comps built outside the parser (NewComposition builds it
//     explicitly via the re-parse closed loop).
//   - opaque is reserved for a future phase that needs to round-trip
//     unrecognized sibling chunks under the comp's owning Item LIST
//     (to satisfy the opaque-preservation invariant); currently nil.
type compositionBackrefs struct {
	// compName is stored for error-message context.
	compName string

	// tickRate is the per-composition TickRate captured at parse time; used
	// by SetDisplayStartTime to compute the cdta divisor. TickRate is derived
	// from cdta @0x08 / @0xA8 and does not change with any current setter, so
	// a snapshot is safe.
	tickRate float64

	// FrameRateHz is kept in sync by SetFrameRate and used by SetDuration and
	// SetDisplayStartFrame to convert frame/second quantities.
	FrameRateHz float64

	// cdta is the underlying cdta chunk reference, captured by
	// parseComposition. Used by SetBGColor / SetShutterAngle /
	// SetMotionBlur* / SetWorkArea for length-preserving writes.
	cdta *rifx.Chunk

	// nameChunk is the comp's Utf8 name chunk (length-variable Set name).
	nameChunk *rifx.Chunk

	// prinChunk / prdaChunk are the comp's renderer chunks under the PRin
	// LIST sibling. prin is a fixed 104-byte chunk (renderer match-name +
	// display name, NUL-padded); prda carries renderer-specific options and
	// is variable-length. SetRenderer patches prin in place (length-
	// preserving name fields) and replaces prda wholesale (structural).
	// Both nil when the comp has no PRin LIST.
	prinChunk *rifx.Chunk
	prdaChunk *rifx.Chunk

	// itemList is the cached owning Item LIST chunk; populated by
	// parseComposition. Used by NewComposition (re-parse closed loop) +
	// future structural mutations. derived cache, never owned (see
	// Invariant #8).
	itemList *rifx.Chunk

	// Item-level chunk refs shared with Footage / Folder. Populated by
	// parseItem from the surrounding Item LIST.
	itemCmtaChunk  *rifx.Chunk // cmta sibling under the Item LIST
	itemIdtaChunk  *rifx.Chunk // idta sibling — Label byte at payload @0x3A
	itemLayrParent *rifx.Chunk // Item LIST itself — needed for cmta insertion when missing

	opaque map[rifx.ChunkID]*rifx.Chunk
}

var _ CompositionWriter = (*compositionBackrefs)(nil)

// compositionBack returns the concrete back-refs behind a Composition's writer
// interface for serializer-stage raw chunk access. Returns nil when the comp
// was built outside the parser. Free function (the receiver is a scene type
// post package-split).
func compositionBack(c *Composition) *compositionBackrefs {
	if cb, ok := scene.CompositionBack(c).(*compositionBackrefs); ok {
		return cb
	}
	return nil
}

func (b *compositionBackrefs) CdtaData() []byte {
	if b == nil || b.cdta == nil {
		return nil
	}
	return b.cdta.Data
}

func (b *compositionBackrefs) PrdaData() []byte {
	if b == nil || b.prdaChunk == nil {
		return nil
	}
	return b.prdaChunk.Data
}

// GuideLdatData returns the live ldat byte slice under the comp's Item-level
// Gide → list container — the single source of truth for ruler guides.
// syncGuides copies each guide's scene-owned 16-byte block back into this slice
// (paired by index) at WriteAEP time. Returns nil when the comp has no guides
// container (built outside the parser, or no guides).
func (b *compositionBackrefs) GuideLdatData() []byte {
	if b == nil || b.itemList == nil {
		return nil
	}
	gide := b.itemList.FindFirstList(rifx.IDGide)
	if gide == nil {
		return nil
	}
	list := gide.FindFirstList(rifx.IDkfl)
	if list == nil {
		return nil
	}
	ldat := list.FindFirst(rifx.IDLdat)
	if ldat == nil {
		return nil
	}
	return ldat.Data
}

func (b *compositionBackrefs) SetBGColor(rgb [3]uint8) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) < codec.CdtaBGColorB+1 {
		return fmt.Errorf("comp %q: cdta too short for BGColor write (len=%d)", b.compName, len(b.cdta.Data))
	}
	b.cdta.Data[codec.CdtaBGColorR] = rgb[0]
	b.cdta.Data[codec.CdtaBGColorG] = rgb[1]
	b.cdta.Data[codec.CdtaBGColorB] = rgb[2]
	return nil
}

func (b *compositionBackrefs) SetSize(width, height uint16) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) < codec.CdtaHeight+2 {
		return fmt.Errorf("comp %q: cdta too short for Size write (len=%d)", b.compName, len(b.cdta.Data))
	}
	if width == 0 || height == 0 {
		return fmt.Errorf("comp %q: SetSize requires non-zero width and height (got %dx%d)", b.compName, width, height)
	}
	binary.BigEndian.PutUint16(b.cdta.Data[codec.CdtaWidth:codec.CdtaWidth+2], width)
	binary.BigEndian.PutUint16(b.cdta.Data[codec.CdtaHeight:codec.CdtaHeight+2], height)
	return nil
}

func (b *compositionBackrefs) SetResolutionFactor(x, y uint16) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) < codec.CdtaResolutionFactorY+2 {
		return fmt.Errorf("comp %q: cdta too short for ResolutionFactor write (len=%d)", b.compName, len(b.cdta.Data))
	}
	if x == 0 || y == 0 {
		return fmt.Errorf("comp %q: SetResolutionFactor requires non-zero X and Y (got %dx%d)", b.compName, x, y)
	}
	binary.BigEndian.PutUint16(b.cdta.Data[codec.CdtaResolutionFactorX:codec.CdtaResolutionFactorX+2], x)
	binary.BigEndian.PutUint16(b.cdta.Data[codec.CdtaResolutionFactorY:codec.CdtaResolutionFactorY+2], y)
	return nil
}

func (b *compositionBackrefs) SetShutterAngle(degrees uint16) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) < codec.CdtaShutterAngle+2 {
		return fmt.Errorf("comp %q: cdta too short for ShutterAngle write (len=%d)", b.compName, len(b.cdta.Data))
	}
	binary.BigEndian.PutUint16(b.cdta.Data[codec.CdtaShutterAngle:codec.CdtaShutterAngle+2], degrees)
	return nil
}

func (b *compositionBackrefs) SetShutterPhase(phase int32) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) < codec.CdtaShutterPhase+4 {
		return fmt.Errorf("comp %q: cdta too short for ShutterPhase write (len=%d)", b.compName, len(b.cdta.Data))
	}
	binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaShutterPhase:codec.CdtaShutterPhase+4], uint32(phase))
	return nil
}

func (b *compositionBackrefs) SetMotionBlurAdaptiveSampleLimit(limit int32) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) < codec.CdtaMotionBlurAdaptive+4 {
		return fmt.Errorf("comp %q: cdta too short for MotionBlurAdaptiveSampleLimit write (len=%d)", b.compName, len(b.cdta.Data))
	}
	binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaMotionBlurAdaptive:codec.CdtaMotionBlurAdaptive+4], uint32(limit))
	return nil
}

func (b *compositionBackrefs) SetMotionBlurSamplesPerFrame(n int32) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) < codec.CdtaMotionBlurSamples+4 {
		return fmt.Errorf("comp %q: cdta too short for MotionBlurSamplesPerFrame write (len=%d)", b.compName, len(b.cdta.Data))
	}
	binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaMotionBlurSamples:codec.CdtaMotionBlurSamples+4], uint32(n))
	return nil
}

func (b *compositionBackrefs) SetName(newName string) error {
	if b.nameChunk == nil {
		return fmt.Errorf("comp %q: no Utf8 name chunk", b.compName)
	}
	b.nameChunk.Data = []byte(newName)
	return nil
}

// SetFrameRate writes fps to cdta and keeps b.FrameRateHz in sync so
// SetDuration / SetDisplayStartFrame can use the current rate.
func (b *compositionBackrefs) SetFrameRate(fps float64) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) < codec.CdtaFrameRateFrac+2 {
		return fmt.Errorf("comp %q: cdta too short for FrameRate write (len=%d)", b.compName, len(b.cdta.Data))
	}
	if fps <= 0 {
		return fmt.Errorf("comp %q: FrameRate %g not positive", b.compName, fps)
	}
	whole := uint16(fps)
	frac := uint16(math.Round((fps - float64(whole)) * 65536.0))
	binary.BigEndian.PutUint16(b.cdta.Data[codec.CdtaFrameRateWhole:codec.CdtaFrameRateWhole+2], whole)
	binary.BigEndian.PutUint16(b.cdta.Data[codec.CdtaFrameRateFrac:codec.CdtaFrameRateFrac+2], frac)
	b.FrameRateHz = float64(whole) + float64(frac)/65536.0
	return nil
}

func (b *compositionBackrefs) SetDuration(seconds float64) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) < codec.CdtaMasterTicks+4 {
		return fmt.Errorf("comp %q: cdta too short for Duration write (len=%d)", b.compName, len(b.cdta.Data))
	}
	if b.FrameRateHz <= 0 {
		return fmt.Errorf("comp %q: cannot SetDuration without a positive FrameRate (call SetFrameRate first)", b.compName)
	}
	if seconds < 0 {
		return fmt.Errorf("comp %q: Duration %g must be non-negative", b.compName, seconds)
	}
	// Authoritative duration = MasterTicks @0x2C = round(seconds × nominalTickRate)
	// where nominalTickRate = ticksPerFrame@0x06 × round(fps). @0xB0 is the 360
	// shutter reference, not duration. Fall back to legacy @0xB0 frame-count only for
	// synthetic/legacy cdta that lacks ticksPerFrame@0x06 (mirrors the parser).
	// See incident cdta-0xB0-shutter-ref-not-duration.
	ticksPerFrame := uint32(binary.BigEndian.Uint16(b.cdta.Data[0x06:0x08]))
	nominalTickRate := ticksPerFrame * uint32(b.FrameRateHz+0.5)
	if nominalTickRate > 0 {
		ticks := uint32(math.Round(seconds * float64(nominalTickRate)))
		binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaMasterTicks:codec.CdtaMasterTicks+4], ticks)
	} else {
		frames := uint32(math.Round(seconds * b.FrameRateHz))
		binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaShutterAngleMax:codec.CdtaShutterAngleMax+4], frames)
	}
	return nil
}

type cdtaFlagBit struct {
	off  int
	mask byte
}

var (
	flagDraft3D                  = cdtaFlagBit{codec.CdtaFlagsByte8A, 0x01}
	flagHideShyLayers            = cdtaFlagBit{codec.CdtaFlagsByte8B, 0x01}
	flagCompMotionBlur           = cdtaFlagBit{codec.CdtaFlagsByte8B, 0x08}
	flagFrameBlending            = cdtaFlagBit{codec.CdtaFlagsByte8B, 0x10}
	flagPreserveNestedFrameRate  = cdtaFlagBit{codec.CdtaFlagsByte8B, 0x20}
	flagPreserveNestedResolution = cdtaFlagBit{codec.CdtaFlagsByte8B, 0x80}
)

func (b *compositionBackrefs) setCdtaFlagBit(bit cdtaFlagBit, v bool) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) <= bit.off {
		return fmt.Errorf("comp %q: cdta @%#x out of range (len=%d)", b.compName, bit.off, len(b.cdta.Data))
	}
	if v {
		b.cdta.Data[bit.off] |= bit.mask
	} else {
		b.cdta.Data[bit.off] &^= bit.mask
	}
	return nil
}

func (b *compositionBackrefs) SetHideShyLayers(v bool) error {
	return b.setCdtaFlagBit(flagHideShyLayers, v)
}

func (b *compositionBackrefs) SetCompMotionBlur(v bool) error {
	return b.setCdtaFlagBit(flagCompMotionBlur, v)
}

func (b *compositionBackrefs) SetPreserveNestedFrameRate(v bool) error {
	return b.setCdtaFlagBit(flagPreserveNestedFrameRate, v)
}

func (b *compositionBackrefs) SetDraft3D(v bool) error {
	return b.setCdtaFlagBit(flagDraft3D, v)
}

func (b *compositionBackrefs) SetFrameBlending(v bool) error {
	return b.setCdtaFlagBit(flagFrameBlending, v)
}

func (b *compositionBackrefs) SetPreserveNestedResolution(v bool) error {
	return b.setCdtaFlagBit(flagPreserveNestedResolution, v)
}

func (b *compositionBackrefs) SetPixelAspect(par float64) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) < codec.CdtaPixelAspectDen+4 {
		return fmt.Errorf("comp %q: cdta too short for PixelAspect write (len=%d)", b.compName, len(b.cdta.Data))
	}
	if par <= 0 {
		return fmt.Errorf("comp %q: PixelAspect %g must be positive", b.compName, par)
	}
	var num, den uint32
	if par == float64(uint32(par)) {
		num, den = uint32(par), 1
	} else {
		num = uint32(math.Round(par * 100))
		den = 100
	}
	binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaPixelAspectNum:codec.CdtaPixelAspectNum+4], num)
	binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaPixelAspectDen:codec.CdtaPixelAspectDen+4], den)
	return nil
}

func (b *compositionBackrefs) SetWorkArea(startSeconds, endSeconds float64) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) < codec.CdtaWorkAreaEndDiv+4 {
		return fmt.Errorf("comp %q: cdta too short for WorkArea write (len=%d)", b.compName, len(b.cdta.Data))
	}
	if startSeconds < 0 || endSeconds < 0 {
		return fmt.Errorf("comp %q: WorkArea times must be non-negative (got start=%g end=%g)", b.compName, startSeconds, endSeconds)
	}
	startDivisor := binary.BigEndian.Uint32(b.cdta.Data[codec.CdtaWorkAreaStartDiv : codec.CdtaWorkAreaStartDiv+4])
	if startDivisor == 0 {
		startDivisor = 600
	}
	endDivisor := binary.BigEndian.Uint32(b.cdta.Data[codec.CdtaWorkAreaEndDiv : codec.CdtaWorkAreaEndDiv+4])
	if endDivisor == 0 {
		endDivisor = 600
	}
	startDividend := uint32(math.Round(startSeconds * float64(startDivisor)))
	endDividend := uint32(math.Round(endSeconds * float64(endDivisor)))
	binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaWorkAreaStart:codec.CdtaWorkAreaStart+4], startDividend)
	binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaWorkAreaStartDiv:codec.CdtaWorkAreaStartDiv+4], startDivisor)
	binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaWorkAreaEnd:codec.CdtaWorkAreaEnd+4], endDividend)
	binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaWorkAreaEndDiv:codec.CdtaWorkAreaEndDiv+4], endDivisor)
	return nil
}

func (b *compositionBackrefs) SetDisplayStartTime(seconds float64) error {
	if b.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", b.compName)
	}
	if len(b.cdta.Data) < codec.CdtaDisplayStartDiv+4 {
		return fmt.Errorf("comp %q: cdta too short for DisplayStartTime write (len=%d)", b.compName, len(b.cdta.Data))
	}
	if seconds < 0 {
		return fmt.Errorf("comp %q: DisplayStartTime must be non-negative (got %g)", b.compName, seconds)
	}
	if seconds == 0 {
		binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaDisplayStartTime:codec.CdtaDisplayStartTime+4], 0)
		binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaDisplayStartDiv:codec.CdtaDisplayStartDiv+4], 0)
		return nil
	}
	tickRate := b.tickRate
	if tickRate <= 0 {
		tickRate = aeLegacyTimeBase
	}
	divisor := uint32(math.Round(tickRate))
	dividend := uint32(math.Round(seconds * float64(divisor)))
	binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaDisplayStartTime:codec.CdtaDisplayStartTime+4], dividend)
	binary.BigEndian.PutUint32(b.cdta.Data[codec.CdtaDisplayStartDiv:codec.CdtaDisplayStartDiv+4], divisor)
	return nil
}

// SetDisplayStartFrame delegates to SetDisplayStartTime; FrameRateHz is read
// from b.FrameRateHz which SetFrameRate keeps in sync.
func (b *compositionBackrefs) SetDisplayStartFrame(frame int) error {
	if b.FrameRateHz <= 0 {
		return fmt.Errorf("comp %q: FrameRate not set; cannot convert frame to seconds", b.compName)
	}
	return b.SetDisplayStartTime(float64(frame) / b.FrameRateHz)
}

func (b *compositionBackrefs) SetRenderer(name string) error {
	matchName := normalizeRendererName(name)
	tmpl, ok := rendererTemplates[matchName]
	if !ok {
		return fmt.Errorf("SetRenderer: unknown renderer %q (known: %s)", name, knownRenderers())
	}
	if b.prinChunk == nil || b.prdaChunk == nil {
		return fmt.Errorf("SetRenderer: comp %q has no prin/prda back-ref (built outside parser, or has no PRin LIST)", b.compName)
	}
	prin := b.prinChunk
	prda := b.prdaChunk
	if len(prin.Data) != prinSize {
		return fmt.Errorf("SetRenderer: comp %q prin is %d bytes, expected %d", b.compName, len(prin.Data), prinSize)
	}
	writePrinField(prin.Data[prinMatchNameOff:prinDisplayOff], matchName)
	writePrinField(prin.Data[prinDisplayOff:prinDisplayEnd], tmpl.display)
	prda.Data = append([]byte(nil), tmpl.prda...)
	return nil
}

func (b *compositionBackrefs) SetComment(comment string) error {
	if err := setItemComment(b.itemLayrParent, b.itemIdtaChunk, &b.itemCmtaChunk, comment); err != nil {
		return fmt.Errorf("comp %q: %w", b.compName, err)
	}
	return nil
}

func (b *compositionBackrefs) SetLabel(index uint8) error {
	if err := setItemLabel(b.itemIdtaChunk, index); err != nil {
		return fmt.Errorf("comp %q: %w", b.compName, err)
	}
	return nil
}
