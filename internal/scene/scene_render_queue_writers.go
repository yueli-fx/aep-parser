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
// Alpha: this write surface has not been double-version ship-gated. Values
// round-trip byte-stably; AE acceptance is presumed (fields AE itself writes)
// but not yet validated in-app.
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

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="quality,render quality,渲染质量,画质"
// SetQuality sets the render quality (-1 current / 0 wireframe / 1 draft /
// 2 best). Alpha.
func (it *RenderQueueItem) SetQuality(v int) {
	if it.patchU16(codec.RsQuality, sentinelU16(v)) {
		it.RenderSettings.Quality = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="color depth,bit depth,颜色深度,位深"
// SetColorDepth sets the color depth (-1 current / 0 8bpc / 1 16bpc / 2 32bpc).
func (it *RenderQueueItem) SetColorDepth(v int) {
	if it.patchU16(codec.RsColorDepth, sentinelU16(v)) {
		it.RenderSettings.ColorDepth = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="effects,render effects,特效渲染,效果开关"
// SetEffects sets the effects render setting (0 all-off / 1 all-on / 2 current).
func (it *RenderQueueItem) SetEffects(v int) {
	if it.patchU16(codec.RsEffects, sentinelU16(v)) {
		it.RenderSettings.Effects = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="field render,场渲染,场序,interlace"
// SetFieldRender sets field rendering (0 off / 1 upper-first / 2 lower-first).
func (it *RenderQueueItem) SetFieldRender(v int) {
	if it.patchU16(codec.RsFieldRender, sentinelU16(v)) {
		it.RenderSettings.FieldRender = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="pulldown,3:2 pulldown,下拉扫描,帧率转换"
// SetPulldown sets the 3:2 pulldown phase (0 off / 1..5).
func (it *RenderQueueItem) SetPulldown(v int) {
	if it.patchU16(codec.RsPulldown, sentinelU16(v)) {
		it.RenderSettings.Pulldown = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="frame blending,帧混合,运动模糊插帧"
// SetFrameBlending sets frame blending (0 off-all / 1 on-checked / 2 current).
func (it *RenderQueueItem) SetFrameBlending(v int) {
	if it.patchU16(codec.RsFrameBlending, sentinelU16(v)) {
		it.RenderSettings.FrameBlending = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="motion blur,运动模糊"
// SetMotionBlur sets motion blur (0 off-all / 1 on-checked / 2 current).
func (it *RenderQueueItem) SetMotionBlur(v int) {
	if it.patchU16(codec.RsMotionBlur, sentinelU16(v)) {
		it.RenderSettings.MotionBlur = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="proxy,proxy use,代理,代理使用"
// SetProxyUse sets proxy use (0 none / 1 all / 2 current / 3 comp-only).
func (it *RenderQueueItem) SetProxyUse(v int) {
	if it.patchU16(codec.RsProxyUse, sentinelU16(v)) {
		it.RenderSettings.ProxyUse = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="solo switches,独奏开关,solo"
// SetSoloSwitches sets solo switches (0 off / 2 current).
func (it *RenderQueueItem) SetSoloSwitches(v int) {
	if it.patchU16(codec.RsSoloSwitches, sentinelU16(v)) {
		it.RenderSettings.SoloSwitches = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="guide layers,参考层,辅助层"
// SetGuideLayers sets guide layers (0 off / 2 current).
func (it *RenderQueueItem) SetGuideLayers(v int) {
	if it.patchU16(codec.RsGuideLayers, sentinelU16(v)) {
		it.RenderSettings.GuideLayers = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="disk cache,磁盘缓存,缓存"
// SetDiskCache sets disk cache (0 read-only / 2 current).
func (it *RenderQueueItem) SetDiskCache(v int) {
	if it.patchU16(codec.RsDiskCache, sentinelU16(v)) {
		it.RenderSettings.DiskCache = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="frame rate,帧率,fps"
// SetFrameRate sets the frame-rate source (0 use comp / 1 use this).
func (it *RenderQueueItem) SetFrameRate(v int) {
	if it.patchU16(codec.RsUseThisFrameRate, sentinelU16(v)) {
		it.RenderSettings.FrameRate = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="resolution,分辨率,画面尺寸"
// SetResolution sets the [x, y] resolution divisors (>= 1).
func (it *RenderQueueItem) SetResolution(x, y int) {
	if x < 1 || y < 1 {
		return
	}
	if it.patchU16(codec.RsResolutionX, uint16(x)) && it.patchU16(codec.RsResolutionY, uint16(y)) {
		it.RenderSettings.Resolution = [2]int{x, y}
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="skip existing files,跳过已有文件,增量渲染"
// SetSkipExistingFiles toggles "skip existing files".
func (it *RenderQueueItem) SetSkipExistingFiles(v bool) {
	n := uint16(0)
	if v {
		n = 1
	}
	if it.patchU16(codec.RsSkipExistingFiles, n) {
		it.RenderSettings.SkipExistingFiles = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险(固定 64B 字段)" alias="name,template name,渲染设置名称,模板名"
// SetName sets the render-settings template name (template_name @0x5A, a fixed
// 64-byte windows-1252 NUL-padded field). Names longer than 64 bytes are
// truncated; non-latin-1 runes are dropped. Length-preserving.
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

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="notify,completion notify,完成通知"
// SetQueueItemNotify toggles the notify-on-completion flag (flag byte @0x07
// bit 2).
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

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="log type,日志类型,渲染日志"
// SetLogType sets the raw log-type code (@0x50).
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

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="time span start,render start,开始时间,渲染起点"
// SetTimeSpanStart sets the render start time (seconds), switching the time
// span source to CUSTOM. Length-preserving.
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

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="time span duration,render duration,渲染时长,持续时间"
// SetTimeSpanDuration sets the render duration (seconds), switching the time
// span source to CUSTOM. Length-preserving.
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

//aep:cap domain=render-queue tier=alpha verify=ae-accept gate=TestRenderQueueComment_AEShipGate_AE2020,TestRenderQueueComment_AEShipGate_AE2025 boundary="length-variable(RCom chunk 插入/替换);AE 接受+resave-preservation 已验证;无 ScriptingAPI readback" alias="comment,rq comment,渲染队列注释,备注"
// SetComment sets the render queue item's comment (shown in the Render Queue
// panel). length-variable, unlike the fixed-width setters above: the comment
// lives in an RCom wrapper chunk holding a single Utf8 child. When the item
// already has an RCom its payload is replaced; otherwise a fresh RCom is
// inserted into the LItm LIST immediately before the item's settings list
// (matching AE / py-aep per-item ordering). WriteAEP recomputes the LItm/LRdr
// LIST sizes. Setting "" on an item with no RCom is a no-op (AE writes no RCom
// for empty comments).
//
// Double-version ship-gated (AE 2020 + AE 2025): both accept the inserted RCom
// and preserve it byte-identically on resave. The comment is a binary-only
// field — RenderQueueItem has no `comment` ScriptingAPI in any version — so the
// gate verifies acceptance + resave-preservation, not script readback. See
// render_queue_comment_shipgate_test.go.
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

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="channels,output channels,输出通道,RGB,RGBA,Alpha"
// SetChannels sets the output channels (0 RGB / 1 RGBA / 2 Alpha).
func (om *OutputModule) SetChannels(v int) {
	if om.omPatchU8(codec.OmsChannels, byte(v)) {
		om.Settings.Channels = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="resize quality,缩放质量"
// SetResizeQuality sets the resize quality.
func (om *OutputModule) SetResizeQuality(v int) {
	if om.omPatchU8(codec.OmsResizeQuality, byte(v)) {
		om.Settings.ResizeQuality = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="resize,缩放,输出缩放"
// SetResize toggles resize.
func (om *OutputModule) SetResize(v bool) {
	if om.omPatchU8(codec.OmsResize, boolByte(v)) {
		om.Settings.Resize = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="lock aspect ratio,锁定宽高比,等比缩放"
// SetLockAspectRatio toggles lock-aspect-ratio.
func (om *OutputModule) SetLockAspectRatio(v bool) {
	if om.omPatchU8(codec.OmsLockAspectRatio, boolByte(v)) {
		om.Settings.LockAspectRatio = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="crop,裁剪,输出裁剪"
// SetCrop toggles crop (flag byte @0x1F bit 0).
func (om *OutputModule) SetCrop(v bool) {
	if om.omSetBit(codec.OmsFlagByte22, 0, v) {
		om.Settings.Crop = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="crop top,上裁剪,顶部裁剪"
// SetCropTop/Left/Bottom/Right set the crop insets (px).
func (om *OutputModule) SetCropTop(v int) {
	if om.omPatchU16BE(codec.OmsCropTop, uint16(v)) {
		om.Settings.CropTop = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="crop left,左裁剪"
func (om *OutputModule) SetCropLeft(v int) {
	if om.omPatchU16BE(codec.OmsCropLeft, uint16(v)) {
		om.Settings.CropLeft = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="crop bottom,下裁剪,底部裁剪"
func (om *OutputModule) SetCropBottom(v int) {
	if om.omPatchU16BE(codec.OmsCropBottom, uint16(v)) {
		om.Settings.CropBottom = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="crop right,右裁剪"
func (om *OutputModule) SetCropRight(v int) {
	if om.omPatchU16BE(codec.OmsCropRight, uint16(v)) {
		om.Settings.CropRight = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="include project link,项目链接,嵌入项目链接"
// SetIncludeProjectLink toggles the "include project link" flag.
func (om *OutputModule) SetIncludeProjectLink(v bool) {
	if om.omPatchU8(codec.OmsIncludeProjectLink, boolByte(v)) {
		om.Settings.IncludeProjectLink = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="post render action,渲染后操作,完成动作"
// SetPostRenderAction sets the raw post-render action code.
func (om *OutputModule) SetPostRenderAction(v uint32) {
	if om.omPatchU32BE(codec.OmsPostRenderAction, v) {
		om.Settings.PostRenderAction = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险;raw u8(py-aep derives ON/OFF/AUTO)" alias="output audio,输出音频,渲染音频开关"
// SetOutputAudio sets the raw output-audio code (OutputModule @0x2A).
func (om *OutputModule) SetOutputAudio(v int) {
	if om.omPatchU8(codec.OmsOutputAudio, byte(v)) {
		om.Settings.OutputAudio = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险;CMS 联动 enum,round-trip 绿不保证 AE 渲染采用" alias="convert to linear,转线性光,linearize output"
// SetConvertToLinear sets the raw convert-to-linear code (OutputModule @0x5B).
func (om *OutputModule) SetConvertToLinear(v int) {
	if om.omPatchU8(codec.OmsConvertLinear, byte(v)) {
		om.Settings.ConvertToLinear = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="use comp frame number,使用合成帧编号,帧编号"
// SetUseCompFrameNumber toggles "use comp frame number" (flag byte @0x07 bit 3).
func (om *OutputModule) SetUseCompFrameNumber(v bool) {
	if om.omSetBit(codec.OmsFlagByte07, 3, v) {
		om.Settings.UseCompFrameNumber = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="region of interest,ROI,感兴趣区域,局部渲染"
// SetUseRegionOfInterest toggles "use region of interest" (bit 4).
func (om *OutputModule) SetUseRegionOfInterest(v bool) {
	if om.omSetBit(codec.OmsFlagByte07, 4, v) {
		om.Settings.UseRegionOfInterest = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="XMP,source XMP,XMP元数据,元数据"
// SetIncludeSourceXMP toggles "include source XMP metadata" (bit 6).
func (om *OutputModule) SetIncludeSourceXMP(v bool) {
	if om.omSetBit(codec.OmsFlagByte07, 6, v) {
		om.Settings.IncludeSourceXMP = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险" alias="preserve RGB,保留RGB,色彩保留"
// SetPreserveRGB toggles "preserve RGB" (bit 7).
func (om *OutputModule) SetPreserveRGB(v bool) {
	if om.omSetBit(codec.OmsFlagByte07, 7, v) {
		om.Settings.PreserveRGB = v
	}
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险(roouData 字节)" alias="depth,output depth,输出色深,颜色深度"
// SetDepth sets the output color depth (Roou @0x47), e.g. 24/32/48/64/96/128.
func (om *OutputModule) SetDepth(v int) {
	if om == nil || len(om.roouData) <= int(codec.RouoDepth) {
		return
	}
	om.roouData[codec.RouoDepth] = byte(v)
	om.Settings.Depth = v
}

//aep:cap domain=render-queue tier=alpha verify=roundtrip boundary="Alpha,未 AE-gate;length-preserving 低风险(roouData 字节)" alias="starting number,image sequence start,序列起始帧,帧序号"
// SetStartingNumber sets the image-sequence starting frame number (Roou @0x10).
func (om *OutputModule) SetStartingNumber(v uint32) {
	if om == nil || len(om.roouData) < int(codec.RouoStartingNumber)+4 {
		return
	}
	binary.BigEndian.PutUint32(om.roouData[codec.RouoStartingNumber:], v)
	om.Settings.StartingNumber = v
}
