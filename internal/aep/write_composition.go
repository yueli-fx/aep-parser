package aep

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Composition setters live here. All length-preserving — they mutate
// fixed-offset fields inside the comp's cdta chunk. Each setter:
//
//   - validates the comp has an underlying cdta chunk (else returns an
//     error — happens for comps built outside the parser);
//   - validates the chunk is long enough for the targeted offset;
//   - mutates bytes in place;
//   - syncs the matching Go field.
//
// Next Project.WriteAEP call serializes the change.

// SetBGColor writes a new background color (R, G, B), each 0..255, to
// cdta @0x34/@0x35/@0x36.
// length-preserving (3 bytes).
func (c *Composition) SetBGColor(rgb [3]uint8) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) < 0x37 {
		return fmt.Errorf("comp %q: cdta too short for BGColor write (len=%d)", c.Name, len(c.cdta.Data))
	}
	c.cdta.Data[0x34] = rgb[0]
	c.cdta.Data[0x35] = rgb[1]
	c.cdta.Data[0x36] = rgb[2]
	c.BGColor = rgb
	return nil
}

// SetSize writes a new canvas pixel size (width, height) to cdta
// @0x8C / @0x8E (uint16 BE pair). length-preserving (4 bytes).
// Does not touch pixel aspect ratio at @0x90/@0x94 — set that
// separately via SetPixelAspect.
func (c *Composition) SetSize(width, height uint16) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) < 0x90 {
		return fmt.Errorf("comp %q: cdta too short for Size write (len=%d)", c.Name, len(c.cdta.Data))
	}
	if width == 0 || height == 0 {
		return fmt.Errorf("comp %q: SetSize requires non-zero width and height (got %dx%d)", c.Name, width, height)
	}
	binary.BigEndian.PutUint16(c.cdta.Data[0x8C:0x8E], width)
	binary.BigEndian.PutUint16(c.cdta.Data[0x8E:0x90], height)
	c.Width = width
	c.Height = height
	return nil
}

// SetResolutionFactor writes the comp's preview-resolution downsample
// factors (X, Y) to cdta @0x00 / @0x02 (two uint16 BE). Mirrors AE
// Scripting's CompItem.resolutionFactor. [1,1] = Full, [2,2] = Half,
// [4,4] = Quarter; custom non-square pairs allowed. Both factors must
// be ≥ 1 (AE clamps; we refuse 0 to surface caller bugs).
// length-preserving (4 bytes).
func (c *Composition) SetResolutionFactor(x, y uint16) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) < 0x04 {
		return fmt.Errorf("comp %q: cdta too short for ResolutionFactor write (len=%d)", c.Name, len(c.cdta.Data))
	}
	if x == 0 || y == 0 {
		return fmt.Errorf("comp %q: SetResolutionFactor requires non-zero X and Y (got %dx%d)", c.Name, x, y)
	}
	binary.BigEndian.PutUint16(c.cdta.Data[0x00:0x02], x)
	binary.BigEndian.PutUint16(c.cdta.Data[0x02:0x04], y)
	c.ResolutionFactor = [2]uint16{x, y}
	return nil
}

// SetShutterAngle writes the motion-blur shutter angle (uint16 BE,
// degrees, AE UI range 0..720, default 180) to cdta @0xAE.
// length-preserving (2 bytes).
func (c *Composition) SetShutterAngle(degrees uint16) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) < 0xB0 {
		return fmt.Errorf("comp %q: cdta too short for ShutterAngle write (len=%d)", c.Name, len(c.cdta.Data))
	}
	binary.BigEndian.PutUint16(c.cdta.Data[0xAE:0xB0], degrees)
	c.ShutterAngle = degrees
	return nil
}

// SetShutterPhase writes the motion-blur shutter phase (int32 BE) to
// cdta @0xB4. Unit is likely degrees (AE UI shows -90 / +90 / etc.) but
// not independently UI-verified — caller passes the raw int32 value.
// length-preserving (4 bytes).
func (c *Composition) SetShutterPhase(phase int32) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) < 0xB8 {
		return fmt.Errorf("comp %q: cdta too short for ShutterPhase write (len=%d)", c.Name, len(c.cdta.Data))
	}
	binary.BigEndian.PutUint32(c.cdta.Data[0xB4:0xB8], uint32(phase))
	c.ShutterPhase = phase
	return nil
}

// SetMotionBlurAdaptiveSampleLimit writes the motion-blur adaptive
// sample limit (int32 BE, AE default 128) to cdta @0xC4.
// length-preserving (4 bytes).
func (c *Composition) SetMotionBlurAdaptiveSampleLimit(limit int32) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) < 0xC8 {
		return fmt.Errorf("comp %q: cdta too short for MotionBlurAdaptiveSampleLimit write (len=%d)", c.Name, len(c.cdta.Data))
	}
	binary.BigEndian.PutUint32(c.cdta.Data[0xC4:0xC8], uint32(limit))
	c.MotionBlurAdaptiveSampleLimit = limit
	return nil
}

// SetMotionBlurSamplesPerFrame writes the per-frame motion-blur sample
// count (int32 BE, AE default 16) to cdta @0xC8.
// length-preserving (4 bytes).
func (c *Composition) SetMotionBlurSamplesPerFrame(n int32) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) < 0xCC {
		return fmt.Errorf("comp %q: cdta too short for MotionBlurSamplesPerFrame write (len=%d)", c.Name, len(c.cdta.Data))
	}
	binary.BigEndian.PutUint32(c.cdta.Data[0xC8:0xCC], uint32(n))
	c.MotionBlurSamplesPerFrame = n
	return nil
}

// SetName rewrites the composition's display name (length-variable
// Utf8 chunk replacement; WriteAEP recomputes parent Item LIST size).
// Returns an error if the comp has no Utf8 name chunk (rare).
func (c *Composition) SetName(newName string) error {
	if c.nameChunk == nil {
		return fmt.Errorf("comp %q: no Utf8 name chunk", c.Name)
	}
	c.nameChunk.Data = []byte(newName)
	c.Name = newName
	return nil
}

// SetFrameRate writes a new frame rate (fps) to cdta @0x9C-0x9F. AE
// splits fps into a uint16 whole part and a uint16 fractional part
// (numerator over 65536); we use the same encoding so partial frame
// rates like 29.97 round-trip exactly.
// length-preserving (4 bytes).
//
// `Duration` is recomputed from the existing frame count so the
// in-memory value stays consistent.
func (c *Composition) SetFrameRate(fps float64) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) < 0xA0 {
		return fmt.Errorf("comp %q: cdta too short for FrameRate write (len=%d)", c.Name, len(c.cdta.Data))
	}
	if fps <= 0 {
		return fmt.Errorf("comp %q: FrameRate %g not positive", c.Name, fps)
	}
	whole := uint16(fps)
	frac := uint16(math.Round((fps - float64(whole)) * 65536.0))
	binary.BigEndian.PutUint16(c.cdta.Data[0x9C:0x9E], whole)
	binary.BigEndian.PutUint16(c.cdta.Data[0x9E:0xA0], frac)
	c.FrameRate = float64(whole) + float64(frac)/65536.0
	// Frame count at @0xB0 is unchanged; recompute Duration from
	// frames / new fps.
	if len(c.cdta.Data) >= 0xB4 && c.FrameRate > 0 {
		frames := binary.BigEndian.Uint32(c.cdta.Data[0xB0:0xB4])
		c.Duration = float64(frames) / c.FrameRate
	}
	return nil
}

// SetDuration writes a new composition duration (seconds) to cdta @0xB0
// as a uint32 frame count (= round(seconds × FrameRate)). Requires
// `FrameRate > 0`.
// length-preserving (4 bytes).
func (c *Composition) SetDuration(seconds float64) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) < 0xB4 {
		return fmt.Errorf("comp %q: cdta too short for Duration write (len=%d)", c.Name, len(c.cdta.Data))
	}
	if c.FrameRate <= 0 {
		return fmt.Errorf("comp %q: cannot SetDuration without a positive FrameRate (call SetFrameRate first)", c.Name)
	}
	if seconds < 0 {
		return fmt.Errorf("comp %q: Duration %g must be non-negative", c.Name, seconds)
	}
	frames := uint32(math.Round(seconds * c.FrameRate))
	binary.BigEndian.PutUint32(c.cdta.Data[0xB0:0xB4], frames)
	c.Duration = float64(frames) / c.FrameRate
	return nil
}

// Composition boolean flag bits live in cdta @0x8A / @0x8B. Reverse-
// engineered by diffing AE 2020 fixture variants (see
// test_data/re_batch.aep + tmp_debug/re_batch.jsx).
//
//	0x8A bit 0 (0x01) — Draft3D
//	0x8B bit 0 (0x01) — HideShyLayers
//	0x8B bit 3 (0x08) — MotionBlur (comp-level master switch)
//	0x8B bit 5 (0x20) — PreserveNestedFrameRate
//
// Note: `dropFrame` doesn't have a clean bit in cdta — AE encodes it
// indirectly through work-area divisor choice (600 vs 30720), so it
// isn't exposed as a setter here. Use SetFrameRate / SetWorkArea
// instead if you need to switch the comp's time base.

type cdtaFlagBit struct {
	off  int
	mask byte
}

var (
	flagDraft3D                  = cdtaFlagBit{0x8A, 0x01}
	flagHideShyLayers            = cdtaFlagBit{0x8B, 0x01}
	flagCompMotionBlur           = cdtaFlagBit{0x8B, 0x08}
	flagFrameBlending            = cdtaFlagBit{0x8B, 0x10}
	flagPreserveNestedFrameRate  = cdtaFlagBit{0x8B, 0x20}
	flagPreserveNestedResolution = cdtaFlagBit{0x8B, 0x80}
)

// setCdtaFlagBit flips a single cdta flag bit length-preserving.
func (c *Composition) setCdtaFlagBit(b cdtaFlagBit, v bool) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) <= b.off {
		return fmt.Errorf("comp %q: cdta @%#x out of range (len=%d)", c.Name, b.off, len(c.cdta.Data))
	}
	if v {
		c.cdta.Data[b.off] |= b.mask
	} else {
		c.cdta.Data[b.off] &^= b.mask
	}
	return nil
}

// SetHideShyLayers toggles "Hide Shy Layers" on the comp (cdta @0x8B bit 0).
func (c *Composition) SetHideShyLayers(v bool) error {
	return c.setCdtaFlagBit(flagHideShyLayers, v)
}

// SetCompMotionBlur toggles the comp-level motion-blur master switch
// (cdta @0x8B bit 3). Note: this is independent of per-layer
// `Layer.MotionBlur` — the layer renders motion blur only when both
// switches are on.
func (c *Composition) SetCompMotionBlur(v bool) error {
	return c.setCdtaFlagBit(flagCompMotionBlur, v)
}

// SetPreserveNestedFrameRate toggles "Preserve frame rate when nested
// or in render queue" on the comp (cdta @0x8B bit 5).
func (c *Composition) SetPreserveNestedFrameRate(v bool) error {
	return c.setCdtaFlagBit(flagPreserveNestedFrameRate, v)
}

// SetDraft3D toggles the comp's "Draft 3D" preview switch (cdta @0x8A
// bit 0). Disables shadows / motion blur / DOF in viewport for speed.
func (c *Composition) SetDraft3D(v bool) error {
	return c.setCdtaFlagBit(flagDraft3D, v)
}

// SetFrameBlending toggles the comp-level frame-blend master switch
// (cdta @0x8B bit 4). Layers also need their own FrameBlendEnabled on
// to render with blending.
func (c *Composition) SetFrameBlending(v bool) error {
	return c.setCdtaFlagBit(flagFrameBlending, v)
}

// SetPreserveNestedResolution toggles "Preserve resolution when nested"
// (cdta @0x8B bit 7).
func (c *Composition) SetPreserveNestedResolution(v bool) error {
	return c.setCdtaFlagBit(flagPreserveNestedResolution, v)
}

// SetPixelAspect writes the pixel aspect ratio (PAR) to cdta as a
// numerator/denominator pair at @0x90 / @0x94 (uint32 BE each).
//
// AE writes simple integer ratios for common presets (1/1 = 1.0,
// 2/1 = 2.0). For fractional ratios this setter picks
// num = round(par × 100) and den = 100 — accurate enough for AE's
// built-in PAR list (0.91, 1.09, 1.21, 1.33, 1.46, 1.5, 2.0).
// length-preserving (8 bytes).
func (c *Composition) SetPixelAspect(par float64) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) < 0x98 {
		return fmt.Errorf("comp %q: cdta too short for PixelAspect write (len=%d)", c.Name, len(c.cdta.Data))
	}
	if par <= 0 {
		return fmt.Errorf("comp %q: PixelAspect %g must be positive", c.Name, par)
	}
	var num, den uint32
	if par == float64(uint32(par)) {
		num, den = uint32(par), 1
	} else {
		num = uint32(math.Round(par * 100))
		den = 100
	}
	binary.BigEndian.PutUint32(c.cdta.Data[0x90:0x94], num)
	binary.BigEndian.PutUint32(c.cdta.Data[0x94:0x98], den)
	c.PixelAspect = float64(num) / float64(den)
	return nil
}

// SetWorkArea writes new work-area start/end times (seconds) to cdta
// @0x1C-0x2B. Both values are encoded as dividend/divisor pairs; we
// reuse the existing dividend divisor at @0x20 and @0x28 when non-zero
// (typically 600 — AE's standard work-area divisor) and pick 600 when
// the existing divisor is zero.
//
// startSeconds and endSeconds must be non-negative; endSeconds may be
// less than startSeconds (AE allows that visually — the work area
// simply has zero/negative span — but most workflows want
// endSeconds > startSeconds).
//
// length-preserving (16 bytes total).
func (c *Composition) SetWorkArea(startSeconds, endSeconds float64) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) < 0x2C {
		return fmt.Errorf("comp %q: cdta too short for WorkArea write (len=%d)", c.Name, len(c.cdta.Data))
	}
	if startSeconds < 0 || endSeconds < 0 {
		return fmt.Errorf("comp %q: WorkArea times must be non-negative (got start=%g end=%g)", c.Name, startSeconds, endSeconds)
	}
	startDivisor := binary.BigEndian.Uint32(c.cdta.Data[0x20:0x24])
	if startDivisor == 0 {
		startDivisor = 600
	}
	endDivisor := binary.BigEndian.Uint32(c.cdta.Data[0x28:0x2C])
	if endDivisor == 0 {
		endDivisor = 600
	}
	startDividend := uint32(math.Round(startSeconds * float64(startDivisor)))
	endDividend := uint32(math.Round(endSeconds * float64(endDivisor)))
	binary.BigEndian.PutUint32(c.cdta.Data[0x1C:0x20], startDividend)
	binary.BigEndian.PutUint32(c.cdta.Data[0x20:0x24], startDivisor)
	binary.BigEndian.PutUint32(c.cdta.Data[0x24:0x28], endDividend)
	binary.BigEndian.PutUint32(c.cdta.Data[0x28:0x2C], endDivisor)
	c.WorkAreaStart = float64(startDividend) / float64(startDivisor)
	c.WorkAreaEnd = float64(endDividend) / float64(endDivisor)
	return nil
}

// SetDisplayStartTime rewrites the comp's display-start-time origin
// (seconds) at cdta @0xA4 (dividend) / @0xA8 (divisor). length-
// preserving (8 bytes). Pass 0 to clear back to default.
//
// The divisor stored is the comp's TickRate (computed from cdta @0x08 /
// @0xA8). When seconds == 0, both dividend and divisor are written as 0
// (matching AE's "unset" encoding so re-parsing yields DisplayStartTime
// == 0).
//
// AE 25 sometimes writes a divisor (e.g. 23976 for 29.97 fps) that
// doesn't exactly satisfy ticks_per_frame × frame_count math (small
// rounding artifacts). The setter / getter use the stored pair verbatim
// so round-tripping a Set call gives back exactly what was set.
func (c *Composition) SetDisplayStartTime(seconds float64) error {
	if c.cdta == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if len(c.cdta.Data) < 0xAC {
		return fmt.Errorf("comp %q: cdta too short for DisplayStartTime write (len=%d)", c.Name, len(c.cdta.Data))
	}
	if seconds < 0 {
		return fmt.Errorf("comp %q: DisplayStartTime must be non-negative (got %g)", c.Name, seconds)
	}
	if seconds == 0 {
		binary.BigEndian.PutUint32(c.cdta.Data[0xA4:0xA8], 0)
		binary.BigEndian.PutUint32(c.cdta.Data[0xA8:0xAC], 0)
		c.DisplayStartTime = 0
		return nil
	}
	tickRate := c.TickRate
	if tickRate <= 0 {
		tickRate = aeLegacyTimeBase
	}
	divisor := uint32(math.Round(tickRate))
	dividend := uint32(math.Round(seconds * float64(divisor)))
	binary.BigEndian.PutUint32(c.cdta.Data[0xA4:0xA8], dividend)
	binary.BigEndian.PutUint32(c.cdta.Data[0xA8:0xAC], divisor)
	c.DisplayStartTime = float64(dividend) / float64(divisor)
	return nil
}

// SetDisplayStartFrame is a frame-count convenience wrapper around
// SetDisplayStartTime. Computes seconds = frame / FrameRate and writes
// the cdta pair. Errors when FrameRate <= 0.
func (c *Composition) SetDisplayStartFrame(frame int) error {
	if c.FrameRate <= 0 {
		return fmt.Errorf("comp %q: FrameRate not set; cannot convert frame to seconds", c.Name)
	}
	return c.SetDisplayStartTime(float64(frame) / c.FrameRate)
}
