package scene

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/codec"
)

// Composition setters live here. All length-preserving — they delegate byte
// writes to compositionBackrefs (via CompositionWriter) and then sync the
// matching Go scene fields.
//
// Next Project.WriteAEP call serializes the change.

// @summary     Set the composition's background color
// @param       rgb  the new background color, each channel 0..255
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, 3 bytes at cdta offsets 0x34/0x35/0x36
// @alias       background color,背景色,bg color,comp background
func (c *Composition) SetBGColor(rgb [3]uint8) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if err := c.back.SetBGColor(rgb); err != nil {
		return err
	}
	c.BGColor = rgb
	return nil
}

// @summary     Set the composition's canvas pixel size
// @param       width   the new canvas width in pixels
// @param       height  the new canvas height in pixels
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a uint16 BE pair at cdta offsets 0x8C/0x8E;
//   does not touch pixel aspect ratio at 0x90/0x94 — set that separately via
//   SetPixelAspect
// @alias       canvas size,comp size,width,height,分辨率,合成尺寸,宽高
func (c *Composition) SetSize(width, height uint16) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if err := c.back.SetSize(width, height); err != nil {
		return err
	}
	c.Width = width
	c.Height = height
	return nil
}

// @summary     Set the composition's preview-resolution downsample factors
// @description Mirrors AE Scripting's CompItem.resolutionFactor. [1,1] = Full,
//   [2,2] = Half, [4,4] = Quarter; custom non-square pairs are allowed.
// @param       x  the horizontal downsample factor, must be >= 1
// @param       y  the vertical downsample factor, must be >= 1
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, two uint16 BE values at cdta offsets
//   0x00/0x02; AE clamps factors below 1, this setter rejects 0 outright to
//   surface caller bugs
// @alias       resolution factor,preview resolution,分辨率因子,预览分辨率,half quarter full
func (c *Composition) SetResolutionFactor(x, y uint16) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if err := c.back.SetResolutionFactor(x, y); err != nil {
		return err
	}
	c.ResolutionFactor = [2]uint16{x, y}
	return nil
}

// @summary     Set the composition's motion-blur shutter angle
// @param       degrees  the new shutter angle in degrees, AE UI range 0..720, default 180
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a uint16 BE value at cdta offset 0xAE
// @alias       shutter angle,快门角,motion blur angle,运动模糊快门
func (c *Composition) SetShutterAngle(degrees uint16) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if err := c.back.SetShutterAngle(degrees); err != nil {
		return err
	}
	c.ShutterAngle = degrees
	return nil
}

// @summary     Set the composition's motion-blur shutter phase
// @description Unit is likely degrees (AE UI shows values like -90 / +90) but
//   that has not been independently UI-verified — the caller passes the raw
//   int32 value as-is.
// @param       phase  the new shutter phase value
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, an int32 BE value at cdta offset 0xB4
// @alias       shutter phase,快门相位,motion blur phase,运动模糊相位
func (c *Composition) SetShutterPhase(phase int32) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if err := c.back.SetShutterPhase(phase); err != nil {
		return err
	}
	c.ShutterPhase = phase
	return nil
}

// @summary     Set the composition's motion-blur adaptive sample limit
// @param       limit  the new adaptive sample limit, AE default 128
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, an int32 BE value at cdta offset 0xC4
// @alias       motion blur adaptive sample limit,运动模糊自适应采样上限,adaptive samples
func (c *Composition) SetMotionBlurAdaptiveSampleLimit(limit int32) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if err := c.back.SetMotionBlurAdaptiveSampleLimit(limit); err != nil {
		return err
	}
	c.MotionBlurAdaptiveSampleLimit = limit
	return nil
}

// @summary     Set the composition's per-frame motion-blur sample count
// @param       n  the new per-frame sample count, AE default 16
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, an int32 BE value at cdta offset 0xC8
// @alias       motion blur samples per frame,每帧运动模糊采样数,samples per frame
func (c *Composition) SetMotionBlurSamplesPerFrame(n int32) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if err := c.back.SetMotionBlurSamplesPerFrame(n); err != nil {
		return err
	}
	c.MotionBlurSamplesPerFrame = n
	return nil
}

// @summary     Set the composition's display name
// @param       newName  the new display name
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable: replaces the whole Utf8 name chunk and
//   recomputes the parent Item LIST size; returns an error when the
//   composition has no Utf8 name chunk (rare)
// @alias       comp name,合成名,重命名合成,rename composition
func (c *Composition) SetName(newName string) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no Utf8 name chunk", c.Name)
	}
	if err := c.back.SetName(newName); err != nil {
		return err
	}
	c.Name = newName
	return nil
}

// @summary     Set the composition's frame rate
// @description AE splits fps into a uint16 whole part and a uint16
//   fractional part (numerator over 65536); this setter uses the same
//   encoding so partial frame rates like 29.97 round-trip exactly. `Duration`
//   is recomputed from the existing frame count so the in-memory value stays
//   consistent. Do not call this and SetDuration on the same composition: this
//   setter does not rescale the stored duration ticks, so AE will read back
//   the wrong duration at the new rate.
// @param       fps  the new frame rate
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, 4 bytes at cdta offsets 0x9C-0x9F
// @alias       frame rate,帧率,fps,合成帧率
func (c *Composition) SetFrameRate(fps float64) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if err := c.back.SetFrameRate(fps); err != nil {
		return err
	}
	whole := uint16(fps)
	frac := uint16(math.Round((fps - float64(whole)) * 65536.0))
	c.FrameRate = float64(whole) + float64(frac)/65536.0
	// Recompute Duration from authoritative MasterTicks @0x2C (duration ticks /
	// nominalTickRate). @0xB0 is the 360 shutter reference, not duration.
	if d := c.back.CdtaData(); len(d) >= codec.CdtaMasterTicks+4 && c.FrameRate > 0 {
		durTicks := binary.BigEndian.Uint32(d[codec.CdtaMasterTicks : codec.CdtaMasterTicks+4])
		ticksPerFrame := uint32(binary.BigEndian.Uint16(d[0x06:0x08]))
		nominalTickRate := ticksPerFrame * uint32(c.FrameRate+0.5)
		if durTicks > 0 && nominalTickRate > 0 {
			c.Duration = float64(durTicks) / float64(nominalTickRate)
		}
	}
	return nil
}

// @summary     Set the composition's duration
// @description Writes the duration as a uint32 frame count (round(seconds x
//   FrameRate)) and requires FrameRate to already be positive. Do not call
//   this and SetFrameRate on the same composition: SetFrameRate does not
//   rescale the stored duration ticks, so changing the rate afterward makes
//   AE read back a shorter or longer duration than intended.
// @param       seconds  the new duration in seconds, must be non-negative
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a uint32 BE frame count at cdta offset 0xB0
// @alias       duration,合成时长,comp duration,时长,持续时间
func (c *Composition) SetDuration(seconds float64) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if c.FrameRate <= 0 {
		return fmt.Errorf("comp %q: cannot SetDuration without a positive FrameRate (call SetFrameRate first)", c.Name)
	}
	if seconds < 0 {
		return fmt.Errorf("comp %q: Duration %g must be non-negative", c.Name, seconds)
	}
	if err := c.back.SetDuration(seconds); err != nil {
		return err
	}
	frames := uint32(math.Round(seconds * c.FrameRate))
	c.Duration = float64(frames) / c.FrameRate
	return nil
}

// @summary     Set the composition's "Hide Shy Layers" toggle
// @param       v  the new toggle state
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at cdta offset 0x8B bit 0
// @alias       hide shy layers,隐藏羞涩图层,shy layers
func (c *Composition) SetHideShyLayers(v bool) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	return c.back.SetHideShyLayers(v)
}

// @summary     Set the composition's motion-blur master switch
// @description This switch is independent of the per-layer `Layer.MotionBlur`
//   flag — a layer renders motion blur only when both switches are on.
// @param       v  the new toggle state
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at cdta offset 0x8B bit 3
// @alias       comp motion blur,合成运动模糊开关,motion blur master switch
func (c *Composition) SetCompMotionBlur(v bool) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	return c.back.SetCompMotionBlur(v)
}

// @summary     Set the composition's "preserve frame rate when nested" toggle
// @param       v  the new toggle state
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at cdta offset 0x8B bit 5
// @alias       preserve nested frame rate,嵌套帧率保留,preserve frame rate
func (c *Composition) SetPreserveNestedFrameRate(v bool) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	return c.back.SetPreserveNestedFrameRate(v)
}

// @summary     Set the composition's "Draft 3D" preview switch
// @description Disables shadows, motion blur, and depth of field in the
//   viewport for faster preview playback.
// @param       v  the new toggle state
// @domain      comp
// @stability   stable
// @verify      roundtrip
// @since       AE2020
// @boundary    length-preserving, a single bit at cdta offset 0x8A bit 0; the
//   written byte matches AE's own draft3d=true encoding bit-for-bit, but the
//   DOM does not reflect this bit back on reopen (a derived/runtime state, not
//   independently value-verifiable through the DOM) — acceptance is capped at
//   byte-preservation rather than a DOM readback check
// @incident    comp-setdraft3d-false-green
// @alias       draft 3D,草稿3D,3D预览,快速3D
func (c *Composition) SetDraft3D(v bool) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	return c.back.SetDraft3D(v)
}

// @summary     Set the composition's frame-blend master switch
// @description Layers also need their own FrameBlendEnabled flag on to
//   actually render with blending.
// @param       v  the new toggle state
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at cdta offset 0x8B bit 4
// @alias       frame blending,帧融合开关,frame blend master
func (c *Composition) SetFrameBlending(v bool) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	return c.back.SetFrameBlending(v)
}

// @summary     Set the composition's "Preserve resolution when nested" toggle
// @param       v  the new toggle state
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a single bit at cdta offset 0x8B bit 7
// @alias       preserve nested resolution,嵌套分辨率保留,preserve resolution
func (c *Composition) SetPreserveNestedResolution(v bool) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	return c.back.SetPreserveNestedResolution(v)
}

// @summary     Set the composition's pixel aspect ratio
// @description AE writes simple integer ratios for common presets (1/1 = 1.0,
//   2/1 = 2.0). For fractional ratios this setter picks num =
//   round(par x 100) and den = 100, which is accurate enough for AE's
//   built-in PAR presets (0.91, 1.09, 1.21, 1.33, 1.46, 1.5, 2.0).
// @param       par  the new pixel aspect ratio
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, a numerator/denominator uint32 BE pair at
//   cdta offsets 0x90/0x94
// @alias       pixel aspect ratio,像素纵横比,PAR,像素比
func (c *Composition) SetPixelAspect(par float64) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if err := c.back.SetPixelAspect(par); err != nil {
		return err
	}
	var num, den uint32
	if par == float64(uint32(par)) {
		num, den = uint32(par), 1
	} else {
		num = uint32(math.Round(par * 100))
		den = 100
	}
	c.PixelAspect = float64(num) / float64(den)
	return nil
}

// @summary     Set the composition's work-area start and end times
// @description Both values are encoded as dividend/divisor pairs; this setter
//   reuses the existing divisors when non-zero (typically 600, AE's standard
//   work-area divisor) and falls back to 600 when the existing divisor is
//   zero. endSeconds may be less than startSeconds — AE allows that visually,
//   giving the work area a zero or negative span — but most workflows want
//   endSeconds greater than startSeconds.
// @param       startSeconds  the new work-area start time in seconds, must be non-negative
// @param       endSeconds    the new work-area end time in seconds, must be non-negative
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, 16 bytes total at cdta offsets 0x1C-0x2B
// @alias       work area,工作区域,work area start end,in point out point,工作范围
func (c *Composition) SetWorkArea(startSeconds, endSeconds float64) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if err := c.back.SetWorkArea(startSeconds, endSeconds); err != nil {
		return err
	}
	// Recompute scene values using the same divisor logic as the backrefs
	// implementation, reading the live cdta bytes through the writer interface.
	if d := c.back.CdtaData(); len(d) >= codec.CdtaWorkAreaEndDiv+4 {
		startDivisor := binary.BigEndian.Uint32(d[codec.CdtaWorkAreaStartDiv : codec.CdtaWorkAreaStartDiv+4])
		endDivisor := binary.BigEndian.Uint32(d[codec.CdtaWorkAreaEndDiv : codec.CdtaWorkAreaEndDiv+4])
		startDividend := binary.BigEndian.Uint32(d[codec.CdtaWorkAreaStart : codec.CdtaWorkAreaStart+4])
		endDividend := binary.BigEndian.Uint32(d[codec.CdtaWorkAreaEnd : codec.CdtaWorkAreaEnd+4])
		if startDivisor != 0 {
			c.WorkAreaStart = float64(startDividend) / float64(startDivisor)
		}
		if endDivisor != 0 {
			c.WorkAreaEnd = float64(endDividend) / float64(endDivisor)
		}
	}
	return nil
}

// @summary     Set the composition's display-start-time origin
// @description The divisor stored is the composition's TickRate (computed
//   from cdta offsets 0x08/0xA8). Passing 0 clears the value back to default
//   by writing both dividend and divisor as 0, matching AE's "unset"
//   encoding so re-parsing yields DisplayStartTime == 0. AE 2025 sometimes
//   writes a divisor (e.g. 23976 for 29.97 fps) that doesn't exactly satisfy
//   the ticks-per-frame times frame-count math due to small rounding
//   artifacts; this setter and its getter use the stored pair verbatim so a
//   round trip of a Set call gives back exactly what was set.
// @param       seconds  the new display-start-time origin in seconds, 0 clears it
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving, 8 bytes at cdta offsets 0xA4 (dividend) /
//   0xA8 (divisor)
// @alias       display start time,显示起始时间,timecode origin,起始时码
func (c *Composition) SetDisplayStartTime(seconds float64) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	if err := c.back.SetDisplayStartTime(seconds); err != nil {
		return err
	}
	if seconds == 0 {
		c.DisplayStartTime = 0
		return nil
	}
	// Recompute stored value from cdta to match backrefs rounding exactly.
	if d := c.back.CdtaData(); len(d) >= codec.CdtaDisplayStartDiv+4 {
		dividend := binary.BigEndian.Uint32(d[codec.CdtaDisplayStartTime : codec.CdtaDisplayStartTime+4])
		divisor := binary.BigEndian.Uint32(d[codec.CdtaDisplayStartDiv : codec.CdtaDisplayStartDiv+4])
		if divisor != 0 {
			c.DisplayStartTime = float64(dividend) / float64(divisor)
		}
	}
	return nil
}

// @summary     Set the composition's display-start-time origin by frame count
// @description A frame-count convenience wrapper around SetDisplayStartTime:
//   computes seconds = frame / FrameRate and delegates the cdta write to it.
//   Errors when FrameRate is not yet set to a positive value.
// @param       frame  the new display-start frame count
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    delegates to SetDisplayStartTime; requires FrameRate > 0
// @alias       display start frame,起始帧,timecode start frame,frame offset
func (c *Composition) SetDisplayStartFrame(frame int) error {
	if c.FrameRate <= 0 {
		return fmt.Errorf("comp %q: FrameRate not set; cannot convert frame to seconds", c.Name)
	}
	return c.SetDisplayStartTime(float64(frame) / c.FrameRate)
}
