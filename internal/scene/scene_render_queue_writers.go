package scene

import (
	"encoding/binary"
	"fmt"

	"github.com/example/aep-parser/internal/codec"
)

// write_render_queue.go — length-preserving setters for render queue item +
// output module settings. Each patches a fixed-width field inside the item's
// scene-owned settings copy (settingsBlock / roouData — the single source of
// truth), so no chunk size changes. syncRenderQueue copies the mutated buffers
// back into the owning RIFX chunks at WriteAEP time. Non-structural, so no AE
// ship-gate (same model as the other Set* value patches).
//
// Most of this write surface is now double-version ship-gated for AE acceptance +
// resave-preservation: TestRenderQueueSettings_AEShipGate_AE2020/AE2025 sets every
// value setter to a non-default, has AE open the file (acceptance) and resave, and
// re-parses to confirm the bytes survive. 36 fields preserve in both versions →
// verify=ae-accept. Two exceptions stay verify=roundtrip because AE normalizes them
// on resave (not a write bug, an AE-semantics reset): PreserveRGB (AE2025 clears the
// color-management-gated bit) and QueueItemNotify (AE2020 clears the notify bit).
// There is no cross-version ScriptingAPI readback for these binary settings, so
// acceptance + resave-preservation is the verification ceiling (same as SetComment).
//
// Concurrency: like all Set* patches these mutate shared scene buffers; callers
// serialize their own access (see incidents/concurrency-unsafe-shared-chunk-bytes).

// patchU16 writes a big-endian u16 at the given field offset in the item's
// settings block, no-op when the item has no backing block.
func (it *RenderQueueItem) patchU16(fieldOffset codec.RenderSettingOffset, v uint16) bool {
	if it == nil || len(it.settingsBlock) < int(fieldOffset)+2 {
		return false
	}
	binary.BigEndian.PutUint16(it.settingsBlock[fieldOffset:], v)
	return true
}

// sentinelU16 maps the "current settings" value -1 back to the binary 0xFFFF
// sentinel; other values pass through truncated to u16.
func sentinelU16(v int) uint16 {
	if v < 0 {
		return 0xFFFF
	}
	return uint16(v)
}

// --- render settings setters ------------------------------------------------

// @summary     Set the render quality
// @param       v  -1 current, 0 wireframe, 1 draft, 2 best
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       quality,render quality,渲染质量,画质
func (it *RenderQueueItem) SetQuality(v int) {
	if it.patchU16(codec.RsQuality, sentinelU16(v)) {
		it.RenderSettings.Quality = v
	}
}

// @summary     Set the render color depth
// @param       v  -1 current, 0 8bpc, 1 16bpc, 2 32bpc
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       color depth,bit depth,颜色深度,位深
func (it *RenderQueueItem) SetColorDepth(v int) {
	if it.patchU16(codec.RsColorDepth, sentinelU16(v)) {
		it.RenderSettings.ColorDepth = v
	}
}

// @summary     Set the effects render setting
// @param       v  0 all-off, 1 all-on, 2 current
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       effects,render effects,特效渲染,效果开关
func (it *RenderQueueItem) SetEffects(v int) {
	if it.patchU16(codec.RsEffects, sentinelU16(v)) {
		it.RenderSettings.Effects = v
	}
}

// @summary     Set the field rendering mode
// @param       v  0 off, 1 upper-first, 2 lower-first
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       field render,场渲染,场序,interlace
func (it *RenderQueueItem) SetFieldRender(v int) {
	if it.patchU16(codec.RsFieldRender, sentinelU16(v)) {
		it.RenderSettings.FieldRender = v
	}
}

// @summary     Set the 3:2 pulldown phase
// @param       v  0 off, 1 through 5 select the pulldown phase
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       pulldown,3:2 pulldown,下拉扫描,帧率转换
func (it *RenderQueueItem) SetPulldown(v int) {
	if it.patchU16(codec.RsPulldown, sentinelU16(v)) {
		it.RenderSettings.Pulldown = v
	}
}

// @summary     Set the frame blending mode
// @param       v  0 off for all layers, 1 on for checked layers, 2 current
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       frame blending,帧混合,运动模糊插帧
func (it *RenderQueueItem) SetFrameBlending(v int) {
	if it.patchU16(codec.RsFrameBlending, sentinelU16(v)) {
		it.RenderSettings.FrameBlending = v
	}
}

// @summary     Set the motion blur mode
// @param       v  0 off for all layers, 1 on for checked layers, 2 current
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       motion blur,运动模糊
func (it *RenderQueueItem) SetMotionBlur(v int) {
	if it.patchU16(codec.RsMotionBlur, sentinelU16(v)) {
		it.RenderSettings.MotionBlur = v
	}
}

// @summary     Set proxy use for the render queue item
// @param       v  0 none, 1 use all proxies, 2 current, 3 comp proxies only
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       proxy,proxy use,代理,代理使用
func (it *RenderQueueItem) SetProxyUse(v int) {
	if it.patchU16(codec.RsProxyUse, sentinelU16(v)) {
		it.RenderSettings.ProxyUse = v
	}
}

// @summary     Set the solo switches setting
// @param       v  0 off, 2 current
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       solo switches,独奏开关,solo
func (it *RenderQueueItem) SetSoloSwitches(v int) {
	if it.patchU16(codec.RsSoloSwitches, sentinelU16(v)) {
		it.RenderSettings.SoloSwitches = v
	}
}

// @summary     Set the guide layers render setting
// @param       v  0 off, 2 current
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       guide layers,参考层,辅助层
func (it *RenderQueueItem) SetGuideLayers(v int) {
	if it.patchU16(codec.RsGuideLayers, sentinelU16(v)) {
		it.RenderSettings.GuideLayers = v
	}
}

// @summary     Set the disk cache render setting
// @param       v  0 read-only, 2 current
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       disk cache,磁盘缓存,缓存
func (it *RenderQueueItem) SetDiskCache(v int) {
	if it.patchU16(codec.RsDiskCache, sentinelU16(v)) {
		it.RenderSettings.DiskCache = v
	}
}

// @summary     Set the frame-rate source
// @param       v  0 use comp frame rate, 1 use this item's frame rate
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       frame rate,帧率,fps
func (it *RenderQueueItem) SetFrameRate(v int) {
	if it.patchU16(codec.RsUseThisFrameRate, sentinelU16(v)) {
		it.RenderSettings.FrameRate = v
	}
}

// @summary     Set the render resolution divisors
// @param       x  horizontal resolution divisor (>= 1)
// @param       y  vertical resolution divisor (>= 1)
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       resolution,分辨率,画面尺寸
func (it *RenderQueueItem) SetResolution(x, y int) {
	if x < 1 || y < 1 {
		return
	}
	if it.patchU16(codec.RsResolutionX, uint16(x)) && it.patchU16(codec.RsResolutionY, uint16(y)) {
		it.RenderSettings.Resolution = [2]int{x, y}
	}
}

// @summary     Toggle "skip existing files" for the render queue item
// @param       v  whether to skip files that already exist on disk
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       skip existing files,跳过已有文件,增量渲染
func (it *RenderQueueItem) SetSkipExistingFiles(v bool) {
	n := uint16(0)
	if v {
		n = 1
	}
	if it.patchU16(codec.RsSkipExistingFiles, n) {
		it.RenderSettings.SkipExistingFiles = v
	}
}

// @summary     Set the render-settings template name
// @description The name is stored in a fixed 64-byte windows-1252 NUL-padded
//   field (template_name @0x5A). Names longer than 64 bytes are truncated;
//   non-latin-1 runes are dropped.
// @param       name  the template name to store
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field; length-preserving edit against a fixed 64-byte field is low risk
// @alias       name,template name,渲染设置名称,模板名
func (it *RenderQueueItem) SetName(name string) {
	if it == nil || len(it.settingsBlock) < int(codec.RsTemplateName)+codec.RsTemplateNameLen {
		return
	}
	field := it.settingsBlock[codec.RsTemplateName : codec.RsTemplateName+codec.RsTemplateNameLen]
	for i := range field {
		field[i] = 0
	}
	n := 0
	for _, r := range name {
		if n >= codec.RsTemplateNameLen {
			break
		}
		if r < 0x100 { // windows-1252 / latin-1 representable
			field[n] = byte(r)
			n++
		}
	}
	it.Name = codec.DecodeWin1252(field)
}

// @summary     Toggle the notify-on-completion flag (flag byte @0x07 bit 2)
// @param       v  whether to notify when the render queue item completes
// @domain      render-queue
// @stability   alpha
// @verify      roundtrip
// @since       AE2020
// @boundary    AE accepts the written bit, but AE2020 clears it on resave (a version-specific normalization, not a write bug); AE2025 preserves it. Since preservation is not guaranteed across both versions, verification stays at roundtrip
// @alias       notify,completion notify,完成通知
func (it *RenderQueueItem) SetQueueItemNotify(v bool) {
	if it == nil || len(it.settingsBlock) <= int(codec.RsFlagByte) {
		return
	}
	mask := byte(1 << 2)
	if v {
		it.settingsBlock[codec.RsFlagByte] |= mask
	} else {
		it.settingsBlock[codec.RsFlagByte] &^= mask
	}
	it.QueueItemNotify = v
}

// @summary     Set the raw log-type code (@0x50)
// @param       v  the raw log-type code
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       log type,日志类型,渲染日志
func (it *RenderQueueItem) SetLogType(v uint16) {
	if it.patchU16(codec.RsLogType, v) {
		it.LogType = v
	}
}

// patchU32 writes a big-endian u32 at fieldOffset in the settings block.
func (it *RenderQueueItem) patchU32(fieldOffset codec.RenderSettingOffset, v uint32) bool {
	if it == nil || len(it.settingsBlock) < int(fieldOffset)+4 {
		return false
	}
	binary.BigEndian.PutUint32(it.settingsBlock[fieldOffset:], v)
	return true
}

// secondsToFraction reduces a non-negative seconds value to a dividend/divisor
// pair (scale 1e6 + gcd reduction). Exact for values with ≤6 decimal places;
// denominator stays ≤ 1e6 so both fit u32.
func secondsToFraction(v float64) (uint32, uint32) {
	if v <= 0 {
		return 0, 1
	}
	const scale = 1000000
	num := int64(v*scale + 0.5)
	den := int64(scale)
	g := gcdInt64(num, den)
	return uint32(num / g), uint32(den / g)
}

func gcdInt64(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	if a == 0 {
		return 1
	}
	return a
}

// @summary     Set the render start time, switching the time span to custom
// @param       seconds  the render start time in seconds
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       time span start,render start,开始时间,渲染起点
func (it *RenderQueueItem) SetTimeSpanStart(seconds float64) {
	if seconds < 0 {
		return
	}
	num, den := secondsToFraction(seconds)
	if !it.patchU16(codec.RsTimeSpanSource, codec.TimeSpanCustom) {
		return
	}
	it.patchU32(codec.RsTimeSpanStartDividend, num)
	it.patchU32(codec.RsTimeSpanStartDivisor, den)
	it.TimeSpanStart = seconds
}

// @summary     Set the render duration, switching the time span to custom
// @param       seconds  the render duration in seconds
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       time span duration,render duration,渲染时长,持续时间
func (it *RenderQueueItem) SetTimeSpanDuration(seconds float64) {
	if seconds <= 0 {
		return
	}
	num, den := secondsToFraction(seconds)
	if !it.patchU16(codec.RsTimeSpanSource, codec.TimeSpanCustom) {
		return
	}
	it.patchU32(codec.RsTimeSpanDurDividend, num)
	it.patchU32(codec.RsTimeSpanDurDivisor, den)
	it.TimeSpanDuration = seconds
}

// @summary     Set the render queue item's comment shown in the Render Queue panel
// @description Unlike the fixed-width setters above, this field is
//   length-variable: the comment lives in an RCom wrapper chunk holding a
//   single Utf8 child. When the item already has an RCom its payload is
//   replaced; otherwise a fresh RCom is inserted into the LItm LIST
//   immediately before the item's settings list, matching AE's per-item
//   ordering. WriteAEP recomputes the LItm/LRdr LIST sizes. Setting "" on an
//   item with no RCom is a no-op, since AE writes no RCom for empty comments.
//   Both AE versions accept the inserted RCom and preserve it byte-identically
//   on resave. The comment is a binary-only field with no ScriptingAPI in any
//   version, so the gate verifies acceptance and resave-preservation rather
//   than script readback.
// @param       comment  the comment text to store
// @domain      render-queue
// @stability   alpha
// @verify      ae-accept
// @gate        TestRenderQueueComment_AEShipGate_AE2020,TestRenderQueueComment_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-variable edit that inserts or replaces an RCom chunk; acceptance and resave-preservation are verified, but there is no ScriptingAPI readback to cross-check against
// @alias       comment,rq comment,渲染队列注释,备注
func (it *RenderQueueItem) SetComment(comment string) error {
	if it == nil || it.back == nil {
		return fmt.Errorf("render queue item: no chunk backrefs (built outside parser?)")
	}
	if err := it.back.SetComment(comment); err != nil {
		return err
	}
	it.Comment = comment
	return nil
}

// --- output module settings setters -----------------------------------------

func boolByte(v bool) byte {
	if v {
		return 1
	}
	return 0
}

// omPatchU8 / omPatchU16BE / omPatchU32BE / omSetBit patch the 128B settings
// block in place; ok reports whether the block is present and long enough.
func (om *OutputModule) omPatchU8(off codec.RenderSettingOffset, v byte) bool {
	if om == nil || len(om.settingsBlock) <= int(off) {
		return false
	}
	om.settingsBlock[off] = v
	return true
}

func (om *OutputModule) omPatchU16BE(off codec.RenderSettingOffset, v uint16) bool {
	if om == nil || len(om.settingsBlock) < int(off)+2 {
		return false
	}
	binary.BigEndian.PutUint16(om.settingsBlock[off:], v)
	return true
}

func (om *OutputModule) omPatchU32BE(off codec.RenderSettingOffset, v uint32) bool {
	if om == nil || len(om.settingsBlock) < int(off)+4 {
		return false
	}
	binary.BigEndian.PutUint32(om.settingsBlock[off:], v)
	return true
}

func (om *OutputModule) omSetBit(byteOff codec.RenderSettingOffset, bit int, v bool) bool {
	if om == nil || len(om.settingsBlock) <= int(byteOff) {
		return false
	}
	mask := byte(1 << bit)
	if v {
		om.settingsBlock[byteOff] |= mask
	} else {
		om.settingsBlock[byteOff] &^= mask
	}
	return true
}

// @summary     Set the output channels
// @param       v  0 RGB, 1 RGBA, 2 Alpha
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       channels,output channels,输出通道,RGB,RGBA,Alpha
func (om *OutputModule) SetChannels(v int) {
	if om.omPatchU8(codec.OmsChannels, byte(v)) {
		om.Settings.Channels = v
	}
}

// @summary     Set the output resize quality
// @param       v  the raw resize quality code
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       resize quality,缩放质量
func (om *OutputModule) SetResizeQuality(v int) {
	if om.omPatchU8(codec.OmsResizeQuality, byte(v)) {
		om.Settings.ResizeQuality = v
	}
}

// @summary     Toggle output resize
// @param       v  whether the output module resizes the frame
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       resize,缩放,输出缩放
func (om *OutputModule) SetResize(v bool) {
	if om.omPatchU8(codec.OmsResize, boolByte(v)) {
		om.Settings.Resize = v
	}
}

// @summary     Toggle lock-aspect-ratio for output resize
// @param       v  whether resizing locks the aspect ratio
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       lock aspect ratio,锁定宽高比,等比缩放
func (om *OutputModule) SetLockAspectRatio(v bool) {
	if om.omPatchU8(codec.OmsLockAspectRatio, boolByte(v)) {
		om.Settings.LockAspectRatio = v
	}
}

// @summary     Toggle output crop (flag byte @0x1F bit 0)
// @param       v  whether the output module crops the frame
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       crop,裁剪,输出裁剪
func (om *OutputModule) SetCrop(v bool) {
	if om.omSetBit(codec.OmsFlagByte22, 0, v) {
		om.Settings.Crop = v
	}
}

// @summary     Set the top crop inset in pixels
// @param       v  the top crop inset in pixels
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       crop top,上裁剪,顶部裁剪
func (om *OutputModule) SetCropTop(v int) {
	if om.omPatchU16BE(codec.OmsCropTop, uint16(v)) {
		om.Settings.CropTop = v
	}
}

// @summary     Set the left crop inset in pixels
// @param       v  the left crop inset in pixels
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       crop left,左裁剪
func (om *OutputModule) SetCropLeft(v int) {
	if om.omPatchU16BE(codec.OmsCropLeft, uint16(v)) {
		om.Settings.CropLeft = v
	}
}

// @summary     Set the bottom crop inset in pixels
// @param       v  the bottom crop inset in pixels
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       crop bottom,下裁剪,底部裁剪
func (om *OutputModule) SetCropBottom(v int) {
	if om.omPatchU16BE(codec.OmsCropBottom, uint16(v)) {
		om.Settings.CropBottom = v
	}
}

// @summary     Set the right crop inset in pixels
// @param       v  the right crop inset in pixels
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       crop right,右裁剪
func (om *OutputModule) SetCropRight(v int) {
	if om.omPatchU16BE(codec.OmsCropRight, uint16(v)) {
		om.Settings.CropRight = v
	}
}

// @summary     Toggle the "include project link" flag
// @param       v  whether to embed a project link in the output
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       include project link,项目链接,嵌入项目链接
func (om *OutputModule) SetIncludeProjectLink(v bool) {
	if om.omPatchU8(codec.OmsIncludeProjectLink, boolByte(v)) {
		om.Settings.IncludeProjectLink = v
	}
}

// @summary     Set the raw post-render action code
// @param       v  the raw post-render action code
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       post render action,渲染后操作,完成动作
func (om *OutputModule) SetPostRenderAction(v uint32) {
	if om.omPatchU32BE(codec.OmsPostRenderAction, v) {
		om.Settings.PostRenderAction = v
	}
}

// @summary     Set the raw output-audio code (OutputModule @0x2A)
// @param       v  the raw output-audio code (encodes on/off/auto)
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit of a raw byte field is low risk
// @alias       output audio,输出音频,渲染音频开关
func (om *OutputModule) SetOutputAudio(v int) {
	if om.omPatchU8(codec.OmsOutputAudio, byte(v)) {
		om.Settings.OutputAudio = v
	}
}

// @summary     Set the raw convert-to-linear code (OutputModule @0x5B)
// @param       v  the raw convert-to-linear code
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; this enum interacts with color management, so a green value round-trip does not guarantee AE applies it during rendering
// @alias       convert to linear,转线性光,linearize output
func (om *OutputModule) SetConvertToLinear(v int) {
	if om.omPatchU8(codec.OmsConvertLinear, byte(v)) {
		om.Settings.ConvertToLinear = v
	}
}

// @summary     Toggle "use comp frame number" (flag byte @0x07 bit 3)
// @param       v  whether output filenames use the comp's frame number
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       use comp frame number,使用合成帧编号,帧编号
func (om *OutputModule) SetUseCompFrameNumber(v bool) {
	if om.omSetBit(codec.OmsFlagByte07, 3, v) {
		om.Settings.UseCompFrameNumber = v
	}
}

// @summary     Toggle "use region of interest" (bit 4)
// @param       v  whether rendering is limited to the region of interest
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       region of interest,ROI,感兴趣区域,局部渲染
func (om *OutputModule) SetUseRegionOfInterest(v bool) {
	if om.omSetBit(codec.OmsFlagByte07, 4, v) {
		om.Settings.UseRegionOfInterest = v
	}
}

// @summary     Toggle "include source XMP metadata" (bit 6)
// @param       v  whether source XMP metadata is embedded in the output
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit is low risk
// @alias       XMP,source XMP,XMP元数据,元数据
func (om *OutputModule) SetIncludeSourceXMP(v bool) {
	if om.omSetBit(codec.OmsFlagByte07, 6, v) {
		om.Settings.IncludeSourceXMP = v
	}
}

// @summary     Toggle "preserve RGB" (bit 7)
// @param       v  whether RGB channels are preserved through color management
// @domain      render-queue
// @stability   alpha
// @verify      roundtrip
// @since       AE2020
// @boundary    AE accepts the written bit, but AE2025 clears it on resave when the output's color-management context gates it (not a write bug). Since preservation is not guaranteed, verification stays at roundtrip
// @alias       preserve RGB,保留RGB,色彩保留
func (om *OutputModule) SetPreserveRGB(v bool) {
	if om.omSetBit(codec.OmsFlagByte07, 7, v) {
		om.Settings.PreserveRGB = v
	}
}

// @summary     Set the output color depth (Roou @0x47)
// @param       v  the output color depth, e.g. 24, 32, 48, 64, 96, or 128
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit of the roou byte is low risk
// @alias       depth,output depth,输出色深,颜色深度
func (om *OutputModule) SetDepth(v int) {
	if om == nil || len(om.roouData) <= int(codec.RouoDepth) {
		return
	}
	om.roouData[codec.RouoDepth] = byte(v)
	om.Settings.Depth = v
}

// @summary     Set the image-sequence starting frame number (Roou @0x10)
// @param       v  the starting frame number for an image sequence
// @domain      render-queue
// @stability   stable
// @verify      ae-accept
// @gate        TestRenderQueueSettings_AEShipGate_AE2020,TestRenderQueueSettings_AEShipGate_AE2025
// @since       AE2020
// @boundary    accepted and resave-preserved on both AE versions; no ScriptingAPI readback exists for this binary field, so the gate verifies acceptance plus resave-preservation; length-preserving edit of the roou bytes is low risk
// @alias       starting number,image sequence start,序列起始帧,帧序号
func (om *OutputModule) SetStartingNumber(v uint32) {
	if om == nil || len(om.roouData) < int(codec.RouoStartingNumber)+4 {
		return
	}
	binary.BigEndian.PutUint32(om.roouData[codec.RouoStartingNumber:], v)
	om.Settings.StartingNumber = v
}
