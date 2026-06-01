package aep

import "encoding/binary"

// write_render_queue.go — length-preserving in-place setters for render queue
// item settings. Each patches a fixed-width field inside the 2246-byte
// render-settings ldat block (settingsBlock aliases the chunk bytes), so no
// chunk size changes and WriteAEP re-emits the mutation. Non-structural, so no
// AE ship-gate (same model as the other Set* value patches).
//
// Alpha: this write surface has not been double-version ship-gated. Values
// round-trip byte-stably; AE acceptance is presumed (fields AE itself writes)
// but not yet validated in-app.
//
// Concurrency: like all Set* patches these mutate shared chunk bytes; callers
// serialize their own access (see incidents/concurrency-unsafe-shared-chunk-bytes).

// patchU16 writes a big-endian u16 at the given field offset in the item's
// settings block, no-op when the item has no backing block.
func (it *RenderQueueItem) patchU16(fieldOffset int, v uint16) bool {
	if it == nil || len(it.settingsBlock) < fieldOffset+2 {
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

// SetQuality sets the render quality (-1 current / 0 wireframe / 1 draft /
// 2 best). Alpha.
func (it *RenderQueueItem) SetQuality(v int) {
	if it.patchU16(rsQuality, sentinelU16(v)) {
		it.RenderSettings.Quality = v
	}
}

// SetColorDepth sets the color depth (-1 current / 0 8bpc / 1 16bpc / 2 32bpc).
func (it *RenderQueueItem) SetColorDepth(v int) {
	if it.patchU16(rsColorDepth, sentinelU16(v)) {
		it.RenderSettings.ColorDepth = v
	}
}

// SetEffects sets the effects render setting (0 all-off / 1 all-on / 2 current).
func (it *RenderQueueItem) SetEffects(v int) {
	if it.patchU16(rsEffects, sentinelU16(v)) {
		it.RenderSettings.Effects = v
	}
}

// SetFieldRender sets field rendering (0 off / 1 upper-first / 2 lower-first).
func (it *RenderQueueItem) SetFieldRender(v int) {
	if it.patchU16(rsFieldRender, sentinelU16(v)) {
		it.RenderSettings.FieldRender = v
	}
}

// SetPulldown sets the 3:2 pulldown phase (0 off / 1..5).
func (it *RenderQueueItem) SetPulldown(v int) {
	if it.patchU16(rsPulldown, sentinelU16(v)) {
		it.RenderSettings.Pulldown = v
	}
}

// SetFrameBlending sets frame blending (0 off-all / 1 on-checked / 2 current).
func (it *RenderQueueItem) SetFrameBlending(v int) {
	if it.patchU16(rsFrameBlending, sentinelU16(v)) {
		it.RenderSettings.FrameBlending = v
	}
}

// SetMotionBlur sets motion blur (0 off-all / 1 on-checked / 2 current).
func (it *RenderQueueItem) SetMotionBlur(v int) {
	if it.patchU16(rsMotionBlur, sentinelU16(v)) {
		it.RenderSettings.MotionBlur = v
	}
}

// SetProxyUse sets proxy use (0 none / 1 all / 2 current / 3 comp-only).
func (it *RenderQueueItem) SetProxyUse(v int) {
	if it.patchU16(rsProxyUse, sentinelU16(v)) {
		it.RenderSettings.ProxyUse = v
	}
}

// SetSoloSwitches sets solo switches (0 off / 2 current).
func (it *RenderQueueItem) SetSoloSwitches(v int) {
	if it.patchU16(rsSoloSwitches, sentinelU16(v)) {
		it.RenderSettings.SoloSwitches = v
	}
}

// SetGuideLayers sets guide layers (0 off / 2 current).
func (it *RenderQueueItem) SetGuideLayers(v int) {
	if it.patchU16(rsGuideLayers, sentinelU16(v)) {
		it.RenderSettings.GuideLayers = v
	}
}

// SetDiskCache sets disk cache (0 read-only / 2 current).
func (it *RenderQueueItem) SetDiskCache(v int) {
	if it.patchU16(rsDiskCache, sentinelU16(v)) {
		it.RenderSettings.DiskCache = v
	}
}

// SetFrameRate sets the frame-rate source (0 use comp / 1 use this).
func (it *RenderQueueItem) SetFrameRate(v int) {
	if it.patchU16(rsUseThisFrameRate, sentinelU16(v)) {
		it.RenderSettings.FrameRate = v
	}
}

// SetResolution sets the [x, y] resolution divisors (>= 1).
func (it *RenderQueueItem) SetResolution(x, y int) {
	if x < 1 || y < 1 {
		return
	}
	if it.patchU16(rsResolutionX, uint16(x)) && it.patchU16(rsResolutionY, uint16(y)) {
		it.RenderSettings.Resolution = [2]int{x, y}
	}
}

// SetSkipExistingFiles toggles "skip existing files".
func (it *RenderQueueItem) SetSkipExistingFiles(v bool) {
	n := uint16(0)
	if v {
		n = 1
	}
	if it.patchU16(rsSkipExistingFiles, n) {
		it.RenderSettings.SkipExistingFiles = v
	}
}

// SetName sets the render-settings template name (template_name @0x5A, a fixed
// 64-byte windows-1252 NUL-padded field). Names longer than 64 bytes are
// truncated; non-latin-1 runes are dropped. Length-preserving.
func (it *RenderQueueItem) SetName(name string) {
	if it == nil || len(it.settingsBlock) < rsTemplateName+rsTemplateNameLen {
		return
	}
	field := it.settingsBlock[rsTemplateName : rsTemplateName+rsTemplateNameLen]
	for i := range field {
		field[i] = 0
	}
	n := 0
	for _, r := range name {
		if n >= rsTemplateNameLen {
			break
		}
		if r < 0x100 { // windows-1252 / latin-1 representable
			field[n] = byte(r)
			n++
		}
	}
	it.Name = decodeWin1252(field)
}

// SetQueueItemNotify toggles the notify-on-completion flag (flag byte @0x07
// bit 2).
func (it *RenderQueueItem) SetQueueItemNotify(v bool) {
	if it == nil || len(it.settingsBlock) <= rsFlagByte {
		return
	}
	mask := byte(1 << 2)
	if v {
		it.settingsBlock[rsFlagByte] |= mask
	} else {
		it.settingsBlock[rsFlagByte] &^= mask
	}
	it.QueueItemNotify = v
}

// SetLogType sets the raw log-type code (@0x50).
func (it *RenderQueueItem) SetLogType(v uint16) {
	if it.patchU16(rsLogType, v) {
		it.LogType = v
	}
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
func (om *OutputModule) omPatchU8(off int, v byte) bool {
	if om == nil || len(om.settingsBlock) <= off {
		return false
	}
	om.settingsBlock[off] = v
	return true
}

func (om *OutputModule) omPatchU16BE(off int, v uint16) bool {
	if om == nil || len(om.settingsBlock) < off+2 {
		return false
	}
	binary.BigEndian.PutUint16(om.settingsBlock[off:], v)
	return true
}

func (om *OutputModule) omPatchU32BE(off int, v uint32) bool {
	if om == nil || len(om.settingsBlock) < off+4 {
		return false
	}
	binary.BigEndian.PutUint32(om.settingsBlock[off:], v)
	return true
}

func (om *OutputModule) omSetBit(byteOff, bit int, v bool) bool {
	if om == nil || len(om.settingsBlock) <= byteOff {
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

// SetChannels sets the output channels (0 RGB / 1 RGBA / 2 Alpha).
func (om *OutputModule) SetChannels(v int) {
	if om.omPatchU8(omsChannels, byte(v)) {
		om.Settings.Channels = v
	}
}

// SetResizeQuality sets the resize quality.
func (om *OutputModule) SetResizeQuality(v int) {
	if om.omPatchU8(omsResizeQuality, byte(v)) {
		om.Settings.ResizeQuality = v
	}
}

// SetResize toggles resize.
func (om *OutputModule) SetResize(v bool) {
	if om.omPatchU8(omsResize, boolByte(v)) {
		om.Settings.Resize = v
	}
}

// SetLockAspectRatio toggles lock-aspect-ratio.
func (om *OutputModule) SetLockAspectRatio(v bool) {
	if om.omPatchU8(omsLockAspectRatio, boolByte(v)) {
		om.Settings.LockAspectRatio = v
	}
}

// SetCrop toggles crop (flag byte @0x1F bit 0).
func (om *OutputModule) SetCrop(v bool) {
	if om.omSetBit(omsFlagByte22, 0, v) {
		om.Settings.Crop = v
	}
}

// SetCropTop/Left/Bottom/Right set the crop insets (px).
func (om *OutputModule) SetCropTop(v int) {
	if om.omPatchU16BE(omsCropTop, uint16(v)) {
		om.Settings.CropTop = v
	}
}

func (om *OutputModule) SetCropLeft(v int) {
	if om.omPatchU16BE(omsCropLeft, uint16(v)) {
		om.Settings.CropLeft = v
	}
}

func (om *OutputModule) SetCropBottom(v int) {
	if om.omPatchU16BE(omsCropBottom, uint16(v)) {
		om.Settings.CropBottom = v
	}
}

func (om *OutputModule) SetCropRight(v int) {
	if om.omPatchU16BE(omsCropRight, uint16(v)) {
		om.Settings.CropRight = v
	}
}

// SetIncludeProjectLink toggles the "include project link" flag.
func (om *OutputModule) SetIncludeProjectLink(v bool) {
	if om.omPatchU8(omsIncludeProjectLink, boolByte(v)) {
		om.Settings.IncludeProjectLink = v
	}
}

// SetPostRenderAction sets the raw post-render action code.
func (om *OutputModule) SetPostRenderAction(v uint32) {
	if om.omPatchU32BE(omsPostRenderAction, v) {
		om.Settings.PostRenderAction = v
	}
}

// SetUseCompFrameNumber toggles "use comp frame number" (flag byte @0x07 bit 3).
func (om *OutputModule) SetUseCompFrameNumber(v bool) {
	if om.omSetBit(omsFlagByte07, 3, v) {
		om.Settings.UseCompFrameNumber = v
	}
}

// SetUseRegionOfInterest toggles "use region of interest" (bit 4).
func (om *OutputModule) SetUseRegionOfInterest(v bool) {
	if om.omSetBit(omsFlagByte07, 4, v) {
		om.Settings.UseRegionOfInterest = v
	}
}

// SetIncludeSourceXMP toggles "include source XMP metadata" (bit 6).
func (om *OutputModule) SetIncludeSourceXMP(v bool) {
	if om.omSetBit(omsFlagByte07, 6, v) {
		om.Settings.IncludeSourceXMP = v
	}
}

// SetPreserveRGB toggles "preserve RGB" (bit 7).
func (om *OutputModule) SetPreserveRGB(v bool) {
	if om.omSetBit(omsFlagByte07, 7, v) {
		om.Settings.PreserveRGB = v
	}
}

// SetDepth sets the output color depth (Roou @0x47), e.g. 24/32/48/64/96/128.
func (om *OutputModule) SetDepth(v int) {
	if om == nil || len(om.roouData) <= roouDepth {
		return
	}
	om.roouData[roouDepth] = byte(v)
	om.Settings.Depth = v
}

// SetStartingNumber sets the image-sequence starting frame number (Roou @0x10).
func (om *OutputModule) SetStartingNumber(v uint32) {
	if om == nil || len(om.roouData) < roouStartingNumber+4 {
		return
	}
	binary.BigEndian.PutUint32(om.roouData[roouStartingNumber:], v)
	om.Settings.StartingNumber = v
}
