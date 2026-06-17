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

// SetBGColor writes a new background color (R, G, B), each 0..255, to
// cdta @0x34/@0x35/@0x36.
// length-preserving (3 bytes).
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(3B);双版本 AE gated(comp-settings from-scratch fixture)" alias="background color,背景色,bg color,comp background"
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

// SetSize writes a new canvas pixel size (width, height) to cdta
// @0x8C / @0x8E (uint16 BE pair). length-preserving (4 bytes).
// Does not touch pixel aspect ratio at @0x90/@0x94 — set that
// separately via SetPixelAspect.
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(4B);不改 PAR;双版本 AE gated(comp-settings from-scratch fixture)" alias="canvas size,comp size,width,height,分辨率,合成尺寸,宽高"
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

// SetResolutionFactor writes the comp's preview-resolution downsample
// factors (X, Y) to cdta @0x00 / @0x02 (two uint16 BE). Mirrors AE
// Scripting's CompItem.resolutionFactor. [1,1] = Full, [2,2] = Half,
// [4,4] = Quarter; custom non-square pairs allowed. Both factors must
// be ≥ 1 (AE clamps; we refuse 0 to surface caller bugs).
// length-preserving (4 bytes).
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(4B);拒绝 0 值;双版本 AE gated(comp-settings from-scratch fixture)" alias="resolution factor,preview resolution,分辨率因子,预览分辨率,half quarter full"
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

// SetShutterAngle writes the motion-blur shutter angle (uint16 BE,
// degrees, AE UI range 0..720, default 180) to cdta @0xAE.
// length-preserving (2 bytes).
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(2B);AE UI 范围 0..720;双版本 AE gated(comp-settings from-scratch fixture)" alias="shutter angle,快门角,motion blur angle,运动模糊快门"
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

// SetShutterPhase writes the motion-blur shutter phase (int32 BE) to
// cdta @0xB4. Unit is likely degrees (AE UI shows -90 / +90 / etc.) but
// not independently UI-verified — caller passes the raw int32 value.
// length-preserving (4 bytes).
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(4B);双版本 AE gated(comp-settings fixture,AE DOM shutterPhase=-90 实读)" alias="shutter phase,快门相位,motion blur phase,运动模糊相位"
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

// SetMotionBlurAdaptiveSampleLimit writes the motion-blur adaptive
// sample limit (int32 BE, AE default 128) to cdta @0xC4.
// length-preserving (4 bytes).
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(4B);AE 默认 128;双版本 AE gated(comp-settings from-scratch fixture)" alias="motion blur adaptive sample limit,运动模糊自适应采样上限,adaptive samples"
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

// SetMotionBlurSamplesPerFrame writes the per-frame motion-blur sample
// count (int32 BE, AE default 16) to cdta @0xC8.
// length-preserving (4 bytes).
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(4B);AE 默认 16;双版本 AE gated(comp-settings from-scratch fixture)" alias="motion blur samples per frame,每帧运动模糊采样数,samples per frame"
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

// SetName rewrites the composition's display name (length-variable
// Utf8 chunk replacement; WriteAEP recomputes parent Item LIST size).
// Returns an error if the comp has no Utf8 name chunk (rare).
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-variable(Utf8 整片替换+父 LIST size 重算,CLAUDE.md #1 例外);双版本 AE gated(comp-settings from-scratch fixture)" alias="comp name,合成名,重命名合成,rename composition"
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

// SetFrameRate writes a new frame rate (fps) to cdta @0x9C-0x9F. AE
// splits fps into a uint16 whole part and a uint16 fractional part
// (numerator over 65536); we use the same encoding so partial frame
// rates like 29.97 round-trip exactly.
// length-preserving (4 bytes).
//
// `Duration` is recomputed from the existing frame count so the
// in-memory value stays consistent.
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(4B);29.97 等分数帧率精确 round-trip;双版本 AE gated(comp-settings fixture,单独 comp 验);⚠ 勿与 SetDuration 同 comp(不 rescale duration ticks→AE 读错率,见 incident)" alias="frame rate,帧率,fps,合成帧率"
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

// SetDuration writes a new composition duration (seconds) to cdta @0xB0
// as a uint32 frame count (= round(seconds × FrameRate)). Requires
// `FrameRate > 0`.
// length-preserving (4 bytes).
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(4B);写帧数=round(sec×FrameRate);⚠ 勿与 SetFrameRate 同 comp(SetFrameRate 不 rescale duration ticks→AE 按旧率读,8s→6.4s,见 incident);双版本 AE gated(comp-settings fixture,单独 30fps comp 验)" alias="duration,合成时长,comp duration,时长,持续时间"
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

// SetHideShyLayers toggles "Hide Shy Layers" on the comp (cdta @0x8B bit 0).
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(1bit);双版本 AE gated(comp-settings from-scratch fixture)" alias="hide shy layers,隐藏羞涩图层,shy layers"
func (c *Composition) SetHideShyLayers(v bool) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	return c.back.SetHideShyLayers(v)
}

// SetCompMotionBlur toggles the comp-level motion-blur master switch
// (cdta @0x8B bit 3). Note: this is independent of per-layer
// `Layer.MotionBlur` — the layer renders motion blur only when both
// switches are on.
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(1bit);层级开关需与 Layer.MotionBlur 配合;双版本 AE gated(comp-settings from-scratch fixture)" alias="comp motion blur,合成运动模糊开关,motion blur master switch"
func (c *Composition) SetCompMotionBlur(v bool) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	return c.back.SetCompMotionBlur(v)
}

// SetPreserveNestedFrameRate toggles "Preserve frame rate when nested
// or in render queue" on the comp (cdta @0x8B bit 5).
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(1bit);双版本 AE gated(comp-settings from-scratch fixture)" alias="preserve nested frame rate,嵌套帧率保留,preserve frame rate"
func (c *Composition) SetPreserveNestedFrameRate(v bool) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	return c.back.SetPreserveNestedFrameRate(v)
}

// SetDraft3D toggles the comp's "Draft 3D" preview switch (cdta @0x8A
// bit 0). Disables shadows / motion blur / DOF in viewport for speed.
//
//aep:cap domain=comp tier=stable verify=roundtrip boundary="length-preserving(1bit);无专门 AE gate→round-trip" alias="draft 3D,草稿3D,3D预览,快速3D"
func (c *Composition) SetDraft3D(v bool) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	return c.back.SetDraft3D(v)
}

// SetFrameBlending toggles the comp-level frame-blend master switch
// (cdta @0x8B bit 4). Layers also need their own FrameBlendEnabled on
// to render with blending.
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(1bit);层级开关需与 Layer.FrameBlendEnabled 配合;双版本 AE gated(comp-settings from-scratch fixture)" alias="frame blending,帧融合开关,frame blend master"
func (c *Composition) SetFrameBlending(v bool) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	return c.back.SetFrameBlending(v)
}

// SetPreserveNestedResolution toggles "Preserve resolution when nested"
// (cdta @0x8B bit 7).
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(1bit);双版本 AE gated(comp-settings from-scratch fixture)" alias="preserve nested resolution,嵌套分辨率保留,preserve resolution"
func (c *Composition) SetPreserveNestedResolution(v bool) error {
	if c.back == nil {
		return fmt.Errorf("comp %q: no cdta chunk", c.Name)
	}
	return c.back.SetPreserveNestedResolution(v)
}

// SetPixelAspect writes the pixel aspect ratio (PAR) to cdta as a
// numerator/denominator pair at @0x90 / @0x94 (uint32 BE each).
//
// AE writes simple integer ratios for common presets (1/1 = 1.0,
// 2/1 = 2.0). For fractional ratios this setter picks
// num = round(par × 100) and den = 100 — accurate enough for AE's
// built-in PAR list (0.91, 1.09, 1.21, 1.33, 1.46, 1.5, 2.0).
// length-preserving (8 bytes).
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(8B);分数用 round(par×100)/100 近似;双版本 AE gated(comp-settings fixture,PAR=2.0 实读)" alias="pixel aspect ratio,像素纵横比,PAR,像素比"
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
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(16B);divisor 复用已存值或取 600;双版本 AE gated(comp-settings from-scratch fixture)" alias="work area,工作区域,work area start end,in point out point,工作范围"
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
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="length-preserving(8B);seconds=0 写双零清除;AE25 divisor 舍入不一致但 round-trip 精确;双版本 AE gated(comp-settings from-scratch fixture)" alias="display start time,显示起始时间,timecode origin,起始时码"
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

// SetDisplayStartFrame is a frame-count convenience wrapper around
// SetDisplayStartTime. Computes seconds = frame / FrameRate and writes
// the cdta pair. Errors when FrameRate <= 0.
//
//aep:cap domain=comp tier=stable verify=ae-accept gate=TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025 boundary="委托 SetDisplayStartTime;需 FrameRate>0;双版本 AE gated(comp-settings fixture,frame 15→0.5s 实读)" alias="display start frame,起始帧,timecode start frame,frame offset"
func (c *Composition) SetDisplayStartFrame(frame int) error {
	if c.FrameRate <= 0 {
		return fmt.Errorf("comp %q: FrameRate not set; cannot convert frame to seconds", c.Name)
	}
	return c.SetDisplayStartTime(float64(frame) / c.FrameRate)
}
