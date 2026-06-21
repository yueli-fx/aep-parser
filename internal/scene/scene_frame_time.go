package scene

import (
	"fmt"
	"math"
)

// Frame-time companion accessors — mirror of py-aep's `layer.frame_in_point`, //nolint:jargon
// `keyframe.frame_time`, `comp.work_area_start_frame`, etc. All conversions
// use the owning composition's FrameRate (not TickRate). Setters round the
// frame integer back to seconds via `seconds = frame / FrameRate` before
// delegating to the existing seconds-based setter.
//
// No new RIFX writes are introduced here — every Set* helper delegates to
// the corresponding seconds-based setter (SetInPoint / SetOutPoint /
// SetStartTime / SetWorkArea / SetDisplayStartTime / Keyframe.SetTime /
// Marker.SetTime / Marker.SetDuration).

// secondsToFrames rounds `seconds × fps` to the nearest integer frame.
// Returns 0 when fps is zero or negative (avoids divide-by-zero in
// reverse setters; the converted seconds will be 0 anyway).
func secondsToFrames(seconds, fps float64) int {
	if fps <= 0 {
		return 0
	}
	return int(math.Round(seconds * fps))
}

// framesToSeconds returns the float seconds equivalent of an integer
// frame count at the given fps. Used by every Set*Frame helper.
func framesToSeconds(frame int, fps float64) float64 {
	if fps <= 0 {
		return 0
	}
	return float64(frame) / fps
}

// ──────────────────────────────────────────────
// Layer frame-time accessors
// ──────────────────────────────────────────────

// InPoint returns the layer's in-point in seconds (ldta @0x14/@0x18,
// dividend/divisor pair). Returns 0 if the underlying ldta chunk is
// absent or too short.
//
// Companion to existing SetInPoint(seconds).
func (l *Layer) InPoint() float64 {
	return l.readLdtaFrac(0x14)
}

// OutPoint returns the layer's out-point in seconds (ldta @0x1C/@0x20).
// Returns 0 if the underlying ldta chunk is absent or too short.
//
// Companion to existing SetOutPoint(seconds).
func (l *Layer) OutPoint() float64 {
	return l.readLdtaFrac(0x1C)
}

// readLdtaFrac reads a dividend/divisor int32/uint32 pair from ldta
// at the given offset (8 bytes total) and returns dividend/divisor.
// Returns 0 when ldta is nil, too short, or the divisor is zero.
func (l *Layer) readLdtaFrac(off int) float64 {
	if l.back == nil {
		return 0
	}
	v, _ := l.back.LdtaFrac(off)
	return v
}

// layerFps returns the layer's owning composition FrameRate, or 0 when
// the layer was built outside the parser (no owning comp).
func (l *Layer) layerFps() float64 {
	if l.comp == nil {
		return 0
	}
	return l.comp.FrameRate
}

// FrameInPoint returns the layer in-point as an integer frame count
// (seconds × FrameRate, rounded). Returns 0 when the owning comp's
// FrameRate is unknown.
func (l *Layer) FrameInPoint() int {
	return secondsToFrames(l.InPoint(), l.layerFps())
}

// @summary     Set a layer's in-point from an integer frame count
// @description Converts frame back to seconds via the owning composition's
//   FrameRate, then delegates to the seconds-based in-point setter. The
//   in-point is source-relative; AE's UI displays startTime plus this value.
// @param       frame  the new in-point as a frame count
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerXform_AEShipGate_AE2020,TestLayerXform_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; requires the owning composition's
//   FrameRate to be > 0
// @alias       frame in point,入点帧,layer in frame,图层入点
func (l *Layer) SetFrameInPoint(frame int) error {
	fps := l.layerFps()
	if fps <= 0 {
		return fmt.Errorf("layer %q: cannot set frame in-point — no owning comp (FrameRate unknown)", l.Name)
	}
	return l.SetInPoint(framesToSeconds(frame, fps))
}

// FrameOutPoint returns the layer out-point as an integer frame count.
func (l *Layer) FrameOutPoint() int {
	return secondsToFrames(l.OutPoint(), l.layerFps())
}

// @summary     Set a layer's out-point from an integer frame count
// @description Converts frame back to seconds via the owning composition's
//   FrameRate, then delegates to the seconds-based out-point setter. The
//   out-point is source-relative; AE's UI displays startTime plus this value.
// @param       frame  the new out-point as a frame count
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerXform_AEShipGate_AE2020,TestLayerXform_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; requires the owning composition's
//   FrameRate to be > 0
// @alias       frame out point,出点帧,layer out frame,图层出点
func (l *Layer) SetFrameOutPoint(frame int) error {
	fps := l.layerFps()
	if fps <= 0 {
		return fmt.Errorf("layer %q: cannot set frame out-point — no owning comp (FrameRate unknown)", l.Name)
	}
	return l.SetOutPoint(framesToSeconds(frame, fps))
}

// FrameStartTime returns the layer start-time as an integer frame count.
func (l *Layer) FrameStartTime() int {
	return secondsToFrames(l.StartTime, l.layerFps())
}

// @summary     Set a layer's start time from an integer frame count
// @description Converts frame back to seconds via the owning composition's
//   FrameRate, then delegates to the seconds-based start-time setter.
// @param       frame  the new start time as a frame count
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerXform_AEShipGate_AE2020,TestLayerXform_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; requires the owning composition's
//   FrameRate to be > 0
// @alias       frame start time,起始帧,layer start frame,图层起始时间
func (l *Layer) SetFrameStartTime(frame int) error {
	fps := l.layerFps()
	if fps <= 0 {
		return fmt.Errorf("layer %q: cannot set frame start-time — no owning comp (FrameRate unknown)", l.Name)
	}
	return l.SetStartTime(framesToSeconds(frame, fps))
}

// ──────────────────────────────────────────────
// Composition frame-time accessors
// ──────────────────────────────────────────────

// DisplayStartFrame returns the comp's display-start frame
// (DisplayStartTime × FrameRate, rounded). Setter exists already as
// SetDisplayStartFrame.
func (c *Composition) DisplayStartFrame() int {
	return secondsToFrames(c.DisplayStartTime, c.FrameRate)
}

// WorkAreaStartFrame returns the comp's work-area start as a frame count.
func (c *Composition) WorkAreaStartFrame() int {
	return secondsToFrames(c.WorkAreaStart, c.FrameRate)
}

// @summary     Set a composition's work-area start from an integer frame count
// @description Converts frame back to seconds via FrameRate, then delegates
//   to the seconds-based work-area setter, preserving the existing end.
// @param       frame  the new work-area start as a frame count
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; requires FrameRate to be > 0
// @alias       work area start frame,工作区起始帧,render range start,渲染范围
func (c *Composition) SetWorkAreaStartFrame(frame int) error {
	if c.FrameRate <= 0 {
		return fmt.Errorf("composition %q: SetWorkAreaStartFrame requires FrameRate > 0", c.Name)
	}
	startSec := framesToSeconds(frame, c.FrameRate)
	return c.SetWorkArea(startSec, c.WorkAreaEnd)
}

// WorkAreaEndFrame returns the comp's work-area end as a frame count.
func (c *Composition) WorkAreaEndFrame() int {
	return secondsToFrames(c.WorkAreaEnd, c.FrameRate)
}

// @summary     Set a composition's work-area end from an integer frame count
// @description Converts frame back to seconds via FrameRate, then delegates
//   to the seconds-based work-area setter, preserving the existing start.
// @param       frame  the new work-area end as a frame count
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; requires FrameRate to be > 0
// @alias       work area end frame,工作区结束帧,render range end,渲染结束
func (c *Composition) SetWorkAreaEndFrame(frame int) error {
	if c.FrameRate <= 0 {
		return fmt.Errorf("composition %q: SetWorkAreaEndFrame requires FrameRate > 0", c.Name)
	}
	endSec := framesToSeconds(frame, c.FrameRate)
	return c.SetWorkArea(c.WorkAreaStart, endSec)
}

// WorkAreaDurationFrame returns the work-area span (end − start) in frames.
func (c *Composition) WorkAreaDurationFrame() int {
	return c.WorkAreaEndFrame() - c.WorkAreaStartFrame()
}

// @summary     Set a composition's work-area duration from an integer frame count
// @description Recomputes the work-area end so that end minus start equals
//   the given frame span, keeping start fixed, then delegates to the
//   seconds-based work-area setter.
// @param       frame  the new work-area duration as a frame count
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestCompSettings_AEShipGate_AE2020,TestCompSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; requires FrameRate to be > 0
// @alias       work area duration,工作区时长帧数,render duration frames
func (c *Composition) SetWorkAreaDurationFrame(frame int) error {
	if c.FrameRate <= 0 {
		return fmt.Errorf("composition %q: SetWorkAreaDurationFrame requires FrameRate > 0", c.Name)
	}
	endSec := c.WorkAreaStart + framesToSeconds(frame, c.FrameRate)
	return c.SetWorkArea(c.WorkAreaStart, endSec)
}

// FrameDuration returns the comp's total Duration as a frame count.
//
// Note: this is the total duration in frames. It is NOT the duration of
// a single frame in seconds (= 1/FrameRate).
func (c *Composition) FrameDuration() int {
	return secondsToFrames(c.Duration, c.FrameRate)
}

// ──────────────────────────────────────────────
// Keyframe frame-time accessors
// ──────────────────────────────────────────────

// FrameTime returns the keyframe's time as an integer frame count
// (Time × owning-comp FrameRate, rounded). Returns 0 when the keyframe
// was built outside the parser (FrameRate unknown).
func (k *Keyframe) FrameTime() int {
	if k.back == nil {
		return 0
	}
	return secondsToFrames(k.Time, k.back.FrameRateHz())
}

// @summary     Set a keyframe's time from an integer frame count
// @description Converts frame back to seconds via the owning composition's
//   FrameRate, then delegates to the seconds-based time setter.
// @param       frame  the new keyframe time as a frame count
// @domain      keyframe
// @stability   stable
// @verify      ae-accept
// @gate        TestKeyframeMutate_AEShipGate_AE2020,TestKeyframeMutate_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; requires the owning composition's
//   FrameRate to be known
// @alias       keyframe frame time,关键帧帧编号,kf frame,帧时间
func (k *Keyframe) SetFrameTime(frame int) error {
	if k.back == nil || k.back.FrameRateHz() <= 0 {
		return fmt.Errorf("keyframe: SetFrameTime requires owning composition FrameRate > 0")
	}
	return k.SetTime(framesToSeconds(frame, k.back.FrameRateHz()))
}

// ──────────────────────────────────────────────
// Marker frame-time accessors
// ──────────────────────────────────────────────

// FrameTime returns the marker's time as an integer frame count.
func (m *Marker) FrameTime() int {
	return secondsToFrames(m.Time, m.compFps)
}

// @summary     Set a marker's time from an integer frame count
// @description Converts frame back to seconds via the owning composition's
//   FrameRate, then delegates to the seconds-based time setter.
// @param       frame  the new marker time as a frame count
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestMarkerFields_AEShipGate_AE2020,TestMarkerFields_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; requires the owning composition's
//   FrameRate to be > 0
// @alias       marker frame time,标记帧时间,marker frame,标记帧编号
func (m *Marker) SetFrameTime(frame int) error {
	if m.compFps <= 0 {
		return fmt.Errorf("marker: SetFrameTime requires owning composition FrameRate > 0")
	}
	return m.SetTime(framesToSeconds(frame, m.compFps))
}

// FrameDuration returns the marker's duration as an integer frame count.
// 0 for point markers (Duration == 0).
func (m *Marker) FrameDuration() int {
	return secondsToFrames(m.Duration, m.compFps)
}

// @summary     Set a marker's duration from an integer frame count
// @description Converts frame back to seconds via the owning composition's
//   FrameRate, then delegates to the seconds-based duration setter.
// @param       frame  the new marker duration as a frame count
// @domain      comp
// @stability   stable
// @verify      ae-accept
// @gate        TestMarkerFields_AEShipGate_AE2020,TestMarkerFields_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; requires the owning composition's
//   FrameRate to be > 0
// @alias       marker frame duration,标记时长帧数,cue duration frames
func (m *Marker) SetFrameDuration(frame int) error {
	if m.compFps <= 0 {
		return fmt.Errorf("marker: SetFrameDuration requires owning composition FrameRate > 0")
	}
	return m.SetDuration(framesToSeconds(frame, m.compFps))
}
